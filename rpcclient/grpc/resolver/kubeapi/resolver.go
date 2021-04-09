package kubeapi

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"

	"github.com/k8s-practice/octopus/kubectl"
	"github.com/k8s-practice/octopus/xlog"
	"google.golang.org/grpc/resolver"
	"k8s.io/apimachinery/pkg/fields"
)

const (
	scheme = "kubapi"
)

var (
	errIllegalTarget = errors.New("kubeapi resolver: illegal target.")
	logger           = xlog.Component("KubeapiResolver")
)

func init() {
	resolver.Register(&builder{})
}

type builder struct{}

// Build creates and starts a Kubernetes ApiServer resolver
// that watches the name resolution of the target.
func (b *builder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOptions) (resolver.Resolver, error) {
	namespace, endpoint, port, err := parseTarget(target.Endpoint)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &kubeapiResolver{
		namespace: namespace,
		endpoint:  endpoint,
		port:      port,
		ctx:       ctx,
		cancel:    cancel,
		cc:        cc,
		endpoints: make(chan []string, 0),
		stopWatch: make(chan struct{}, 0),
	}

	callback := func(i interface{}) {
		select {
		case r.endpoints <- i.([]string):
		case <-r.ctx.Done():
		}
	}

	if err = kubectl.Subscribe(r.namespace, r.endpoint,
		fields.Everything(), r.stopWatch, callback); err != nil {
		return nil, err
	}

	r.wg.Add(1)
	go r.watcher()

	return r, nil
}

// Scheme returns the naming scheme of this resolver builder, which is "kubeapi".
func (b *builder) Scheme() string {
	return scheme
}

type kubeapiResolver struct {
	namespace string
	endpoint  string
	port      string

	ctx    context.Context
	cancel context.CancelFunc

	cc resolver.ClientConn

	endpoints chan []string
	stopWatch chan struct{}
	wg        sync.WaitGroup
}

func (r *kubeapiResolver) ResolveNow(resolver.ResolveNowOptions) {
}

func (r *kubeapiResolver) Close() {
	close(r.stopWatch)
	r.cancel()
	r.wg.Wait()
}

func (r *kubeapiResolver) watcher() {
	defer r.wg.Done()
	for {
		select {
		case <-r.ctx.Done():
			return
		case endpoints := <-r.endpoints:
			var addrs []resolver.Address
			for i := 0; i < len(endpoints); i++ {
				addrs = append(addrs,
					resolver.Address{Addr: endpoints[i] + ":" + r.port})
			}
			r.cc.UpdateState(resolver.State{Addresses: addrs})
		}
	}
}

// parseTarget takes the user input target string, returns formatted namespace,
// endpoint and port info.
// target: "prod.gate:8080" returns namespace: "prod", endpoint: "gate", port: "8080"
func parseTarget(target string) (namespace, endpoint, port string, err error) {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		return "", "", "", err
	}

	subs := strings.Split(host, ".")
	if len(subs) != 2 {
		return "", "", "", errIllegalTarget
	}

	namespace = subs[0]
	endpoint = subs[1]

	return namespace, endpoint, port, nil
}
