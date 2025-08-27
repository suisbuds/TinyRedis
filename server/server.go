package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/suisbuds/TinyRedis/log"
	"github.com/suisbuds/TinyRedis/pkg"
)

/*
TinyRedis 服务端，响应客户端请求
*/

// 指令分发handler：处理tcp连接
type Handler interface {
	Start() error                              // 启动 handler
	Handle(ctx context.Context, conn net.Conn) // 处理tcp连接
	Close()                                    // 关闭 handler
}

type Server struct {
	run      sync.Once     // server 只启动一次
	stop     sync.Once     // server 只关闭一次
	handler  Handler       // 处理器
	logger   log.Logger    // 自定义日志
	stopChan chan struct{} // 停止信号
}

func NewServer(handler Handler, logger log.Logger) *Server {
	return &Server{
		handler:  handler,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// server运行时，启动tcp服务并监听端口，将tcp连接交给handler处理

func (s *Server) Start(address string) error {

	// 1. 启动 handler
	if err := s.handler.Start(); err != nil {
		return err
	}
	var _err error
	s.run.Do(func() {
		// 2. 监听退出信号
		exitSignal := []os.Signal{syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT}
		signalChan := make(chan os.Signal, 1)
		closeChan := make(chan struct{}, 4)

		signal.Notify(signalChan, exitSignal...)
		// 启动一个 goroutine 监听信号，然后传递给 closeChan 通知关闭
		pkg.Submit(func() {
			for {
				select {
				case signal := <-signalChan:
					switch signal {
					case exitSignal[0], exitSignal[1], exitSignal[2], exitSignal[3]:
						closeChan <- struct{}{}
						return
					default:
					}
				case <-s.stopChan:
					closeChan <- struct{}{}
					return
				}
			}
		})
		// 3. 监听TCP端口
		listener, err := net.Listen("tcp", address)
		if err != nil {
			_err = err
			return
		}

		// 运行 tcp 服务
		s.listenAndServe(listener, closeChan)
	})

	return _err
}

func (s *Server) Stop() {
	s.stop.Do(func() {
		close(s.stopChan)
	})
}

func (s *Server) listenAndServe(listener net.Listener, closeChan chan struct{}) {
	errChan := make(chan error, 1)
	defer close(errChan)

	// 1. ctx控制所有连接的生命周期
	ctx, cancel := context.WithCancel(context.Background())
	// 2. 协程池启动goroutine监听关闭或错误信号
	pkg.Submit(
		func() {
			select {
			case <-closeChan:
				s.logger.Infof("[server] server closing")
			case err := <-errChan:
				s.logger.Errorf("[server] server error: %s", err.Error())
			}
			// 取消ctx，通知所有handler关闭
			cancel()
			s.handler.Close()
			s.logger.Warnf("[server] server closed")
			if err := listener.Close(); err != nil {
				s.logger.Errorf("[server] server listener close error: %s", err.Error())
			}
		})

	s.logger.Warnf("[server] server starting")
	var wg sync.WaitGroup

	// 3. io 多路复用, 不断从listener中接收连接请求
	for {
		conn, err := listener.Accept()
		if err != nil {
			// 超时类错误，忽略
			if err, ok := err.(net.Error); ok && err.Timeout() {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			//  其他错误，通知errChan
			errChan <- err
			break
		}

		// 为每个 TCP 连接分配 goroutine 处理
		wg.Add(1)
		pkg.Submit(func() {
			defer wg.Done()
			s.handler.Handle(ctx, conn)
		})
	}

	// 等待所有goroutine处理完成
	wg.Wait()
}
