package app

import "github.com/suisbuds/TinyRedis/server"

/*
TinyRedis 客户端
*/

type Application struct {
	server *server.Server
	config *Config
}

func NewApplication(server *server.Server, config *Config) *Application {
	return &Application{
		server: server,
		config: config,
	}
}

func (a *Application) Run() error {
	return a.server.Start(a.config.Address())
}

func (a *Application) Stop() {
	a.server.Stop()
}
