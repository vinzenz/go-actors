package actors

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// Registry is where all actors, schema, types etc get registered
type Registry struct {
	actors         map[string]Actor
	actorDirectory string
}

// ActorDirectory returns the directory which was loaded to fill this registry instance
func (r Registry) ActorDirectory() string {
	return r.actorDirectory
}

// MissingDependenciesError is used to report missing dependencies
type MissingDependenciesError struct {
	Missing []string
}

func (m MissingDependenciesError) Error() string {
	return fmt.Sprintf("Missing the following items: %s", strings.Join(m.Missing, ", "))
}

// LoadRegistry creates a registry object with the give path
func LoadRegistry(path string) (*Registry, error) {
	depCheck := map[string]struct{}{}

	r := &Registry{
		actors:         map[string]Actor{},
		actorDirectory: path,
	}

	err := filepath.Walk(path, func(entryPath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == "_actor.yaml" {
			actorDir := filepath.Dir(entryPath)
			actorName := filepath.Base(actorDir)
			data, err := ioutil.ReadFile(entryPath)
			if err != nil {
				return err
			}
			var d Definition
			if err = yaml.Unmarshal(data, &d); err == nil {
				d.Name = actorName
				d.Registry = r
				d.Directory = actorDir
				r.actors[actorName] = Actor{d}
				if _, ok := depCheck[actorName]; ok {
					delete(depCheck, actorName)
				}
				if d.Group != nil {
					for _, item := range d.Group {
						if _, ok := r.actors[item]; !ok {
							depCheck[item] = struct{}{}
						}
					}
				}
			} else {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(depCheck) > 0 {
		missing := []string{}
		for k := range depCheck {
			missing = append(missing, k)
		}
		sort.Strings(missing)
		return nil, MissingDependenciesError{missing}
	}
	return r, nil
}

// Get returns the actor passed by the name or returns nil
func (r *Registry) Get(name string) *Actor {
	if actor, ok := r.actors[name]; ok {
		return &actor
	}
	return nil
}
