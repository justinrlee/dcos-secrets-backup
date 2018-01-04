// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"bytes"
	"io/ioutil"
	"crypto/tls"
	"os"
	"net/http"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
)

// User
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
// func (c *Cluster) Get() 

// type sToken struct{
// 	Token	string
// }

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

// func jsonPost (url string, )

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("backup called")

		// Hardcoded for testing
		hostname	:= "54.245.74.53"
		username	:= "admin"
		password	:= "thisismypassword"
		keystring := "ThisIsAMagicKeyString12345667890"
		directory := "temp"
		// cluster_url	:= "https://" + cluster

		cluster, err := NewCluster(hostname, username, password)
		if err != nil {
			fmt.Println("TODO: error handling here")
			panic(err)
		}
		// fmt.Println(cluster)

		b, err := cluster.Get("/secrets/v1/secret/default/?list=true")
		// fmt.Println("b")
		// fmt.Println(string(b))

		var secrets SecretList

		json.Unmarshal(b, &secrets)
		// fmt.Println("secrets")
		fmt.Println(secrets.Array)

		for _, secretPath := range secrets.Array {
			fmt.Printf("Getting secret '%s'\n", secretPath)
			secretValue, err := cluster.Get("/secrets/v1/secret/default/" + secretPath)
			if err != nil {
				fmt.Println("TODO: error handling here")
				panic(err)
			}
			// var secret Secret
			fmt.Println(secretValue)
			fmt.Printf("%T\n", secretValue)
			// json.Unmarshal(secretValue, &secret)
			// fmt.Println(secretValue)
			// fmt.Println(string(secretValue))
			// fmt.Println(secret.Value)
			// fmt.Printf("%T\n", secret.Value)
			filePath := directory + "/" + secretPath
			e := encrypt(string(secretValue), keystring)
			// fmt.Println(e)
			fmt.Println("writing to " + filePath)
			createDirFor(filePath)
			writeToFile(e, filePath)
		}

		list := strings.Join(secrets.Array, "\n")
		fmt.Println(list)
		fmt.Println(directory + "/secrets.list")
		writeToFile(list, directory + "/secrets.list")

		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
