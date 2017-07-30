package main

import (
	"log"
	"net/http"
	"text/template"

	"github.com/justinawrey/stevechatv2/controller"
	"github.com/justinawrey/stevechatv2/server"
)

func main() {
	var s = server.NewServer()                                             // initialize new server
	var t = template.Must(template.ParseGlob("assets/templates/*.gohtml")) // parse all needed templates
	var c = controller.NewController(s, t)                                 // initialize new controller

	// let message handling server run in background
	go s.Serve()

	// routes
	http.HandleFunc("/", c.HandleIndex)
	http.HandleFunc("/login", c.HandleLogin)
	http.HandleFunc("/ws", c.HandleWebsocket)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	// serve files
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
