// native/desktop/papercloud-cli/cmd/remote/healthcheck.go
package remote

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

func HealthCheckCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "healthcheck",
		Short: "Check server status",
		Long:  `Command will execute call to backend server to check the status of the server.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Performing health check...")

			serverURL := "http://localhost:8000" // TODO: Replace with code that takes it from preference file.

			// Make a GET request to the healthcheck endpoint
			healthCheckURL := fmt.Sprintf("%s/healthcheck", serverURL)
			fmt.Printf("Connecting to: %s\n", healthCheckURL)

			resp, err := http.Get(healthCheckURL)
			if err != nil {
				fmt.Printf("Error connecting to server: %v\n", err)
				return
			}
			defer resp.Body.Close()

			// Check if the response was successful
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Server returned error status: %s\n", resp.Status)
				return
			}

			// Read and display the response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			// Parse the JSON response
			var healthResponse struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal(body, &healthResponse); err != nil {
				fmt.Printf("Error parsing response: %v\n", err)
				fmt.Printf("Raw response: %s\n", string(body))
				return
			}

			// Display the status
			fmt.Printf("Server status: %s\n", healthResponse.Status)
		},
	}

	return cmd
}
