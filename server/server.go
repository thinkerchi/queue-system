package server

import (
	"fmt"
	"log"
	"net"
	"os"
	def "thinkerchi/queue-system/define"
	que "thinkerchi/queue-system/queue"
	"time"
)

func Run() {
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return

	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go NewHandler(conn).Handle()
	}
}

func NewHandler(conn net.Conn) *Handler {
	return &Handler{
		conn: conn,
		stop: make(chan struct{}),
	}
}

type Handler struct {
	conn    net.Conn
	timeout int64
	stop    chan struct{}
}

func (h *Handler) Stop() {
	h.stop <- struct{}{}
}

func (h *Handler) SetTimeout(timeout int64) {
	h.conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

// 协议 4(命令) + 20(id)
// 命令： SHUT/QUIT/OPEN
func (h *Handler) Handle() {
	defer h.conn.Close()

	clientInfo, err := h.InitPacket()
	if err != nil {
		log.Println(err)
		return
	}

	go h.WriteBack(clientInfo)

	h.KeepReading()

}

func (h *Handler) WriteBack(clientInfo *def.ClientInfo) {
	for {
		select {
		case nofityInfo := <-clientInfo.NotifyInfoChan:
			go func() {
				if err := h.WritePacket(&nofityInfo); err != nil {
					log.Printf("user %s write error: %v\n", clientInfo.Id, err)
				}
			}()
		case <-h.stop:
			return
		}
	}
}

func (h *Handler) KeepReading() {
	defer func() {
		h.Stop()

	}()
	for {
		readInfo, err := h.ReadPacket()
		if err != nil {
			log.Println("conn is close..., error: ", err)
			return
		}

		switch readInfo.Cmd {
		case "SHUT":
			go func() {
				que.QuitQueueChan <- readInfo.Id
			}()
			log.Printf("user %s is quitting queuing\n", readInfo.Id)
		case "QUIT":
			go func() {
				que.QuitGameChan <- struct{}{}
			}()
			log.Printf("user %s is quitting game\n", readInfo.Id)
		default:
			log.Printf("Unknown cmd %v\n", readInfo)
		}

	}
}

func (h *Handler) InitPacket() (clientInfo *def.ClientInfo, err error) {
	readInfo, err := h.ReadPacket()
	if err != nil {
		log.Printf("read conn error:%v\n", err)
		return
	}
	if readInfo.Cmd != "OPEN" {
		err = fmt.Errorf("expected %s, got %s", "OPEN", readInfo.Cmd)
		log.Printf("%#v", *readInfo)
		return
	}
	clientInfo = &def.ClientInfo{
		Id:             readInfo.Id,
		NotifyInfoChan: make(chan def.NofityInfo),
	}

	go func() {
		que.EnqueueChan <- *clientInfo
	}()

	return
}

func (h *Handler) ReadPacket() (content *def.ReadInfo, err error) {
	rb := make([]byte, 36)
	n, err := h.conn.Read(rb)
	if err != nil {
		return
	} else if n != 36 {
		err = fmt.Errorf("read not enough bytes")
		return
	}

	content = new(def.ReadInfo)
	content.ReadFromBytes(rb)

	return
}

func (h *Handler) WritePacket(info *def.NofityInfo) (err error) {
	rb := info.ToBytes()
	_, err = h.conn.Write(rb)

	return
}
