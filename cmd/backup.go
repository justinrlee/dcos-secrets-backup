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
	"encoding/json"

	"github.com/spf13/cobra"
)


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
		validateCipher()

		cluster, err := NewCluster(hostname, username, password)
		if err != nil {
			fmt.Println("Unable to connect to cluster")
			os.Exit(1)
		}

		b, err := cluster.Get("/secrets/v1/secret/default/?list=true")
		if err != nil {
			fmt.Println("Unable to obtain list of secrets")
			os.exit(1)
		}

		var secrets struct{
			Array []string `json:array`
		}

		json.Unmarshal(b, &secrets)

		files := []File {}

		// Get all secrets, add them to the files array
		for _, secretPath := range secrets.Array {
			fmt.Printf("Getting secret '%s'\n", secretPath)
			secretValue, err := cluster.Get("/secrets/v1/secret/default/" + secretPath)
			if err != nil {
				fmt.Println("TODO: error handling here")
				panic(err)
			}

			e := encrypt(string(secretValue), cipherkey)
			files = append(files, File{Path: secretPath, Body: e})
		}

		fmt.Println("Writing to tar at " + destfile)
		writeTar(files, destfile)

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
