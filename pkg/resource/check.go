package resource

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

//Check Query Image list from Openstack
func Check(request CheckRequest) ([]Version, error) {

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: request.Resource.OsAuthURL,
		Username:         request.Resource.OsUsername,
		Password:         request.Resource.OsPassword,
		DomainName:       request.Resource.OsUserDomainName,
		Scope: &gophercloud.AuthScope{
			ProjectName: request.Resource.OsProjectName,
			DomainName: request.Resource.OsProjectDomainName,
		},
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		log.Fatalln(err)
	}
	listOpts := images.ListOpts{
		Name:    request.Resource.Imagename,
		SortKey: "created_at",
		SortDir: "desc",
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
		if request.Version.Ref == image.ID {
			break
		}

	}
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		log.Fatalln(err)
	}
	return response, nil
}
