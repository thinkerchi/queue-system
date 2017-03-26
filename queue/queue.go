package queue

import (
	"container/list"
	def "thinkerchi/queue-system/define"
)

//type def.NofityInfo struct {
//	Front  int // 前面排队人数
//	Status int // 用户状态，1：排队中，2：游戏中
//}

//type def.ClientInfo struct {
//	Id             string
//	NofityInfoChan chan def.NofityInfo
//}

var (
	EnqueueChan   chan def.ClientInfo // 进入排队时发送
	QuitQueueChan chan string         // 退出排队时发送
	QuitGameChan  chan struct{}       //  退出游戏时发送
)

var (
	WaitNumMap map[string]int
	WaitList   *list.List
	PlayChan   chan struct{} // 正在游戏中的人，用于控制正在游戏中的人数
)

var N = 2

func Init() {
	EnqueueChan = make(chan def.ClientInfo)
	QuitQueueChan = make(chan string)
	QuitGameChan = make(chan struct{})

	PlayChan = make(chan struct{}, N)
	WaitNumMap = make(map[string]int)
	WaitList = list.New()
}

func OperateWaitList() {
	for {
		select {
		case clientInfo := <-EnqueueChan:
			Enqueue(clientInfo)
		case id := <-QuitQueueChan:
			QuitQueue(id)
		case <-QuitGameChan:
			QuitGame()
		}
	}
}

func Enqueue(clientInfo def.ClientInfo) {
	WaitNumMap[clientInfo.Id] = WaitList.Len()
	WaitList.PushFront(clientInfo)

	go QueuerChanged(WaitList.Len())
}

func QuitQueue(id string) {

	delete(WaitNumMap, id)
	for e := WaitList.Front(); e != nil; e = e.Next() {
		clientInfo := e.Value.(def.ClientInfo)
		if clientInfo.Id == id {
			WaitList.Remove(e)
			break
		} else {
			wait := WaitNumMap[clientInfo.Id] - 1
			WaitNumMap[clientInfo.Id] = wait
			go func(clientInfo def.ClientInfo, wait int) {
				notifyInfo := def.NofityInfo{
					Front:  wait,
					Status: 1,
				}
				clientInfo.NotifyInfoChan <- notifyInfo
			}(clientInfo, wait)
		}
	}

	go QueuerChanged(WaitList.Len())
}

func QuitGame() {

	e := WaitList.Back()
	if e == nil {
		<-PlayChan
		return
	} else {
		go PlayerChanged(len(PlayChan))
	}

	WaitList.Remove(e)
	go QueuerChanged(WaitList.Len())

	clientInfo := e.Value.(def.ClientInfo)
	go func(clientInfo def.ClientInfo) {
		notifyInfo := def.NofityInfo{
			Front:  0,
			Status: 2,
		}
		clientInfo.NotifyInfoChan <- notifyInfo
	}(clientInfo)

	for e = WaitList.Front(); e != nil; e = e.Next() {
		info := e.Value.(def.ClientInfo)
		wait := WaitNumMap[clientInfo.Id] - 1
		WaitNumMap[clientInfo.Id] = wait
		go func(info def.ClientInfo, wait int) {
			notifyInfo := def.NofityInfo{
				Front:  wait,
				Status: 1,
			}
			clientInfo.NotifyInfoChan <- notifyInfo
		}(info, wait)
	}

}

func PlayerChanged(leng int) {
	PlayerChangedChan <- leng
}

func QueuerChanged(leng int) {
	QueuerChangedChan <- leng
}

var (
	PlayerChangedChan chan int
	QueuerChangedChan chan int
)

func EnterGame() {
	for {
		PlayChan <- struct{}{}
		QuitGameChan <- struct{}{}
	}
}
