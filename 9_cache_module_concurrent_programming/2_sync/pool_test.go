package __sync

import (
	"sync"
	"testing"
)

func TestPool(t *testing.T) {
	p := sync.Pool{
		New: func() interface{} {
			t.Log("生成资源")
			return "hello"
		},
	}

	//只会调用一次，生成资源
	s := p.Get()
	t.Log(s.(string))
	p.Put(s)
	s = p.Get()
	t.Log(s.(string))
}
