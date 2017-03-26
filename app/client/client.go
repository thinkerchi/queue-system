package main

import (
	"thinkerchi/queue-system/client"
)

func main() {
	client.Init()
	go client.NoticeUser()

	client.Run()
}
