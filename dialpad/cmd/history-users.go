/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
	"github.com/clickonetwo/automations/dialpad/internal/users"
)

// usersCmd represents the users command
var usersCmd = &cobra.Command{
	Use:   "users [flags] path-to-file.csv",
	Short: "Manage history users",
	Long: `This command is used to upload and download users of the history server.
The command downloads to the given file unless the --upload flag is given.
The command downloads "reader" users unless the --admin flag is given.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		env, _ := cmd.InheritedFlags().GetString("env")
		if err := storage.PushConfig(env); err != nil {
			log.Fatal(err)
		}
		defer storage.PopConfig()
		admin, _ := cmd.Flags().GetCount("admin")
		upload, _ := cmd.Flags().GetCount("upload")
		historyUsers(args[0], admin > 0, upload > 0)
	},
}

func init() {
	historyCmd.AddCommand(usersCmd)
	historyCmd.Args = cobra.ExactArgs(1)
	usersCmd.Flags().Count("admin", "manage admin users")
	usersCmd.Flags().Count("upload", "upload users")
}

func historyUsers(path string, doAdmin bool, doUpload bool) {
	env := storage.GetConfig()
	var userType = "reader"
	if doAdmin {
		userType = "admin"
	}
	verb := "GET"
	body := io.Reader(nil)
	if doUpload {
		verb = "POST"
		u, err := users.ImportUserIdsEmails(path)
		if err != nil {
			log.Fatal(err)
		}
		b, err := json.Marshal(u)
		if err != nil {
			log.Fatal(err)
		}
		body = bytes.NewReader(b)
	}
	u := fmt.Sprintf("%s/users/%s", env.HerokuHostUrl, userType)
	req, err := http.NewRequest(verb, u, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+env.MasterAdminId)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	rb, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode > 201 {
		log.Fatal(fmt.Errorf("upload failed with status code %d, body: %s", resp.StatusCode, string(rb)))
	}
	if doUpload {
		log.Printf("Upload of users from %q succeeded.", path)
		return
	}
	var userMap map[string]string
	err = json.Unmarshal(rb, &userMap)
	if err != nil {
		log.Fatal(err)
	}
	err = users.ExportUserIdsEmails(path, userMap)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%d users exported to %q.", len(userMap), path)
}
