package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mongodb/mongo-go-driver/mongo"
)

// Server struct
// neccesary variables
type Server struct {
	Host   string
	DBHost string
	mongo  *mongo.Client
}

type Log struct {
	Key     string
	Service string
	Type    string
	Tag     string
	Value   string
	Time    time.Time `json:"_id" bson:"_id"`
}

func main() {
	s := Server{}
	s.Host = ":17000"
	s.DBHost = "mongodb://127.0.0.1:27027"

	var err error
	s.mongo, err = mongo.NewClient(s.DBHost)
	if err != nil {
		log.Fatal(err)
	}

	err = s.mongo.Connect(nil)
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}

		log.Printf("New client: %v\n", conn.RemoteAddr())

		go s.handler(conn)
	}
}

func (s *Server) handler(conn net.Conn) {
	for {
		ctx := context.Background()
		l := Log{}

		d := json.NewDecoder(conn)
		err := d.Decode(&l)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client disconnecting: %v\n", conn.RemoteAddr())
				break
			}

			log.Printf("error receiving log.\n%v", err)
			continue
		}

		c := s.mongo.Database("log").Collection(l.Service)
		rslt, err := c.InsertOne(ctx, l)
		if err != nil {
			log.Printf("error logging log.\n%v", err)
			continue
		}

		log.Printf("log logged: %v\n", rslt.InsertedID)
	}

	conn.Close()
}
