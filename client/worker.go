package client

import "github.com/yangchenxing/go-dds/utility"

func worker(ch <-chan *task, servers *utility.EtcdGrpcClientGroup, peers *utility.EtcdGrpcClientGroup) {
	for {
		task := <-ch
        if task.fromServer {
            
        }
	}
}
