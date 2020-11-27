package main

import (
	"sync"
	"fmt"
	"log"
	"io/ioutil"
	conf "coordinator/config"
	"gopkg.in/yaml.v2"
	"coordinator/web"
	"coordinator/sync_proxy"
	"os"
)

var config *conf.Config = &conf.Config{}

func main() {
	// loading config

	yamlFile, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Println("Read config file error. Exit.")
		return
	}

	log.Printf("%v\n", *config)

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Println("Parsing config file error. Exit.")
		return
	}

	d, err := yaml.Marshal(&config)
	if err != nil {
			log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	// configure logger
	logfile, err := os.OpenFile("coordinator.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer logfile.Close()

    log.SetOutput(logfile)
    fmt.Println("Log file set to coordinator.log")

	var wg sync.WaitGroup

	// start proxy
	wg.Add(1)
	go syncProxy.StartProxy(config)

	// start coordinator web server
	wg.Add(1)
	go web.StartWeb(config)

	fmt.Printf("Finished\n")
	wg.Wait()
}