package commitments

import (
	"fmt"
	"math/big"

	"bytes"
	"encoding/base64"
	"encoding/binary"

	"github.com/Nik-U/pbc"
)

//DefaultPairing is the pairing for all notaries.
var DefaultPairing *pbc.Pairing

func init() {
	var err error
	DefaultPairing, err = pbc.NewPairingFromString(ConsensusParamOfPBC)
	if err != nil {
		panic(fmt.Sprintf("NewPairingFromString from ConsensusParamOfPBC err %s", err))
	}
}

type MultiTrapdoorMasterPublicKey struct {
	G *pbc.Element
	Q *big.Int
	H *pbc.Element
}

func (mtmp *MultiTrapdoorMasterPublicKey) Constructor(g *pbc.Element,
	q *big.Int, h *pbc.Element) {
	mtmp.G = g
	mtmp.Q = q
	mtmp.H = h
}

func (mtmp *MultiTrapdoorMasterPublicKey) UnmarshalJSON(b []byte) (err error) {
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
	mtmp.G, err = deSerializeElement(buf)
	if err != nil {
		return
	}

	//process h
	mtmp.H, err = deSerializeElement(buf)
	if err != nil {
		return
	}

	//process q
	err = binary.Read(buf, binary.BigEndian, &l)
	if err != nil {
		return
	}
	tmpBytes := make([]byte, l)
	_, err = buf.Read(tmpBytes)
	if err != nil {
		return err
	}
	mtmp.Q = new(big.Int)
	mtmp.Q.SetBytes(tmpBytes)
	return nil

}

//MarshalJSON 自定义结构体json编码
func (mtmp *MultiTrapdoorMasterPublicKey) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := serilializeElement(buf, mtmp.G, pbc.G1) //暂时认为只可能是G1
	err = serilializeElement(buf, mtmp.H, pbc.G1)
	qbytes := mtmp.Q.Bytes()
	err = binary.Write(buf, binary.BigEndian, int32(len(qbytes)))
	_, err = buf.Write(qbytes)
	s := base64.StdEncoding.EncodeToString(buf.Bytes())
	return []byte("\"" + s + "\""), err
}
