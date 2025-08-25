package __sync

import "sync"

type MyData struct {
	once sync.Once
}

func (d *MyData) Init() {
	d.once.Do(func() {
		//初始化
	})
}

type Single struct {
}

var s *Single
var sOnce sync.Once

func GetSingle() *Single {
	sOnce.Do(func() {
		s = &Single{}
	})
	return s
}
