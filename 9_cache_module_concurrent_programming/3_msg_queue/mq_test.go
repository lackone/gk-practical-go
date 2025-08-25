package __msg_queue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMqSend(t *testing.T) {
	b := &Broker{}

	//模拟发送者
	go func() {
		for {
			err := b.Send(Msg{Data: time.Now().String()})
			if err != nil {
				t.Log(err)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(3)

	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("消费者 %d", i)
		go func() {
			defer wg.Done()
			msgs, err := b.Subscribe(100)
			if err != nil {
				t.Log(err)
				return
			}
			for msg := range msgs {
				t.Log(name, msg)
			}
		}()
	}

	wg.Wait()
}
