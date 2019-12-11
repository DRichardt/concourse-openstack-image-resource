package resource

import (
	"io"
	"os"
	"path"
	"strconv"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

//In downloads a image from image store and save it to disk
func In(request InRequest, destinationDir string) (*InResponse, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: request.Resource.OsAuthURL,
		Username:         request.Resource.OsUsername,
		Password:         request.Resource.OsPassword,
		DomainName:       request.Resource.OsUserDomainName,
		Scope: &gophercloud.AuthScope{
			ProjectName: request.Resource.OsProjectName,
			DomainName:  request.Resource.OsProjectDomainName,
		},
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

	out, err := os.Create(savepath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, image)
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
				Value: strconv.Itoa(myimage.MinDiskGigabytes),
			},
			Metadata{
				Name:  "minimal RAM",
				Value: strconv.Itoa(myimage.MinRAMMegabytes),
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
