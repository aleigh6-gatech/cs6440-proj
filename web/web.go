package web



import (
	"github.com/martini-contrib/render"
	"github.com/go-martini/martini"
	"log"
	conf "coordinator/config"
)

var m martini.ClassicMartini

var config *conf.Config

// StartWeb starts web app
func StartWeb(_config *conf.Config) {
	config = _config

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Get("/admin", func(r render.Render){

		inst := struct {
			HealthcheckInterval int
		}{
			HealthcheckInterval: config.HealthCheckInterval,
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