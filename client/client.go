package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
	"github.com/yangchenxing/go-dds/file"
	"github.com/yangchenxing/go-dds/utility"
)

var (
	errNoEtcdClient = errors.New("No Etcd Client")
	errNoServersKey = errors.New("No Servers Key")
)

type Client struct {
	sync.Mutex
	keysAPI        etcd.KeysAPI
	serversClients *utility.EtcdGrpcClientGroup
	peersClients   *utility.EtcdGrpcClientGroup
}

func NewClient(keysAPI etcd.KeysAPI, serversGroupConfig, peersGroupConfig utility.EtcdGrpcClientGroupConfig) (*Client, error) {
	client := &Client{
		keysAPI: keysAPI,
	}
	serversClients, err := utility.NewEtcdGrpcClientGroup(keysAPI, serversGroupConfig)
	if err != nil {
		return nil, fmt.Errorf("create etcd grpc client group for servers fail: %s", err.Error())
	}
	client.serversClients = serversClients
	peersClients, err := utility.NewEtcdGrpcClientGroup(keysAPI, peersGroupConfig)
	if err != nil {
		return nil, fmt.Errorf("create etcd grpc client group for peers fail: %s", err.Error())
	}
	client.peersClients = peersClients
	return client, nil
}

func (client *Client) Download(key string, output file.File) error {
	resp, err := client.keysAPI.Get(context.Background(), key, nil)
	if err != nil {
		return err
	}
	var fileInfo file.FileInfo
	if err := json.Unmarshal([]byte(resp.Node.Value), &fileInfo); err != nil {
		return fmt.Errorf("unmarshal file info fail: %s", err.Error())
	}

	return nil
}
