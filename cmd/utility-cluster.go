package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
func (c *Cluster) GetSecret(secretID string, cipherKey string, pool chan int, secretChan chan<- Secret) {
		<- pool // Wait for there to be an open spot in the queue
		defer func() {
			pool <- 0
		}()

		fmt.Printf("Getting secret '%s'\n", secretID)
		secretJSON, returnCode, err := c.Call("GET", "/secrets/v1/secret/default/"+secretID, nil)
		if err != nil || returnCode != http.StatusOK {
			fmt.Println("Unable to retrieve secret '%s'\n. [%d]: %s", secretID, returnCode, err.Error)
			secretChan <- Secret{ID: "", EncryptedJSON: ""}
		} else {
			e := encrypt(string(secretJSON), cipherKey)
			secretChan <- Secret{ID: secretID, EncryptedJSON: e}
		}
}

func (c *Cluster) GetSecrets(secrets []string, cipherKey string, secretChan chan Secret, psize int) {
	pool := make(chan int, psize)
	for i:= 0; i < psize; i++ {
		// fmt.Println("Writing 0 to queue")
		pool <- 0
	}
	// Spins off a bunch of goroutines to get secrets and add them to secretChan.  Should be rate limited by psize
	for _, secretID := range secrets {
		go c.GetSecret(secretID, cipherKey, pool, secretChan)
	}
}

// Will attempt to PUT; if it gets a 409 back (i.e., a 'conflict'), will then attempt a PATCH
func (c * Cluster) PushSecret(secret Secret, cipherKey string, pool chan int, rchan chan<- int) {
	
	// We don't really need to throttle decryption / unmarshalling
	plaintext := decrypt(secret.EncryptedJSON, cipherkey)

	var t struct {
		Value string `json:"value"`
	}
	err := json.Unmarshal([]byte(plaintext), &t)
	if err != nil || t.Value == "" {
		fmt.Printf("Unable to decrypt [%s].  You likely have an invalid cipherkey.\n", secret.ID)
		os.Exit(1)
	}

	fmt.Printf("Queueing secret [%s] ...\n", secret.ID)
	secretPath := "/secrets/v1/secret/default/" + secret.ID

	<- pool // throttle
	defer func() {
		pool <- 0
		rchan <- 0
	}()

	resp, code, err := c.Call("PUT", secretPath, []byte(plaintext))
	if code == 201 {
		fmt.Println("Secret [" + secret.ID + "] successfully created.")
	} else if code == 409 {
		// fmt.Printf("[%s] already exists, updating ...\n", secret.ID)
		presp, pcode, perr := c.Call("PATCH", secretPath, []byte(plaintext))
		if pcode == 204 {
			fmt.Println("Secret [" + secret.ID + "] successfully updated.")
		} else if perr != nil {
			fmt.Printf("Error when attempting to update [%s]: %s\n", secret.ID, perr.Error())
		} else {
			fmt.Printf("Error when attempting to update [%s]. [%s]: %s\n", secret.ID, pcode, string(presp))
		}
	} else if err != nil {
		fmt.Printf("Error when attempting to create [%s]: %s\n", secret.ID, err.Error())
	} else {
		fmt.Printf("Error when attempting to create [%s]. [%s]: %s\n", secret.ID, code, string(resp))
	}


}
