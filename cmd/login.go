package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/turnerlabs/harbor-auth-client"
)

//Auth represents a user authentication token
type Auth struct {
	Version  string `json:"version"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:    "login",
	Short:  "Login to harbor",
	Long:   "The login command prompts for your credentials, obtains a temporary token, and stores it your machine so that you don't have to authenticate when running each command.  You can run the logout command to remove your temporary token.",
	Run:    login,
	PreRun: preRunHook,
}

var authURL = "https://auth.services.dmtio.net"

func init() {
	RootCmd.AddCommand(loginCmd)
}

func login(cmd *cobra.Command, args []string) {
	_, _, err := Login()
	if err != nil {
		log.Fatal(err)
	}
}

func writeAuthFile(version string, username string, token string) (bool, error) {
	home, err := homedir.Dir()
	if err != nil {
		return false, err
	}

	curpath, _ := os.Getwd()

	var path = home + "/.harbor"
	err = os.Chdir(path)
	if err != nil {
		err = os.Mkdir(path, 0700)
		if err != nil {
			return false, err
		}
	}

	auth := new(Auth)
	auth.Version = version
	auth.Username = username
	auth.Token = token

	b, _ := json.Marshal(auth)

	authByte := []byte(string(b))
	var credPath = path + "/credentials"
	err = ioutil.WriteFile(credPath, authByte, 0600)
	if err != nil {
		return false, err
	}

	_ = os.Chdir(curpath)

	return true, nil
}

func readAuthFile() (*Auth, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	curpath, _ := os.Getwd()

	var path = home + "/.harbor"
	err = os.Chdir(path)
	if err != nil {
		return nil, err
	}

	var credPath = path + "/credentials"
	byteData, err := ioutil.ReadFile(credPath)
	if err != nil {
		return nil, err
	}

	var serializedAuth Auth
	err = json.Unmarshal(byteData, &serializedAuth)
	if err != nil {
		return nil, err
	}

	_ = os.Chdir(curpath)

	return &serializedAuth, nil
}

//Login -
func Login() (string, string, error) {
	serializedAuth, _ := readAuthFile()

	if serializedAuth != nil {
		isvalid, errHarborAuth := harborAuthenticated(serializedAuth.Username, serializedAuth.Token)
		if errHarborAuth != nil {
			// write it out, isvalid will be false and continue on to force login
			fmt.Println("Unable to verify token: " + errHarborAuth.Error())
		}
		if isvalid {
			setCurrentUser(serializedAuth.Username)
			writeMetric(currentCommand, currentUser)
			return serializedAuth.Username, serializedAuth.Token, nil
		}
	}

	fmt.Println("Login with your Argonauts Login ID to run harbor compose commands. If you don't have a Argonauts Login ID, please reach out in slack to the cloud architecture team.")
	fmt.Print("Username: ")
	var harborUsername string
	fmt.Scanln(&harborUsername)

	fmt.Print("Password: ")
	bytePassword, err := gopass.GetPasswdMasked()
	harborPassword := string(bytePassword)

	harborToken, err := harborLogin(strings.TrimSpace(harborUsername), strings.TrimSpace(harborPassword))
	fmt.Println("")
	if err == nil && len(harborToken) > 1 {
		fmt.Println("Login Succeeded")
		successfullyWritten, errWriteFile := writeAuthFile("v1", harborUsername, harborToken)
		if errWriteFile == nil && successfullyWritten == true {
			setCurrentUser(harborUsername)
			writeMetric(currentCommand, currentUser)
			return harborUsername, harborToken, nil
		}
	}
	return "", "", err
}

//harborLogin -
func harborLogin(username string, password string) (string, error) {
	client, err := harborauth.NewAuthClient(authURL)
	if err != nil {
		if Verbose {
			fmt.Println(err)
		}

		return "", err
	}

	tokenIn, successOut, err := client.Login(username, password)
	if err != nil {
		if Verbose {
			fmt.Println(err)
		}

		return "", err
	}

	if successOut {
		return tokenIn, nil
	}

	return "", err
}

func harborAuthenticated(username string, token string) (bool, error) {
	client, err := harborauth.NewAuthClient(authURL)
	if err != nil {
		return false, err
	}

	isAuth, err := client.IsAuthenticated(username, token)
	if err != nil {
		return false, err
	}

	return isAuth, nil
}
