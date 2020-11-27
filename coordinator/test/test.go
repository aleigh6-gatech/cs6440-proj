package main

import (
	// "net/url"
	// // "encoding/binary"
	"fmt"
	// "net/http"
	// "database/sql"
	// "github.com/go-sql-driver/mysql"
	"time"
)

// func checkEndpoint(address string, path string) bool {
// 	fullURL := fmt.Sprintf("http://%v/%v", address, path)
// 	resp, err := http.Get(fullURL)
// 	if err != nil {
// 		return false
// 	}
// 	return resp.StatusCode < 400
// }


func main() {
	uptimeTicker := time.NewTicker(5 * time.Second)
	dateTicker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-uptimeTicker.C:
			fmt.Println("uptime")
		case <-dateTicker.C:
			fmt.Println("date")
		}
	}
}











