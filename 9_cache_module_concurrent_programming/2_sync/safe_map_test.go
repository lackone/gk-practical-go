package __sync

import (
	"fmt"
	"testing"
	"time"
)

func TestSafeMap(t *testing.T) {
	sm := SafeMap[string, string]{
		data: map[string]string{},
	}

	go func() {
		val, loaded := sm.LoadOrStore("key", "value")
		fmt.Println("g1:", val, loaded)
	}()

	go func() {
		val, loaded := sm.LoadOrStore("key", "value")
		fmt.Println("g2:", val, loaded)
	}()

	time.Sleep(time.Second)
}
