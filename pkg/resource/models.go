package resource

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
