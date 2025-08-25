package __context

import (
	"context"
	"testing"
	"time"
)

type mykey struct {
}

func TestContext(t *testing.T) {
	// 一般是链路起点，或者调用的起点
	ctx1 := context.Background()

	//在不确定 context 该用啥的时候，用 TODO()
	//ctx2 := context.TODO()

	v := context.WithValue(ctx1, mykey{}, "my-key")
	println(v)

	ctx3, cancelFunc := context.WithCancel(ctx1)
	//用完ctx再去调用
	cancelFunc()

	println(ctx3)
}

func TestContextWithCancel(t *testing.T) {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	//用ctx
	<-ctx.Done()
	t.Log("cancel", ctx.Err())
}

func TestContextWithDeadline(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(3*time.Second))

	deadline, _ := ctx.Deadline()
	t.Log("deadline: ", deadline)

	defer cancelFunc()

	<-ctx.Done()

	t.Log("deadline: ", ctx.Err())
}

func TestContextWithTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, 3*time.Second)

	deadline, _ := ctx.Deadline()
	t.Log("deadline: ", deadline)

	defer cancelFunc()

	<-ctx.Done()

	t.Log("deadline: ", ctx.Err())
}

func TestContextParent(t *testing.T) {
	ctx := context.Background()
	parent := context.WithValue(ctx, "mykey", "myval")
	child := context.WithValue(parent, "mykey", "myval2")

	t.Log("parent my key", parent.Value("mykey"))
	t.Log("child my key", child.Value("mykey"))

	child2, cancel := context.WithCancel(parent)
	defer cancel()
	t.Log("child2 my key", child2.Value("mykey"))

	child3 := context.WithValue(parent, "newkey", "newval")
	t.Log("parent newkey", parent.Value("newkey"))
	t.Log("child3 newkey", child3.Value("newkey"))

}
