package utility

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	etcdGetOptions = client.GetOptions{
		Recursive: true,
		Sort:      true,
	}
	etcdWatcherOptions = client.WatcherOptions{
		Recursive: true,
	}
)

type namedGrpcClientConn struct {
	*grpc.ClientConn
	name string
}

type EtcdGrpcClientGroupConfig struct {
	Key                     string
	DialOptions             []grpc.DialOption
	WatcherRetryDelay       time.Duration
	WatcherErrorHandler     func(error)
	WatcherUpdateRetry      int
	WatcherUpdateRetryDelay time.Duration
}

type EtcdGrpcClientGroup struct {
	EtcdGrpcClientGroupConfig
	keysAPI client.KeysAPI
	clients []namedGrpcClientConn
}

func NewEtcdGrpcClientGroup(keysAPI client.KeysAPI, config EtcdGrpcClientGroupConfig) (*EtcdGrpcClientGroup, error) {
	group := &EtcdGrpcClientGroup{
		EtcdGrpcClientGroupConfig: config,
		keysAPI:                   keysAPI,
	}
	if err := group.update(); err != nil {
		return nil, err
	}
	go group.watch()
	return group, nil
}

func (group *EtcdGrpcClientGroup) Len() int {
	return len(group.clients)
}

func (group *EtcdGrpcClientGroup) Client(i int) *grpc.ClientConn {
	return group.clients[i].ClientConn
}

func (group *EtcdGrpcClientGroup) update() error {
	resp, err := group.keysAPI.Get(context.Background(), group.Key, &etcdGetOptions)
	if err != nil {
		return err
	}
	old := make(map[string]*grpc.ClientConn)
	for _, client := range group.clients {
		old[client.name] = client.ClientConn
	}
	clients := make([]namedGrpcClientConn, len(resp.Node.Nodes))
	for i, node := range resp.Node.Nodes {
		if client := old[node.Value]; client != nil {
			clients[i] = namedGrpcClientConn{
				ClientConn: client,
				name:       node.Value,
			}
		} else if client, err := grpc.Dial(node.Value, group.DialOptions...); err != nil {
			return fmt.Errorf("dial %q fail: %s", node.Value, err.Error())
		} else {
			clients[i] = namedGrpcClientConn{
				ClientConn: client,
				name:       node.Value,
			}
		}
	}
	group.clients = clients
	return nil
}

func (group *EtcdGrpcClientGroup) watch() {
	for {
		watcher := group.keysAPI.Watcher(group.Key, &etcdWatcherOptions)
		for {
			_, err := watcher.Next(context.Background())
			if err != nil {
				group.handleWatchError(err)
				break
			}
			retry := group.WatcherUpdateRetry
			if retry <= 0 {
				retry = 1
			}
			for i := 0; i < retry; i++ {
				if err := group.update(); err != nil {
					group.handleWatchError(err)
					time.Sleep(group.WatcherUpdateRetryDelay)
					continue
				}
				break
			}
		}
		time.Sleep(group.WatcherRetryDelay)
	}

}

func (group *EtcdGrpcClientGroup) handleWatchError(err error) {
	if group.WatcherErrorHandler != nil {
		group.WatcherErrorHandler(err)
	}
}
