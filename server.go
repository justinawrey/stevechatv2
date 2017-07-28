package main

type server struct {
	notifiers map[chan string]bool
	sessions  map[string]bool
	messagec  chan string
}

func newServer() *server {
	return &server{
		notifiers: make(map[chan string]bool),
		sessions:  make(map[string]bool),
		messagec:  make(chan string),
	}
}

func (s *server) serve() {
	for {
		msg := <-s.messagec
		for n := range s.notifiers {
			n <- msg
		}
	}
}
