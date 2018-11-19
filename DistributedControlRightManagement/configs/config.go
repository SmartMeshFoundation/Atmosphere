package configs

import "github.com/ethereum/go-ethereum/crypto/secp256k1"

var SvrSocketIP="127.0.0.1"

var SvrSocketPort="9999"

type TcpStateObject struct {
	BufferSize uint
	Buffer []byte
	Counter uint
}

var(
	Ip1="192.168.124.13"
	Ip2="192.168.124.15"
	Ip3="192.168.124.2"
	Ip4="192.168.124.10"
)

var ThresholdNum=8

var G=secp256k1.S256()