package queue

import (
	"container/list"
	"fmt"
	"sync/atomic"
	def "thinkerchi/queue-system/define"
)

var N int

var (
	EnqueueChan    chan def.ClientInfo // 进入排队时发送
	QuitQueueChan  chan string         // 退出排队时发送
	QuitGameChan   chan struct{}       //  退出游戏时发送
	QuitChan       chan string         // 链接断开时发送
	ChangeInfoChan chan def.ChangeInfo // 在线人数变化时发送
	IsQueuedChan   chan struct{}       // 是否有人在排队
)

var (
	WaitNumMap    map[string]int // 用户id和其当前排队位置的Map
	WaitList      *list.List     // 正在排队中的用户
	PlayChan      chan struct{}  // 正在游戏中的人，用于控制正在游戏中的人数
	OnlinePlayers int32          // 正在游戏中的人数
)

func GetOnlinePlayers() int {
	n := atomic.LoadInt32(&OnlinePlayers)
	return int(n)
}

func IncrOnlinePlayers() int {
	n := atomic.AddInt32(&OnlinePlayers, 1)
	return int(n)
}

func DecrOnlinePlayers() int {
	n := atomic.AddInt32(&OnlinePlayers, -1)
	return int(n)
}

func Init() {
	EnqueueChan = make(chan def.ClientInfo)
	QuitQueueChan = make(chan string)
	QuitGameChan = make(chan struct{})
	QuitChan = make(chan string)
	ChangeInfoChan = make(chan def.ChangeInfo)
	IsQueuedChan = make(chan struct{}, 100)

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
		case id := <-QuitChan:
			Quit(id)
		}
	}
}

func Enqueue(clientInfo def.ClientInfo) {
	WaitNumMap[clientInfo.Id] = WaitList.Len()
	WaitList.PushFront(clientInfo)

	notifyInfo := def.NofityInfo{
		Front:  WaitNumMap[clientInfo.Id],
		Status: 1,
	}

	clientInfo.NotifyInfoChan <- notifyInfo

	go func() {
		IsQueuedChan <- struct{}{}
	}()

	//	logs.Logger.Infof("coming id: %s, wait: %d", clientInfo.Id, WaitNumMap[clientInfo.Id])

	go EndInfoChanged(GetOnlinePlayers(), WaitList.Len())
}

func Quit(id string) {
	waitN, ok := WaitNumMap[id]
	if !ok {
		return
	}

	//	logs.Logger.Infof("id: %s, waitN: %d, ok: %v", id, waitN, ok)

	if waitN == -1 {
		go func() {
			<-PlayChan
			EndInfoChanged(DecrOnlinePlayers(), WaitList.Len())
		}()
	} else {
		QuitQueue(id)
	}

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
				//				logs.Logger.Infof("front: %d, status: %d", notifyInfo.Front, notifyInfo.Status)
				clientInfo.NotifyInfoChan <- notifyInfo
			}(clientInfo, wait)
		}
	}

	go EndInfoChanged(GetOnlinePlayers(), WaitList.Len())
}

func QuitGame() {

	e := WaitList.Back()
	if e == nil {
		<-PlayChan
		return
	} else {
		IncrOnlinePlayers()
	}

	WaitList.Remove(e)

	clientInfo := e.Value.(def.ClientInfo)

	WaitNumMap[clientInfo.Id] = -1

	go EndInfoChanged(GetOnlinePlayers(), WaitList.Len())

	go func(clientInfo def.ClientInfo) {
		notifyInfo := def.NofityInfo{
			Front:  0,
			Status: 2,
		}
		clientInfo.NotifyInfoChan <- notifyInfo
	}(clientInfo)

	for e = WaitList.Front(); e != nil; e = e.Next() {
		info := e.Value.(def.ClientInfo)
		wait := WaitNumMap[info.Id] - 1
		WaitNumMap[info.Id] = wait
		go func(info def.ClientInfo, wait int) {
			notifyInfo := def.NofityInfo{
				Front:  wait,
				Status: 1,
			}
			//			logs.Logger.Infof("front: %d, status: %d", notifyInfo.Front, notifyInfo.Status)
			info.NotifyInfoChan <- notifyInfo
		}(info, wait)
	}

}

func EndInfoChanged(players, queuers int) {
	changeInfo := def.ChangeInfo{
		Players: players,
		Queuers: queuers,
	}
	ChangeInfoChan <- changeInfo
}

func ListenChanges() {
	for {
		select {
		case info := <-ChangeInfoChan:
			fmt.Println("正在游戏人数：", info.Players, "\t正在排队人数：", info.Queuers)
		}
	}
}

func EnterGame() {
	EndInfoChanged(0, 0)
	for {
		PlayChan <- struct{}{}
		<-IsQueuedChan
		QuitGameChan <- struct{}{}
	}
}
