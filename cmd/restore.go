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
		if (destfile != "secrets.tar" && sourcefile == "secrets.tar") {
			fmt.Println("You specified a destination file in a restore command.  Did you mean to specify a source file?")
			os.Exit(1)
		}

		cluster, err := NewCluster(hostname, username, password)
		if err != nil {
			fmt.Println("Unable to connect to cluster")
			os.Exit(1)
		}

		secrets := readTar(sourcefile)
		for _, item := range secrets {
			plaintext := decrypt(item.EncryptedJSON, cipherkey)

			// Prior to restore, will try to unmarshal.  If unable to unmarshal, we have invalid cipherkey.
			var t struct{
				Value string `json:"value"`
			}
			err := json.Unmarshal([]byte(plaintext), &t)
			if (err != nil || t.Value == "")  {
				fmt.Println("Unable to decrypt.  You likely have an invalid cipherkey.")
				os.Exit(1)
			}

			fmt.Printf("Processing [%s] ...\n", item.ID)
			secretPath := "/secrets/v1/secret/default/" + item.ID
			
			resp, code, err := cluster.Call("PUT", secretPath, []byte(plaintext))
			if code == 201 {
				fmt.Println("Secret" + item.ID + "successfully created.")
			} else if code == 409 {
				presp, pcode, perr := cluster.Call("PATCH", secretPath, []byte(plaintext))
				if pcode == 204 {
					fmt.Println("Secret [" + item.ID + "] successfully updated.")
				} else if perr != nil {
					fmt.Println("Error:")
					fmt.Println(perr)
				} else {
					fmt.Println(string(presp))
					fmt.Println(pcode)
				}
			} else if err != nil {
				fmt.Println("Error:")
				fmt.Println(err)
			} else {
				fmt.Println("Something else happened:")
				fmt.Printf("HTTP %s: %s\n", code, string(resp))
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
