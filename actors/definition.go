package actors

var (
	defaultSchemaVersion = "latest"
	defaultHost          = "localhost"
	defaultUser          = "root"
)

type typeDefinition struct {
	Name    string  `yaml:"name"`
	Version *string `yaml:"version,omitempty"`
}

type defaultTypeDefinition typeDefinition

func (d *defaultTypeDefinition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal((*typeDefinition)(d)); err != nil {
		return err
	}
	if d.Version == nil {
		d.Version = &defaultSchemaVersion
	}
	return nil
}

type channelDefinition struct {
	Name string                 `yaml:"name"`
	Type *defaultTypeDefinition `yaml:"type"`
}

type remoteDefinition struct {
	Host *string `yaml:"host,omitempty"`
	User *string `yaml:"user,omitempty"`
}

type optionalRemoteDefinition remoteDefinition

func (o *optionalRemoteDefinition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal((*remoteDefinition)(o)); err != nil {
		return err
	}
	if o.Host == nil {
		o.Host = &defaultHost
	}
	if o.User == nil {
		o.User = &defaultUser
	}
	return nil
}

// Definition contains the definition of actors in yaml files
type Definition struct {
	Registry    *Registry                 `yaml:"-"`
	Directory   string                    `yaml:"-"`
	Name        string                    `yaml:"-"`
	Tags        []string                  `yaml:"tags"`
	Inputs      []channelDefinition       `yaml:"inputs"`
	Outputs     []channelDefinition       `yaml:"outputs"`
	Description string                    `yaml:"description"`
	Execute     *execute                  `yaml:"execute,omitempty"`
	Group       []string                  `yaml:"group,omitempty"`
	Remote      *optionalRemoteDefinition `yaml:"remote,omitempty"`
}

type outputProcessor struct {
	Type   string `yaml:"type"`
	Target string `yaml:"target"`
}

type execute struct {
	Executable      string           `yaml:"executable"`
	ScriptFile      *string          `yaml:"script-file,omitempty"`
	Arguments       []string         `yaml:"arguments"`
	OutputProcessor *outputProcessor `yaml:"output-processor,omitempty"`
}
