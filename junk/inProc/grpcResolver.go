package inProc

import (
	"google.golang.org/grpc/resolver"
	"sync"
)

type grpcResolverBuilder struct {
	state *sync.Mutex
}

func (g grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		target: target,
		cc:     cc,
	}
	r.start()
	return r, nil
}

func (g grpcResolverBuilder) Scheme() string {
	return "in-proc"
}

type grpcResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *grpcResolver) start() {
	r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{
		resolver.Address{
			Addr:       r.target.URL.Host,
			ServerName: r.target.URL.Host,
		},
	}})
}

func (r *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {}

func (r *grpcResolver) Close() {}
