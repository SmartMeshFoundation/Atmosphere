package blockchain

import "github.com/ethereum/go-ethereum/common"

// RPCModuleDependency :
// should provide by rpc module
type RPCModuleDependency interface {
	// GetTokenNetworkAddress get contract address
	GetTokenNetworkAddress() common.Address
	// GetSecretRegistryAddress get contract address
	GetSecretRegistryAddress() common.Address
}
