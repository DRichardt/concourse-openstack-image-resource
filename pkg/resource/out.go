package resource

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

//Out Uploads image to openstack
func Out(request OutRequest, BuildDir string) (*OutResponse, error) {
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

	filepath := request.Params.File
	filepath = path.Clean(filepath)

	//case statement visability
	createOpts := images.CreateOpts{
		Name:            request.Resource.Imagename,
		Protected:       &request.Params.Protected,
		ContainerFormat: request.Params.ContainerFormat,
		DiskFormat:      request.Params.DiskFormat,
		MinDisk:         request.Params.MinDisk,
		MinRAM:          request.Params.MinRAM,
		Visibility:      &request.Params.Visibility,
	}
	switch request.Params.PropertiesBy {
	case "direct":
		err = json.Unmarshal([]byte(request.Params.Properties), &createOpts.Properties)
		if err != nil {
			return nil, err
		}
	case "file":
		propertiesfile, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		defer propertiesfile.Close()
		if err := json.NewDecoder(propertiesfile).Decode(&request.Params.Properties); err != nil {
			return nil, err
		}
	}

	imageresult := images.Create(imageClient, createOpts)

	image, err := imageresult.Extract()
	if err != nil {
		return nil, err
	}
	checksumdata, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer checksumdata.Close()

	imageData, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer imageData.Close()

	h := md5.New()
	if _, err := io.Copy(h, checksumdata); err != nil {
		return nil, err
	}

	err = imagedata.Upload(imageClient, image.ID, imageData).ExtractErr()
	if err != nil {
		return nil, err
	}

	filechecksum := hex.EncodeToString(h.Sum(nil))

	myimage, err := images.Get(imageClient, image.ID).Extract()
	if err != nil {
		return nil, err
	}

	if myimage.Checksum != filechecksum {
		err = errors.New("Checksum doesn't match after upload")
		return nil, err

		if request.Params.DeleteBrokenImages == true {
			images.Delete(imageClient, image.ID)
		}
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

	response := OutResponse{
		Version: Version{Ref: myimage.ID},
		Metadata: []Metadata{
			Metadata{
				Name:  "ImageID",
				Value: myimage.ID,
			},
			Metadata{
				Name:  "name",
				Value: myimage.Name,
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
