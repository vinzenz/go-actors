package main

import (
	"flag"
	"fmt"

	"github.com/vinzenz/go-actor/actors"
)

var (
	actorName string
	actorPath string
)

func init() {
	flag.StringVar(&actorPath, "actorpath", "/usr/share/snactor/actors", "...")
}

func main() {
	flag.Parse()

	actorName = flag.Arg(0)
	data := actors.NewChannelManagerWithInitialData(map[string]interface{}{})
	registry, err := actors.LoadRegistry(actorPath)
	if err != nil {
		panic(".o~O~o._.o~O~o.")
	}
	fmt.Printf("O.o~> actorName => %s\n", actorName)
	actor := registry.Get(actorName)
	result := actor.ExecuteRemote(data, "10.34.76.245", "root", true)
	fmt.Printf("Execution of actor %s (%s) result %t\n", actor.Definition.Name, actor.Definition.Directory, result)
	fmt.Printf("data:\n%v\n", data)
}
