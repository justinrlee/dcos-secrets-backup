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
	"os"
	// "strings"

	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		validateCipher()
		// fmt.Println("restore called")
		
		// hostname	:= "54.245.74.53"
		// username	:= "admin"
		// password	:= "thisismypassword"
		// cipherkey := "ThisIsAMagicKeyString12345667890"
		// sourcefile := "out.tar"
		// directory := "temp"

		cluster, err := NewCluster(hostname, username, password)
		if err != nil {
			fmt.Println("Unable to connect to cluster")
			os.Exit(1)
		}

		// fmt.Println(cluster)

		// secretList, err := readFromFile(directory + "/secrets.list")
		// if err != nil {
		// 	fmt.Println("TODO: error handling here")
		// 	panic(err)
		// }
		// secrets := strings.Split(string(secretList), "\n")
		// fmt.Println(secrets)
		files := readTar(sourcefile)
		for _, item := range files {
			plaintext := decrypt(item.Body, cipherkey)
			fmt.Printf("Processing [%s] ...\n", item.Path)
			// fmt.Println(plaintext)

			secretPath := "/secrets/v1/secret/default/" + item.Path
			resp, code, err := cluster.PCall(secretPath, "PUT", []byte(plaintext))
			// fmt.Printf("Code [%s]")
			if code == 201 {
				// fmt.Println("In 201")
				fmt.Println("Secret" + item.Path + "successfully created.")
			} else if code == 409 {
				// fmt.Println("In 409")
				presp, pcode, perr := cluster.PCall(secretPath, "PATCH", []byte(plaintext))
				if pcode == 204 {
					fmt.Println("Secret [" + item.Path + "] successfully updated.")
				} else if perr != nil {
					fmt.Println("Error:")
					fmt.Println(perr)
				} else {
					fmt.Println(string(presp))
					fmt.Println(pcode)
				}
			} else if err != nil {
				// fmt.Println("In Error")
				fmt.Println("Error:")
				fmt.Println(err)
			} else {
				// fmt.Println("In SE")
				fmt.Println("Something else happened:")
				fmt.Printf("HTTP %s: %s\n", code, string(resp))
				fmt.Println(err)
			}
			

			// func (c *Cluster) PCall(path string, verb string, buf []byte) (body []byte, returnCode int, err error) {
		}

		// for _, secretPath := range secrets {
		// 	// fmt.Println(index)
		// 	filePath := directory + "/" + secretPath
		// 	fmt.Println(filePath)

		// 	secret, err := readFromFile(filePath)
		// 	if err != nil {
		// 		fmt.Println("TODO: error handling here")
		// 		panic(err)
		// 	}
		// 	plaintext := decrypt(string(secret), keystring)
		// 	// fmt.Println(plaintext)

		// 	// fmt.Println("Starting put")
		// 	secretUrlPath := "/secrets/v1/secret/default/" + secretPath
		// 	fmt.Println(secretUrlPath)
		// 	fmt.Println("PUT")
		// 	body, statusCode, err := cluster.Put(secretUrlPath, []byte(plaintext))
		// 	// fmt.Println("Ending put")
		// 	if err != nil {
		// 		fmt.Println("PUT failed")
		// 		panic(err)
		// 	}
		// 	fmt.Println(string(body))
		// 	if statusCode != 201 {
		// 		fmt.Println("PATCH")
		// 		body, statusCode, err := cluster.Patch(secretUrlPath, []byte(plaintext))
		// 		fmt.Println(string(body))
		// 		fmt.Println(string(statusCode))
		// 		fmt.Println(err)
		// 	}
		// }
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
