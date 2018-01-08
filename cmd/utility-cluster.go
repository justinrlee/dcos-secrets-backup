package cmd

import (
	"fmt"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"errors"
)

type User struct{
	Username	string	`json:"uid"`
	Password	string	`json:"password"`
	Token	string	`json:"token,omitempty"`
}

type Cluster struct{
	cluster_url string
	client *http.Client
	user User
}

// Consists of the path to the secret ("ID") and the AES-encrypted JSON definition.  
//JSON format is dependent on DC/OS version, but generally will have a 'value' field.
type Secret struct {
	ID, EncryptedJSON string
}

func createClient() *http.Client {
	// // Create transport to skip verify TODO: add certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client	:= &http.Client{
		Transport: tr,
	} // TODO: add timeouts here

	return client
}

func NewCluster(hostname string, username string, password string) (cluster *Cluster, err error) {
	if (hostname == "" || username == "" || password == "") {
		fmt.Println("Please provide hostname, username, and password")
		return nil, errors.New("")
	}
	var c Cluster
	c.cluster_url	= "https://" + hostname
	c.user = User{Username: username, Password: password}

	// Create JSON to login
	j, err := json.Marshal(c.user)
	if err != nil {
		fmt.Println("TODO: error handling here utility-cluster NewCluster")
		return nil, err
	}

	// Create client
	c.client = createClient()

	// Login and get token
	err = c.Login("/acs/api/v1/auth/login", j)

	return &c, err

}

func (c *Cluster) Login(path string, buf []byte) (err error) {
	fmt.Printf("Logging into cluster [%s]\n", c.cluster_url)
	url := c.cluster_url + path
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")

	resp, err	:= c.client.Do(req)

	if err != nil {
		fmt.Println("TODO: error handling here utility-cluster Login1")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unable to login (Invalid credentials?)")
		return errors.New("Unable to login (Invalid credentials?)")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("TODO: error handling here utility-cluster Login3")
		return err
	}

	// Will add token to user
	err = json.Unmarshal(body, &c.user)

	return err
}

// Basic wrapper that includes specifying the auth token
func (c *Cluster) Call(verb string, path string, buf []byte) (body []byte, returnCode int, err error) {
	url := c.cluster_url + path
	req, err := http.NewRequest(verb, url, bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token=" + string(c.user.Token))

	resp, err	:= c.client.Do(req)

	if err != nil {
		fmt.Println("TODO: error handling here: request failed")
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)

	return body, resp.StatusCode, err
}