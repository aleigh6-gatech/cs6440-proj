package main

import (
	"net/url"
	// "encoding/binary"
	"fmt"
	"net/http"
)

func checkEndpoint(address string, path string) bool {
	fullURL := fmt.Sprintf("http://%v/%v", address, path)
	resp, err := http.Get(fullURL)
	if err != nil {
		return false
	}
	return resp.StatusCode < 400
}


func main() {
	// s := "rtest afdasfdas "

	// bm := BinaryMarshaler{}
	// bu := BinaryUnmarshaler{}

	// resp, _ := http.Get("http://google.com")
	// fmt.Printf("%v\n", checkEndpoint("google.com", ""))

	url, _ := url.Parse("google.com")
	fmt.Printf("%v\n", url)
}











