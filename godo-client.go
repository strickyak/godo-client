/*
 2095  GOPATH=~/go/ go run godo-client.go -f=c --fingerprint ~/keys/id_rsa.pub.fingerprint
 2096  GOPATH=~/go/ go run godo-client.go -f=t temp
 2097  GOPATH=~/go/ go run godo-client.go -f=n temp-2024-03-20--011619
 2098  GOPATH=~/go/ go run godo-client.go -f=d temp-2024-03-20--011619
 2099  GOPATH=~/go/ go run godo-client.go -f=t temp
*/
package main

import (
	"github.com/digitalocean/godo"

	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var Fingerprint = flag.String("fingerprint", "", "filename containing ssh key fingerprint, for droplet creation")

var Command = flag.String("f", "", "Use `c` for create, `l` for long lookup, `d` for delete, `t` for by-tag, `n` to get public IP")

func main() {
	flag.Parse()
	ctx := context.TODO()

	token, err := ioutil.ReadFile("/home/strick/keys/carbon-one")
	if err != nil {
		panic(err)
	}
	client := godo.NewFromToken(string(token))

	switch *Command {

	case "c":
		epoch := time.Now().Format("temp-2006-01-02--150405")

		fp, err := ioutil.ReadFile(*Fingerprint)
		if err != nil {
			log.Fatalf("FATAL: cannot ReadFile fingerprint %q: %v\n\n", *Fingerprint, err)
		}
		fpStr := strings.TrimRight(string(fp), "\r\n")

		sshKey := godo.DropletCreateSSHKey{
			Fingerprint: fpStr,
		}

		_true := true
		createRequest := &godo.DropletCreateRequest{
			Name:   epoch,
			Region: "sfo3",
			Size:   "s-2vcpu-2gb",
			Image: godo.DropletCreateImage{
				Slug: "ubuntu-22-04-x64",
			},
			Tags:     []string{"temp", epoch},
			UserData: "epoch=" + epoch,
			SSHKeys:  []godo.DropletCreateSSHKey{sshKey},
			WithDropletAgent: &_true,
			PrivateNetworking: false,
		}

		newDroplet, _, err := client.Droplets.Create(ctx, createRequest)

		if err != nil {
			log.Fatalf("FATAL: cannot create droplet %q: %v\n\n", epoch, err)
		}

		fmt.Printf("%s\n", newDroplet.Name)

	case "n":
		name := flag.Args()[0]
		droplets, _, err := client.Droplets.ListByName(ctx, name, &godo.ListOptions{})
		if err != nil {
			log.Fatalf("FATAL: cannot ListByName droplet %q: %v\n\n", name, err)
		}
		for _, e := range droplets {
			//  Networks:godo.Networks{V4:[godo.NetworkV4{IPAddress:"165.232.152.245",
			for _, n := range e.Networks.V4 {
				if n.Type == "public" {
					fmt.Printf("%s\n", n.IPAddress)
				}
			}
		}

	case "l":
		name := flag.Args()[0]
		droplets, _, err := client.Droplets.ListByName(ctx, name, &godo.ListOptions{})
		if err != nil {
			log.Fatalf("FATAL: cannot ListByName droplet %q: %v\n\n", name, err)
		}
		fmt.Printf("%v\n", droplets)

	case "t":
		name := flag.Args()[0]
		droplets, _, err := client.Droplets.ListByTag(ctx, name, &godo.ListOptions{})
		if err != nil {
			log.Fatalf("FATAL: cannot ListByTag droplet %q: %v\n\n", name, err)
		}
		for _, e := range droplets {
			fmt.Printf("%s\n", e.Name)
		}

	case "d":
		tag := flag.Args()[0]
		if !strings.HasPrefix(tag, "temp") {
			log.Fatalf("FATAL: Only allowed to delete temp* tags: %q", tag)
		}
		_, err := client.Droplets.DeleteByTag(ctx, tag)
		if err != nil {
			log.Fatalf("FATAL: cannot DeleteByTag droplet %q: %v\n\n", tag, err)
		}
		fmt.Printf("OK\n")

	default:
		panic("TODO:" + *Command)

	}
}
