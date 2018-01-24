package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/user"
	"runtime"
)

type metric struct {
	Source  string `json:"source,omitempty"`
	Action  string `json:"action,omitempty"`
	Error   string `json:"error,omitempty"`
	OS      string `json:"os,omitempty"`
	Arch    string `json:"arch,omitempty"`
	User    string `json:"user,omitempty"`
	Version string `json:"version,omitempty"`
}

const (
	metricInit       = "init"
	metricPs         = "ps"
	metricLogs       = "logs"
	metricUp         = "up"
	metricDown       = "down"
	metricLogin      = "login"
	metricLogout     = "logout"
	metricRestart    = "restart"
	metricDeploy     = "deploy"
	metricCatalog    = "catalog"
	metricBuildToken = "buildtoken"
	metricGenerate   = "generate"
	metricTerraform  = "terraform"
)

func writeMetric(action string) {
	writeMetricErrorString(action, "")
}

func writeMetricError(action string, err error) {
	writeMetricErrorString(action, err.Error())
}

func writeMetricErrorString(action string, err string) {

	// HARBOR_TELEMETRY=0 disables telemetry
	if setting := os.Getenv("HARBOR_TELEMETRY"); setting != "0" {

		user, e := user.Current()
		check(e)

		m := metric{
			Source:  "harbor-compose",
			Action:  action,
			Error:   err,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
			User:    user.Username,
			Version: Version,
		}

		if Verbose {
			log.Println("posting telemetry data")
		}

		//talk to the server in the background to keep things moving
		go postTelemetryData(m)
	}
}

func getTelemetryEndpoint() string {
	return GetConfig().TelemetryURI + "/v1/api/metric"
}

func postTelemetryData(m metric) {
	json, _ := json.Marshal(m)
	req, err := http.NewRequest("POST", getTelemetryEndpoint(), bytes.NewBuffer(json))
	req.Header.Set("X-key", "0vgKlex4EUckdHYCJq2BPBCyJ5E")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error posting telemetry data: %s\n", err)
		return
	}
	resp.Body.Close()
}
