package logger

import (
	"io"
	"log"
	"os"
)

var (
	Debug *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func init() {
	if _, err := os.Stat("logs"); err != nil {
		if err = os.Mkdir("logs", 755); err != nil {
			log.Fatalln("Failed to create logs dir: ", err)
		}
	}

	// 日志输出文件
	file, err := os.OpenFile("logs/master.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error logger file:", err)
	}
	//自定义日志格式
	Debug = log.New(io.MultiWriter(file, os.Stderr), "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(io.MultiWriter(file, os.Stderr), "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(file, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
