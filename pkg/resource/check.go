package resource

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

//Source Strcut with Data for Concourse resource
type Source struct {
	OsUsername          string `json:"OS_USERNAME"`
	OsPassword          string `json:"OS_PASSWORD"`
	OsRegion            string `json:"OS_REGION"`
	OsProjectName       string `json:"OS_PROJECT_NAME"`
	OsAuthURL           string `json:"OS_AUTH_URL"`
	OsUserDomainName    string `json:"OS_USER_DOMAIN_NAME"`
	OsProjectDomainName string `json:"OS_PROJECT_DOMAIN_NAME"`
	Imagename           string `json:"Imagename"`
}

//Version Struct with Data for Version
type Version struct {
	Ref string `json:"ref,omitempty"`
}

//CheckRequest Struct Data Structure for Check Ressource
type CheckRequest struct {
	Resource Source  `json:"source"`
	Version  Version `json:"version"`
}

//Check Query Image list from Openstack
func Check(request CheckRequest) ([]Version, error) {

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: request.Resource.OsAuthURL,
		Username:         request.Resource.OsUsername,
		Password:         request.Resource.OsPassword,
		DomainName:       request.Resource.OsUserDomainName,
		TenantName:       request.Resource.OsProjectName,
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Fatalln(err)
	}
	listOpts := images.ListOpts{
		Name: request.Resource.Imagename,
	}

	imagesClient, err := openstack.NewImageServiceV2(provider, gophercloud.EndpointOpts{
		Region: request.Resource.OsRegion,
	})
	if err != nil {
		log.Fatalln(err)
	}

	allPages, err := images.List(imagesClient, listOpts).AllPages()
	if err != nil {
		panic(err)
	}

	allImages, err := images.ExtractImages(allPages)
	if err != nil {
		log.Fatalln(err)
	}

	response := []Version{}

	for _, image := range allImages {

		response = append(response, Version{Ref: image.ID})

	}
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		log.Fatalln(err)
	}
	return response, nil
}
