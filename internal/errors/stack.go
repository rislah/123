package errors

import (
	"fmt"
	"hash/crc32"
	"os"
	"runtime"
	"strings"
)

var (
	knownFilePathPatterns = []string{
		"github.com/",
	}
)

type Frame struct {
	Filename string `json:"filename"`
	Method   string `json:"method"`
	Line     int    `json:"lineno"`
}

type Stack []Frame

func BuildStack(skip int) Stack {
	stack := make(Stack, 0)

	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		file = shortenFilePath(file)
		stack = append(stack, Frame{file, functionName(pc), line})
	}

	return stack
}

func (s Stack) Fingerprint() string {
	hash := crc32.NewIEEE()
	for _, frame := range s {
		fmt.Fprintf(hash, "%s%s%d", frame.Filename, frame.Method, frame.Line)
	}
	return fmt.Sprintf("%x", hash.Sum32())
}

func shortenFilePath(s string) string {
	idx := strings.Index(s, "/src/pkg/")
	if idx != -1 {
		return s[idx+5:]
	}
	for _, pattern := range knownFilePathPatterns {
		idx = strings.Index(s, pattern)
		if idx != -1 {
			return s[idx:]
		}
	}
	return s
}

func functionName(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "???"
	}
	name := fn.Name()
	end := strings.LastIndex(name, string(os.PathSeparator))
	return name[end+1 : len(name)]
}
