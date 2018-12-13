package commitments

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/Nik-U/pbc"
)

type Commitment struct {
	pubkey      *pbc.Element //zr
	committment *pbc.Element //g1
}

func (c *Commitment) Constructor(pubkey *pbc.Element, a *pbc.Element) {
	c.pubkey = pubkey
	c.committment = a
}

func (c *Commitment) UnmarshalJSON(b []byte) (err error) {
	var l int32
	l = int32(len(b))
	if l < 10 {
		err = fmt.Errorf("data  too short")
		return
	}
	b = b[1 : l-1] //跳过双引号
	b, err = base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		err = fmt.Errorf("UnmarshalJSON err %s", err)
		return
	}
	buf := bytes.NewBuffer(b)

	//process g
	c.pubkey, err = deSerializeElement(buf)
	if err != nil {
		return
	}

	//process h
	c.committment, err = deSerializeElement(buf)
	if err != nil {
		return
	}
	return nil

}

//MarshalJSON 自定义结构体json编码
func (c *Commitment) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := serilializeElement(buf, c.pubkey, pbc.Zr) //暂时认为只可能是G1
	err = serilializeElement(buf, c.committment, pbc.G1)
	s := base64.StdEncoding.EncodeToString(buf.Bytes())
	return []byte("\"" + s + "\""), err
}
