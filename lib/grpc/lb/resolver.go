package lb

import (
	"context"
	"math/rand"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	c      *clientv3.Client
	target string
	cc     resolver.ClientConn
	wch    endpoints.WatchChannel
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (r *Resolver) watch() {
	defer r.wg.Done()

	ends := make(map[string]*endpoints.Update)
	for {
		select {
		case <-r.ctx.Done():
			return

		case ups, ok := <-r.wch:
			if !ok {
				return
			}

			for _, up := range ups {
				switch up.Op {
				case endpoints.Add:
					ends[up.Key] = up
				case endpoints.Delete:
					delete(ends, up.Key)
				}
			}

			r.cc.UpdateState(resolver.State{
				Addresses: shuffle(getAddresses(ends)),
			})
		}
	}
}

func getAddresses(ups map[string]*endpoints.Update) []resolver.Address {
	var addrs = make([]resolver.Address, 0, len(ups))
	for _, up := range ups {
		addr := resolver.Address{
			Addr:     up.Endpoint.Addr,
			Metadata: up.Endpoint.Metadata,
		}
		addrs = append(addrs, addr)
	}

	return addrs
}

func shuffle(addresses []resolver.Address) []resolver.Address {
	// 洗牌算法
	rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	for i := len(addresses); i > 0; i-- {
		last := i - 1
		idx := rand.Intn(i)
		addresses[last], addresses[idx] = addresses[idx], addresses[last]
	}

	return addresses
}

func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {
	r.cancel()
	r.wg.Wait()
}
