package util

import (
	"io/ioutil"
	"bytes"
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


// CloneRequest clones HTTP request
func CloneRequest(req *http.Request) *http.Request {
	// br, _ := req.GetBody()
	body, _ := ioutil.ReadAll(req.Body)

    // you can reassign the body if you need to parse it as multipart
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	newReq, _ := http.NewRequest(req.Method, req.RequestURI, bytes.NewReader(body))

	for k, vv := range req.Header {
		for _, v := range vv {
			newReq.Header.Add(k, v)
		}
	}

	return newReq
}
