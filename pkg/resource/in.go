package resource

import (
	"io/ioutil"
	"path"

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

	myimage, err := images.Get(imageClient, imageID).Extract()
	if err != nil {
		return nil, err
	}
	savepath := destinationDir + request.Resource.Imagename
	savepath = path.Clean(savepath)

	image, err := imagedata.Download(imageClient, imageID).Extract()

	if err != nil {
		return nil, err
	}

	imageData, err := ioutil.ReadAll(image)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(savepath, imageData, 0644)
	if err != nil {
		return nil, err
	}

	var metadataproperties []Metadata
	prop := myimage.Properties
	for k, p := range prop {
		var m Metadata
		m.Name = k
		if str, ok := p.(string); ok {
			m.Value = str
		}
		if _, ok := p.(bool); ok {
			m.Value = "true"
		}
		metadataproperties = append(metadataproperties, m)
	}

	response := InResponse{
		Version: Version{Ref: request.Version.Ref},
		Metadata: []Metadata{
			Metadata{
				Name:  "ImageID",
				Value: imageID,
			},
			Metadata{
				Name:  "name",
				Value: request.Resource.Imagename,
			},
			Metadata{
				Name:  "container format",
				Value: myimage.ContainerFormat,
			},
			Metadata{
				Name:  "disk format",
				Value: myimage.DiskFormat,
			},
			Metadata{
				Name:  "minimal disk",
				Value: string(myimage.MinDiskGigabytes),
			},
			Metadata{
				Name:  "minimal RAM",
				Value: string(myimage.MinRAMMegabytes),
			},
			Metadata{
				Name:  "visibility",
				Value: string(myimage.Visibility),
			},
		},
	}
	for _, mp := range metadataproperties {
		response.Metadata = append(response.Metadata, mp)
	}

	return &response, nil
}
