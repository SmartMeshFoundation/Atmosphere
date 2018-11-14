package params

import (
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"time"

	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
)

/*
	--------------------------------> default param for test <--------------------------------
*/

//DefaultTestXMPPServer xmpp server for test only
const DefaultTestXMPPServer = "193.112.248.133:5222" //"182.254.155.208:5222"

//SpectrumTestNetRegistryAddress Registry contract address
var SpectrumTestNetRegistryAddress = common.HexToAddress("0x52d7167FAD53835a2356C7A872BfbC17C03aD758")

//TestLogServer only for test, enabled if --debug flag is set
const TestLogServer = "http://transport01.smartmesh.cn:8008"

// TestPrivateChainID :
var TestPrivateChainID int64 = 8888

// DefaultEthRPCPollPeriodForTest :
var DefaultEthRPCPollPeriodForTest = 500 * time.Millisecond

/*
	--------------------------------> default param of block chain <--------------------------------
*/

//DefaultChainID :
var DefaultChainID = big.NewInt(0)

//DefaultGasLimit max gas usage for atmosphere tx
const DefaultGasLimit = 3141592 //den's gasLimit.

//DefaultGasPrice from ethereum
const DefaultGasPrice = params.Shannon * 20

//DefaultTxTimeout args
const DefaultTxTimeout = 5 * time.Minute //15seconds for one block,it may take sever minutes

//DefaultMaxRequestTimeout args
const DefaultMaxRequestTimeout = 20 * time.Minute //longest time for a request ,for example ,settle all channles?

//ContractSignaturePrefix for EIP191 https://github.com/ethereum/EIPs/blob/master/EIPS/eip-191.md
var ContractSignaturePrefix = []byte("\x19Spectrum Signed Message:\n")

const (
	//ContractBalanceProofMessageLength balance proof  length
	ContractBalanceProofMessageLength = "176"
	//ContractBalanceProofDelegateMessageLength update balance proof delegate length
	ContractBalanceProofDelegateMessageLength = "144"
	//ContractCooperativeSettleMessageLength cooperative settle channel proof length
	ContractCooperativeSettleMessageLength = "176"
	//ContractDisposedProofMessageLength annouce disposed proof length
	ContractDisposedProofMessageLength = "136"
	//ContractWithdrawProofMessageLength withdraw proof length
	ContractWithdrawProofMessageLength = "156"
	//ContractUnlockDelegateProofMessageLength unlock delegate proof length
	ContractUnlockDelegateProofMessageLength = "188"
)

//GenesisBlockHashToDefaultRegistryAddress :
var GenesisBlockHashToDefaultRegistryAddress = map[common.Hash]common.Address{
	// spectrum
	common.HexToHash("0x57e682b80257aad73c4f3ad98d20435b4e1644d8762ef1ea1ff2806c27a5fa3d"): common.HexToAddress("0x28233F8e0f8Bd049382077c6eC78bE9c2915c7D4"),
	// spectrum test net
	common.HexToHash("0xd011e2cc7f241996a074e2c48307df3971f5f1fe9e1f00cfa704791465d5efc3"): common.HexToAddress("0xa2150A4647908ab8D0135F1c4BFBB723495e8d12"),
	// ethereum
	common.HexToHash("0x88e96d4537bea4d9c05d12549907b32561d3bf31f45aae734cdc119f13406cb6"): utils.EmptyAddress,
	// ethereum test net
	common.HexToHash("0x41800b5c3f1717687d85fc9018faac0a6e90b39deaa0b99e7fe4fe796ddeb26a"): utils.EmptyAddress,
	// ethereum private
	common.HexToHash("0x38a88a9ddffe522df5c07585a7953f8c011c94327a494188bd0cc2410dc40a1a"): common.HexToAddress("0xc2e18A2F6CbF84C1cACf2186Fee6b99405027c07"),
}

/*
	--------------------------------> default param of protocol <--------------------------------
*/

//defaultProtocolRetiesBeforeBackoff
const defaultProtocolRetiesBeforeBackoff = 5
const defaultProtocolRhrottleCapacity = 10.
const defaultProtocolThrottleFillRate = 10.
const defaultprotocolRetryInterval = 1.

//UDPMaxMessageSize message size
const UDPMaxMessageSize = 1200

//DefaultXMPPServer xmpp server
const DefaultXMPPServer = "193.112.248.133:5222"

//DefaultMatrixServerConfig matrix server config
var DefaultMatrixServerConfig = map[string]string{
	"transport01.smartmesh.cn": "http://transport01.smartmesh.cn:8008",
	"transport02.smartmesh.cn": "http://transport02.smartmesh.cn:8008",
	"transport03.smartmesh.cn": "http://transport03.smartmesh.cn:8008",
}

//DefaultMatrixAliasFragment  is discovery DefaultMatrixAliasFragment
const DefaultMatrixAliasFragment = "discovery"

//DefaultMatrixDiscoveryServer is discovery server
const DefaultMatrixDiscoveryServer = "transport01.smartmesh.cn"

//DefaultMatrixNetworkName Specify the network name of the Ethereum network to run Atmosphere on
const DefaultMatrixNetworkName = "ropsten"

/*
	--------------------------------> default param of atmosphere <--------------------------------
*/

//DefaultUDPListenPort listening port for communication bewtween nodes
const DefaultUDPListenPort = 40001

//DefaultRevealTimeout blocks needs to update transfer
const DefaultRevealTimeout = 10

//DefaultSettleTimeout settle time of channel
const DefaultSettleTimeout = 600

//DefaultPollTimeout  request wait time ???
const DefaultPollTimeout = 180 * time.Second

//DefaultChannelSettleTimeoutMin min settle timeout
const DefaultChannelSettleTimeoutMin = 6

/*
DefaultChannelSettleTimeoutMax The maximum settle timeout is chosen as something above
 1 year with the assumption of very fast block times of 12 seconds.
 There is a maximum to avoidpotential overflows as described here:
 https://github.com/Photon/photon/issues/1038
*/
const DefaultChannelSettleTimeoutMax = 2700000

//DefaultDataDir default work directory
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "atmosphere")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "atmosphere")
		} else {
			return filepath.Join(home, ".atmosphere")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

//DefaultKeyStoreDir keystore path of ethereum
func DefaultKeyStoreDir() string {
	return filepath.Join(node.DefaultDataDir(), "keystore")
}
