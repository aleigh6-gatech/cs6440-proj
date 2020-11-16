package main

import (
	"fmt"
	"log"
	"io/ioutil"
	conf "coordinator/config"
	"gopkg.in/yaml.v2"
	Proxy "coordinator/proxy"
	// DataSync "coordinator/data_sync"
)

var config *conf.Config = &conf.Config{}
// var proxy *CoordinatorProxy


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

	// start proxy
	Proxy.StartProxy(config)



	// start coordinator server

	fmt.Printf("Finished\n")
	for { }
}