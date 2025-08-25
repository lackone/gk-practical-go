package _2_network_min_rpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type ConnPool struct {
	idleConns chan *idleConn //空闲连接队列
	reqQueue  []connReq      //请求队列

	maxConnCnt int //最大连接数

	curConnCnt int //当前连接数

	maxIdleCnt int //最大空闲数量

	maxIdleTime time.Duration //最大空闲时间

	initConnCnt int //初始连接数量

	factory func() (net.Conn, error)

	mutex sync.Mutex
}

type idleConn struct {
	conn           net.Conn
	lastActiveTime time.Time
}

type connReq struct {
	connCh chan net.Conn
}

func NewConnPool(initCnt int,
	maxIdleCnt int,
	maxCnt int,
	maxIdleTime time.Duration,
	factory func() (net.Conn, error),
) (*ConnPool, error) {

	if initCnt > maxIdleCnt {
		return nil, fmt.Errorf("初始连接数量不能大于最大空闲数量")
	}

	//空闲连接队列
	idleConns := make(chan *idleConn, maxIdleCnt)

	for i := 0; i < initCnt; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		idleConns <- &idleConn{conn: conn, lastActiveTime: time.Now()}
	}

	return &ConnPool{
		idleConns:   idleConns,
		maxConnCnt:  maxCnt,
		initConnCnt: initCnt,
		maxIdleCnt:  maxIdleCnt,
		curConnCnt:  0,
		maxIdleTime: maxIdleTime,
		factory:     factory,
		mutex:       sync.Mutex{},
	}, nil
}

func (c *ConnPool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done(): //超时
		return nil, ctx.Err()
	default:

	}

	for {
		select {
		case iconn := <-c.idleConns:
			//拿到了空间连接
			if iconn.lastActiveTime.Add(c.maxIdleTime).Before(time.Now()) {
				//如果最后活跃时间加上最大空闲时间，不小于当前时间
				iconn.conn.Close() //关闭连接
				continue
			}
			return iconn.conn, nil
		default:
			//没有空闲连接
			c.mutex.Lock()

			if c.curConnCnt >= c.maxConnCnt {
				//如果连接数超过上限了
				req := connReq{connCh: make(chan net.Conn, 1)}
				//加入请求队列
				c.reqQueue = append(c.reqQueue, req)
				//解锁
				c.mutex.Unlock()

				select {
				case <-ctx.Done():
					//如果超时了
					//1、从队列里面删掉req自已
					//2、在这里转发
					go func() {
						cc := <-req.connCh
						c.Put(context.Background(), cc)
					}()

					return nil, ctx.Err()
				case cc := <-req.connCh: //等别人归还
					return cc, nil
				}
			}

			//没超过上限
			cc, err := c.factory()
			if err != nil {
				c.mutex.Unlock()
				return nil, err
			}

			c.curConnCnt++
			c.mutex.Unlock()

			return cc, nil
		}
	}
}

func (c *ConnPool) Put(ctx context.Context, conn net.Conn) error {
	c.mutex.Lock()

	if len(c.reqQueue) > 0 {
		//有请求
		req := c.reqQueue[0]
		c.reqQueue = c.reqQueue[1:]
		c.mutex.Unlock()
		req.connCh <- conn
		return nil
	}

	c.mutex.Unlock()

	//没有阻塞请求
	iconn := &idleConn{conn: conn, lastActiveTime: time.Now()}

	select {
	case c.idleConns <- iconn:
	default:
		//空闲队列满了
		conn.Close()

		c.mutex.Lock()
		c.curConnCnt--
		c.mutex.Unlock()
	}

	return nil
}
