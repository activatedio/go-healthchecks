package checks_test

import (
	"bufio"
	"context"
	"fmt"
	"github.com/activatedio/go-healthchecks/checks"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

type server struct {
	l net.Listener
}

func (s *server) Run(portCh chan net.Addr) {

	var err error
	s.l, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	portCh <- s.l.Addr()
	for {
		c, err := s.l.Accept()
		if err != nil {
			break
		}

		go func() {
			r := bufio.NewReader(c)
			for {
				msg, err := r.ReadString('\n')
				if err != nil {
					c.Close()
					return
				}
				fmt.Printf("Message incoming: %s", string(msg))
				c.Write([]byte("Message received.\n"))
			}
		}()
	}
}

func (s *server) Stop() {
	err := s.l.Close()
	if err != nil {
		panic(err)
	}
}

func TestTcpChecker(t *testing.T) {

	ctx := context.Background()
	s := &server{}

	addrChan := make(chan net.Addr, 1)

	go s.Run(addrChan)

	addr := <-addrChan

	p := addr.(*net.TCPAddr).Port

	config := &checks.TcpParams{
		Host: "localhost",
		Port: p,
	}

	unit := checks.NewTcpChecker()

	got, err := unit.Check(ctx, config)

	assert.Nil(t, err)
	assert.Equal(t, checks.StatusHealthy, got)

	s.Stop()

	got, err = unit.Check(ctx, config)

	assert.Nil(t, err)
	assert.Equal(t, checks.StatusUnhealthy, got)

	got, err = unit.Check(ctx, &checks.TcpParams{
		Host: "invalid-host",
		Port: 8080,
	})

	assert.Equal(t, checks.StatusUnknown, got)
	assert.EqualError(t, err, "dial tcp: lookup invalid-host: no such host")
}
