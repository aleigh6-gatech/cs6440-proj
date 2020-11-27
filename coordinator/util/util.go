package util

import (
	"fmt"
	"net/http"
	"log"
)

// EndpointFullname returns the key value for check endpoint health status.
func EndpointFullname(clusterName string, endpoint string) string {
	return fmt.Sprintf("%v#%v", clusterName, endpoint)
}


// CheckEndpoint checks the helathiness of endpoint
func CheckEndpoint(enabled bool, address string, path string) bool {
	// Do not check if it is not enabled
	if !enabled {
		return false
	}

	fullURL := fmt.Sprintf("%v/%v", address, path)
	resp, err := http.Get(fullURL)
	log.Printf("check endpoint: %v, %v\n", fullURL, err)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 400
}
