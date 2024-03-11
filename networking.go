package main

import "net"

func NewPipeListener() (client net.Conn, listener net.Listener) {
	clientPipe, listenerPipe := net.Pipe()
	return clientPipe, &pipeListener{
		pipe: listenerPipe,
		done: make(chan struct{}),
	}
}

type pipeListener struct {
	pipe net.Conn
	done chan struct{}
}

func (p *pipeListener) Accept() (net.Conn, error) {
	if p.pipe != nil {
		return p.pipe, nil
	}
	<-p.done
	return nil, net.ErrClosed
}

func (p *pipeListener) Close() error {
	return p.pipe.Close()
}

func (p *pipeListener) Addr() net.Addr {
	return p.pipe.LocalAddr()
}
