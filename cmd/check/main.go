package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/DRichardt/concourse-openstack-image-resource/pkg/resource"
)

func main() {
	var request resource.CheckRequest
	if os.Getenv("Debug") == "true" {
		file, err := os.Open("check.json")
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

	response, err := resource.Check(request)
	if err != nil {
		log.Fatalln(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		log.Fatalln("writing response to stdout", err)
	}
}
