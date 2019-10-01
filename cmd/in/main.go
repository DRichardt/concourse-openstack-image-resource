package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/DRichardt/concourse-openstack-image-resource/pkg/resource"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("usage: %s <dest directory>\n", os.Args[0])
	}
	destinationDir := os.Args[1]
	var request resource.CheckRequest
	if os.Getenv("Debug") == "true" {
		file, err := os.Open("in.json")
		if err != nil {
			log.Fatalln(err)
		}
		if err := json.NewDecoder(file).Decode(&request); err != nil {
			log.Fatalln("reading request from stdin", err)
		}
	} else {
		if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
			log.Fatalln("reading request from stdin", err)
		}
	}
	response, err := resource.In(request, destinationDir)
	if err != nil {
		log.Fatalln(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		log.Fatalln("writing response to stdout", err)
	}
}
