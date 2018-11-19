package p2p

import (
	"net"
	"github.com/DistributedControlRightManagement/configs"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tmlibs/common"
)

type NetAddress struct {
	ID string
	IP net.IP
	Port uint
	Remark string
}

type SvrListenSocket struct {
	listener net.Listener
	localAddress *NetAddress
	connections chan net.Conn
	common.BaseService
}

func InitListenService() (lsn *SvrListenSocket) {
	var listener net.Listener
	tcpListenAddress := configs.SvrSocketIP + ":" + configs.SvrSocketPort
	listener, err := net.Listen("tcp", tcpListenAddress)
	if err != nil {
		logrus.Fatal("Connot set-up tcp listening,error :", err)
	}
	logrus.Info("Start tcp listen on ", tcpListenAddress)
	lsn = &SvrListenSocket{
		listener:     listener,
		localAddress: nil,
		connections:  make(chan net.Conn, 4),
	}
	lsn.BaseService = *common.NewBaseService(nil, "SvrListenSocket", lsn)
	err = lsn.Start()
	if err != nil {
		logrus.Fatal("Connot start tcp listening ,error :", err)
	}
	return
}

func (sls *SvrListenSocket) OnStart() error {
	 err:=sls.BaseService.OnStart()
	 if err!=nil{
	 	return err
	 }
	 go sls.AsyncCallback()
	 return nil
}

func (sls *SvrListenSocket) OnStop() {
	sls.BaseService.OnStop()
	sls.listener.Close()
}

//只接受认可的机器来互相通信
func (sls *SvrListenSocket)AsyncCallback() {
	for {
		if sls.IsRunning() {
			break
		}
		endPointConn, err := sls.listener.Accept()
		if err != nil {
			panic(err)
		}
		//var legalEndPoint=[]string{commons.Ip1,commons.Ip2,commons.Ip3,commons.Ip4}
		sls.connections <- endPointConn
	}
	close(sls.connections)
}