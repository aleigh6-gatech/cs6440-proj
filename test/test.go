package main

import (
	"fmt"
	"time"
	"net/http"
)



func main() {
	client := http.Client{ Timeout: 5 * time.Second }

	resp, _ := client.Get("google.com")
	fmt.Printf("resp %v\n", resp)
}













