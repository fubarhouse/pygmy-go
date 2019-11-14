package library

import (
	"fmt"
	"strings"

	"github.com/fubarhouse/pygmy/v1/service/haproxy_connector"
	model "github.com/fubarhouse/pygmy/v1/service/interface"
	"github.com/fubarhouse/pygmy/v1/service/network"
	"github.com/fubarhouse/pygmy/v1/service/resolv"
)

func Status(c Config) {

	Setup(&c)

	Containers, _ := model.DockerContainerList()
	for _, Container := range Containers {
		if Container.Labels["pygmy"] == "pygmy" {
			name := strings.TrimLeft(Container.Names[0], "/")
			for x, Service := range c.Services {
				if Service.Name == name {
					name = x
				}
			}
			Service := c.Services[name]
			if Service.Name != "" {
				if !Service.Disabled && !Service.Discrete && Service.Name != "" {
					if s, _ := Service.Status(); s {
						fmt.Printf("[*] %v: Running as container %v\n", name, Service.Name)
					} else {
						fmt.Printf("[ ] %v is not running\n", Service.Name)
					}
				}
			} else {
				fmt.Printf("[!] %v: Still running as (no longer configured)\n", name)
			}
		}
	}

	for Network, Containers := range c.Networks {
		netStat, _ := network.Status(Network)
		if netStat {
			for _, Container := range Containers {
				if s, _ := haproxy_connector.Connected(Container, Network); s {
					fmt.Printf("[*] %v is connected to network %v\n", Container, Network)
				} else {
					fmt.Printf("[ ] %v is not connected to network %v\n", Container, Network)
				}
			}
		}
	}

	for _, resolver := range c.Resolvers {
		r := resolv.New(resolv.Resolv{Name: resolver.Name, Data: resolver.Data, Folder: resolver.Folder, File: resolver.File})
		if s := r.Status(); s {
			fmt.Printf("[*] Resolv %v is properly connected\n", resolver.Name)
		} else {
			fmt.Printf("[ ] Resolv %v is not properly connected\n", resolver.Name)
		}
	}

	for _, volume := range c.Volumes {
		if s, _ := model.DockerVolumeExists(volume); s {
			fmt.Printf("[*] Volume %v has been created\n", volume)
		} else {
			fmt.Printf("[ ] Volume %v has not been created\n", volume)
		}
	}

}
