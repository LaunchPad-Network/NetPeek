package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/LaunchPad-Network/NetPeek/internal/version"

	"github.com/lfcypo/viperx"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	Stdout = iota + 1
	File
	FileAndStdout
)

func rotaryFile() *lumberjack.Logger {
	logFile := viperx.GetString("log.file", fmt.Sprintf("./logs/%s.log", version.EntryPoint()))
	logFileSplit := strings.Split(logFile, string(os.PathSeparator))
	if !strings.Contains(logFileSplit[len(logFileSplit)-1], ".") {
		logFile = path.Join(logFile, fmt.Sprintf("%s.log", version.EntryPoint()))
	}

	return &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    viperx.GetInt("log.maxSize", 100),
		MaxBackups: viperx.GetInt("log.maxBackups", 10),
		MaxAge:     viperx.GetInt("log.maxAge", 30),
		Compress:   viperx.GetBool("log.compress", true),
		LocalTime:  true,
	}
}

func Output() io.Writer {
	mode := viperx.GetInt("log.mode", Stdout)
	if mode == FileAndStdout {
		return io.MultiWriter(os.Stdout, rotaryFile())
	} else if mode == File {
		return rotaryFile()
	} else if mode == Stdout {
		return os.Stdout
	} else {
		panic("log mode is invalid")
	}
}
