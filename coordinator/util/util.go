 package util

import (
	"context"
	"io/ioutil"
	"bytes"
	"fmt"
	"net/http"
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
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 400
}


// CloneRequest clones HTTP request
func CloneRequest(req *http.Request) *http.Request {
	// clone body
	newReq := req.Clone(context.TODO())
	*newReq = *req

	var b bytes.Buffer
	b.ReadFrom(req.Body)
	req.Body = ioutil.NopCloser(&b)
	newReq.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))

	// clone url
	// url := fmt.Sprintf("http://%v%v", req.Host, req.URL.Path)

	// clone headers

	for k, vv := range req.Header {
		for _, v := range vv {
			newReq.Header.Add(k, v)
		}
	}

	return newReq
}
