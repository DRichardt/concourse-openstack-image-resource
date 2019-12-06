package resource

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/sapcc/gophercloud-limes/resources"
	limesprojects "github.com/sapcc/gophercloud-limes/resources/v1/projects"
)

//GetProjectIDAndDomainIDByToken Searches for a Project and Domain in Scope of the actual Token.
func GetProjectIDAndDomainIDByToken(identityClient *gophercloud.ServiceClient, Username string, OsUserDomainName string, ProjectName string) (string, string, error) {

	userlistOpts := users.ListOpts{
		Name:     Username,
		DomainID: OsUserDomainName,
	}

	allUserPages, err := users.List(identityClient, userlistOpts).AllPages()
	if err != nil {
		return "", "", err
	}

	allUsers, err := users.ExtractUsers(allUserPages)
	if err != nil {
		return "", "", err
	}

	allProjectPages, err := users.ListProjects(identityClient, allUsers[0].ID).AllPages()
	if err != nil {
		return "", "", err
	}

	allProjects, err := projects.ExtractProjects(allProjectPages)
	if err != nil {
		return "", "", err
	}

	for _, project := range allProjects {
		if project.Name == ProjectName {
			return project.DomainID, project.ID, nil
		}
	}
	err = fmt.Errorf("Could not find Project with the given User. Check Permissions")
	return "nil", "nil", err
}

//GetProjectUUIDByProjectname searchs for the UUID of a Project
func GetProjectUUIDByProjectname(identityClient *gophercloud.ServiceClient, ProjectName string) (string, error) {
	listopts := projects.ListOpts{
		Enabled: gophercloud.Enabled,
		Name:    ProjectName,
	}

	allPages, err := projects.List(identityClient, listopts).AllPages()
	if err != nil {
		return "", err
	}

	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		return "", err
	}

	return allProjects[0].ID, nil

}

func GetDomainUUIDByDomainname(identityClient *gophercloud.ServiceClient, DomainName string) (string, error) {
	listopts := domains.ListOpts{
		Enabled: gophercloud.Enabled,
		Name:    DomainName,
	}

	allPages, err := domains.List(identityClient, listopts).AllPages()
	if err != nil {
		return "", err
	}

	allProjects, err := domains.ExtractDomains(allPages)
	if err != nil {
		return "", err
	}

	return allProjects[0].ID, nil

}

//IsValidUUID checks if string is valid UUID
func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

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

	identitiyClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{
		Region: request.Resource.OsRegion,
	})
	if err != nil {
		return nil, err
	}

	imageClient, err := openstack.NewImageServiceV2(provider, gophercloud.EndpointOpts{
		Region: request.Resource.OsRegion,
	})
	if err != nil {
		return nil, err
	}

	filepath := BuildDir + "/" + request.Params.File
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
		var propertiesfilepath string
		propertiesfilepath = BuildDir + "/" + request.Params.Properties
		propertiesfilepath = path.Clean(propertiesfilepath)
		propertiesfile, err := os.Open(propertiesfilepath)
		if err != nil {
			return nil, err
		}
		defer propertiesfile.Close()

		propertiesdata, err := ioutil.ReadAll(propertiesfile)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(propertiesdata, &createOpts.Properties)
		if err != nil {
			var errstrings []string
			errstrings = append(errstrings, "Error in unmarshal propertiesfile:")
			errstrings = append(errstrings, err.Error())
			errstrings = append(errstrings, "Body is:")
			errstrings = append(errstrings, string(propertiesdata))
			err = fmt.Errorf(strings.Join(errstrings, "\n"))
			return nil, err
		}

	}
	filesizedata, err := os.Open(filepath)
	defer filesizedata.Close()
	if err != nil {
		return nil, err
	}
	fi, err := filesizedata.Stat()
	if err != nil {
		// Could not obtain stat, handle error
	}

	if request.Params.CheckQuota == true {
		limesclient, err := resources.NewLimesV1(provider, gophercloud.EndpointOpts{
			Region: request.Resource.OsRegion,
		})
		if err != nil {
			var errstrings []string
			errstrings = append(errstrings, "Error while creating limes client:")
			errstrings = append(errstrings, err.Error())
			err = fmt.Errorf(strings.Join(errstrings, "\n"))
			return nil, err
		}

		limesopts := limesprojects.GetOpts{Service: "object-store"}
		if err != nil {
			return nil, err
		}

		opts.Scope.DomainID, opts.Scope.ProjectID, err = GetProjectIDAndDomainIDByToken(identitiyClient, request.Resource.OsUsername, request.Resource.OsUserDomainName, request.Resource.OsProjectName)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		limesquotarequest, err := limesprojects.Get(limesclient, opts.Scope.DomainID, opts.Scope.ProjectID, limesopts).Extract()
		if err != nil {
			var errstrings []string
			errstrings = append(errstrings, "Error while Limes request:")
			errstrings = append(errstrings, err.Error())
			err = fmt.Errorf(strings.Join(errstrings, "\n"))
			return nil, err
		}
		limesquota := limesquotarequest.Services["object-store"].Resources["capacity"]
		filesize := fi.Size()

		if int64(limesquota.Usage)+fi.Size() > int64(limesquota.Quota) {
			var errstrings []string

			errstrings = append(errstrings, "Error: not enogh Quota for Uploading Image. Please incease object-store-quota\n")
			errorstring := fmt.Sprintf("Actual Quota: %s (%s)", humanize.Bytes(uint64(limesquota.Quota)), strconv.FormatUint(limesquota.Quota, 10))
			errstrings = append(errstrings, errorstring)
			errorstring = fmt.Sprintf("Usage: %s (%s)", humanize.Bytes(uint64(limesquota.Usage)), strconv.FormatUint(limesquota.Usage, 10))
			errstrings = append(errstrings, errorstring)
			errorstring = fmt.Sprintf("Filesize: %s (%s)", humanize.Bytes(uint64(filesize)), strconv.FormatInt(filesize, 10))
			errstrings = append(errstrings, errorstring)
			err = fmt.Errorf(strings.Join(errstrings, "\n"))
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
