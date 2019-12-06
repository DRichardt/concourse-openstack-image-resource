package resource

import "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"


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

//InRequest Struct Data Structure for In Ressource
type InRequest struct {
	Resource Source  `json:"source"`
	Version  Version `json:"version"`
}

//InResponse Struct contains formed Data for a in in a concourse resource
type InResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}

//Metadata contains key value Pair for Metadata
type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//OutRequest contains formed data for concourse out 
type OutRequest struct {
	Resource Source    `json:"source"`
	Params   OutParams `json:"params"`
}

//OutParams params for out
type OutParams struct {
	File               string                 `json:"file"`
	DiskFormat         string                 `json:"disk_format"`
	ContainerFormat    string                 `json:"container_format"`
	MinDisk            int                    `json:"min_disk"`
	MinRAM             int                    `json:"min_ram"`
	Visibility         images.ImageVisibility `json:"visibility"`
	Protected          bool                   `json:"protected"`
	Properties         string                 `json:"properties"`
	PropertiesBy       string                 `json:"properties_by"`
	DeleteBrokenImages bool                   `json:"delete_broken_images"`
	CheckQuota 		   bool 				  `json:"check_quota"`
}

//OutResponse Response from Image Upload
type OutResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}