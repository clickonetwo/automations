/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/clickonetwo/automations/dialpad/internal/event"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive webhook payloads from Dialpad",
	Long: `This command runs a server process that receives webhook payloads from Dialpad.
The registration of the subscriptions that generate the webhooks must be done separately.
The server process, in addition to serving the webhook endpoint, serves a /status endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		envName, err := cmd.Flags().GetString("env")
		if err != nil {
			panic(err)
		}
		receive(envName)
	},
}

func init() {
	rootCmd.AddCommand(receiveCmd)
	receiveCmd.Flags().StringP("env", "e", "", "processing environment")
}

func receive(envName string) {
	_ = storage.PushConfig(envName)
	defer storage.PopConfig()
	config := storage.GetConfig()
	r := middleware.CreateCoreEngine()
	r.PUT("/hook", event.ReceiveCallWebhook)
	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "receiver running"})
	})
	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
	}
	address := "0.0.0.0:" + port
	if config.Name == "development" {
		address = "127.0.0.1:" + port
	}
	if err := r.Run(address); err != nil {
		panic(err)
	}
}
