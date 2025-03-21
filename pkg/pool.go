package pkg

import (
	"runtime/debug"
	"strings"

	"github.com/panjf2000/ants"
	"github.com/suisbuds/TinyRedis/log"
)

/*
Ants 协程池复用 goroutine, 降低频繁创建销毁的开销，适合处理大量短生命周期任务
*/

// 包级别变量在包初始化时构建
var (
	// 全局协程池
	pool     *ants.Pool
	capacity = 50000
)

func init() {
	// 创建一个容量为 50000 的协程池
	_pool, err := ants.NewPool(capacity, ants.WithPanicHandler(func(i interface{}) {
		// 捕捉异常信息和调用栈，通过日志输出
		stackInfo := strings.Replace(string(debug.Stack()), "\n", "", -1)
		log.GetLogger().Errorf("recover info: %v, stack info: %s", i, stackInfo)
	}))
	if err != nil {
		panic(err)
	}
	pool = _pool
}

// 提交任务到协程池
func Submit(task func()) {
	pool.Submit(task)
}
