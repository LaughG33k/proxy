package client

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

func CustomDial(proxyAddr, requiredService string, dialTimeout time.Duration) (net.Conn, error) {

	conn, err := net.DialTimeout("tcp", proxyAddr, dialTimeout)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 512)

	if err != nil {
		log.Panic(err)
		return nil, err
	}

	if _, err = conn.Write([]byte(fmt.Sprintf("RQS?%s", requiredService))); err != nil {
		log.Panic(err)
		return nil, err
	}

	n, err := conn.Read(buf)

	if err != nil {
		log.Panic(err)
		return nil, err
	}

	if string(buf[:n]) != "success" {
		return nil, errors.New(string(buf[:n]))
	}

	return conn, nil
}
