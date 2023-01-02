package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/nanassito/home/pkg/app"
)

func main() {
	server := app.NewServer()
	router := httprouter.New()
	router.GET("/", server.GetAir())
	router.GET("/air", server.GetAir())
	router.GET("/air/room/:roomID", server.GetRoom())
	router.POST("/air/room/:roomID", server.PostRoom())

	// fs := http.FileServer(http.Dir("static/"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe(":7008", router))
}
