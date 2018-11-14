package params

/*
MobileMode works on mobile device, 移动设备模式,这时候 atmosphere 并不是一个独立的进程,这时候很多工作模式要发生变化.
比如:
1.不能任意退出
2. 对于网络通信的处理要更谨慎
3. 对于资源的消耗如何控制?
*/
/*
 *	MobileMode : a boolean value to adapt with mobile modes.
 *
 *	Note : if true, then atmosphere is not an individual process, work mode is about to change.
 *		1. not support exit arbitrarily.
 *		2. handle internet communication more prudent.
 *		3. How to control amount of resource consumption.
 */
var MobileMode bool

//RevealTimeout blocks needs to update transfer
var RevealTimeout = DefaultRevealTimeout

//ChainID of this tokenNetwork
var ChainID = DefaultChainID

//MatrixServerConfig matrix server config
var MatrixServerConfig = DefaultMatrixServerConfig

// ContractVersionPrefix :
var ContractVersionPrefix = "0.5"

// ForkConfirmNumber : 分叉确认块数量,BlockNumber < 最新块-ForkConfirmNumber的事件被认为无分叉的风险
var ForkConfirmNumber int64 = 17
