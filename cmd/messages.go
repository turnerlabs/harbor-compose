package cmd

const (
	messageHealthCheckRequired              = "a container-level 'HEALTHCHECK' environment variable is required"
	messageEnvvarsCannotBeEmpty             = "environment variable names and value can not be empty"
	messageIntervalGreaterThanTimeout       = "healthcheckIntervalSeconds must be > healthcheckTimeoutSeconds"
	messagePortRequired                     = "at least one port is required"
	messageContainerRequired                = "at least 1 container is required"
	messageReplicaValidation                = "replicas must be between 1 and 1000"
	messageBargeRequired                    = "barge is required for a shipment"
	messageEnvironmentUnderscores           = "environment can not contain underscores ('_')"
	messageShipmentEnvironmentNotFound      = "shipment environment not found"
	messageReplicasMustBeNumber             = "replicas must be a number"
	messageEnableMonitoringTrueFalse        = "please enter true or false for enableMonitoring"
	messageTimeoutValidNumber               = "please enter a valid number for healthcheckTimeoutSeconds"
	messageIntervalValidNumber              = "please enter a valid number for healthcheckIntervalSeconds"
	messageChangeBarge                      = "changing barges involves downtime. Please run the 'down' command first, then change barge and then run 'up' again"
	messageChangePort                       = "port changes involve downtime.  Please run the 'down --delete' command first"
	messageChangeHealthCheck                = "healthcheck changes involve downtime.  Please run the 'down --delete' command first"
	messageChangeContainer                  = "container changes involve downtime.  Please run the 'down --delete' command first"
	messageShipmentEnvironmentFlagsRequired = "both --shipment and --environment flags are required"
)
