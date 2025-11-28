package persist

import (
	"io"
	"os"
	"time"

	"github.com/suisbuds/TinyRedis/database"
	"github.com/suisbuds/TinyRedis/datastore"
	"github.com/suisbuds/TinyRedis/handler"
	"github.com/suisbuds/TinyRedis/log"
	"github.com/suisbuds/TinyRedis/pkg"
	"github.com/suisbuds/TinyRedis/protocol"
)

// 重写 aof 文件
func (a *aofPersister) rewriteAOF() error {
	// 1 重写前处理. 需要短暂加锁
	tmpFile, fileSize, err := a.startRewrite()
	if err != nil {
		return err
	}

	// 2 aof 指令重写. 与主流程并发执行
	if err = a.doRewrite(tmpFile, fileSize); err != nil {
		return err
	}

	// 3 完成重写. 需要短暂加锁
	return a.endRewrite(tmpFile, fileSize)
}

// 记录重写位置
func (a *aofPersister) startRewrite() (*os.File, int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.aofFile.Sync(); err != nil {
		return nil, 0, err
	}

	fileInfo, _ := os.Stat(a.aofFileName)
	fileSize := fileInfo.Size()

	// 创建一个临时的 aof 文件
	tmpFile, err := os.CreateTemp("./", "*.aof")
	if err != nil {
		return nil, 0, err
	}

	return tmpFile, fileSize, nil
}

// 得到数据库快照 forkDB后，遍历 forkDB 并生成 aof 指令（aof持久化并发执行）
func (a *aofPersister) doRewrite(tmpFile *os.File, fileSize int64) error {
	forkedDB, err := a.forkDB(fileSize)
	if err != nil {
		return err
	}

	// 将 db 数据转为 aof cmd
	forkedDB.ForEach(func(key string, adapter database.CmdAdapter, expireAt *time.Time) {
		_, _ = tmpFile.Write(handler.NewMultiBulkReply(adapter.ToCmd()).ToBytes())

		if expireAt == nil {
			return
		}

		expireCmd := [][]byte{[]byte(database.CmdTypeExpireAt), []byte(key), []byte(pkg.TimeSecondFormat(*expireAt))}
		_, _ = tmpFile.Write(handler.NewMultiBulkReply(expireCmd).ToBytes())
	})

	return nil
}

// 读取并执行当前 AOF 文件，新建内存数据库的快照 (只fork一次)
func (a *aofPersister) forkDB(fileSize int64) (database.DataStore, error) {
	file, err := os.Open(a.aofFileName)
	if err != nil {
		return nil, err
	}
	file.Seek(0, io.SeekStart)
	logger := log.GetLogger()
	reloader := readCloserAdapter(io.LimitReader(file, fileSize), file.Close)
	fakePerisister := newFakePersister(reloader)
	tmpKVStore := datastore.NewKVStore(fakePerisister)
	executor := database.NewDBExecutor(tmpKVStore)
	trigger := database.NewDBTrigger(executor)
	h, err := handler.NewHandler(trigger, fakePerisister, protocol.NewParser(logger), logger)
	if err != nil {
		return nil, err
	}
	if err = h.Start(); err != nil {
		return nil, err
	}
	return tmpKVStore, nil
}

// 复制增量数据到 aof 副本
func (a *aofPersister) endRewrite(tmpFile *os.File, fileSize int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// copy commands executed during rewriting to tmpFile
	/* read write commands executed during rewriting */
	src, err := os.Open(a.aofFileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
		_ = tmpFile.Close()
	}()

	if _, err = src.Seek(fileSize, 0); err != nil {
		return err
	}

	// 把老的 aof 文件中后续内容 copy 到 tmp 中
	if _, err = io.Copy(tmpFile, src); err != nil {
		return err
	}

	// 关闭老的 aof 文件，准备废弃
	_ = a.aofFile.Close()
	// 重命名 tmp 文件，作为新的 aof 文件
	if err := os.Rename(tmpFile.Name(), a.aofFileName); err != nil {
		// log
	}

	// 重新开启
	aofFile, err := os.OpenFile(a.aofFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		panic(err)
	}
	a.aofFile = aofFile
	return nil
}
