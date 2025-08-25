package _0_cache_api

import (
	"context"
	"gk-practical-go/10_cache_api/shutdown_server"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestShutdownServer(t *testing.T) {
	s1 := shutdown_server.NewServer("server01", "0.0.0.0:8090")
	s1.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello world"))
	}))
	s2 := shutdown_server.NewServer("server02", "0.0.0.0:8091")

	app := shutdown_server.NewApp(
		[]*shutdown_server.Server{s1, s2},
		shutdown_server.WithShutdownCallback(StoreCacheToDB),
	)

	println(os.Getpid())

	app.StartAndServe()
}

func StoreCacheToDB(ctx context.Context) {
	done := make(chan struct{}, 1)
	go func() {
		println("刷新缓存中...")
		time.Sleep(time.Second)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		println("缓存刷新超时")
	case <-done:
		println("缓存刷新到DB")
	}
}

func TestServer(t *testing.T) {
	s := shutdown_server.NewServer("test", "0.0.0.0:9999")
	s.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello world"))
	}))
	s.Start()
}
