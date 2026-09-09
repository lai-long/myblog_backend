package logx

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/zero-contrib/logx/zerologx"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Path       string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Level      string //debug / info / warn /error
}

func Init(c Config) error {
	if err := os.MkdirAll(filepath.Dir(c.Path), os.ModePerm); err != nil {
		return err
	}

	fileWriter := &lumberjack.Logger{
		Filename:   c.Path,
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxBackups,
		LocalTime:  true,
		Compress:   true,
	}
	comsole := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.DateTime,
		NoColor:    false,
	}
	multi := zerolog.MultiLevelWriter(fileWriter, comsole)

	logger := zerolog.New(multi).
		With().
		Timestamp().
		Caller().
		Logger()

	switch c.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logx.SetWriter(zerologx.NewZeroLogWriter(logger))
	return nil
}
