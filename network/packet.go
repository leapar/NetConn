package network

import (
	"encoding/binary"
	"errors"
	"github.com/davyxu/cellnet"
	"io"
	"fmt"
)

var (
	ErrMaxPacket  = errors.New("packet over size")
	ErrMinPacket  = errors.New("packet short size")
	ErrShortMsgID = errors.New("short msgid")
)

const (
	bodySize  = 24 // 包体大小字段
)

// 接收Length-Type-Value格式的封包流程
func RecvLTVPacket(reader io.Reader, maxPacketSize int) (msg interface{}, err error) {

	// Size为uint16，占2字节
	var sizeBuffer = make([]byte, bodySize)

	// 持续读取Size直到读到为止
	_, err = io.ReadFull(reader, sizeBuffer)

	// 发生错误时返回
	if err != nil {
		return
	}

	if len(sizeBuffer) < bodySize {
		return nil, ErrMinPacket
	}

	// 用小端格式读取Size
	size := binary.LittleEndian.Uint32(sizeBuffer[20:24])

	if maxPacketSize > 0 && size >= uint32(maxPacketSize) {
		return nil, ErrMaxPacket
	}

	// 分配包体大小
	body := make([]byte, bodySize+size)

	// 读取包体数据
	_, err = io.ReadFull(reader, body[bodySize:])

	// 发生错误时返回
	if err != nil {
		return
	}

	copy(body,sizeBuffer)


	// 将字节数组和消息ID用户解出消息

	jjMsg := &JJProtoCodec{}
	err,_ = jjMsg.Decode(body)
	/*msg, _, err = codec.DecodeMessage(int(msgid), msgData)
	 */
	if err != nil {
		// TODO 接收错误时，返回消息
		return nil, err
	}

	return jjMsg,nil
}

// 发送Length-Type-Value格式的封包流程
func SendLTVPacket(writer io.Writer, ctx cellnet.ContextSet, data interface{}) error {
	jjMsg,ok := data.(JJProtoCodec)
	if !ok {
		return fmt.Errorf("invalid jj msg")
	}
	pkt,err := jjMsg.Encode()
	if err!= nil {
		return err
	}
	// 将数据写入Socket
	err = WriteFull(writer, pkt)
	return err
}

// 完整发送所有封包
func WriteFull(writer io.Writer, buf []byte) error {

	total := len(buf)

	for pos := 0; pos < total; {

		n, err := writer.Write(buf[pos:])

		if err != nil {
			return err
		}

		pos += n
	}

	return nil

}
