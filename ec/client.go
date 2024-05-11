package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"path"
	"strings"
)

type Config struct {
	Etcd     *EtcdConfig
	Services []*ServiceConfig
}
type EtcdConfig struct {
	Address []string
}
type ServiceConfig struct {
	Id         string
	ConfigPath string
	WatchKeys  []string
}

var conf string

func init() {
	flag.StringVar(&conf, "conf", "service_config.toml", "config file path")
}
func main() {
	flag.Parse()
	confdata, err := os.ReadFile(conf)
	if err != nil {
		panic(err)
	}
	conf := &Config{}
	if err := toml.Unmarshal(confdata, conf); err != nil {
		panic(err)
	}
	cli, err := clientv3.NewFromURLs(conf.Etcd.Address)
	if err != nil {
		panic(err)
	}
	for _, v := range conf.Services {
		service := fmt.Sprintf("%s/configs/", v.Id)
		resp, err := cli.Get(context.Background(), service, clientv3.WithPrefix())
		if err != nil {
			fmt.Println(err)
		}
		for _, kv := range resp.Kvs {
			filename := strings.TrimPrefix(string(kv.Key), service)
			f, err := os.Create(path.Join(v.ConfigPath, filename))
			if err != nil {

			}
			f.Write(kv.Value)
			f.Close()
		}
		for _, watchkey := range v.WatchKeys {
			ch := cli.Watch(context.Background(), path.Join(service, watchkey))
			for x := range ch {
				for range x.Events {
				}
			}
		}

	}

}
