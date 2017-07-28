package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var tpls = template.Must(template.ParseGlob("assets/templates/*.gohtml"))

func handleIndex(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie("session")
	if err == http.ErrNoCookie { //user must login first
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	if err = tpls.ExecuteTemplate(w, "index.gohtml", c.Value); err != nil {
		log.Fatalln(err)
	}
}

func handleWebsocket(w http.ResponseWriter, req *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatalln(err)
	}

	notifier := make(chan string)
	s.notifiers[notifier] = true

	// sender
	go func(conn *websocket.Conn, noti chan string) {

		// when connection closes, we need to:
		// close connection handler,
		// delete channel from notifier channel list,
		// send a final message through the channel so receiver goroutine stops
		defer func(conn *websocket.Conn, noti chan string) {
			conn.Close()
			delete(s.notifiers, noti)
			noti <- "closed"
		}(conn, noti)

		for {
			_, bmsg, err := conn.ReadMessage()
			if err != nil {
				break //connection is closed
			}
			smsg := string(bmsg)

			// relay message to server
			s.messagec <- smsg
		}
	}(conn, notifier)

	// receiver
	go func(conn *websocket.Conn, noti chan string) {
		defer conn.Close()
		for {
			// receive message back from server
			res := <-noti
			bres := []byte(res)
			if err := conn.WriteMessage(websocket.TextMessage, bres); err != nil {
				break
			}
		}
	}(conn, notifier)
}

func handleLogin(w http.ResponseWriter, req *http.Request) {
	if _, err := req.Cookie("session"); err == nil {
		http.Redirect(w, req, "/", http.StatusSeeOther) //already logged in
		return
	}

	if req.Method == http.MethodPost {
		un := req.FormValue("un")
		// check against existing sessions
		if _, ok := s.sessions[un]; ok {
			if err := tpls.ExecuteTemplate(w, "login.gohtml", "username already exists"); err != nil {
				log.Fatalln(err)
			}
			return
		}

		// username is not taken
		s.sessions[un] = true

		// set up session cookie
		c := &http.Cookie{
			Name:  "session",
			Value: un,
		}
		http.SetCookie(w, c)
		http.Redirect(w, req, "/", http.StatusSeeOther) // go to chat
		return
	}

	if err := tpls.ExecuteTemplate(w, "login.gohtml", nil); err != nil {
		log.Fatalln(err)
	}
}
