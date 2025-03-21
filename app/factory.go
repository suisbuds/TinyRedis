package app

import (
	"github.com/suisbuds/TinyRedis/database"
	"github.com/suisbuds/TinyRedis/datastore"
	"github.com/suisbuds/TinyRedis/handler"
	"github.com/suisbuds/TinyRedis/log"
	"github.com/suisbuds/TinyRedis/persist"
	"github.com/suisbuds/TinyRedis/protocol"
	"github.com/suisbuds/TinyRedis/server"
	"go.uber.org/dig"
)

/*
1. 基于dig实现依赖注入，自动构建对象
2. 工厂模式全局管理应用各个模块
*/

var container = dig.New()

func init() {

	// 加载配置
	_ = container.Provide(SetUpConfig)
	_ = container.Provide(PersistThinker)
	// 日志
	_ = container.Provide(log.GetLogger)

	// 1. 存储引擎层

	// 数据持久化
	_ = container.Provide(persist.NewPersister)
	// 存储介质
	_ = container.Provide(datastore.NewKVStore)
	// DB执行器
	_ = container.Provide(database.NewDBExecutor)
	// DB触发器
	_ = container.Provide(database.NewDBTrigger)

	// 2. 指令层

	// 协议解析器
	_ = container.Provide(protocol.NewParser)
	// 指令分发器
	_ = container.Provide(handler.NewHandler)

	// 3. 服务层
	_ = container.Provide(server.NewServer)
}

func InitServer() (*server.Server, error) {
	
	// 从container中获取已注册的handler和logger

	var handler server.Handler
	if err := container.Invoke(func(_handler server.Handler) {
		handler = _handler
	}); err != nil {
		return nil, err
	}

	var logger log.Logger
	if err := container.Invoke(func(_logger log.Logger) {
		logger = _logger
	}); err != nil {
		return nil, err
	}

	// 创建server
	return server.NewServer(handler, logger), nil
}
