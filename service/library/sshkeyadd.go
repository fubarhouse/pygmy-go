package library

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	model "github.com/fubarhouse/pygmy-go/service/interface"
	"github.com/fubarhouse/pygmy-go/service/ssh/agent"
)

// SshKeyAdd will add a given key to the ssh agent.
func SshKeyAdd(c Config, key string, index int) error {

	Setup(&c)

	if key != "" {
		if _, err := os.Stat(key); err != nil {
			fmt.Printf("%v\n", err)
			return err
		}
	}

	var e error

	for _, Container := range c.Services {
		purpose, _ := Container.GetFieldString("purpose")
		if purpose == "addkeys" {
			if !agent.Search(Container, key) {
				if runtime.GOOS == "windows" {
					Container.Config.Cmd = []string{"ssh-add", "/key"}
					Container.HostConfig.Binds = append(Container.HostConfig.Binds, fmt.Sprintf("%v:/key", key))
				} else {
					Container.Config.Cmd = []string{"ssh-add", key}
					Container.HostConfig.Binds = append(Container.HostConfig.Binds, fmt.Sprintf("%v:%v", key, key))
				}

				if index != 0 {

					// We need a brand new copy of the existing container config.
					var newService model.Service
					b, _ := json.Marshal(Container)
					e := json.Unmarshal(b, &newService)
					if e != nil {
						fmt.Println(e)
					}

					name, _ := newService.GetFieldString("name")
					name = strings.SplitAfter(name, "_")[0]

					// For some reason Container works well here but it should be newService - needs investigation.
					e = Container.SetField("name", fmt.Sprintf("%v_%v", name, index))

					if e != nil {
						fmt.Println(e)
					}

					// Remove & recreate the container, but don't output anything.
					newService.Remove()
					newService.Create()
					e = newService.Start()
					if e != nil {
						return e
					}
					l, e := newService.DockerLogs()
					if e != nil {
						return e
					}

					// We need tighter control on the output of this container...
					for _, line := range strings.Split(string(l), "\n") {
						if strings.Contains(line, "Identity added:") {
							fmt.Println(line)
						}
					}

				} else {

					e := Container.Start()
					if e != nil {
						return e
					}
					l, e := Container.DockerLogs()
					if e != nil {
						return e
					}

					// We need tighter control on the output of this container...
					for _, line := range strings.Split(string(l), "\n") {
						if strings.Contains(line, "Identity added:") {
							fmt.Println(line)
						}
					}

				}

			}
		}

	}
	return e
}
