package cmd

import "time"

// HelmitContainer represents a single contianer instance in harbor
type HelmitContainer struct {
	Name      string   `json:"name"`
	ID        string   `json:"id"`
	Image     string   `json:"image"`
	Logstream string   `json:"log_stream"`
	Logs      []string `json:"logs"`
}

// HelmitReplica represents a single running replica in harbor
type HelmitReplica struct {
	Host       string            `json:"host"`
	Provider   string            `json:"provider"`
	Containers []HelmitContainer `json:"containers"`
}

// HelmitResponse represents a response from helmit
type HelmitResponse struct {
	Error    bool            `json:"error"`
	Replicas []HelmitReplica `json:"replicas"`
}

//ShipmentStatus represents the deployed status of a shipment
type ShipmentStatus struct {
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
	Status    struct {
		Phase      string `json:"phase"`
		Containers []struct {
			Image    string `json:"image"`
			Ready    bool   `json:"ready"`
			Restarts int    `json:"restarts"`
			State    struct {
				Running struct {
					StartedAt time.Time `json:"startedAt"`
				} `json:"running"`
			} `json:"state"`
			Status    string `json:"status"`
			LastState struct {
			} `json:"lastState"`
		} `json:"containers"`
	} `json:"status"`
	AverageRestarts float32 `json:"averageRestarts"`
}
