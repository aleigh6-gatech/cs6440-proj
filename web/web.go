package web



import (
	"github.com/martini-contrib/render"
	"github.com/go-martini/martini"
	"log"

)

var m martini.ClassicMartini


func StartWeb() {
	m := martini.Classic()
	m.Use(render.Renderer())

	m.Get("/admin", func(r render.Render){
		r.HTML(200, "index", "world")
	})

	m.Get("/", func() string {
		return "ok"
	})

	// m.RunOnAddr(":8080")
	m.Run()
	log.Println("Martini started")
}