package resource

import (
	"io/ioutil"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
)

//In downloads a image from image store and save it to disk

func In(request InRequest, destinationDir string) (*InResponse, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: request.Resource.OsAuthURL,
		Username:         request.Resource.OsUsername,
		Password:         request.Resource.OsPassword,
		DomainName:       request.Resource.OsUserDomainName,
		TenantName:       request.Resource.OsProjectName,
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, err
	}

	imageClient, err := openstack.NewImageServiceV2(provider, gophercloud.EndpointOpts{
		Region: request.Resource.OsRegion,
	})
	if err != nil {
		return nil, err
	}

	imageID := request.Version.Ref

	image, err := imagedata.Download(imageClient, imageID).Extract()

	if err != nil {
		return nil, err
	}

	imageData, err := ioutil.ReadAll(image)
	if err != nil {
		return nil, err
	}

	savepath := "/tmp/" + request.Resource.Imagename

	err = ioutil.WriteFile(savepath, imageData, 0644)
	if err != nil {
		return nil, err
	}

	response := InResponse{
		Version: Version{Ref: request.Version.Ref},
		Metadata: []Metadata{
			Metadata{
				Name:  "ImageID",
				Value: imageID,
			},
			Metadata{
				Name:  "Name",
				Value: request.Resource.Imagename,
			},
		},
	}

	return &response, nil
}
