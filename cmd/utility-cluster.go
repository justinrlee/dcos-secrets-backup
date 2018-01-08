package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type User struct {
	Username string `json:"uid"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
}

type Cluster struct {
	cluster_url string
	client      *http.Client
	user        User
}

// Consists of the path to the secret ("ID") and the AES-encrypted JSON definition.
// JSON format is dependent on DC/OS version, but generally will have a 'value' field.
type Secret struct {
	ID, EncryptedJSON string
}

func createClient() *http.Client {
	// // Create transport to skip verify TODO: add certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	} // TODO: add timeouts here

	return client
}

func NewCluster(hostname string, username string, password string) (cluster *Cluster, err error) {
	if hostname == "" || username == "" || password == "" {
		fmt.Println("Please provide hostname, username, and password")
		return nil, errors.New("")
	}
	var c Cluster
	c.cluster_url = "https://" + hostname
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

	resp, err := c.client.Do(req)

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
	req.Header.Set("Authorization", "token="+string(c.user.Token))

	resp, err := c.client.Do(req)

	if err != nil {
		fmt.Println("TODO: error handling here: request failed")
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)

	return body, resp.StatusCode, err
}

// Get secret
func (c *Cluster) GetSecret(secretID string, cipherKey string, secretChan chan<- Secret, queue chan int) {
		<- queue // Wait for there to be an open spot in the queue
		fmt.Printf("Getting secret '%s'\n", secretID)
		secretJSON, returnCode, err := c.Call("GET", "/secrets/v1/secret/default/"+secretID, nil)
		if err != nil || returnCode != http.StatusOK {
			fmt.Println("Unable to retrieve secret '%s'\n. [%d]: %s", secretID, returnCode, err.Error)
			secretChan <- Secret{ID: "", EncryptedJSON: ""}
			queue <- 0 // Release queue spot
		} else {
			e := encrypt(string(secretJSON), cipherKey)
			secretChan <- Secret{ID: secretID, EncryptedJSON: e}
			queue <- 0 // Release queue spot
		}
}

func (c *Cluster) GetSecrets(secrets []string, cipherKey string, secretChan chan Secret, qsize int) {
	queue := make(chan int, qsize)
	for i:= 0; i < qsize; i++ {
		// fmt.Println("Writing 0 to queue")
		queue <- 0
	}
	// Spins off a bunch of goroutines to get secrets and add them to secretChan.  Should be rate limited by qsize
	for _, secretID := range secrets {
		go c.GetSecret(secretID, cipherKey, secretChan, queue)
	}
}

// for _, secretID := range secrets.Array {
// 	fmt.Printf("Getting secret '%s'\n", secretID)
// 	// secretValue, err := cluster.Get("/secrets/v1/secret/default/" + secretPath)
// 	secretJSON, returnCode, err := cluster.Call("GET", "/secrets/v1/secret/default/"+secretID, nil)
// 	if err != nil || returnCode != http.StatusOK {
// 		fmt.Println("TODO: error handling here")
// 		panic(err)
// 	}

// 	e := encrypt(string(secretJSON), cipherkey)
// 	secretSlice = append(secretSlice, Secret{ID: secretID, EncryptedJSON: e})
// }