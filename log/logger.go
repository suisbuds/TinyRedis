package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

/*
基于 zap 和 lumberjack 封装日志库
*/

type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

var (
	zapper Logger
	// zap 日志级别
	Level = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
		"fatal": zapcore.FatalLevel,
	}
)

// 配置选项
type LogConfig struct {
	LogName   string // 日志名称
	LogLevel  string // 日志级别
	FileName  string // 文件名称
	MaxAge    int    // 日志保留时间，以天为单位
	MaxSize   int    // 日志保留大小，以 M 为单位
	MaxBackup int    // 保留文件个数
	Compress  bool   // 是否压缩
}

// 配置方法别名
type Option func(*LogConfig)

// 封装 zap.SugaredLogger 和 config
type zapWrapper struct {
	*zap.SugaredLogger
	config LogConfig
}

func init() {
	zapper = NewLogger(NewLogConfig())
}

// 初始化配置
func NewLogConfig(opts ...Option) LogConfig {
	// 默认配置
	config := LogConfig{
		LogName:   "app",
		LogLevel:  "info",
		FileName:  "app.log",
		MaxAge:    7,
		MaxSize:   50,
		MaxBackup: 3,
		Compress:  true,
	}
	// 可变参数,可选参数模式
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

func WithLogLevel(level string) Option {
	return func(l *LogConfig) {
		l.LogLevel = level
	}
}

func WithFileName(filename string) Option {
	return func(l *LogConfig) {
		l.FileName = filename
	}
}

func NewLogger(config LogConfig) Logger {
	z := &zapWrapper{config: config}
	encoder := z.getEncoder()       // 日志编码器
	writeSyncer := z.getLogWriter() // 日志写入器
	core := zapcore.NewCore(encoder, writeSyncer, Level[config.LogLevel])
	// 创建logger,添加调用者信息和跳过层级
	z.SugaredLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	return z
}

// 获取编码配置
func (w *zapWrapper) getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	// 设置时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 大写字母显示日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// console 格式输出
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// 利用 lumberjack 构造写入器，实现日志切割、压缩、备份和日志轮转，从而避免日志文件过大
func (z *zapWrapper) getLogWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   z.config.FileName,
		MaxAge:     z.config.MaxAge,
		MaxSize:    z.config.MaxSize,
		MaxBackups: z.config.MaxBackup,
		Compress:   z.config.Compress,
	})
}

func GetLogger() Logger {
	return zapper
}
