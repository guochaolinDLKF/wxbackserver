package log

import (
	"fmt"
	stdlog "log"
	"os"
	"runtime/debug"
	"time"
)

type loghelper struct {
	WHour   int            // 当前正在写的小时数
	LogDir  string         // 日志文件目录
	LogFile *os.File       // 日志文件
	Logger  *stdlog.Logger // logger
}

var useStdout bool
var lh [4]*loghelper

func (l *loghelper) chkFile() *stdlog.Logger {
	curH := time.Now().Hour()
	if l.WHour != curH {
		// 时间发生改变，创建新文件
		fileName := l.LogDir + "/" + time.Now().Format("20060102_15") + ".log"
		logFile, err := os.Create(fileName)
		if err != nil {
			if lh[3] != nil && lh[3].Logger != nil {
				lh[3].Logger.Printf("LOG FILE CREATE FAILD! ERROR:%s\n", err.Error())
			} else {
				stdlog.Panic(err)
				return nil
			}
		} else {
			if l.LogFile != nil {
				l.LogFile.Close()
			}
			l.LogFile = logFile
			if l.Logger == nil {
				l.Logger = stdlog.New(l.LogFile, "", stdlog.LstdFlags)
			} else {
				l.Logger.SetOutput(l.LogFile)
			}
		}
		l.WHour = curH
	}
	return l.Logger
}

func Init(root string, stdout bool) {
	fmt.Println("root:",root)
	for i := 0; i < 4; i++ {
		lh[i] = &loghelper{WHour: -1}
	}

	useStdout = stdout

	lh[0].LogDir = root + "/dbg"
	lh[1].LogDir = root + "/cmn"
	lh[2].LogDir = root + "/err"
	lh[3].LogDir = root

	for i := 0; i < 4; i++ {
		lh[i].chkFile().Println("start!")
	}
}

func Dbg(msg interface{}) {
	if useStdout {
		fmt.Printf("%+v\n", msg)
	} else if logger := lh[0].chkFile(); logger != nil {
		logger.Printf("%+v\n", msg)
	}
}

func Cmn(msg interface{}) {
	if useStdout {
		fmt.Printf("%+v\n", msg)
	} else if logger := lh[1].chkFile(); logger != nil {
		logger.Printf("%+v\n", msg)
	}
}

func Err(msg interface{}) {
	if useStdout {
		fmt.Printf("%+v\n%s\n", msg, debug.Stack())
	} else if logger := lh[2].chkFile(); logger != nil {
		logger.Printf("%+v\n%s\n", msg, debug.Stack())
	}
}

func Fatal(msg interface{}) {
	if useStdout {
		fmt.Printf("%+v\n%s\n", msg, debug.Stack())
	} else if logger := lh[3].chkFile(); logger != nil {
		logger.Printf("%+v\n%s\n", msg, debug.Stack())
	}
}
