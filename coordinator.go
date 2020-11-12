package main

import (
	"fmt"
	"log"
	"io/ioutil"
	conf "coordinator/config"
	"gopkg.in/yaml.v2"


)

func main() {
	// proxy.TestProxy()
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Println("Read config file error. Exit.")
		return
	}

	config := conf.Config{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Println("Parsing config file error. Exit.")
		return
	}

	d, err := yaml.Marshal(&config)
	if err != nil {
			log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

}