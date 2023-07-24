package log

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

type contextHook struct {
	Field  string
	Skip   int
	levels []log.Level
}

// NewContextHook use to make an hook
// 根据上面的推断, 我们递归深度可以设置到5即可.
func NewContextHook(levels ...log.Level) log.Hook {
	hook := contextHook{
		Field:  "line",
		Skip:   5,
		levels: levels,
	}
	if len(hook.levels) == 0 {
		hook.levels = log.AllLevels
	}
	return &hook
}

// Levels implement levels
func (hook contextHook) Levels() []log.Level {
	return log.AllLevels
}

// Fire implement fire
func (hook contextHook) Fire(entry *log.Entry) error {
	entry.Data[hook.Field] = findCaller(hook.Skip)
	return nil
}

// 对caller进行递归查询, 直到找到非logrus包产生的第一个调用.
// 因为filename我获取到了上层目录名, 因此所有logrus包的调用的文件名都是 logrus/...
// 因此通过排除logrus开头的文件名, 就可以排除所有logrus包的自己的函数调用
func findCaller(skip int) string {
	file := ""
	line := 0
	for i := 0; i < 10; i++ {
		file, line = getCaller(skip + i)
		if !strings.HasPrefix(file, "logrus") {
			break
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// 这里其实可以获取函数名称的: fnName := runtime.FuncForPC(pc).Name()
// 但是我觉得有 文件名和行号就够定位问题, 因此忽略了caller返回的第一个值:pc
// 在标准库log里面我们可以选择记录文件的全路径或者文件名, 但是在使用过程成并发最合适的,
// 因为文件的全路径往往很长, 而文件名在多个包中往往有重复, 因此这里选择多取一层, 取到文件所在的上层目录那层.
func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	//fmt.Println(file)
	//fmt.Println(line)
	if !ok {
		return "", 0
	}
	n := 0
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			n++
			if n >= 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return file, line
}

func InitLog(logPath string, logFileName string, logLevel int) {
	log.SetLevel(log.Level(logLevel))

	baseLogPaht := path.Join(logPath, logFileName)
	writer, err := rotatelogs.New(
		baseLogPaht+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPaht),   // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(24*time.Hour),    // 文件最大保存时间
		rotatelogs.WithRotationTime(time.Hour), // 日志切割时间间隔
	)

	if err != nil {
		log.Printf("failed to create rotatelogs: %s", err)
		return
	}

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		log.DebugLevel: writer, // 为不同级别设置不同的输出目的
		log.InfoLevel:  writer,
		log.WarnLevel:  writer,
		log.ErrorLevel: writer,
		log.FatalLevel: writer,
		log.PanicLevel: writer,
	}, &log.TextFormatter{})
	log.AddHook(NewContextHook())
	log.AddHook(lfHook)

	//TODO
}
