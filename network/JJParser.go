package network

import "fmt"

type JJParser struct {
	datas []byte
	Flag string
}


func (this *JJParser)ParseJJ(datas []byte) []*JJProtoCodec {
	pucks := make([]byte,len(datas)+len(this.datas))
	copy(pucks,this.datas)
	copy(pucks[len(this.datas):],datas)

	i := 0
	rets := make([]*JJProtoCodec,0)
	for ; ;  {
		jj := JJProtoCodec{}
		if len(pucks) == i {
			this.datas = make([]byte,0)
			return rets
		}
		_,len := jj.Decode(pucks[i:])
		if len < 0 {
			this.datas = pucks[i:]
			return rets
		} else {
			fmt.Printf("%s%d\t%d\t%d\t%X\t%d\t%d\t%v\n",this.Flag,jj.Times,jj.Idx,jj.Unknown1,jj.Type,jj.Unknown3,jj.Len,jj.Data)
			rets = append(rets,&jj)
			i += len
		}
	}

}