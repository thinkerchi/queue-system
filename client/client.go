package client

import (
	"fmt"
	"log"
	"net"
	def "thinkerchi/queue-system/define"
	"thinkerchi/queue-system/utils"
)

var (
	Ip   string
	Port string
)

func Run() {
	opt := fmt.Sprintf("%s:%s", Ip, Port)
	conn, err := net.Dial("tcp", opt)
	if err != nil {
		log.Println(err)
		return
	}

	NewClient(conn).Handle()
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn: conn,
		stop: make(chan struct{}),
	}
}

type Client struct {
	conn net.Conn
	stop chan struct{}
}

func (c *Client) Stop() {
	c.stop <- struct{}{}
}

func (c *Client) Handle() {
	defer func() {
		c.conn.Close()
	}()

	info, err := c.InitPacket()
	if err != nil {
		return
	}

	fmt.Println(info.Id)

	go c.KeepReceiving()

	c.WriteToServer(info)
}

func (c *Client) KeepReceiving() {
	defer c.Stop()

	for {

		notifyInfo, err := c.ReadPacket()
		if err != nil {
			return
		}
		go func() {
			NotifyInfoChan <- *notifyInfo
		}()

	}
}

func (c *Client) WriteToServer(info *def.ReadInfo) {
	select {
	case <-QuitQueueInfoChan:
		c.conn.Write(info.ToBytes())
		fmt.Println("Quitting....")
	case <-c.stop:
		return
	}

}

func (c *Client) InitPacket() (readInfo *def.ReadInfo, err error) {
	var initInfo = def.ReadInfo{
		Cmd: "OPEN",
		Id:  utils.GetGuid(),
	}

	bytes := initInfo.ToBytes()

	n, err := c.conn.Write(bytes)
	if err != nil {
		log.Println(err)
		return
	} else if n != 36 {
		err = fmt.Errorf("expected %d, got %d\n", 36, n)
		log.Println(err)
		return
	}

	readInfo = &initInfo

	return
}

func (c *Client) ReadPacket() (notifyInfo *def.NofityInfo, err error) {
	bytes := make([]byte, 5)

	_, err = c.conn.Read(bytes)
	if err != nil {
		log.Println(err)
		return
	}

	notifyInfo = new(def.NofityInfo)
	notifyInfo.ReadFromBytes(bytes)

	return
}
