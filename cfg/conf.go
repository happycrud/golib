package cfg

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "cfg", "configs", "configuration fils paths default ./configs/")
}

func Config() *Configs {
	return configFn()
}

var configFn = sync.OnceValue[*Configs](initConfig)

func initConfig() *Configs {
	if confPath == "" {
		panic("empty cfg path")
	}
	files, err := os.ReadDir(confPath)
	if err != nil {
		panic(err)
	}
	x := make(map[string][]byte)
	for _, f := range files {
		data, err := os.ReadFile(path.Join(confPath, f.Name()))
		if err != nil {
			panic(err)
		}
		x[f.Name()] = data

	}
	at := atomic.Value{}
	at.Store(x)
	wc, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	ret := &Configs{
		dir:       confPath,
		filesRW:   at,
		fsWatcher: wc,
		closeCh:   make(chan struct{}),
	}
	go ret.watch()
	return ret
}

type EventFn func(event *fsnotify.Event)
type Configs struct {
	// config dir path
	dir string

	filesRW atomic.Value

	fsWatcher *fsnotify.Watcher

	closeCh chan struct{}
}

func (c *Configs) UnmarshalTo(key string, object interface{}) error {
	if v, ok := c.Raw(key); !ok {
		return errors.New("key not exist")
	} else {
		if strings.HasSuffix(strings.ToLower(key), ".toml") {
			return toml.Unmarshal(v, object)
		}

		if strings.HasSuffix(strings.ToLower(key), ".yml") {
			return yaml.Unmarshal(v, object)
		}

		if strings.HasSuffix(strings.ToLower(key), ".json") {
			return json.Unmarshal(v, object)
		}
		return errors.New("not support config file format")
	}

}

func (c *Configs) Watch(key string) error {
	if _, ok := c.Raw(key); !ok {
		return errors.New("key is not exist")
	}
	return c.fsWatcher.Add(path.Join(c.dir, key))
}

func (c *Configs) Raw(key string) ([]byte, bool) {
	v := c.filesRW.Load().(map[string][]byte)
	d, x := v[key]
	return d, x
}
func (c *Configs) RawString(key string) (string, bool) {
	v := c.filesRW.Load().(map[string][]byte)
	d, x := v[key]
	if x {
		return string(d), x
	}
	return "", x
}
func (c *Configs) watch() {
	for {
		select {
		case event, ok := <-c.fsWatcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) {
				// read file name
				f, err := os.ReadFile(event.Name)
				if err != nil {
					panic(err)
				}
				v := c.filesRW.Load().(map[string][]byte)
				_, n := path.Split(event.Name)
				v[n] = f
				c.filesRW.Store(v)
			}

		case e, ok := <-c.fsWatcher.Errors:
			if !ok {
				fmt.Println(e)
				return
			}

		}

	}

}
