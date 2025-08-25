package shutdown_server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type serverMux struct {
	reject bool
	*http.ServeMux
}

func (s *serverMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.reject {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("service unavailable"))
		return
	}
	s.ServeMux.ServeHTTP(w, r)
}

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

func NewServer(name string, addr string) *Server {
	mux := &serverMux{
		ServeMux: http.NewServeMux(),
	}
	return &Server{
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		name: name,
		mux:  mux,
	}
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) rejectReq() {
	s.mux.reject = true
}

func (s *Server) stop() error {
	return s.srv.Shutdown(context.Background())
}

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallback(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = cbs
	}
}

type App struct {
	servers []*Server

	//退出整个超时时间
	shutdownTimeout time.Duration

	//等待处理已有请求时间
	waitTime time.Duration

	//回调超时时间
	cbTimeout time.Duration

	cbs []ShutdownCallback
}

func NewApp(servers []*Server, opts ...Option) *App {
	res := &App{
		waitTime:        10 * time.Second,
		cbTimeout:       3 * time.Second,
		shutdownTimeout: 30 * time.Second,
		servers:         servers,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func (app *App) StartAndServe() {
	for _, s := range app.servers {
		srv := s
		go func() {
			if err := srv.Start(); err != nil {
				log.Printf("server start error: %s", err)
			}
		}()
	}

	//监听信号
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, signals...)

	<-ch //第一次退出
	println("Shutting down...")

	go func() {
		select {
		case <-ch: //第二次，则强制退出
			println("强制退出")
			os.Exit(1)
		case <-time.After(app.shutdownTimeout):
			println("超时强制退出")
			os.Exit(1)
		}
	}()

	app.shutdown()
}

func (app *App) shutdown() {
	println("关闭应用，停止接收新请求")
	for _, s := range app.servers {
		s.rejectReq()
	}

	println("等待执行请求完成")

	//这里可以改造实时统计正在处理的请求数量，如果为0，则下一步
	time.Sleep(app.waitTime)

	wg := sync.WaitGroup{}
	wg.Add(len(app.servers))
	for _, s := range app.servers {
		srv := s
		go func() {
			if err := srv.stop(); err != nil {
				log.Printf("server stop error: %s", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	println("执行回调")

	wg.Add(len(app.cbs))
	for _, cb := range app.cbs {
		c := cb
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), app.cbTimeout)
			c(ctx)
			cancel()
			wg.Done()
		}()
	}
	wg.Wait()

	println("开始释放资源")

	app.close()
}

func (app *App) close() {
	//释放一些可能的资源
	time.Sleep(time.Second)
	println("app close")
}
