package helm

// Repo represents a configured Helm chart repository.
type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Chart represents a Helm chart entry from search results.
type Chart struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}
