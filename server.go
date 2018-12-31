package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mongodb/mongo-go-driver/mongo"
)

// Server struct
// neccesary variables
type Server struct {
	Host   string
	DBHost string
	mongo  *mongo.Client
}

func main() {
	s := Server{}
	s.Host = ":17000"
	s.DBHost = "mongodb://localhost:27017"

	var err error
	s.mongo, err = mongo.NewClient(s.DBHost)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", s.Host)
	if err != nil {
		log.Fatal(err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		listener.Close()
		os.Exit(0)
	}()

	conn, err := listener.Accept()
	if err != nil {
		log.Println(err)
	}

	go s.handler(conn)
}

func (s *Server) handler(conn net.Conn) {

}
