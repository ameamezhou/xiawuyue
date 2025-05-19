package xlog

import (
	"fmt"
	"github.com/ameamezhou/xiawuyue/xconfig"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	errorLog *log.Logger
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog *log.Logger
	loggers  = []*log.Logger{errorLog, debugLog}
	mu       sync.Mutex
)

// log levels
const (
	DebugLevel = iota
	InfoLevel
	ErrorLevel
	Disabled
)

// colorCode
const (
	colorRed    = "31"
	colorGreen  = "32"
	colorYellow = "33"
	colorBlue   = "34"
)

func InitLoger(logFilePath string) {
	if logFilePath == "" {
		errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
		debugLog = log.New(os.Stdout, "\033[32m[debug]\033[0m ", log.LstdFlags|log.Lshortfile)
		infoLog = log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile)
		warnLog = log.New(os.Stdout, "\033[33m[warn]\033[0m ", log.LstdFlags|log.Lshortfile)
	} else {
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		errorLog = log.New(logFile, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
		debugLog = log.New(logFile, "\033[32m[debug]\033[0m ", log.LstdFlags|log.Lshortfile)
		infoLog = log.New(logFile, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile)
		warnLog = log.New(logFile, "\033[33m[warn]\033[0m ", log.LstdFlags|log.Lshortfile)
	}
}

func InitLogerProject(conf *xconfig.WeConfig) {
	logPath := conf.GetValue("project", "logpath", "")
	projname := conf.GetValue("project", "project", "")
	logFile := fmt.Sprintf("%s%s_%s", logPath, projname, time.Now().Format("2006-01-02"))
	InitLoger(logFile)
	go func() {
		for {
			now := time.Now()
			nextMidnight := now.Add(24 * time.Hour)
			midnightTime := nextMidnight.Truncate(24 * time.Hour)

			duration := midnightTime.Sub(now)
			if duration < 0 {
				duration += 24 * time.Hour
			}

			time.Sleep(duration)

			currentDate := time.Now().Format("2006-01-02")
			newLogFilePath := fmt.Sprintf("%s%s_%s", logPath, projname, currentDate)

			outputFile, err := os.OpenFile(newLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("Failed to open new log file: %v", err)
			}

			errorLog.SetOutput(outputFile)
			debugLog.SetOutput(outputFile)
			infoLog.SetOutput(outputFile)

			Infof("日志文件已切换到: %s\n", newLogFilePath)
		}
	}()
}

// 文件写入后续优化可以改为 mmap 写入  不用 write file  效率会更高

func Error(v ...any) {
	logWithPosition(errorLog, fmt.Sprint(v...), colorRed)
}

func Errorf(format string, v ...any) {
	logWithPosition(errorLog, fmt.Sprintf(format, v...), colorRed)
}

func Debug(v ...any) {
	logWithPosition(debugLog, fmt.Sprint(v...), colorGreen)
}

func Debugf(format string, v ...any) {
	logWithPosition(debugLog, fmt.Sprintf(format, v...), colorGreen)
}

func Info(v ...any) {
	logWithPosition(infoLog, fmt.Sprint(v...), colorBlue)
}

func Infof(format string, v ...any) {
	logWithPosition(infoLog, fmt.Sprintf(format, v...), colorBlue)
}

func Warn(v ...any) {
	logWithPosition(infoLog, fmt.Sprint(v...), colorYellow)
}

func Warnf(format string, v ...any) {
	logWithPosition(infoLog, fmt.Sprintf(format, v...), colorYellow)
}

func TestColor(v ...any) {
	logWithPosition(debugLog, fmt.Sprint(v...), "33")
}

// logWithPosition 记录一条带有文件名和行号的日志信息
func logWithPosition(l *log.Logger, msg, colorCode string) {
	// 获取调用者的文件名和行号
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		// 获取函数名
		// fn := runtime.FuncForPC(pc).Name()
		// 提取文件名和行号
		file = file[strings.LastIndex(file, "/")+1:]
	}

	// 格式化日志信息
	message := fmt.Sprintf("\033[%sm [%s:%d] %s \033[0m", colorCode, file, line, msg)
	fmt.Println(message)
	if l != nil {
		l.Println(message)
	}
}

/*
灵感来源
package mylogger

import (
	"log"
	"os"
	"runtime"
)

type MyLogger struct {
	*log.Logger
}

// New 创建一个新的 MyLogger 实例
func New(out *os.File, prefix string, flag int) *MyLogger {
	logger := log.New(out, prefix, flag)
	return &MyLogger{logger}
}

// Trace 记录一条 trace 级别的日志信息
func (l *MyLogger) Trace(v ...interface{}) {
	l.logWithPosition("TRACE", v...)
}

// Debug 记录一条 debug 级别的日志信息
func (l *MyLogger) Debug(v ...interface{}) {
	l.logWithPosition("DEBUG", v...)
}

// Info 记录一条 info 级别的日志信息
func (l *MyLogger) Info(v ...interface{}) {
	l.logWithPosition("INFO", v...)
}

// Warn 记录一条 warn 级别的日志信息
func (l *MyLogger) Warn(v ...interface{}) {
	l.logWithPosition("WARN", v...)
}

// Error 记录一条 error 级别的日志信息
func (l *MyLogger) Error(v ...interface{}) {
	l.logWithPosition("ERROR", v...)
}

// logWithPosition 记录一条带有文件名和行号的日志信息
func (l *MyLogger) logWithPosition(level string, v ...interface{}) {
	// 获取调用者的文件名和行号
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		// 获取函数名
		fn := runtime.FuncForPC(pc).Name()
		// 提取文件名和行号
		file = fn[strings.LastIndex(fn, "/")+1:]
	}

	// 格式化日志信息
	message := fmt.Sprintf("[%s] [%s:%d] ", level, file, line)
	for _, arg := range v {
		message += fmt.Sprintf("%v ", arg)
	}
	message += "\n"

	// 输出日志信息
	l.Logger.Println(message)
}

func logWithPosition(msg string) {
	// 获取当前 goroutine 的栈跟踪信息
	stack := make([]byte, 4096)
	length := runtime.Stack(stack, false)
	stackTrace := string(stack[:length])

	// 查找栈跟踪信息中的文件名和行号
	var fileName string
	var lineNumber int
	_, _ = fmt.Sscanf(stackTrace, "main.go:%d", &lineNumber) // 假设栈顶是 main.go 文件
	for _, line := range strings.Split(stackTrace, "\n") {
		if strings.Contains(line, "logWithPosition") {
			fileName = strings.TrimSpace(strings.Split(line, "(")[0])
			break
		}
	}

	// 打印日志信息和位置
	log.Printf("%s %s:%d %s", msg, fileName, lineNumber, debug.Stack())
}
*/
