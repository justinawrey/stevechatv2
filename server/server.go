package server

import "time"

// Message - a chat formatted message
type Message struct {
	Ts     time.Time // time of message send
	Sender string    // username of message sender
	Text   string    // message text
}

// Session - models a single user session
type Session struct {
	Notifier chan Message // notifier channel unique to session for receiving messages from other users
	Username string       // username of session user
}

// Server - models a chat server which receives messages
// from users and relays these messages to all connected sessions
type Server struct {
	Sessions map[string]Session // contains UUID -> session
	Messages []Message          // a log of all messages submitted by all sessions
	Messagec chan Message       // channel through which server receives messages from users
	Joinc    chan Message       // channel through which server is notified that a user has joined
	Leavec   chan Message       // channel through which server is notified that a user has left
}

// NewServer - initialize new Server
func NewServer() *Server {
	return &Server{
		Sessions: make(map[string]Session),
		Messages: make([]Message, 0),
		Messagec: make(chan Message),
		Joinc:    make(chan Message),
		Leavec:   make(chan Message),
	}
}

// notify all connected sessions of some message msg
func (s *Server) notifyAll(msg Message) {
	for _, ssn := range s.Sessions {
		ssn.Notifier <- msg
	}
}

// Serve - launch the server
func (s *Server) Serve() {
	for {
		select {
		case msg := <-s.Joinc:
			s.notifyAll(msg)
		case msg := <-s.Leavec:
			s.notifyAll(msg)
		case msg := <-s.Messagec:
			s.notifyAll(msg)
		}
	}
}

// ValidateUsername - returns false if username exists, true if not
func (s Server) ValidateUsername(un string) (bool, string) {
	// validate username
	// need to do extra validation here
	if un == "SYSTEM" {
		return false, "username is invalid"
	}
	// check against existing usernames
	for _, ssn := range s.Sessions {
		if ssn.Username == un {
			return false, "username is already in use"
		}
	}
	return true, ""
}

// String - prints formatted message in the form of
// hh:mm:ss username > text
// also formats message with italics, bolds, and adds a line break
func (m Message) String() string {
	smsg := "<b>" + m.Ts.String()[11:19] + " " + m.Sender + " > </b>" + m.Text
	if m.Sender == "SYSTEM" {
		smsg = "<i>" + smsg + "</i>"
	}
	return smsg + "<br>"
}
