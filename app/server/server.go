package main

import (
	que "thinkerchi/queue-system/queue"
	"thinkerchi/queue-system/server"
)

func main() {
	que.Init()
	go que.OperateWaitList()
	go que.EnterGame()
	server.Run()
}
