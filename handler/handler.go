package handler

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/suisbuds/TinyRedis/log"
	"github.com/suisbuds/TinyRedis/server"
)

/*
实现指令分发器handler
*/

type Handler struct {
	sync.Once // Close 只执行一次
	mu        sync.RWMutex
	conns     map[net.Conn]struct{} // TCP连接
	close     atomic.Bool           // Handler是否关闭

	db        DB        // 存储引擎
	parser    Parser    // 协议解析器
	persister Persister // 持久化模块
	logger    log.Logger
}

func NewHandler(db DB, persister Persister, parser Parser, logger log.Logger) (server.Handler, error) {
	h := Handler{
		conns:     make(map[net.Conn]struct{}),
		persister: persister,
		logger:    logger,
		db:        db,
		parser:    parser,
	}

	return &h, nil
}

// 加载持久化数据
func (h *Handler) Start() error {
	reloader, err := h.persister.Reloader()
	if err != nil {
		return err
	}
	defer reloader.Close()
	h.handleStream(SetLoadingPattern(context.Background()), newFakeReaderWriter(reloader))
	return nil
}

// 处理TCP连接
func (h *Handler) Handle(ctx context.Context, conn net.Conn) {
	h.mu.Lock()
	// handler 已关闭
	if h.close.Load() {
		h.mu.Unlock()
		return
	}

	// 缓存当前连接
	h.conns[conn] = struct{}{}
	h.mu.Unlock()

	h.handleStream(ctx, conn)
}

func (h *Handler) handleStream(ctx context.Context, conn io.ReadWriter) {
	// parser 解析数据流
	stream := h.parser.ParseStream(conn)
	for {
		select {
		case <-ctx.Done():
			// ctx 取消
			h.logger.Warnf("[handler] context error: %s", ctx.Err().Error())
			return

		// 从stream中读取数据包
		case droplet := <-stream:
			if err := h.handleDroplet(ctx, conn, droplet); err != nil {
				h.logger.Errorf("[handler] connection terminated: %s", droplet.Err.Error())
				return
			}
		}
	}
}

// 处理数据包
func (h *Handler) handleDroplet(ctx context.Context, conn io.ReadWriter, droplet *Droplet) error {
	if droplet.Terminated() {
		return droplet.Err
	}

	// 记录数据包错误
	if droplet.Err != nil {
		_, _ = conn.Write(droplet.Reply.ToBytes())
		h.logger.Errorf("[handler] connection error: %s", droplet.Err.Error())
		return nil
	}

	if droplet.Reply == nil {
		h.logger.Errorf("[handler] empty request")
		return nil
	}

	// 检查请求参数
	multiReply, ok := droplet.Reply.(MultiReply)
	if !ok {
		h.logger.Errorf("[handler] invalid request: %s", droplet.Reply.ToBytes())
		return nil
	}

	// 执行器执行数据库指令
	if reply := h.db.Do(ctx, multiReply.Args()); reply != nil {
		conn.Write(reply.ToBytes())
		return nil
	}

	// 未知错误
	conn.Write(UnknownErrReplyBytes)
	return nil
}

// 关闭所有TCP连接，释放资源
func (h *Handler) Close() {
	h.Once.Do(func() {
		h.logger.Warnf("[handler] handler close")
		h.close.Store(true)
		h.mu.RLock()
		defer h.mu.RUnlock()
		for conn := range h.conns {
			if err := conn.Close(); err != nil {
				h.logger.Errorf("[handler] close connection error: %s, local address: %s", err.Error(), conn.LocalAddr().String())
			}
		}
		h.conns = nil
		h.db.Close()
		h.persister.Close()
	})
}
