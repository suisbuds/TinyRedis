package main

import "github.com/suisbuds/TinyRedis/app"

/* 
TinyRedis 入口
*/

func main() {

	// 1. 创建 server 实例
	server, err := app.InitServer()
	if err != nil {
		panic(err)
	}

	// 2. 创建 app 实例
	app := app.NewApplication(server, app.SetUpConfig())
	defer app.Stop()

	// 3. 运行 app
	if err := app.Run(); err != nil {
		panic(err)
	}
}
