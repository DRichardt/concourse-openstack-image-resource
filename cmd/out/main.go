package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/DRichardt/concourse-openstack-image-resource/pkg/resource"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <dest directory>\n", os.Args[0])
	}
	buildDir := os.Args[1]
	var request resource.OutRequest
	request.Params.DeleteBrokenImages = true
	if os.Getenv("Debug") == "true" {
		file, err := os.Open("out.json")
		if err != nil {
			log.Fatalln(err)
		}
		if err := json.NewDecoder(file).Decode(&request); err != nil {
			log.Fatalln("reading request from stdin", err)
		}
	} else {
		test, err := os.Create("stdin.txt")
		if err != nil {
			log.Fatalln(err)
		}

		_, err = io.Copy(test, os.Stdin)
		if err != nil {
			log.Fatalln(err)
		}

		helper, err := os.Open("stdin.txt")
		if err != nil {
			log.Fatalln(err)
		}

		if err := json.NewDecoder(helper).Decode(&request); err != nil {
			log.Fatalln("reading request from stdin", err)
		}
	}
	response, err := resource.Out(request, buildDir)
	if err != nil {
		log.Fatalln(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		log.Fatalln("writing response to stdout", err)
	}
}
