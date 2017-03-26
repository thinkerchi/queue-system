package client

import (
	"fmt"
	def "thinkerchi/queue-system/define"
)

var (
	NotifyInfoChan    chan def.NofityInfo
	QuitQueueInfoChan chan struct{}
)

func Init() {
	NotifyInfoChan = make(chan def.NofityInfo, 10)
	QuitQueueInfoChan = make(chan struct{})
}

func NoticeUser() {
	for {
		select {
		case notifyInfo := <-NotifyInfoChan:
			if notifyInfo.Status == 2 {
				fmt.Println("You are in the game now!")
			} else {
				fmt.Println("The number of people before you is ", notifyInfo.Front)
			}
		}
	}
}
