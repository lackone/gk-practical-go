package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"gk-practical-go/14_service_register_discovery/registry"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync"
)

type Registry struct {
	c       *clientV3.Client
	sess    *concurrency.Session
	cancels []func()
	mutex   sync.Mutex
}

func NewRegistry(c *clientV3.Client) (*Registry, error) {
	sess, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	return &Registry{
		c:    c,
		sess: sess,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	fmt.Println(si)

	val, err := json.Marshal(si)
	if err != nil {
		return err
	}

	_, err = r.c.Put(ctx, r.InstanceKey(si), string(val), clientV3.WithLease(r.sess.Lease()))

	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.InstanceKey(si))
	return err
}

func (r *Registry) ListServices(ctx context.Context, serviceName string) ([]registry.ServiceInstance, error) {
	//获取前缀
	resp, err := r.c.Get(ctx, r.ServiceKey(serviceName), clientV3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make([]registry.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var si registry.ServiceInstance
		if err := json.Unmarshal(kv.Value, &si); err != nil {
			return nil, err
		}
		res = append(res, si)
	}
	fmt.Println("11111111")
	fmt.Println(res)
	return res, nil
}

func (r *Registry) Subscribe(serviceName string) (<-chan registry.Event, error) {
	ctx, cancel := context.WithCancel(context.Background())

	r.mutex.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mutex.Unlock()

	ctx = clientV3.WithRequireLeader(ctx)

	watchCh := r.c.Watch(context.Background(), r.ServiceKey(serviceName), clientV3.WithPrefix())

	res := make(chan registry.Event)

	go func() {
		for {
			select {
			case resp := <-watchCh:
				if resp.Err() != nil {
					//return
					continue
				}

				if resp.Canceled {
					return
				}

				for range resp.Events {
					res <- registry.Event{}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return res, nil
}

func (r *Registry) Close() error {
	r.mutex.Lock()
	cancels := r.cancels
	r.cancels = nil
	r.mutex.Unlock()

	for _, cancel := range cancels {
		cancel()
	}

	err := r.sess.Close()
	return err
}

func (r *Registry) InstanceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Addr)
}

func (r *Registry) ServiceKey(name string) string {
	return fmt.Sprintf("/micro/%s", name)
}
