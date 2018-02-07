package cmd

import (
	"time"
)

// HelmitContainer represents a single container instance in harbor
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
			ID        string                        `json:"id"`
			Image     string                        `json:"image"`
			Ready     bool                          `json:"ready"`
			Restarts  int                           `json:"restarts"`
			State     map[string]ContainerState     `json:"state"`
			Status    string                        `json:"status"`
			LastState map[string]ContainerLastState `json:"lastState"`
		} `json:"containers"`
	} `json:"status"`
	AverageRestarts float32 `json:"averageRestarts"`
}

//ShipmentEventResult represents system events for a shipment/environment
type ShipmentEventResult struct {
	Namespace string          `json:"namespace"`
	Version   string          `json:"version"`
	Events    []ShipmentEvent `json:"events"`
}

//ShipmentEvent represents a shipment event
type ShipmentEvent struct {
	Type    string `json:"type"`
	Count   int    `json:"count"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Source  struct {
		Component string `json:"component"`
	} `json:"source"`
	FirstTimestamp time.Time `json:"firstTimestamp"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
	StartTime      string
}

// ContainerState represents a particular state of a container
type ContainerState struct {
	StartedAt time.Time `json:"startedAt"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
}

// ContainerLastState represents the last state of a container
type ContainerLastState struct {
	ExitCode    int       `json:"exitCode"`
	Reason      string    `json:"reason"`
	StartedAt   time.Time `json:"startedAt"`
	FinishedAt  time.Time `json:"finishedAt"`
	ContainerID string    `json:"containerID"`
}
