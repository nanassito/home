package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nanassito/home/pkg/app"
)

var (
	staticRoot = flag.String("static-root", "/github/home/pkg/app/static/", "Root directory of where the static files are stored.")
)

func init() {
	flag.Parse()
}

func main() {
	server := app.NewServer()
	router := httprouter.New()
	router.GET("/", server.GetAir())
	router.GET("/air", server.GetAir())
	router.GET("/air/room/:roomID", server.GetRoom())
	router.POST("/air/room/:roomID", server.PostRoom())

	router.ServeFiles("/static/*filepath", http.Dir(*staticRoot))

	log.Fatal(http.ListenAndServe(":7008", router))
}
