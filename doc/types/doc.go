package types

type DocInfo struct {
	ID       string
	Info     *Info
	Host     string `json:"host,omitempty"`
	BasePath string `json:"basePath,omitempty"`
}

type Info struct {
	Description    string       `json:"description,omitempty"`
	Title          string       `json:"title,omitempty"`
	TermsOfService string       `json:"termsOfService,omitempty"`
	Contact        *ContactInfo `json:"contact,omitempty"`
	Version        string       `json:"version,omitempty"`
}

type ContactInfo struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type MethodItem struct {
	ID          string   `json:"operationId,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}
