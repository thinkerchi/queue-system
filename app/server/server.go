package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	que "thinkerchi/queue-system/queue"
	"thinkerchi/queue-system/server"

	"github.com/dlintw/goconf"
)

func main() {

	confFile := "./server.ini"

	l_conf, err := goconf.ReadConfigFile(confFile)
	if err != nil {
		fmt.Println("read file ", confFile, " error: ", err.Error())
		return
	}

	if err := Init(l_conf); err != nil {
		fmt.Println("init error: ", err.Error())
		return
	}

	que.Init()
	go que.OperateWaitList()
	go que.EnterGame()
	go que.ListenChanges()

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	server.Run()
}

func Init(l_conf *goconf.ConfigFile) (err error) {
	ip, err := l_conf.GetString("queue", "ip")
	if err != nil {
		return
	}

	if ip == "" {
		ip = "localhost"
	}
	server.Ip = ip

	port, err := l_conf.GetString("queue", "port")
	if err != nil {
		return
	}

	if port == "" {
		port = "1234"
	}
	server.Port = port

	maxOnline, err := l_conf.GetInt("queue", "max_online")
	if err != nil {
		return
	}

	if maxOnline <= 0 {
		maxOnline = 100
	}
	que.N = maxOnline

	return
}
