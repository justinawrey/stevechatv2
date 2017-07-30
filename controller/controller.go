package controller

import (
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/justinawrey/stevechatv2/server"
	"github.com/satori/go.uuid"
)

// Controller - handles responding to http requests
type Controller struct {
	srvr *server.Server     // connection to active server
	tpls *template.Template // templates for sending html to client
}

// NewController - initialize new controller
func NewController(s *server.Server, t *template.Template) *Controller {
	return &Controller{
		srvr: s,
		tpls: t,
	}
}

// HandleIndex - handle http requests at route "/"
func (c *Controller) HandleIndex(w http.ResponseWriter, req *http.Request) {
	co, err := req.Cookie("session")
	if err == http.ErrNoCookie { //user must login first
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	// show all past messages, as well as username in bottom left
	data := struct {
		User     string
		Messages []server.Message
	}{
		User:     c.srvr.Sessions[co.Value].Username,
		Messages: c.srvr.Messages,
	}

	if err = c.tpls.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		log.Fatalln(err)
	}
}

// HandleWebsocket - handle ws requests at route "/ws"
func (c *Controller) HandleWebsocket(w http.ResponseWriter, req *http.Request) {
	co, err := req.Cookie("session")
	if err == http.ErrNoCookie {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatalln(err)
	}

	// sender
	go func(conn *websocket.Conn, ssn server.Session) {

		msg := server.Message{
			Ts:     time.Now(),
			Sender: "SYSTEM",
			Text:   ssn.Username + " has joined!",
		}

		c.srvr.Messages = append(c.srvr.Messages, msg) // add message to running message database
		c.srvr.Joinc <- msg                            // broadcast to all connected sessions that user has joined

		// when connection closes, we need to:
		// close connection handler,
		// delete session from active sessions,
		// delete the cookie so user must log back in,
		// and send a final message through the channel so receiver goroutine stops
		defer func(conn *websocket.Conn, ssn server.Session) {
			conn.Close()
			delete(c.srvr.Sessions, co.Value)
			msg := server.Message{
				Ts:     time.Now(),
				Sender: "SYSTEM",
				Text:   ssn.Username + " has left!",
			}
			c.srvr.Messages = append(c.srvr.Messages, msg)
			c.srvr.Leavec <- msg
		}(conn, ssn)

		for {
			_, btxt, err := conn.ReadMessage()
			if err != nil {
				break //connection has been closed
			}

			// relay message to server
			msg := server.Message{
				Ts:     time.Now(),
				Sender: ssn.Username,
				Text:   string(btxt),
			}
			c.srvr.Messages = append(c.srvr.Messages, msg)
			c.srvr.Messagec <- msg
		}
	}(conn, c.srvr.Sessions[co.Value])

	// receiver
	go func(conn *websocket.Conn, ssn server.Session) {
		defer conn.Close()
		for {
			res := <-ssn.Notifier
			bres := []byte(res.String())
			if err := conn.WriteMessage(websocket.TextMessage, bres); err != nil {
				break
			}
		}
	}(conn, c.srvr.Sessions[co.Value])
}

// HandleLogin - handle http requests at "/login"
func (c *Controller) HandleLogin(w http.ResponseWriter, req *http.Request) {
	if _, err := req.Cookie("session"); err == nil {
		http.Redirect(w, req, "/", http.StatusSeeOther) //already logged in
		return
	}

	if req.Method == http.MethodPost {
		un := req.FormValue("un")
		// check username against existing sessions
		if succ, msg := c.srvr.ValidateUsername(un); !succ {
			if err := c.tpls.ExecuteTemplate(w, "login.gohtml", msg); err != nil {
				log.Fatalln(err)
			}
			return
		}

		// username is not taken -- we're good to go
		// set up session cookie
		co := &http.Cookie{
			Name:  "session",
			Value: uuid.NewV4().String(),
		}

		c.srvr.Sessions[co.Value] = server.Session{
			Notifier: make(chan server.Message),
			Username: un,
		}

		http.SetCookie(w, co)
		http.Redirect(w, req, "/", http.StatusSeeOther) // go to chat
		return
	}

	if err := c.tpls.ExecuteTemplate(w, "login.gohtml", nil); err != nil {
		log.Fatalln(err)
	}
}
