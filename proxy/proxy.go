package proxy

import (
	"time"
	 conf "coordinator/config"
	 "log"
)

var config *conf.Config

func startHealthCheck() {
	go func() {
		for {

			log.Println("health check started")

			// sleep
			time.Sleep( time.Duration(config.HealthCheckInterval) * time.Second )
		}
	}()

}

// StartProxy starts a proxy with config
func StartProxy(newConfig *conf.Config) {
	config = newConfig

	startHealthCheck()

}