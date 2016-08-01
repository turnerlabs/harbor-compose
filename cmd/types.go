package cmd

// AuthRequest represents an authentication request
type AuthRequest struct {
	User string `json:"username,omitempty"`
	Pass string `json:"password,omitempty"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Success bool
	Token   string
}

// ComposeShipment represents a harbor shipment
type ComposeShipment struct {
	Env         string            `yaml:"env"`
	Barge       string            `yaml:"barge"`
	Containers  []string          `yaml:"containers"`
	Replicas    int               `yaml:"replicas"`
	Group       string            `yaml:"group"`
	Property    string            `yaml:"property"`
	Project     string            `yaml:"project"`
	Product     string            `yaml:"product"`
	Environment map[string]string `yaml:"environment"`
}

// HarborCompose represents a harbor-compose.yml file
type HarborCompose struct {
	Shipments map[string]ComposeShipment `yaml:"shipments"`
}

// DockerCompose represents a docker-compose.yml file
type DockerCompose struct {
	Version  string                          `yaml:"version"`
	Services map[string]DockerComposeService `yaml:"services"`
}

// DockerComposeService represents a container
type DockerComposeService struct {
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports"`
	Environment map[string]string `yaml:"environment"`
}

// ShipmentEnvironment represents a shipment/environment combination
type ShipmentEnvironment struct {
	Name           string             `json:"name,omitempty"`
	EnvVars        []EnvVarPayload    `json:"envVars,omitempty"`
	Ports          []PortPayload      `json:"ports,omitempty"`
	Containers     []ContainerPayload `json:"containers,omitempty"`
	Providers      []ProviderPayload  `json:"providers,omitempty"`
	ParentShipment struct {
		Name    string          `json:"name,omitempty"`
		EnvVars []EnvVarPayload `json:"envVars,omitempty"`
		Group   string          `json:"group,omitempty"`
	}
}

// EnvVarPayload represents EnvVar
type EnvVarPayload struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Type  string `json:"type,omitempty"`
}

// PortPayload represents a port
type PortPayload struct {
	Name                string `json:"name,omitempty"`
	Value               int    `json:"value,omitempty"`
	Protocol            string `json:"protocol,omitempty"`
	Healthcheck         string `json:"healthcheck,omitempty"`
	Primary             bool   `json:"primary,omitempty"`
	External            bool   `json:"external,omitempty"`
	PublicVip           bool   `json:"public_vip,omitempty"`
	PublicPort          int    `json:"public_port,omitempty"`
	EnableProxyProtocol bool   `json:"enable_proxy_protocol,omitempty"`
	SslArn              string `json:"ssl_arn,omitempty"`
	SslManagementType   string `json:"ssl_management_type,omitempty"`
}

// ContainerPayload represents a container payload
type ContainerPayload struct {
	Name    string          `json:"name,omitempty"`
	Image   string          `json:"image,omitempty"`
	EnvVars []EnvVarPayload `json:"envVars,omitempty"`
	Ports   []PortPayload   `json:"ports,omitempty"`
}

// ProviderPayload represents a provider payload
type ProviderPayload struct {
	Name     string          `json:"name"`
	Replicas int             `json:"replicas"`
	EnvVars  []EnvVarPayload `json:"envVars,omitempty"`
	Barge    string          `json:"barge,omitempty"`
}

// TriggerResponseSingle is the payload returned from the trigger api
type TriggerResponseSingle struct {
	Message string `json:"message,omitempty"`
}

// TriggerResponseMultiple is the payload returned from the trigger api
type TriggerResponseMultiple struct {
	Messages []string `json:"message,omitempty"`
}

// NewShipmentEnvironment is used for bulk-creating a new shipment
type NewShipmentEnvironment struct {
	Username    string          `json:"username"`
	Token       string          `json:"token"`
	Info        NewShipmentInfo `json:"main"`
	Environment NewEnvironment  `json:"environment"`
	Containers  []NewContainer  `json:"containers"`
	Providers   []NewProvider   `json:"providers"`
}

// NewShipmentInfo represents new shipment info
type NewShipmentInfo struct {
	Name  string          `json:"name"`
	Group string          `json:"group"`
	Vars  []EnvVarPayload `json:"vars"`
}

// NewEnvironment representsa new environment
type NewEnvironment struct {
	Name string          `json:"name"`
	Vars []EnvVarPayload `json:"vars"`
}

// NewContainer respresents a new container
type NewContainer struct {
	Name    string          `json:"name"`
	Version string          `json:"version"`
	Image   string          `json:"image"`
	Vars    []EnvVarPayload `json:"vars"`
	Ports   []PortPayload   `json:"ports"`
}

// NewProvider represents a new provider
type NewProvider struct {
	Name     string          `json:"name"`
	Replicas int             `json:"replicas"`
	Vars     []EnvVarPayload `json:"vars,omitempty"`
	Barge    string          `json:"barge,omitempty"`
}
