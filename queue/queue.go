package queue

import (
	"container/list"
	"fmt"
	"sync/atomic"
	def "thinkerchi/queue-system/define"
)

var N int // 最大同时可在线游戏人数

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

// 操作WaitList列表的入口，所有有关操作都在这个函数中
func OperateWaitList() {
	for {
		select {
		// 有用户登陆
		case clientInfo := <-EnqueueChan:
			Enqueue(clientInfo)
		// 有用户退出排队
		case id := <-QuitQueueChan:
			QuitQueue(id)
		// 有用户退出游戏
		case <-QuitGameChan:
			QuitGame()
		// 有用户退出，具体是退出排队还是游戏在Quit(）内进一步判断
		case id := <-QuitChan:
			Quit(id)
		}
	}
}

// 新用户登陆
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

	EndInfoChanged(GetOnlinePlayers(), WaitList.Len())
}

// 用户退出
func Quit(id string) {
	waitN, ok := WaitNumMap[id]
	if !ok {
		return
	}

	if waitN == -1 {
		EndInfoChanged(DecrOnlinePlayers(), WaitList.Len())
		<-PlayChan
	} else {
		QuitQueue(id)
	}

}

// 用户退出排队
// 从WaitList表删除此用户，并且更新受影响的排队用户的等待位置
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

	EndInfoChanged(GetOnlinePlayers(), WaitList.Len())
}

// 有用户退出游戏
// 并不关心是哪个用户退出游戏...
// 让WaitList中时间最靠前的排队用户进入游戏，并且更新所有其他排队用户的排队位置
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

	EndInfoChanged(GetOnlinePlayers(), WaitList.Len())

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

// 向控制台打印实时更新的用户数据
func ListenChanges() {
	for {
		select {
		case info := <-ChangeInfoChan:
			fmt.Println("正在游戏人数：", info.Players, "\t正在排队人数：", info.Queuers)
		}
	}
}

// 检测是否还有游戏空位
// 如果有，并且有人排队，就发送一个有人退出游戏的信息(将有人退出游戏和有空位抽象成一种情况处理)
func EnterGame() {
	EndInfoChanged(0, 0)
	for {
		PlayChan <- struct{}{}
		<-IsQueuedChan
		QuitGameChan <- struct{}{}
	}
}
