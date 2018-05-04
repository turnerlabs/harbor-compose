package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate [shipment] [environment]",
	Short: "Migrate a shipment environment to another platform",
	Long: `Migrate a shipment environment to another platform

The migrate command outputs files that are useful for migrating a shipment/environment to another platform.  
Note that the migrate command only outputs files and does not perform an actual migration.

The migrate command's --build-provider flag allows you to generate build provider-specific files that allow you to build Docker images and do CI/CD.
`,
	Example: `harbor-compose migrate my-shipment dev
harbor-compose migrate my-shipment dev --platform ecsfargate --build-provider circleciv2
harbor-compose migrate my-shipment prod --platform ecsfargate	
harbor-compose migrate my-shipment prod --platform ecsfargate --role admin
`,
	Run:    migrate,
	PreRun: preRunHook,
}

var migrateBuildProvider string
var migratePlatform string
var migrateRole string

func init() {
	migrateCmd.PersistentFlags().StringVarP(&migratePlatform, "platform", "p", "ecsfargate", "target migration platform")

	migrateCmd.PersistentFlags().StringVarP(&migrateBuildProvider, "build-provider", "b", "", "migrate build provider-specific files that allow you to build Docker images do CI/CD")

	migrateCmd.PersistentFlags().StringVarP(&migrateRole, "role", "r", "devops", "migrate using specified aws role")

	RootCmd.AddCommand(migrateCmd)
}

func migrate(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		os.Exit(-1)
	}

	username, token, err := Login()
	check(err)

	shipment := args[0]
	env := args[1]

	if Verbose {
		log.Printf("fetching shipment...")
	}
	shipmentObject := GetShipmentEnvironment(username, token, shipment, env)
	if shipmentObject == nil {
		fmt.Println(messageShipmentEnvironmentNotFound)
		return
	}

	//convert a Shipment object into a HarborCompose object
	harborCompose := transformShipmentToHarborCompose(shipmentObject)

	//convert a Shipment object into a DockerCompose object, with hidden envvars
	dockerCompose, hiddenEnvVars := transformShipmentToDockerCompose(shipmentObject)

	if migratePlatform != "ecsfargate" {
		check(errors.New("ecsfargate is the only platform currently supported"))
	}

	//output customized migration template
	targetDir, migrationData := migrateToEcsFargate(shipmentObject, &harborCompose)

	//update image in docker-compose.yml
	for _, v := range dockerCompose.Services {
		v.Image = migrationData.NewImage
	}

	//if build provider is specified, allow it modify the compose objects and do its thing
	if len(buildProvider) > 0 {
		provider, err := getBuildProvider(buildProvider)
		check(err)

		artifacts, err := provider.ProvideArtifacts(&dockerCompose, &harborCompose, shipmentObject.BuildToken)
		check(err)

		//write artifacts to file system
		if artifacts != nil {
			for _, artifact := range artifacts {
				//create directories if needed
				dirs := filepath.Dir(artifact.FilePath)
				err = os.MkdirAll(dirs, os.ModePerm)
				check(err)

				if _, err := os.Stat(artifact.FilePath); err == nil {
					//exists
					fmt.Print(artifact.FilePath + " already exists. Overwrite? ")
					if askForConfirmation() {
						err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), artifact.FileMode)
						check(err)
					}
				} else {
					//doesn't exist
					err = ioutil.WriteFile(artifact.FilePath, []byte(artifact.FileContents), artifact.FileMode)
					check(err)
				}
			}
		}
	}

	//prompt if the file already exists
	targetDCFile := filepath.Join(targetDir, DockerComposeFile)
	yes := true
	if _, err := os.Stat(targetDCFile); err == nil {
		fmt.Print("docker-compose.yml already exists. Overwrite? ")
		yes = askForConfirmation()
	}
	if yes {
		SerializeDockerCompose(dockerCompose, targetDCFile)
		fmt.Println("wrote " + DockerComposeFile)
	}

	//prompt if the file already exist
	targetHCFile := filepath.Join(targetDir, HarborComposeFile)
	if _, err := os.Stat(targetHCFile); err == nil {
		fmt.Print("harbor-compose.yml already exists. Overwrite? ")
		yes = askForConfirmation()
	}
	if yes {
		SerializeHarborCompose(harborCompose, targetHCFile)
		fmt.Println("wrote " + HarborComposeFile)
	}

	if len(hiddenEnvVars) > 0 {

		//prompt to override hidden env file
		targetHiddenFile := filepath.Join(targetDir, hiddenEnvFileName)
		if _, err := os.Stat(targetHiddenFile); err == nil {
			fmt.Print(targetHiddenFile + " already exists. Overwrite? ")
			yes = askForConfirmation()
		}
		if yes {
			writeEnvFile(hiddenEnvVars, targetHiddenFile)
			fmt.Println("wrote " + targetHiddenFile)
		}

		//add hidden env_file to .gitignore and .dockerignore (to avoid checking in secrets)
		sensitiveFiles := []string{hiddenEnvFileName, ".terraform"}
		appendToFile(".gitignore", sensitiveFiles)
		appendToFile(".dockerignore", sensitiveFiles)
	}

	fmt.Println()
	fmt.Println("Run the following commands to provision an matching infrastructure stack on the target platform:")
	fmt.Println("cd infrastructure/base")
	fmt.Println("terraform init")
	fmt.Println("terraform apply")
	fmt.Println("cd ../env/" + env)
	fmt.Println("terraform init")
	fmt.Println("terraform apply")
	fmt.Println()
	fmt.Println("Then run the following script to copy your docker image to ECR:")
	fmt.Println("./migrate-image.sh")
	fmt.Println()
	fmt.Println("Then run the following command to deploy your application image and environment variables:")
	fmt.Println("fargate service deploy -f docker-compose.yml")
	fmt.Println()
	fmt.Println("To integrate with DOC monitoring:")
	fmt.Println("./doc-monitoring.sh on")
	fmt.Println()
	fmt.Println("Once you're comfortable with your new enviornment, run the following command to delete your harbor environment:")
	fmt.Println("harbor-compose down --delete")
	fmt.Println()
}
