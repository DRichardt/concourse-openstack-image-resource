package resource

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/imagedata"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

var REDACT_HEADERS = []string{"x-auth-token", "x-auth-key", "x-service-token",
	"x-storage-token", "x-account-meta-temp-url-key", "x-account-meta-temp-url-key-2",
	"x-container-meta-temp-url-key", "x-container-meta-temp-url-key-2", "set-cookie",
	"x-subject-token"}

type LogRoundTripper struct {
	rt http.RoundTripper
}

func (lrt *LogRoundTripper) logRequest(original io.ReadCloser, contentType string) (io.ReadCloser, error) {
	defer original.Close()

	var bs bytes.Buffer
	_, err := io.Copy(&bs, original)
	if err != nil {
		return nil, err
	}

	// Handle request contentType
	if strings.HasPrefix(contentType, "application/json") {
		debugInfo := lrt.formatJSON(bs.Bytes())
		log.Printf("[DEBUG] OpenStack Request Body: %s", debugInfo)
	}

	return ioutil.NopCloser(strings.NewReader(bs.String())), nil
}

func (lrt *LogRoundTripper) formatJSON(raw []byte) string {
	var rawData interface{}

	err := json.Unmarshal(raw, &rawData)
	if err != nil {
		log.Printf("[DEBUG] Unable to parse OpenStack JSON: %s", err)
		return string(raw)
	}

	data, ok := rawData.(map[string]interface{})
	if !ok {
		pretty, err := json.MarshalIndent(rawData, "", "  ")
		if err != nil {
			log.Printf("[DEBUG] Unable to re-marshal OpenStack JSON: %s", err)
			return string(raw)
		}

		return string(pretty)
	}

	// Mask known password fields
	if v, ok := data["auth"].(map[string]interface{}); ok {
		if v, ok := v["identity"].(map[string]interface{}); ok {
			if v, ok := v["password"].(map[string]interface{}); ok {
				if v, ok := v["user"].(map[string]interface{}); ok {
					v["password"] = "***"
				}
			}
			if v, ok := v["application_credential"].(map[string]interface{}); ok {
				v["secret"] = "***"
			}
			if v, ok := v["token"].(map[string]interface{}); ok {
				v["id"] = "***"
			}
		}
	}

	// Ignore the catalog
	if v, ok := data["token"].(map[string]interface{}); ok {
		if _, ok := v["catalog"]; ok {
			return ""
		}
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("[DEBUG] Unable to re-marshal OpenStack JSON: %s", err)
		return string(raw)
	}

	return string(pretty)
}
func (lrt *LogRoundTripper) logResponse(original io.ReadCloser, contentType string) (io.ReadCloser, error) {
	if strings.HasPrefix(contentType, "application/json") {
		var bs bytes.Buffer
		defer original.Close()
		_, err := io.Copy(&bs, original)
		if err != nil {
			return nil, err
		}
		debugInfo := lrt.formatJSON(bs.Bytes())
		if debugInfo != "" {
			log.Printf("[DEBUG] OpenStack Response Body: %s", debugInfo)
		}
		return ioutil.NopCloser(strings.NewReader(bs.String())), nil
	}

	log.Printf("[DEBUG] Not logging because OpenStack response body isn't JSON")
	return original, nil
}

func (lrt *LogRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	defer func() {
		if request.Body != nil {
			request.Body.Close()
		}
	}()

	var err error

	log.Printf("[DEBUG] OpenStack Request URL: %s %s", request.Method, request.URL)
	log.Printf("[DEBUG] OpenStack request Headers:\n%s", formatHeaders(request.Header))

	if request.Body != nil {
		request.Body, err = lrt.logRequest(request.Body, request.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
	}

	response, err := lrt.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	log.Printf("[DEBUG] OpenStack Response Code: %d", response.StatusCode)
	log.Printf("[DEBUG] OpenStack Response Headers:\n%s", formatHeaders(response.Header))

	response.Body, err = lrt.logResponse(response.Body, response.Header.Get("Content-Type"))

	return response, err
}

func newlHTTPClient() http.Client {
	return http.Client{
		Transport: &LogRoundTripper{
			rt: http.DefaultTransport,
		},
	}

}

func formatHeaders(headers http.Header) string {
	redactedHeaders := redactHeaders(headers)
	sort.Strings(redactedHeaders)

	return strings.Join(redactedHeaders, "\n")
}
func redactHeaders(headers http.Header) (processedHeaders []string) {
	for name, header := range headers {
		var sensitive bool

		for _, redact_header := range REDACT_HEADERS {
			if strings.ToLower(name) == strings.ToLower(redact_header) {
				sensitive = true
			}
		}

		for _, v := range header {
			if sensitive {
				processedHeaders = append(processedHeaders, fmt.Sprintf("%v: %v", name, "***"))
			} else {
				processedHeaders = append(processedHeaders, fmt.Sprintf("%v: %v", name, v))
			}
		}
	}
	return
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

	provider, err := openstack.NewClient(request.Resource.OsAuthURL)
	if err != nil {
		return nil, err
	}
	provider.HTTPClient = newlHTTPClient()
	err = openstack.Authenticate(provider, opts)
	if err != nil {
		return nil, err
	}

	/*	provider, err := openstack.AuthenticatedClient(opts)
		if err != nil {
			return nil, err
		}*/

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
