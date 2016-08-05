package cmd

import (
	"fmt"
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
	Logout()
	return
}

//Logout --
func Logout() {
	serializedAuth := readFile()
	if serializedAuth != nil {
		_, err := harborLogout(strings.TrimSpace(serializedAuth.Username), strings.TrimSpace(serializedAuth.Token))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	deleteFile()
	fmt.Println("Logout Succeeded")
	return
}

func deleteFile() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current user info: " + err.Error())
		return
	}

	var path = usr.HomeDir + "/.harbor"
	var credPath = path + "/credentials"

	err = os.Remove(credPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.Remove(path)
	if err != nil {
		fmt.Println(err)
	}

	return
}

func harborLogout(username string, token string) (bool, error) {
	client, err := harborauth.NewAuthClient(authURL)
	isLoggedOut, err := client.Logout(username, token)
	if err != nil || isLoggedOut != true {
		return false, err
	}
	return isLoggedOut, nil
}
