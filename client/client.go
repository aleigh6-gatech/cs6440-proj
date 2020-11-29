package main

import (
	"sync"
	"os"
	"net/http"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
	"fmt"
	"flag"
	 "math/rand"
)

// Config structure for configuration file
type Config struct {
	UpdateConfigInterval int `yaml:"update_config_interval" default:"5"`
	Concurrency int  `yaml:"concurrency" default:"1"`
	WriteRatio float32 `yaml:"write_ratio" default:"0"`
	APIHost string `yaml:"api_host"`
	RequestLogInterval int `yaml:"request_log_interval" default:"30"`
	ReadPath string `yaml:"read_path"`
	WritePath string `yaml:"write_path"`
	ExtraDataDir string `yaml:"extra_data_dir"`
}

var config Config

var extraDataFilepaths []os.FileInfo
var extraDataPtr int

func updateConfig(configPath string) {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Println("Read config file error. Keep last working config.")
		return
	}

	newConfig := Config{}

	err = yaml.Unmarshal(yamlFile, &newConfig)
	if err != nil {
		log.Println("Parsing config file error. Keep last working config.")
		return
	}

	if newConfig != config {
		d, _ := yaml.Marshal(&newConfig)
		log.Printf("Config changed: Dump: %v\n", string(d))
	}

	config = newConfig
}

func sendRandReadRequest() error {
	fullPath := fmt.Sprintf("%s%s", config.APIHost, config.ReadPath)

	_, err := http.Get(fullPath)
	return err
}

func sendRandWriteRequest() error {
	fullPath := fmt.Sprintf("%s%s", config.APIHost, config.WritePath)

	// get content from extra data
	if extraDataPtr == len(extraDataFilepaths) {
		log.Printf("Processed all extra data")
		extraDataPtr++
		return nil
	}

	if extraDataPtr > len(extraDataFilepaths) {
		return nil
	}

	filepath := fmt.Sprintf("%s/%s", config.ExtraDataDir , extraDataFilepaths[extraDataPtr].Name())
	extraDataPtr++

	body, err := os.Open(filepath)
	if err != nil {
		log.Printf("File read error: %v, %v", filepath, err)
		return nil
	}
	defer body.Close()

	_, err = http.Post(fullPath, "application/json", body)
	return err
}

// RequestMetrics stores metrics about requests
type RequestMetrics struct {
	ReadCount int
	WriteCount int
	ReadFailed int
	WriteFailed int
}

var requestMetrics RequestMetrics

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: ")
		flag.PrintDefaults()
	}

	// prepare
	var err error

	updateConfig("./config.yaml")
	extraDataFilepaths, err =  ioutil.ReadDir(config.ExtraDataDir)

	if err != nil {
		log.Printf("Failed to load extra data for writing\n")
	}

	var wg sync.WaitGroup

	// send request according to config
	go func() {
		var interval float32

		for {
			randNum := rand.Float32()
			if randNum < config.WriteRatio {
				// write request
				reqErr := sendRandWriteRequest()
				requestMetrics.WriteCount++

				if reqErr != nil {
					log.Printf("%v\n", reqErr)
					requestMetrics.WriteFailed++
				}
			} else{
				// read request
				reqErr := sendRandReadRequest()
				requestMetrics.ReadCount++

				if reqErr != nil {
					requestMetrics.ReadFailed++
				}
			}

			// sleep
			interval = 1/float32(config.Concurrency)
			time.Sleep( time.Duration(interval*1000) * time.Millisecond )
		}
	}()
	wg.Add(1)

	// read config
	go func() {
		for {
			updateConfig("./config.yaml")
			time.Sleep(time.Duration(config.UpdateConfigInterval) * time.Second)
		}
	}()
	wg.Add(1)

	go func() {
		for {
			log.Printf(
				"Read %d (failed %d), write %d (failed %d)\n",
				requestMetrics.ReadCount,
				requestMetrics.ReadFailed,
				requestMetrics.WriteCount,
				requestMetrics.WriteFailed)
			time.Sleep(time.Duration(config.RequestLogInterval) * time.Second)
		}
	}()
	wg.Add(1)

	wg.Wait()
}