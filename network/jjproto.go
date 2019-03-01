package network

import (
	"github.com/davyxu/cellnet"
	"io"
	"net"
	"github.com/davyxu/cellnet/proc"
	"github.com/davyxu/cellnet/proc/tcp"
	"encoding/binary"
	"fmt"
)

type JJProtoCodec struct {
	Times uint32//4字节 GetTickCount + time32
	Idx uint32//4字节数据包序号
	Unknown1 uint32//跟0x75比较
	Type uint32//是否等于0x40000000  如果不等于则 and 0xBFFFFFFF
	Unknown3 uint32
	Len uint32
	Data []byte
	RawData []byte
}

// 编码器的名称
func (self *JJProtoCodec) Name() string {
	return "jjgame"
}

func (self *JJProtoCodec) MimeType() string {
	return "application/binary"
}



// 将结构体编码为JSON的字节数组
func (self *JJProtoCodec) Encode() ( []byte,  error) {
	pkt := make([]byte, bodySize+self.Len)

	binary.LittleEndian.PutUint32(pkt, self.Times)
	binary.LittleEndian.PutUint32(pkt, self.Idx)
	binary.LittleEndian.PutUint32(pkt, self.Unknown1)
	binary.LittleEndian.PutUint32(pkt, self.Type)
	binary.LittleEndian.PutUint32(pkt, self.Unknown3)
	binary.LittleEndian.PutUint32(pkt, self.Len)

	copy(pkt[24:],self.Data)
	return pkt,nil

}

// 将JSON的字节数组解码为结构体
//int -1 不够
// >=0 d
func (self *JJProtoCodec) Decode(data []byte) (error,int) {
	if len(data) < 24 {
		return fmt.Errorf("data not enough"),-1
	}
	self.Type = binary.LittleEndian.Uint32(data[0xC:])
	self.Len = binary.LittleEndian.Uint32(data[0x14:])
	self.Idx = binary.LittleEndian.Uint32(data[0x4:])
	var i int = 0

	if self.Type == 0x40000000 {

		self.Times = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		i += 4
		self.Unknown1 = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		self.Type = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		self.Unknown3 = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		i += 4
		i += int(self.Len)

		if i > len(data) {
			return fmt.Errorf("data not enough"),-1
		}

	} else {

		self.Times = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		i += 4
		self.Unknown1 = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		self.Type = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		self.Unknown3 = binary.LittleEndian.Uint32(data[i:i+4])
		i += 4
		i += 4
		i += int(self.Len)

		if i > len(data) {
			return fmt.Errorf("data not enough"),-1
		}

		self.Type  = self.Type  & 0xBFFFFFFF
		if self.Len > 0 {
			for j := 0; j < int(self.Len); {
				temp := int(self.Idx)
				temp = temp ^ j
				temp = temp & 3

				if j+ 24 >= len(data) {
					fmt.Println("error")
				}
				data[j+24] = data[j+24] ^ data[temp]
				j++
			}
		}
	}
	
/*
signed int __thiscall sub_42FAE0(_DWORD *this, _DWORD *a2)
{
  int v2; // edx
  int v3; // ecx

  if ( a2[3] & 0x40000000 )
  {
    this[128] = 1;
    v2 = a2[5];
    a2[3] &= 0xBFFFFFFF;
    if ( v2 )
    {
      v3 = 0;
      if ( v2 > 0 )
      {
        do
        {
          *((_BYTE *)a2 + v3 + 24) ^= *((_BYTE *)a2 + (((unsigned __int8)v3 ^ (unsigned __int8)a2[1]) & 3));
          ++v3;
        }
        while ( v3 < a2[5] );
      }
    }
  }
  return 1;
}

*/





	self.Data = make([]byte,self.Len)
	copy(self.Data,data[24:i])
	//self.RawData = data



	return nil,i
}

func init() {

	// 注册编码器

	proc.RegisterProcessor("jj.ltv", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(JJMessageTransmitter))
		bundle.SetHooker(new(tcp.MsgHooker))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))


	})

}


type JJMessageTransmitter struct   {

}

type socketOpt interface {
	MaxPacketSize() int
	ApplySocketReadTimeout(conn net.Conn, callback func())
	ApplySocketWriteTimeout(conn net.Conn, callback func())
}

func (JJMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {
	reader, ok := ses.Raw().(io.Reader)

	// 转换错误，或者连接已经关闭时退出
	if !ok || reader == nil {
		return nil, nil
	}

	opt := ses.Peer().(socketOpt)

	if conn, ok := reader.(net.Conn); ok {

		// 有读超时时，设置超时
		opt.ApplySocketReadTimeout(conn, func() {

			msg, err = RecvLTVPacket(reader, opt.MaxPacketSize())

		})
	}

	return
}

func (JJMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) (err error) {
	data,ok := msg.([]byte)
	if !ok {
		return nil
	}

	writer, ok := ses.Raw().(io.Writer)

	// 转换错误，或者连接已经关闭时退出
	if !ok || writer == nil {
		return nil
	}

	opt := ses.Peer().(socketOpt)

	// 有写超时时，设置超时
	opt.ApplySocketWriteTimeout(writer.(net.Conn), func() {
		err = WriteFull(writer,data)
		if err != nil {
			fmt.Println(err)
		}

	//	err = SendLTVPacket(writer, ses.(cellnet.ContextSet), msg)

	})

	return
}