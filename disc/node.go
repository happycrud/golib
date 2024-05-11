package disc

import (
	"context"
	"fmt"

	"github.com/rs/xid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

type Node struct {
	ServiceID       string
	NodeID          string
	Endpoint        string
	EndpointManager endpoints.Manager
	Ec              *clientv3.Client
}

func GetServiceInstance(serviceID string, myaddr string) *Node {
	i := &Node{
		ServiceID:       serviceID,
		NodeID:          xid.New().String(),
		Endpoint:        myaddr,
		EndpointManager: nil,
		Ec:              GetEc(),
	}
	i.EndpointManager, _ = endpoints.NewManager(i.Ec, i.ServiceID)
	return i
}
func (i *Node) DiscID() string {
	return fmt.Sprintf("%s/node/%s", i.ServiceID, i.NodeID)
}
func (i *Node) Register() error {
	ctx := context.Background()
	lease := clientv3.NewLease(i.Ec)
	tick, err := lease.Grant(ctx, 30)
	if err != nil {
		return err
	}
	ch, err := lease.KeepAlive(ctx, tick.ID)
	if err != nil {
		return err
	}
	go func() {
		for range ch {
			// fmt.Println("lease ", v)
		}
	}()

	return i.EndpointManager.AddEndpoint(ctx, i.DiscID(), endpoints.Endpoint{Addr: i.Endpoint}, clientv3.WithLease(tick.ID))

}
func (i *Node) Deregister() error {
	return i.EndpointManager.DeleteEndpoint(context.Background(), i.DiscID())
}
