package cmd

// HelmitContainer represents a single contianer instance in harbor
type HelmitContainer struct {
	Name string `json:"name"`
	Id string `json:"id"`
	Image string `json:"image"`
	Log_stream string `json:"log_stream"`
	Logs []string `json:"logs"`
}

// HelmitReplica represents a single running replica in harbor
type HelmitReplica struct {
	Host string `json:"host"`
	Provider string `json:"provider"`
	Containers []HelmitContainer `json:"containers"`
}

// HelmitResponse represents a response from helmit
type HelmitResponse struct {
	Error bool `json:"error"`
	Replicas []HelmitReplica `json:"replicas"`
}
