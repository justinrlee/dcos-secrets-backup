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
	"strings"

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
		// fmt.Println("restore called")
		
		hostname	:= "54.245.74.53"
		username	:= "admin"
		password	:= "thisismypassword"
		keystring := "ThisIsAMagicKeyString12345667890"
		directory := "temp"

		cluster, err := NewCluster(hostname, username, password)
		if err != nil {
			fmt.Println("TODO: error handling here")
			panic(err)
		}

		fmt.Println(cluster)

		secretList, err := readFromFile(directory + "/secrets.list")
		if err != nil {
			fmt.Println("TODO: error handling here")
			panic(err)
		}
		secrets := strings.Split(string(secretList), "\n")
		fmt.Println(secrets)
		for _, secretPath := range secrets {
			// fmt.Println(index)
			filePath := directory + "/" + secretPath
			fmt.Println(filePath)

			secret, err := readFromFile(filePath)
			if err != nil {
				fmt.Println("TODO: error handling here")
				panic(err)
			}
			plaintext := decrypt(string(secret), keystring)
			// fmt.Println(plaintext)

			// fmt.Println("Starting put")
			secretUrlPath := "/secrets/v1/secret/default/" + secretPath
			fmt.Println(secretUrlPath)
			fmt.Println("PUT")
			body, statusCode, err := cluster.Put(secretUrlPath, []byte(plaintext))
			// fmt.Println("Ending put")
			if err != nil {
				fmt.Println("PUT failed")
				panic(err)
			}
			fmt.Println(string(body))
			if statusCode != 201 {
				fmt.Println("PATCH")
				body, statusCode, err := cluster.Patch(secretUrlPath, []byte(plaintext))
				fmt.Println(string(body))
				fmt.Println(string(statusCode))
				fmt.Println(err)
			}
		}

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
