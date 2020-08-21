package registercenter

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"log"
	"sync"
	"time"
)

type etcdBuilder struct {
	address     string
	client      *clientv3.Client
	serviceName string
}

func (e *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	e.serviceName = target.Endpoint
	adds, serviceConfig, err := e.resolve()
	if err != nil {
		return nil, err
	}
	cc.NewAddress(adds)
	cc.NewServiceConfig(serviceConfig)
	etcdResolver := NewEtcdResolver(&cc, e, opts)
	etcdResolver.wg.Add(1)
	go etcdResolver.watcher()
	return etcdResolver, nil
}

func (e *etcdBuilder) Scheme() string {
	return "etcd"
}

func NewEtcdBuilder(address string) resolver.Builder {
	client, err := clientv3.NewFromURL(address)
	if err != nil {
		log.Fatal("LearnGrpc: create etcd client error", err.Error())
		return nil
	}
	return &etcdBuilder{address: address, client: client}
}

func (e *etcdBuilder) resolve() ([]resolver.Address, string, error) {
	withRange := clientv3.WithRange(fmt.Sprintf("%s%s", e.serviceName, "/end"))
	res, err := e.client.Get(context.Background(), e.serviceName, withRange)
	if err != nil {
		return nil, "", err
	}
	adds := make([]resolver.Address, 0)
	for i := range res.Kvs {
		if v := res.Kvs[i].Value; v != nil {
			temp := resolver.Address{Addr: string(v), ServerName: e.serviceName}
			adds = append(adds, temp)
		}
	}
	return adds, "", nil
}

type etcdResolver struct {
	clientConn           *resolver.ClientConn
	etcdBuilder          *etcdBuilder
	t                    *time.Ticker
	wg                   sync.WaitGroup
	rn                   chan struct{}
	ctx                  context.Context
	cancel               context.CancelFunc
	disableServiceConfig bool
}

func NewEtcdResolver(cc *resolver.ClientConn, cb *etcdBuilder, opts resolver.BuildOptions) *etcdResolver {
	ctx, cancel := context.WithCancel(context.Background())
	return &etcdResolver{
		clientConn:           cc,
		etcdBuilder:          cb,
		t:                    time.NewTicker(time.Second),
		ctx:                  ctx,
		cancel:               cancel,
		disableServiceConfig: opts.DisableServiceConfig}
}

func (cr *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case cr.rn <- struct{}{}:
	default:
	}
}

func (cr *etcdResolver) Close() {
	cr.cancel()
	cr.wg.Wait()
	cr.t.Stop()
}

func (cr *etcdResolver) watcher() {
	cr.wg.Done()
	for {
		select {
		case <-cr.ctx.Done():
			return
		case <-cr.rn:
		case <-cr.t.C:
		}
		adds, serviceConfig, err := cr.etcdBuilder.resolve()
		if err != nil {
			log.Fatal("query service entries error:", err.Error())
		}
		(*cr.clientConn).NewAddress(adds)
		(*cr.clientConn).NewServiceConfig(serviceConfig)
	}
}

type etcdClientConn struct {
	adds  []resolver.Address
	sc    string
	state resolver.State
}

func NewEtcdClientConn() resolver.ClientConn {
	return &etcdClientConn{}
}
func (cc *etcdClientConn) ReportError(error) {
	panic("implement me")
}

func (cc *etcdClientConn) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	panic("implement me")
}

func (cc *etcdClientConn) NewAddress(addresses []resolver.Address) {
	cc.adds = addresses
}

func (cc *etcdClientConn) NewServiceConfig(serviceConfig string) {
	cc.sc = serviceConfig
}

func (cc *etcdClientConn) UpdateState(state resolver.State) {
	cc.state = state
}

func GenerateAndRegisterEtcdResolver(address string, serviceName string) (schema string, err error) {
	builder := NewEtcdBuilder(address)
	target := resolver.Target{Scheme: builder.Scheme(), Endpoint: serviceName}
	_, err = builder.Build(target, NewEtcdClientConn(), resolver.BuildOptions{})
	if err != nil {
		return builder.Scheme(), err
	}
	resolver.Register(builder)
	schema = builder.Scheme()
	return schema, nil
}
