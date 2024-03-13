package main

import (
	"context"
	"net"
)

func NewPipeListener(ctx context.Context) (client net.Conn, listener net.Listener) {
	clientPipe, listenerPipe := net.Pipe()
	return clientPipe, &pipeListener{
		pipe: listenerPipe,
		ctx:  ctx,
		done: ctx.Done(),
	}
}

type pipeListener struct {
	pipe net.Conn
	ctx  context.Context
	done <-chan struct{}
}

func (p *pipeListener) Accept() (net.Conn, error) {
	if p.pipe != nil {
		tmp := p.pipe
		p.pipe = nil
		return tmp, nil
	}
	<-p.done
	return nil, net.ErrClosed
}

func (p *pipeListener) Close() error {
	if p.pipe == nil {
		return net.ErrClosed
	}

	return p.pipe.Close()
}

func (p *pipeListener) Addr() net.Addr {
	return p.pipe.LocalAddr()
}
