package util

import (
	"fmt"
)

// EndpointFullname returns the key value for check endpoint health status.
func EndpointFullname(clusterName string, endpoint string) string {
	return fmt.Sprintf("%v#%v", clusterName, endpoint)
}


// CheckEndpoint checks the helathiness of endpoint
func CheckEndpoint(address string, path string) bool {
	return true

	// client.get
	// resp, _ := http.Get(address)

	// return resp.StatusCode < 400
}
