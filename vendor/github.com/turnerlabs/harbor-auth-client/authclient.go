package harborauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"
)

type harborAuthClient struct {
	url string
}

// NewAuthClient -
func NewAuthClient(url string) (Auth, error) {
	if len(url) == 0 {
		return nil, errors.New("The servers url is required.")
	}

	c := newHarborAuthClient(url)
	return &c, nil
}

func newHarborAuthClient(url string) harborAuthClient {
	s := harborAuthClient{url: url}
	return s
}

// Login

type loginIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginOut struct {
	Token   string `json:"token"`
	Success bool   `json:"success"`
}

// Login -
func (s *harborAuthClient) Login(username string, password string) (string, bool, error) {
	if !govalidator.IsByteLength(password, 8, 128) {
		return "", false, errors.New("Password is either less than the minimum of 8 or over the maximum number of 128 characters")
	}

	in := loginIn{}
	in.Username = username
	in.Password = password
	inBuf, err := json.Marshal(in)

	fullURL := s.url + "/v1/auth/gettoken"
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(inBuf))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return "", false, err
	}

	// defer the close until we've read all response body data
	defer resp.Body.Close()

	// anything other than 200 is bad
	status := resp.StatusCode
	if status != 200 {
		return "", false, errors.New("Invalid Status Code: " + resp.Status)
	}

	// read the response body data
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false, errors.New("Error in ReadAll")
	}

	// marshall out the byte array data to something we can return
	var out loginOut
	json.Unmarshal(responseBody, &out)

	// Need to verify output before returning

	return out.Token, out.Success, nil
}

// Logout

type logoutIn struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type logoutOut struct {
	Success bool `json:"success"`
}

// Logout -
func (s *harborAuthClient) Logout(username string, token string) (bool, error) {
	if !govalidator.IsAlpha(username) {
		return false, errors.New("Usernames must be alphabetical")
	}

	if len(token) == 0 {
		return false, errors.New("Empty token")
	}

	in := logoutIn{}
	in.Username = username
	in.Token = token
	inBuf, err := json.Marshal(in)

	fullURL := s.url + "/v1/auth/destroytoken"
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(inBuf))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	defer resp.Body.Close()

	status := resp.StatusCode
	if status != 200 {
		return false, errors.New("Invalid Status Code: " + resp.Status)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.New("Error in ReadAll")
	}

	var out logoutOut
	json.Unmarshal(responseBody, &out)

	// Need to verify output before returning

	return out.Success, nil
}

// IsAuthenticated

type isAuthenticatedIn struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type isAuthenticatedOut struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
}

// IsAuthenticated -
func (s *harborAuthClient) IsAuthenticated(username string, token string) (bool, error) {
	if len(token) == 0 {
		return false, errors.New("Empty token")
	}

	in := isAuthenticatedIn{}
	in.Username = username
	in.Token = token
	inBuf, err := json.Marshal(in)

	fullURL := s.url + "/v1/auth/checktoken"
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(inBuf))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	defer resp.Body.Close()

	status := resp.StatusCode
	if status != 200 {
		return false, errors.New("Invalid Status Code: " + resp.Status)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.New("Error in ReadAll")
	}

	var out isAuthenticatedOut
	json.Unmarshal(responseBody, &out)

	// Need to verify output before returning

	return out.Success, nil
}
