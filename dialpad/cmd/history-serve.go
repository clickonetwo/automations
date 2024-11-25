/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/clickonetwo/automations/dialpad/internal/history"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
	"github.com/clickonetwo/automations/dialpad/internal/users"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve SMS history from Dialpad",
	Long: `This command runs a server for SMS history from Dialpad.
The server has an HTML interface that is served from the root.`,
	Run: func(cmd *cobra.Command, args []string) {
		envName, err := cmd.InheritedFlags().GetString("env")
		if err != nil {
			panic(err)
		}
		serveHistory(envName)
		fmt.Println("serve called")
	},
}

func init() {
	historyCmd.AddCommand(serveCmd)
}

func serveHistory(envName string) {
	startTime := time.Now()
	_ = storage.PushConfig(envName)
	defer storage.PopConfig()
	config := storage.GetConfig()
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	if err = history.LoadEventHistory(); err != nil {
		panic(err)
	}
	if err = history.LoadAllContacts(); err != nil {
		panic(err)
	}
	defer logger.Sync()
	if config.Name == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := middleware.CreateCoreEngine(logger)
	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "history server running",
			"env":          config.Name,
			"started":      startTime.String(),
			"time":         time.Since(startTime).String(),
			"event_count":  len(history.EventHistory),
			"reader_count": len(users.ListUsers("reader")),
			"admin_count":  len(users.ListUsers("admin")),
		})
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/search")
	})
	r.GET("/history", users.CheckLoginMiddleware, history.RequestHandler)
	r.GET("/search", users.CheckLoginMiddleware, history.SearchHandler)
	r.GET("/login", users.LoginHandler)
	r.GET("/logout", users.LogoutHandler)
	r.GET("/users/:type", users.DownloadUsers)
	r.POST("/users/:type", users.UploadUsers)
	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
	}
	address := "0.0.0.0:" + port
	if config.Name == "development" {
		address = "127.0.0.1:" + port
	}
	logger.Info(
		"Listening on address",
		zap.String("endpoint", config.HerokuHostUrl),
		zap.String("address", address),
	)
	if err := r.Run(address); err != nil {
		panic(err)
	}
}