package network

import (
	"time"

	"fmt"

	"net"

	"errors"
	"sync"

	"github.com/SmartMeshFoundation/Atmosphere/encoding"
	"github.com/SmartMeshFoundation/Atmosphere/internal/rpanic"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/network/xmpptransport"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
)

//Policier to control the sending speed of transporter
type Policier interface {
	//Consume tokens.
	//Args:
	//tokens (float): number of transport tokens to consume
	//Returns:
	//wait_time (float): waiting time for the consumer
	Consume(tokens float64) time.Duration
}

//DeviceTypeMobile if you are a Photon running on a mobile phone
var DeviceTypeMobile = xmpptransport.TypeMobile

//DeviceTypeMeshBox if you are a Photon running on a meshbox
var DeviceTypeMeshBox = xmpptransport.TypeMeshBox

//DeviceTypeOther if you don't known the type,and is not a mobile phone, then other
var DeviceTypeOther = xmpptransport.TypeOtherDevice

//Transporter denotes a communication transport used by protocol
type Transporter interface {
	//Send a message to receiver
	Send(receiver common.Address, data []byte) error
	//Start ,ready for send and receive
	Start()
	//Stop send and receive
	Stop()
	//StopAccepting stops receiving
	StopAccepting()
	//RegisterProtocol a receiver
	RegisterProtocol(protcol ProtocolReceiver)
	//NodeStatus get node's status and is online right now
	NodeStatus(addr common.Address) (deviceType string, isOnline bool)
}

type dummyPolicy struct {
}

//Consume mocker
func (dp *dummyPolicy) Consume(tokens float64) time.Duration {
	time.Now()
	return 0
}

type timeFunc func() time.Time

//TokenBucket Implementation of the token bucket throttling algorithm.
type TokenBucket struct {
	Capacity  float64
	FillRate  float64
	Tokens    float64
	timeFunc  timeFunc
	Timestamp time.Time
}

//NewTokenBucket create a TokenBucket
func NewTokenBucket(capacity, fillRate float64, timeFunc ...timeFunc) *TokenBucket {
	tb := &TokenBucket{
		Capacity: capacity,
		FillRate: fillRate,
		Tokens:   capacity,
	}
	if len(timeFunc) == 1 {
		tb.timeFunc = timeFunc[0]
	} else {
		tb.timeFunc = time.Now
	}
	tb.Timestamp = tb.timeFunc()
	return tb
}

//Consume calc wait time.
func (tb *TokenBucket) Consume(tokens float64) time.Duration {
	waitTime := 0.0
	tb.Tokens -= tokens
	if tb.Tokens < 0 {
		tb.getTokens()
	}
	if tb.Tokens < 0 {
		waitTime = -tb.Tokens / tb.FillRate
	}
	return time.Duration(waitTime * float64(time.Second))
}
func (tb *TokenBucket) getTokens() {
	now := tb.timeFunc()
	fill := float64(now.Sub(tb.Timestamp)) / float64(time.Second)
	tb.Tokens += tb.FillRate * fill
	if tb.Tokens > tb.Capacity {
		tb.Tokens = tb.Capacity
	}
	tb.Timestamp = tb.timeFunc()
}

//ProtocolReceiver receive
type ProtocolReceiver interface {
	receive(data []byte)
}

//
/*
UDPTransport represents a UDP server
but how to handle listen error?
we need stop listen when switch to background
restart listen when switch foreground
*/
type UDPTransport struct {
	protocol      ProtocolReceiver
	conn          *SafeUDPConnection
	UAddr         *net.UDPAddr
	policy        Policier
	stopped       bool
	stopReceiving bool //todo use atomic to replace
	intranetNodes map[common.Address]*net.UDPAddr
	lock          sync.RWMutex
	name          string
	log           log.Logger
}

//NewUDPTransport create UDPTransport
func NewUDPTransport(name, host string, port int, protocol ProtocolReceiver, policy Policier) (t *UDPTransport, err error) {
	t = &UDPTransport{
		UAddr: &net.UDPAddr{
			IP:   net.ParseIP(host),
			Port: port,
		},
		protocol:      protocol,
		policy:        policy,
		log:           log.New("name", name),
		intranetNodes: make(map[common.Address]*net.UDPAddr),
	}
	return
}

//Start udp listening
func (ut *UDPTransport) Start() {
	go func() {
		data := make([]byte, 4096)
		defer rpanic.PanicRecover("udptransport Start")
		for {
			conn, err := NewSafeUDPConnection("udp", ut.UAddr)
			if err != nil {
				log.Error(fmt.Sprintf("listen udp %s error %v", ut.UAddr.String(), err))
				time.Sleep(time.Second)
				continue
			}
			log.Info(fmt.Sprintf("udp server listening on %s", ut.UAddr.String()))
			ut.conn = conn
			ut.log.Info(fmt.Sprintf(" listen udp on %s", ut.UAddr))
			for {
				if ut.stopReceiving {
					return
				}
				read, remoteAddr, err := ut.conn.ReadFromUDP(data)
				if err != nil {
					if !ut.stopped {
						ut.log.Error(fmt.Sprintf("udp read data failure! %s", err))
						err = ut.conn.Close()
						break
					} else {
						return
					}

				}
				ut.log.Trace(fmt.Sprintf("receive from %s ,message=%s,hash=%s", remoteAddr,
					encoding.MessageType(data[0]), utils.HPex(utils.Sha3(data[:read]))))
				err = ut.Receive(data[:read])
			}
		}

	}()
	time.Sleep(time.Millisecond)
}

//Receive a message
func (ut *UDPTransport) Receive(data []byte) error {
	//ut.log.Trace(fmt.Sprintf("recevied data\n%s", hex.Dump(data)))
	if ut.stopReceiving {
		return errors.New("stop receive")
	}
	if ut.protocol != nil { //receive data before register a protocol
		ut.protocol.receive(data)
	}
	return nil
}

/*
Send `bytes_` to `host_port`.
Args:
    sender (address): The address of the running node.
    host_port (Tuple[(str, int)]): Tuple with the Host name and Port number.
    bytes_ (bytes): The bytes that are going to be sent through the wire.
*/
func (ut *UDPTransport) Send(receiver common.Address, data []byte) error {
	if ut.stopped {
		return fmt.Errorf("%s closed", ut.name)
	}
	ua, err := ut.getHostPort(receiver)
	if err != nil {
		return err
	}
	ut.log.Trace(fmt.Sprintf("%s send to %s %s:%d, message=%s,response hash=%s", ut.name,
		utils.APex2(receiver), ua.IP, ua.Port, encoding.MessageType(data[0]),
		utils.HPex(utils.Sha3(data, receiver[:]))))
	//ut.log.Trace(fmt.Sprintf("send data  \n%s", hex.Dump(data)))
	//only comment this line,if you want to test.
	//time.Sleep(ut.policy.Consume(1)) //force to wait,
	//todo need one lock for write?
	_, err = ut.conn.WriteToUDP(data, ua)
	return err
}

func (ut *UDPTransport) getHostPort(addr common.Address) (ua *net.UDPAddr, err error) {
	ut.lock.RLock()
	defer ut.lock.RUnlock()
	ua, ok := ut.intranetNodes[addr]
	if ok {
		return
	}
	err = fmt.Errorf("%s host port not found", utils.APex(addr))
	return
}
func (ut *UDPTransport) setHostPort(nodes map[common.Address]*net.UDPAddr) {
	ut.lock.Lock()
	defer ut.lock.Unlock()
	ut.intranetNodes = nodes
}

//RegisterProtocol register receiver
func (ut *UDPTransport) RegisterProtocol(proto ProtocolReceiver) {
	ut.protocol = proto
}

//Stop UDP connection
func (ut *UDPTransport) Stop() {
	ut.stopReceiving = true
	ut.stopped = true
	ut.intranetNodes = make(map[common.Address]*net.UDPAddr)
	if ut.conn != nil {
		err := ut.conn.Close()
		if err != nil {
			log.Warn(fmt.Sprintf("close err %s ", err))
		}
	}
}

//StopAccepting stop receiving
func (ut *UDPTransport) StopAccepting() {
	ut.stopReceiving = true
}

//NodeStatus always mark the node offline
func (ut *UDPTransport) NodeStatus(addr common.Address) (deviceType string, isOnline bool) {
	ut.lock.RLock()
	defer ut.lock.RUnlock()
	if _, ok := ut.intranetNodes[addr]; ok {
		return DeviceTypeOther, true
	}
	return DeviceTypeOther, false
}
