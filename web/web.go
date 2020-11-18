package web

import (
	"strings"
	"github.com/martini-contrib/render"
	"github.com/go-martini/martini"
	"log"
	conf "coordinator/config"
	"coordinator/proxy"
)

var m martini.ClassicMartini

var config *conf.Config

type healthRow struct {
	Cluster string
	Endpoint string
	Health bool
}


// StartWeb starts web app
func StartWeb(_config *conf.Config) {
	config = _config

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Get("/admin", func(r render.Render){

		// prepare servers health
		rows := []healthRow{}
		for endpointFullname, health := range proxy.HealthStatus {
			tokens := strings.Split(endpointFullname, "#")
			log.Printf("%v\n", tokens)
			row := healthRow{
				Cluster: tokens[0],
				Endpoint: tokens[1],
				Health: health,
			}
			log.Printf("%v\n", row)
			rows = append(rows, row)
		}

		inst := struct {
			HealthcheckInterval int
			ServersHealth []healthRow
		}{
			HealthcheckInterval: config.HealthCheckInterval,
			ServersHealth: rows,
		}

		r.HTML(200, "index", inst)
	})

	m.Get("/", func() string {
		return "ok"
	})

	// m.RunOnAddr(":8080")
	m.Run()
	log.Println("Martini started")
}