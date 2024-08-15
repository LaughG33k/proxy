package server

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LaughG33k/proxy"
)

type bind struct {
	conn1, conn2        net.Conn
	i                   proxy.Instance
	l                   proxy.Logger
	idleTimeout         time.Time
	idleTimeoutDuration time.Duration
	clsoeBind           atomic.Bool
	sync.RWMutex
}

func initChain(c1, c2 net.Conn, i proxy.Instance, logger proxy.Logger, idleTimeout time.Duration, readTimeout, writeTimeout int64) *bind {

	if readTimeout > 0 {
		c2.SetReadDeadline(time.Unix(readTimeout, 0))
	}

	if writeTimeout > 0 {
		c2.SetWriteDeadline(time.Unix(writeTimeout, 0))
	}

	b := &bind{
		conn1:               c1,
		conn2:               c2,
		idleTimeout:         time.Now().Add(idleTimeout),
		idleTimeoutDuration: idleTimeout,
		clsoeBind:           atomic.Bool{},
		i:                   i,
		l:                   logger,
	}

	return b

}

func (b *bind) StartDataForwarding() {

	b.log(proxy.Log, fmt.Sprintf("bind of %s and %s start data forwarding", b.conn1.RemoteAddr().String(), b.conn2.RemoteAddr().String()))

	b.i.AddConn()

	go func() {
		for {

			if b.clsoeBind.Load() {
				return
			}

			time.Sleep(b.idleTimeoutDuration)

			if err := b.checkIdleBind(); err != nil {
				b.close()
				return
			}

		}
	}()
	go b.readFromSentTo(b.conn1, b.conn2)
	go b.readFromSentTo(b.conn2, b.conn1)

}

func (b *bind) readFromSentTo(from, to net.Conn) error {

	buf := make([]byte, 8192)

	defer b.close()

	for {

		if b.clsoeBind.Load() {
			return nil
		}

		n, err := from.Read(buf)

		if err != nil {
			return err
		}

		b.log(proxy.Log, fmt.Sprintf("bind of %s and %s received data size of %d", b.conn1.RemoteAddr().String(), b.conn2.RemoteAddr().String(), n))

		b.resetIdle()

		if _, err := to.Write(buf[:n]); err != nil {
			return err
		}

	}

}

func (b *bind) checkIdleBind() error {

	b.Lock()
	defer b.Unlock()

	if b.idleTimeout.Unix() <= time.Now().Unix() {
		return errors.New("exceeded idle timeout")
	}

	return nil

}

func (b *bind) resetIdle() {

	b.Lock()
	defer b.Unlock()

	b.idleTimeout = b.idleTimeout.Add(b.idleTimeoutDuration)

}

func (b *bind) log(level proxy.LevelLog, args ...any) {

	if b.l == nil {
		return
	}

	b.l.Log(level, args)

}

func (b *bind) close() error {

	b.log(proxy.Log, fmt.Sprintf("bind of %s and %s start closing", b.conn1.RemoteAddr().String(), b.conn2.RemoteAddr().String()))

	if b.clsoeBind.Load() {
		b.log(proxy.Log, fmt.Sprintf("bind of %s and %s also closed", b.conn1.RemoteAddr().String(), b.conn2.RemoteAddr().String()))
		return nil
	}

	b.i.RemoveConn()

	b.clsoeBind.Swap(true)

	err1 := b.conn1.Close()
	err2 := b.conn2.Close()

	if err1 != nil {
		return err1
	}

	if err2 != nil {
		return err2
	}

	b.log(proxy.Log, fmt.Sprintf("bind of %s and %s closed", b.conn1.RemoteAddr().String(), b.conn2.RemoteAddr().String()))

	return nil

}
