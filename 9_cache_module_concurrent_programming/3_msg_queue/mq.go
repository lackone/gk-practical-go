package __msg_queue

import (
	"errors"
	"sync"
)

type Broker struct {
	mutex sync.RWMutex
	chans []chan Msg
}

func (b *Broker) Send(msg Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, ch := range b.chans {
		select {
		case ch <- msg:
		default:
			//如果满了走这里
			return errors.New("broker channel full")
		}
	}
	return nil
}

func (b *Broker) Subscribe(cap int) (<-chan Msg, error) {
	res := make(chan Msg, cap)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.chans = append(b.chans, res)
	return res, nil
}

func (b *Broker) Close() error {
	b.mutex.Lock()
	chans := b.chans
	b.chans = nil
	b.mutex.Unlock()

	for _, ch := range chans {
		close(ch)
	}

	return nil
}

type Msg struct {
	Data string
}

type BrokerV2 struct {
	mutex     sync.RWMutex
	consumers []func(Msg)
}

func (b *BrokerV2) Send(msg Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, ch := range b.consumers {
		go ch(msg)
	}
	return nil
}

func (b *BrokerV2) Subscribe(cb func(Msg)) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.consumers = append(b.consumers, cb)
	return nil
}
