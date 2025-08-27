package database

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/suisbuds/TinyRedis/handler"
)

type DBTrigger struct {
	once     sync.Once
	executor Executor
}

func NewDBTrigger(executor Executor) handler.DB {
	return &DBTrigger{executor: executor}
}

func (d *DBTrigger) Do(ctx context.Context, cmdLine [][]byte) handler.Reply {
	if len(cmdLine) < 2 {
		return handler.NewErrReply(fmt.Sprintf("invalid cmd line: %v", cmdLine))
	}

	// 将输入命令转换为小写以支持大写命令
	cmdType := CmdType(strings.ToLower(string(cmdLine[0])))
	if !d.executor.ValidCommand(cmdType) {
		return handler.NewErrReply(fmt.Sprintf("unknown cmd '%s'", cmdLine[0]))
	}

	cmd := Command{
		ctx:      ctx,
		cmd:      cmdType,
		args:     cmdLine[1:],
		receiver: make(CmdReceiver),
	}

	// 投递给到 executor
	d.executor.Entrance() <- &cmd

	// 监听 chan，直到接收到返回的 reply
	return <-cmd.Receiver()
}

func (d *DBTrigger) Close() {
	d.once.Do(d.executor.Close)
}
