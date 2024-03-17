package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
)

func NewPipeListener(n int, ctx context.Context) (clients []net.Conn, listener net.Listener) {
	listenPipes := make([]net.Conn, n)
	clients = make([]net.Conn, n)

	for i := 0; i < n; i++ {
		clients[i], listenPipes[i] = uniqueAddr(net.Pipe())
	}

	ctx, done := context.WithCancel(ctx)

	return clients, &pipeListener{
		n:     n,
		pipes: listenPipes,
		ctx:   ctx,
		done:  done,
	}
}

type pipeListener struct {
	pipes []net.Conn
	n     int
	ctx   context.Context
	done  func()
}

func (p *pipeListener) Accept() (net.Conn, error) {
	if p.n > 0 {
		p.n--
		return p.pipes[p.n], nil
	}
	<-p.ctx.Done()
	return nil, net.ErrClosed
}

func (p *pipeListener) Close() error {
	if p.pipes == nil {
		return net.ErrClosed
	}

	p.done()

	for _, pipe := range p.pipes {
		pipe.Close()
	}

	p.pipes = nil
	return nil
}

func (p *pipeListener) Addr() net.Addr {
	return pipeAddr{}
}

func uniqueAddr(conn1 net.Conn, conn2 net.Conn) (net.Conn, net.Conn) {
	addr1 := pipeAddr{}
	addr2 := pipeAddr{}
	rand.Read(addr1[:])
	rand.Read(addr2[:])

	return &wrapAddr{
			Conn:  conn1,
			laddr: addr1,
			raddr: addr2,
		}, &wrapAddr{
			Conn:  conn2,
			laddr: addr2,
			raddr: addr1,
		}
}

type wrapAddr struct {
	net.Conn
	laddr net.Addr
	raddr net.Addr
}

func (w *wrapAddr) RemoteAddr() net.Addr { return w.raddr }

func (w *wrapAddr) LocalAddr() net.Addr { return w.laddr }

type pipeAddr [128 / 8]byte

func (p pipeAddr) Network() string { return "pipe" }
func (p pipeAddr) String() string  { return fmt.Sprintf("pipe_%v", [128 / 8]byte(p)) }
