package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jtacoma/uritemplates"
)

func check(e error) {
	if e != nil {
		writeMetricError(currentCommand, currentUser, e)
		//pause here to allow async telemetry call to go through
		time.Sleep(2 * time.Second)

		log.Fatal("ERROR: ", e)
	}
}

//find the ec2 provider
func ec2Provider(providers []ProviderPayload) *ProviderPayload {
	for _, provider := range providers {
		if provider.Name == providerEc2 {
			return &provider
		}
	}
	log.Fatal("ec2 provider is missing")
	return nil
}

func appendToFile(file string, lines []string) {
	if _, err := os.Stat(file); err == nil {
		//update
		file, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
		check(err)
		defer file.Close()
		for _, line := range lines {
			_, err = file.WriteString("\n" + line)
			check(err)
		}
	} else {
		//create
		data := ""
		for _, line := range lines {
			data += line + "\n"
		}
		err := ioutil.WriteFile(file, []byte(data), 0644)
		check(err)
	}
}

type tuple struct {
	Item1 string
	Item2 string
}

func param(item1 string, item2 string) tuple {
	return tuple{
		Item1: item1,
		Item2: item2,
	}
}

func buildURI(baseURI string, template string, params ...tuple) string {
	uriTemplate, err := uritemplates.Parse(baseURI + template)
	check(err)
	values := make(map[string]interface{})
	for _, v := range params {
		values[v.Item1] = v.Item2
	}
	uri, err := uriTemplate.Expand(values)
	check(err)
	return uri
}

func findContainer(container string, containers []ContainerPayload) *ContainerPayload {
	for _, c := range containers {
		if c.Name == container {
			return &c
		}
	}
	return nil
}

func findEnvVar(name string, envVars []EnvVarPayload) EnvVarPayload {
	for _, e := range envVars {
		if e.Name == name {
			return e
		}
	}
	return EnvVarPayload{}
}

//returns a tuple slice containing shipment/environments from user input (cli flags or compose file)
func getShipmentEnvironmentsFromInput(shipmentFlag string, envFlag string) ([]tuple, *HarborCompose) {
	result := []tuple{}
	var hc *HarborCompose

	//either use the shipment/environment flags or the yaml file
	if shipmentFlag != "" && envFlag != "" {
		result = append(result, tuple{Item1: shipmentFlag, Item2: envFlag})
	} else if psShipment != "" && psEnvironment == "" {
		check(errors.New(messageShipmentEnvironmentFlagsRequired))
	} else if shipmentFlag == "" && envFlag != "" {
		check(errors.New(messageShipmentEnvironmentFlagsRequired))
	} else {
		//read the compose file to get the shipment/environment list
		harborComposeConfig := DeserializeHarborCompose(HarborComposeFile)
		hc = &harborComposeConfig
		for shipmentName, shipment := range hc.Shipments {
			result = append(result, tuple{Item1: shipmentName, Item2: shipment.Env})
		}
	}

	return result, hc
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func debug(a ...interface{}) {
	if Verbose {
		log.Println(a...)
	}
}
