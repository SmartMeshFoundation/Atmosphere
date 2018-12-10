package commitments

import (
	"encoding/binary"
	"io"

	"github.com/Nik-U/pbc"
)

func serilializeElement(w io.Writer, e *pbc.Element, f pbc.Field) (err error) {
	b := e.Bytes()
	err = binary.Write(w, binary.BigEndian, int32(len(b)))
	err = binary.Write(w, binary.BigEndian, int32(f))
	_, err = w.Write(b)
	return err
}

func deSerializeElement(r io.Reader) (e *pbc.Element, err error) {
	var l int32
	err = binary.Read(r, binary.BigEndian, &l)
	if err != nil {
		return
	}
	var f int32
	err = binary.Read(r, binary.BigEndian, &f)
	if err != nil {
		return
	}
	switch pbc.Field(f) {
	case pbc.G1:
		e = DefaultPairing.NewG1()
	case pbc.G2:
		e = DefaultPairing.NewG2()
	case pbc.GT:
		e = DefaultPairing.NewGT()
	case pbc.Zr:
		e = DefaultPairing.NewZr()
	}
	tmpBytes := make([]byte, l)
	_, err = r.Read(tmpBytes)
	if err != nil {
		return
	}
	e.SetBytes(tmpBytes)
	return
}
