package actors

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/anmitsu/go-shlex"
)

var (
	unsafeQuote      *regexp.Regexp
	playbookTemplate *template.Template
)

type playbookData struct {
	Host                string
	User                string
	Playbook            string
	ActorCommand        string
	InputFile           string
	OutputFile          string
	ActorName           string
	ActorCWD            string
	ActorRepoPath       string
	ActorRemoteRepoPath string
	SyncRepo            string
}

func init() {
	var err error
	unsafeQuote, err = regexp.CompilePOSIX(`[^A-Za-z0-9@%+=:,./-]`)
	if err != nil {
		panic(err)
	}
	playbookTemplate, err = template.New("playbookTemplate").Parse(`
	ansible-playbook
        {{.Playbook}}
        -i {{.Host}},
        -u {{.User}}
        -e actor_repository="{{.ActorRepoPath}}"
        -e sync_repo={{.SyncRepo}}
        -e remote_host={{.Host}}
        -e actor_output_file="{{.OutputFile}}"
        -e actor_input_file="{{.InputFile}}"
        -e actor_command="'{{.ActorCommand}}'"
        -e actor_name="{{.ActorName}}"
        -e actor_cwd="{{.ActorCWD}}"
        -e actor_remote_repo_path="{{.ActorRemoteRepoPath}}"
    `)
	if err != nil {
		panic(err)
	}
}

// Actor is an interface to represent the loaded Actor implementation
type Actor struct {
	Definition Definition
}

func (actor *Actor) handleStdin(w io.Writer, data *ChannelManager) error {
	inputs, err := data.GetFiltered(actor.Definition.Inputs)
	if err == nil {
		err = json.NewEncoder(w).Encode(inputs)
	}
	return err
}

func (actor *Actor) handleRemoteOutput(r io.Reader, data *ChannelManager) bool {
	var ansibleResult map[string]interface{}
	f, _ := os.Open(r.(*os.File).Name())
	err := json.NewDecoder(f).Decode(&ansibleResult)
	if err == nil {
		json.NewEncoder(os.Stdout).Encode(ansibleResult)
		actor.handleStdout(bytes.NewBufferString(ansibleResult["stdout"].(string)), data)
		actor.handleStderr(bytes.NewBufferString(ansibleResult["stderr"].(string)), data)
		return ansibleResult["rc"].(float64) == 0
	}
	return false
}

func (actor *Actor) handleStdout(r io.Reader, data *ChannelManager) error {
	if actor.Definition.Execute.OutputProcessor != nil {
		lines, _ := ioutil.ReadAll(r)
		data.AssignToVariable(actor.Definition.Execute.OutputProcessor.Target, strings.Split(string(lines), "\n"))
		return nil
	}
	var output map[string]interface{}
	json.NewDecoder(r).Decode(&output)
	return data.AssignFiltered(actor.Definition.Outputs, output)
}

func (actor *Actor) handleStderr(r io.Reader, data *ChannelManager) error {
	return nil
}

// Execute executes the actor
func (actor *Actor) Execute(data *ChannelManager, syncRepo bool) bool {
	if actor.Definition.Remote != nil {
		return actor.ExecuteRemote(data, *actor.Definition.Remote.Host, *actor.Definition.Remote.User, syncRepo)
	}
	execute := actor.Definition.Execute

	cmd := exec.Command(execute.Executable)
	if execute.ScriptFile != nil {
		cmd.Args = append(cmd.Args, *execute.ScriptFile)
	}
	if execute.Arguments != nil {
		cmd.Args = append(cmd.Args, execute.Arguments...)
	}

	stdInPipe, _ := cmd.StdinPipe()
	stdOutPipe, _ := cmd.StdoutPipe()
	stdErrPipe, _ := cmd.StdoutPipe()

	actor.handleStdin(stdInPipe, data)
	stdInPipe.Close()

	cmd.Run()

	actor.handleStdout(stdOutPipe, data)
	stdOutPipe.Close()
	actor.handleStderr(stdErrPipe, data)
	stdErrPipe.Close()

	return cmd.ProcessState.Success()
}

func quote(s string) string {
	if len(s) == 0 {
		return "''"
	}

	if !unsafeQuote.Match([]byte(s)) {
		return s
	}

	return strings.Replace(s, "'", "'\"'\"'", -1)
}

// ExecuteRemote executes the actor remotely
func (actor *Actor) ExecuteRemote(data *ChannelManager, host string, user string, syncRepo bool) bool {
	execute := actor.Definition.Execute

	actorCommand := []string{execute.Executable}
	if execute.ScriptFile != nil {
		actorCommand = append(actorCommand, *execute.ScriptFile)
	}
	if execute.Arguments != nil {
		actorCommand = append(actorCommand, execute.Arguments...)
	}

	for idx := range actorCommand {
		actorCommand[idx] = quote(actorCommand[idx])
	}

	actorRelPath, _ := filepath.Rel(actor.Definition.Registry.ActorDirectory(), actor.Definition.Directory)
	actorRemotePath := filepath.Clean(filepath.Join("/tmp/actors", actorRelPath))

	tempInputFile, err := ioutil.TempFile("", "actorOutput")
	if err != nil {
		return false
	}
	defer os.Remove(tempInputFile.Name())
	tempOutputFile, err := ioutil.TempFile("", "actorInput")
	if err != nil {
		return false
	}
	defer os.Remove(tempOutputFile.Name())

	syncRepoParam := "False"
	if syncRepo {
		syncRepoParam = "True"
	}

	playbookParams := playbookData{
		Host:                host,
		User:                user,
		Playbook:            filepath.Clean(filepath.Join(actor.Definition.Registry.ActorDirectory(), "../playbooks/remote-execute-actor.yaml")),
		ActorCommand:        strings.Join(actorCommand, " "),
		InputFile:           tempInputFile.Name(),
		OutputFile:          tempOutputFile.Name(),
		ActorName:           actor.Definition.Name,
		ActorCWD:            actorRemotePath,
		ActorRepoPath:       actor.Definition.Registry.ActorDirectory(),
		ActorRemoteRepoPath: "/tmp/actors",
		SyncRepo:            syncRepoParam,
	}
	output := bytes.NewBufferString("")
	playbookTemplate.ExecuteTemplate(output, "playbookTemplate", &playbookParams)
	cmdArgs, _ := shlex.Split(output.String(), true)

	actor.handleStdin(tempInputFile, data)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	cmd.Run()
	tempOutputFile.Seek(0, io.SeekStart)
	success := actor.handleRemoteOutput(tempOutputFile, data)

	if !cmd.ProcessState.Success() {
		return false
	}

	return success
}
