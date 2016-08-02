package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	_, _ = Login()
	return
}

//WriteFile -
func WriteFile(version string, username string, token string) {
	auth := new(Auth)
	auth.Version = version
	auth.Username = username
	auth.Token = token

	b, _ := json.Marshal(auth)

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current user info: " + err.Error())
		fmt.Println("Unable to write credentials file to ~/.harbor")
		return
	}

	var path = usr.HomeDir + "/.harbor"
	err = os.Chdir(path)
	if err != nil {
		err = os.Mkdir(path, 0700)
		if err != nil {
			fmt.Println("Unable to create directory ~/.harbor: " + err.Error())
			fmt.Println("Unable to write credentials file to ~/.harbor")
		}
	}

	d1 := []byte(string(b))
	var credPath = path + "/credentials"
	err = ioutil.WriteFile(credPath, d1, 0600)
	if err != nil {
		fmt.Println("Unable to write credentials file to ~/.harbor: " + err.Error())
	}

	return
}

//ReadFile -
func ReadFile() *Auth {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get Current User Info: " + err.Error())
		fmt.Println("Unable to read credentials file in ~/.harbor")
		return nil
	}

	var path = usr.HomeDir + "/.harbor"
	err = os.Chdir(path)
	if err != nil {
		fmt.Println("Unable to read credentials file in ~/.harbor")
		return nil
	}

	var credPath = path + "/credentials"
	byteData, err := ioutil.ReadFile(credPath)
	if err != nil {
		fmt.Println("Unable to write credentials file to ~/.harbor: " + err.Error())
		return nil
	}

	var serializedAuth Auth
	err = json.Unmarshal(byteData, &serializedAuth)
	if err != nil {
		fmt.Println("Unable to write credentials file to ~/.harbor: " + err.Error())
		return nil
	}

	return &serializedAuth
}

//Login -
func Login() (string, string) {
	serializedAuth := ReadFile()
	if serializedAuth != nil {
		return serializedAuth.Username, serializedAuth.Token
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
		WriteFile("v1", harborUsername, harborToken)
		return harborUsername, harborToken
	}
	fmt.Println(err)
	fmt.Println("Login Failed")
	return "", ""
}

//harborLogin -
func harborLogin(username string, password string) (string, error) {
	client, err := harborauth.NewAuthClient(authURL)
	tokenIn, successOut, err := client.Login(username, password)
	if err != nil && successOut != true {
		return "", err
	}
	return tokenIn, nil
}
