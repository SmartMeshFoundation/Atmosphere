package blockchain

import "time"

// DefaultEthRPCPollPeriodForTest :
var DefaultEthRPCPollPeriodForTest = 500 * time.Millisecond

// DefaultEthRPCPollPeriod :
var DefaultEthRPCPollPeriod = 7500 * time.Millisecond

// DefaultEthRPCTimeout :
var DefaultEthRPCTimeout = 3 * time.Second

// DefaultEnableForkConfirm :
var DefaultEnableForkConfirm = false

// DefaultForkConfirmNumber : 分叉确认块数量,BlockNumber < 最新块-ForkConfirmNumber的事件被认为无分叉的风险
var DefaultForkConfirmNumber int64 = 17

type config struct {
	RPCPollPeriod     time.Duration // 轮询周期
	RPCTimeOut        time.Duration // 事件轮询调用超时时间
	ForkConfirmNumber int64         // 分叉确认块数量,BlockNumber < 最新块-ForkConfirmNumber的事件被认为无分叉的风险
	EnableForkConfirm bool          // 事件延迟上报开关
	LogPeriod         int64         // new block number 日志间隔
}

// newDefaultConfig :
func newDefaultConfig() *config {
	return &config{
		RPCPollPeriod:     DefaultEthRPCPollPeriod,
		RPCTimeOut:        DefaultEthRPCTimeout,
		ForkConfirmNumber: DefaultForkConfirmNumber,
		EnableForkConfirm: DefaultEnableForkConfirm,
		LogPeriod:         1,
	}
}
