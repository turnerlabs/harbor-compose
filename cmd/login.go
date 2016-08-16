package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"
	"github.com/turnerlabs/harbor-auth-client"
)

//Auth struc
type Auth struct {
	Version  string `json:"version"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to harbor",
	Long:  ``,
	Run:   login,
}

var authURL = "https://auth.services.dmtio.net"

func init() {
	RootCmd.AddCommand(loginCmd)
}

func login(cmd *cobra.Command, args []string) {
	_, _, err := Login()
	if err != nil {
		//log.Printf(err.Error())
		fmt.Println("Login Failed")
	}
	return
}

func writeFile(version string, username string, token string) (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, err
	}

	var path = usr.HomeDir + "/.harbor"
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

	return true, nil
}

func readFile() (*Auth, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	var path = usr.HomeDir + "/.harbor"
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

	return &serializedAuth, nil
}

//Login -
func Login() (string, string, error) {
	serializedAuth, err := readFile()
	if err != nil {
		log.Printf(err.Error())
	}

	if serializedAuth != nil {
		isvalid, errHarborAuth := harborAuthenticated(serializedAuth.Username, serializedAuth.Token)
		if errHarborAuth != nil {
			// write it out, isvalid will be false and continue on to force login
			fmt.Println("Unable to verify token: " + err.Error())
		}
		if isvalid {
			return serializedAuth.Username, serializedAuth.Token, nil
		}
	}

	fmt.Println("Login with your Argonauts Login ID to run harbor compose commands. If you don't have a Argonauts Login ID, please reach out in slack to the cloud architecture team.")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	harborUsername, _ := reader.ReadString('\n')

	fmt.Print("Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	harborPassword := string(bytePassword)

	harborToken, err := harborLogin(strings.TrimSpace(harborUsername), strings.TrimSpace(harborPassword))
	fmt.Println("")
	if err == nil && len(harborToken) > 1 {
		fmt.Println("Login Succeeded")
		successfullyWritten, errWriteFile := writeFile("v1", harborUsername, harborToken)
		if errWriteFile == nil && successfullyWritten == true {
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
