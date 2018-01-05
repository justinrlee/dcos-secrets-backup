package cmd

import (
	"fmt"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"bytes"
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

type SecretList struct{
	Array []string `json:array`
}

type Secret struct{
	Value string `json:value` // Cannot unmarshal to []byte
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
	var c Cluster
	c.cluster_url	= "https://" + hostname
	c.user = User{Username: username, Password: password}

	// Create JSON to login
	j, err := json.Marshal(c.user)
	if err != nil {
		fmt.Println("TODO: error handling here")
		return nil, err
	}

	// Create client
	c.client = createClient()

	// Login and get token
	err = c.Login("/acs/api/v1/auth/login", j)

	return &c, err

}

func (c *Cluster) Login(path string, buf []byte) (err error) {
	fmt.Println("Logging in to cluster.")
	url := c.cluster_url + path
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")

	resp, err	:= c.client.Do(req)

	if err != nil {
		fmt.Println("TODO: error handling here")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("TODO: error handling here")
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	// Will add token to user
	err = json.Unmarshal(body, &c.user)

	return err
}

func (c *Cluster) Get(path string) (body []byte, err error) {
	url := c.cluster_url + path
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "token=" + string(c.user.Token))

	resp, err	:= c.client.Do(req)

	if err != nil {
		fmt.Println("TODO: error handling here")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("TODO: error handling here")
		return nil, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("TODO: error handling here")
		return nil, err
	}

	return body, nil
}

func (c *Cluster) Post(path string, buf []byte) (body []byte, err error) {
	url := c.cluster_url + path
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token=" + string(c.user.Token))

	resp, err	:= c.client.Do(req)

	if err != nil {
		fmt.Println("TODO: error handling here")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("TODO: error handling here")
		return nil, err
	}

	body, _ = ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &c.user)
	fmt.Println(c.user)
	// fmt.Println(body)

	return body, nil
}

func (c *Cluster) Put(path string, buf []byte) (body []byte, err error) {
	url := c.cluster_url + path
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token=" + string(c.user.Token))

	resp, err	:= c.client.Do(req)

	if err != nil {
		fmt.Println("TODO: error handling here")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("TODO: error handling here")
		fmt.Println(resp.StatusCode)
		fmt.Println(resp)
		return nil, err
	}

	body, _ = ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &c.user)
	fmt.Println(c.user)
	// fmt.Println(body)

	return body, nil
}
