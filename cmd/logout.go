package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/harbor-auth-client"
)

// logoutCmd represents the login command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of harbor",
	Long:  ``,
	Run:   logout,
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}

func logout(cmd *cobra.Command, args []string) {
	serializedAuth, err := readFile()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	if serializedAuth != nil {
		_, err := harborLogout(strings.TrimSpace(serializedAuth.Username), strings.TrimSpace(serializedAuth.Token))
		if err != nil {
			log.Fatalf(err.Error())
			return
		}
		deleteFile()
		fmt.Println("Logout Succeeded")
	}
}

func deleteFile() (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, err
	}

	var path = usr.HomeDir + "/.harbor"
	var credPath = path + "/credentials"

	// removing file
	err = os.Remove(credPath)
	if err != nil {
		return false, err
	}

	// removing directory
	err = os.Remove(path)
	if err != nil {
		return false, err
	}

	if Verbose {
		log.Printf("Credentials file removed successfully.")
	}

	return true, nil
}

func harborLogout(username string, token string) (bool, error) {
	client, err := harborauth.NewAuthClient(authURL)
	if err != nil {
		log.Fatalf(err.Error())
		return false, err
	}

	isLoggedOut, err := client.Logout(username, token)
	if err != nil {
		log.Fatalf(err.Error())
		return false, err
	}

	return isLoggedOut, nil
}
