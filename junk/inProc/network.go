package inProc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync"
)

type Network struct {
	changes   sync.Mutex
	listeners map[string]*listener
}

func NewNetwork() *Network {
	return &Network{
		changes:   sync.Mutex{},
		listeners: make(map[string]*listener),
	}
}

func (n *Network) Listen(ctx context.Context, address string) (net.Listener, error) {
	n.changes.Lock()
	defer n.changes.Unlock()

	_, exists := n.listeners[address]
	if exists {
		return nil, &AlreadyBound{Address: address}
	}

	pipe := bufconn.Listen(1024 * 1024)
	l := &listener{pipe: pipe}

	n.listeners[address] = l
	return pipe, nil
}

func (n *Network) Dial(ctx context.Context, address string) (net.Conn, error) {
	n.changes.Lock()
	defer n.changes.Unlock()
	l, exists := n.listeners[address]
	if !exists {
		return nil, &NoSuchListener{Address: address}
	}
	return l.pipe.DialContext(ctx)
}

type listener struct {
	pipe *bufconn.Listener
}

type AlreadyBound struct {
	Address string
}

func (a *AlreadyBound) Error() string {
	return fmt.Sprintf("already bound to %s", a.Address)
}

type NoSuchListener struct {
	Address string
}

func (n *NoSuchListener) Error() string {
	return fmt.Sprintf("in-proc: no such listener: %s", n.Address)
}
