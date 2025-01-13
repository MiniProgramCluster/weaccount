package main

import (
	"flag"
	"fmt"
	"net/http"

	"weaccount/internal/conf"
	"weaccount/internal/handlers"
	"weaccount/utils/db"
	"weaccount/utils/log"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// 定义命令行标志
	logLevel := flag.String("log-level", "debug", "Set the log level (debug, info, warn, error, fatal, panic)")
	configFile := flag.String("env", ".env.json", "Set the env file path")
	listenPort := flag.Int("listen", 9001, "Set the listen port")
	flag.Parse()
	// 初始化日志
	log.Init(&lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    50, // 兆
		MaxBackups: 0,  // 不删除文件
		MaxAge:     28, // 保留28天
	}, *logLevel)
	conf.Init(*configFile)
	// 初始化数据库
	db.Initialize()
	// 初始化路由
	http.HandleFunc("/account/auth", handlers.AuthHandler)
	// -----------------调试代码 结束
	listenAddr := fmt.Sprintf("127.0.0.1:%d", *listenPort)
	fmt.Printf("Server starting on %s...", listenAddr)
	log.Logger().Info().Msgf("Server starting on %s...", listenAddr)
	log.Logger().Info().Msg("Hello, World!")
	http.ListenAndServe(listenAddr, nil)
	log.Logger().Info().Msg("Server started")
}
