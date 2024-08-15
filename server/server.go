package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/LaughG33k/proxy"
)

type Server struct {
	host      string
	port      string
	tlsConfig *tls.Config
	listener  net.Listener
	closed    *atomic.Bool
	b         proxy.IBalancer
	Logger    proxy.Logger
}

func InitServer(host, port string, balancer proxy.IBalancer) *Server {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))

	if err != nil {
		log.Panic(err)
	}

	return &Server{
		host:     host,
		port:     port,
		listener: listener,
		closed:   &atomic.Bool{},
		b:        balancer,
	}

}

func (s *Server) log(level proxy.LevelLog, args ...interface{}) {
	if s.Logger == nil {
		return
	}

	s.Logger.Log(level, args...)

}

func (s *Server) Close() error {

	if s.closed.Load() {
		return errors.New("server also closed")
	}

	return s.listener.Close()

}

func (s *Server) AcceptConn() {

	for {

		if s.closed.Load() {
			break
		}

		conn, err := s.listener.Accept()

		if err != nil {
			log.Panic(err)

		}

		go s.initProxyDial(conn)

	}

}

func (s *Server) initProxyDial(conn net.Conn) {

	buf := make([]byte, 1024)

	n, err := conn.Read(buf)

	if err != nil {
		s.log(proxy.Info, err)
		return
	}

	rqs := string(buf[:n])
	var service string

	fmt.Println(rqs[:4])

	if len(rqs) > 4 && "RQS?" == rqs[:4] {
		service = rqs[4:]
	} else {
		conn.Close()
		return
	}

	fmt.Println(service)

	var conn2 net.Conn
	var i proxy.Instance

	err = proxy.Retry(func() error {
		i = s.b.GetInstance(service)
		c, err := net.DialTimeout("tcp", i.GetAddr(), time.Second*3)

		if err != nil {

			s.log(proxy.Info, err)

			return err

		}

		conn2 = c

		return nil
	}, 5, 0)

	if err != nil {

		s.log(proxy.Info, err)

		if _, err = conn.Write([]byte("error: failed to establish connection")); err != nil {
			s.log(proxy.Info, err)
		}

		conn.Close()

		return
	}

	if _, err = conn.Write([]byte("success")); err != nil {
		s.log(proxy.Info, err)
		conn.Close()
		return
	}

	b := initChain(conn, conn2, i, s.Logger, 10*time.Second, 0, 0)

	b.StartDataForwarding()

}
