package main

import (
	"log"
	"time"

	"github.com/LaughG33k/proxy"
	"github.com/LaughG33k/proxy/client"
	"github.com/LaughG33k/proxy/server"
)

type Log struct {
	L *log.Logger
}

func (l *Log) Log(level proxy.LevelLog, args ...interface{}) {

	switch level {

	case proxy.Info:
		l.L.Println(args)

	case proxy.Error:
		l.L.Fatal(args...)

	case proxy.Panic:
		l.L.Panic(args...)

	case proxy.Log:
		l.L.Println(args)

	}

}

func main() {

	l := &Log{L: log.Default()}

	balancer := proxy.NewBalancer(100)

	service := proxy.InitInstance("127.0.0.1:8081", 10000, proxy.Low)

	balancer.AddService("auth")
	balancer.AddInstance("auth", service)

	srv := server.InitServer("127.0.0.1", "8080", balancer)

	if srv == nil {
		log.Panic("server is nil")
	}

	srv.Logger = l

	go srv.AcceptConn()

	conn, err := client.CustomDial("127.0.0.1:8080", "auth", 10*time.Second)

	if err != nil {
		log.Panic(err)
	}

	conn.Write([]byte("hello"))

	for {

	}
}
