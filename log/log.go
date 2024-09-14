package log

import (
	"io"
	"log/slog"
	"os"

	"github.com/happycrud/golib/log/rotate"
)

type Config struct {
	Level      int
	AddSource  bool
	JSONFORMAT bool
	StdOutput  bool
	Root       string
	Prefix     string
}

func Init(c *Config) {
	var l slog.Handler
	var w io.Writer
	opt := &slog.HandlerOptions{
		Level:     slog.Level(c.Level),
		AddSource: c.AddSource,
	}
	if c.StdOutput {
		w = os.Stdout
	} else {
		var err error
		w, err = rotate.New(c.Root, c.Prefix)
		if err != nil {
			panic(err)
		}
	}

	if c.JSONFORMAT {
		l = slog.NewJSONHandler(w, opt)
	} else {
		l = slog.NewTextHandler(w, opt)
	}

	slog.SetDefault(slog.New(l))
}
