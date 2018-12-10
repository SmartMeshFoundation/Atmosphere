package commitments

import (
	"fmt"
	"math/big"

	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/Nik-U/pbc"
)

type Open struct {
	secrets    []*big.Int
	randomness *pbc.Element
}

func (open *Open) Constructor(randomness *pbc.Element, secrets []*big.Int) {
	open.secrets = secrets
	open.randomness = randomness
}

func (open *Open) GetSecrets() []*big.Int {
	return open.secrets
}

func (open *Open) getRandomness() *pbc.Element {
	return open.randomness
}

func (c *Open) UnmarshalJSON(b []byte) (err error) {
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
	c.randomness, err = deSerializeElement(buf)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf.Bytes(), &c.secrets)
	return nil

}

//MarshalJSON 自定义结构体json编码
func (c *Open) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := new(bytes.Buffer)
	err := serilializeElement(buf, c.randomness, pbc.Zr) //暂时认为只可能是Zr
	b, err = json.Marshal(c.secrets)
	_, err = buf.Write(b)
	s := base64.StdEncoding.EncodeToString(buf.Bytes())
	return []byte("\"" + s + "\""), err
}
