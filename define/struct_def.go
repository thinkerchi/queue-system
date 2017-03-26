package define

import (
	"encoding/binary"
)

type ReadInfo struct {
	Cmd string
	Id  string
}

type NofityInfo struct {
	Front  int // 前面排队人数
	Status int // 用户状态，1：排队中，2：游戏中
}

type ClientInfo struct {
	Id             string
	NotifyInfoChan chan NofityInfo
}

type ChangeInfo struct {
	Players int
	Queuers int
}

func (t *NofityInfo) ReadFromBytes(src []byte) {
	if len(src) < 5 {
		return
	}
	front := binary.BigEndian.Uint32(src)
	status := int(src[4])

	t.Front = int(front)
	t.Status = status

	return
}

func (t *NofityInfo) ToBytes() []byte {
	bytes := make([]byte, 5)
	binary.BigEndian.PutUint32(bytes, uint32(t.Front))
	bytes[4] = byte(t.Status)

	return bytes
}

func (t *ReadInfo) ReadFromBytes(src []byte) {
	if len(src) < 36 {
		return
	}
	t.Cmd = string(src[:4])
	t.Id = string(src[4:])

	return
}

func (t *ReadInfo) ToBytes() []byte {
	cmdBytes := []byte(t.Cmd)
	idBytes := []byte(t.Id)

	var bytes []byte
	bytes = append(bytes, cmdBytes...)
	bytes = append(bytes, idBytes...)
	return bytes
}
