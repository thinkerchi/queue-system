package main

import (
	"fmt"
	//	"sync"
	"thinkerchi/queue-system/client"

	"github.com/dlintw/goconf"
)

func main() {
	run()
}

func run() {
	conf := "./client.ini"

	l_conf, err := goconf.ReadConfigFile(conf)
	if err != nil {
		fmt.Println("read file ", conf, " error: ", err.Error())
		return
	}

	if err := Init(l_conf); err != nil {
		fmt.Println("init error ", err.Error())
		return
	}

	client.Init()
	go client.NoticeUser()

	client.Run()
}

func Init(l_conf *goconf.ConfigFile) (err error) {
	ip, err := l_conf.GetString("server", "ip")
	if err != nil {
		return
	}

	port, err := l_conf.GetString("server", "port")
	if err != nil {
		return
	}

	client.Ip = ip
	client.Port = port

	return
}
