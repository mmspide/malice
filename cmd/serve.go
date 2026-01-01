// Copyright © 2017 blacktop <https://github.com/blacktop>
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
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/maliceio/malice/api/server"
	"github.com/maliceio/malice/api/server/router"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start Malice API server",
	Long: `Start the Malice REST API server for data access and analysis.
	
The server listens on the configured port and provides endpoints for:
- Scanning files
- Querying results
- Managing plugins
- Accessing analysis data`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe()
	},
}

func runServe() error {
	// Create server configuration
	cfg := &server.Config{
		Logging:     true,
		EnableCors:  true,
		CorsHeaders: "Content-Type, X-Malice-API-Version",
		Version:     "0.4.0",
	}

	// Create new server
	srv := server.New(cfg)
	defer srv.Close()

	// Initialize routes
	routes, err := router.GetRoutes()
	if err != nil {
		return fmt.Errorf("failed to get routes: %w", err)
	}
	srv.InitRouter(routes)

	// Accept connections on localhost:8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	defer listener.Close()

	srv.Accept("localhost:8080", listener)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Infof("Received signal: %v, shutting down gracefully", sig)
		srv.Close()
	}()

	// Run server
	errChan := make(chan error, 1)
	go func() {
		srv.Wait(errChan)
	}()

	// Wait for error or context cancellation
	select {
	case err := <-errChan:
		if err != nil {
			log.Errorf("Server error: %v", err)
			return err
		}
	case <-context.Background().Done():
		log.Info("Server stopped")
	}

	log.Info("Malice API server started successfully on :8080")
	log.Info("Visit http://localhost:8080 to access the API")
	return nil
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 8080, "API server port")
	serveCmd.Flags().BoolP("debug", "d", false, "Enable debug logging")
}
