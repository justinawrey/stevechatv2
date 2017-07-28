package main

import (
	"log"
	"net/http"
)

var s = newServer()

func main() {
	// let message handling server run in background
	go s.serve()

	// routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/ws", handleWebsocket)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	// serve files
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
