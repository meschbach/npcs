package inproc

import (
	"sync"

	"google.golang.org/grpc/resolver"
)

type grpcResolverBuilder struct {
	state *sync.Mutex
}

//nolint:gocritic // hugeParam: interface method signature, cannot change
func (g grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		target: target,
		cc:     cc,
	}
	err := r.start()
	return r, err
}

func (g grpcResolverBuilder) Scheme() string {
	return "in-proc"
}

type grpcResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *grpcResolver) start() error {
	return r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{
		{Addr: r.target.URL.Host, ServerName: r.target.URL.Host},
	}})
}

func (r *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {}

func (r *grpcResolver) Close() {}
