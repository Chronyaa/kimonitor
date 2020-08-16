package util

import (
	"fmt"
	"log"
	"os"
)

type logger struct {
	*log.Logger
}

// Debug 用于Debug Level的Logger
var Debug *logger

// Warning 用于Warning Level的Logger
var Warning *logger

// Error 用于Error Level的Logger
var Error *logger

// 文件名，需要关闭该等级LOG的时候设置为os.DevNull即可
var debugFile = "kimonitor.debug.log"
var warningFile = "kimonitor.warning.log"
var errorFile = "kimonitor.error.log"

func init() {
	d, err := os.OpenFile(debugFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Println("Open", debugFile, " failed.")
		return
	}
	Debug = &logger{log.New(d, "[DEBUG]", log.LstdFlags)}
	w, err := os.OpenFile(warningFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Println("Open", warningFile, " failed.")
		return
	}
	Warning = &logger{log.New(w, "[WARNING]", log.LstdFlags|log.Llongfile)}
	e, err := os.OpenFile(errorFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Println("Open", errorFile, "failed.")
		return
	}
	Error = &logger{log.New(e, "[ERROR]", log.LstdFlags|log.Llongfile)}

}

// 为了简化error处理的一堆实验功能，函数开始时defer Recover()，然后就可以通过panic-recover来达成返回的效果
// 配合命名返回参数让原本四行的error处理变成一行

// Recover 配合Return使用
func Recover() {
	recover()
}

// Returnln err != nil时panic，然后让Recover接住，造成返回的效果
func (l *logger) Returnln(err error, msg ...interface{}) {
	if err != nil {
		l.Output(2, fmt.Sprintln(msg...))
		panic(nil)
	}
}

// Returnf err != nil时panic，然后让Recover接住，造成返回的效果
func (l *logger) Returnf(err error, format string, msg ...interface{}) {
	if err != nil {
		l.Output(2, fmt.Sprintf(format, msg...))
		panic(nil)
	}
}

// Assertln !p时panic，然后让Recover接住，造成返回的效果
func (l *logger) Assertln(p bool, msg ...interface{}) {
	if !p {
		l.Output(2, fmt.Sprintln(msg...))
		panic(nil)
	}
}

// Assertf !p时panic，然后让Recover接住，造成返回的效果
func (l *logger) Assertf(p bool, format string, msg ...interface{}) {
	if !p {
		l.Output(2, fmt.Sprintf(format, msg...))
		panic(nil)
	}
}

// Checkln err != nil时打印msg
func (l *logger) Checkln(err error, msg ...interface{}) {
	if err != nil {
		l.Output(2, fmt.Sprintln(msg...))
	}
}

// Checkf err != nil时打印msg
func (l *logger) Checkf(err error, format string, msg ...interface{}) {
	if err != nil {
		l.Output(2, fmt.Sprintf(format, msg...))
	}
}

// PrintlnIfNot !p时打印msg
func (l *logger) PrintlnIfNot(p bool, msg ...interface{}) {
	if !p {
		l.Output(2, fmt.Sprintln(msg...))
	}
}

// PrintfIfNot err == true打印msg
func (l *logger) PrintfIfNot(p bool, format string, msg ...interface{}) {
	if !p {
		l.Output(2, fmt.Sprintf(format, msg...))
	}
}
