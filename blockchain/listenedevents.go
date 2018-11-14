package blockchain

import (
	"github.com/SmartMeshFoundation/Atmosphere/contracts"
	"github.com/ethereum/go-ethereum/core/types"
)

func newEventChannelOpenAndDeposit(el *types.Log) (event *contracts.TokenNetworkChannelOpenedAndDeposit, err error) {
	event = &contracts.TokenNetworkChannelOpenedAndDeposit{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelOpenedAndDeposit, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventChannelNewDeposit(el *types.Log) (event *contracts.TokenNetworkChannelNewDeposit, err error) {
	event = &contracts.TokenNetworkChannelNewDeposit{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelNewDeposit, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}
func newEventChannelClosed(el *types.Log) (event *contracts.TokenNetworkChannelClosed, err error) {
	event = &contracts.TokenNetworkChannelClosed{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelClosed, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventChannelUnlocked(el *types.Log) (event *contracts.TokenNetworkChannelUnlocked, err error) {
	event = &contracts.TokenNetworkChannelUnlocked{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelUnlocked, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventBalanceProofUpdated(el *types.Log) (event *contracts.TokenNetworkBalanceProofUpdated, err error) {
	event = &contracts.TokenNetworkBalanceProofUpdated{}
	err = UnpackLog(&tokenNetworkAbi, event, nameBalanceProofUpdated, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventChannelPunished(el *types.Log) (event *contracts.TokenNetworkChannelPunished, err error) {
	event = &contracts.TokenNetworkChannelPunished{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelPunished, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventChannelSettled(el *types.Log) (event *contracts.TokenNetworkChannelSettled, err error) {
	event = &contracts.TokenNetworkChannelSettled{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelSettled, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventChannelCooperativeSettled(el *types.Log) (event *contracts.TokenNetworkChannelCooperativeSettled, err error) {
	event = &contracts.TokenNetworkChannelCooperativeSettled{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelCooperativeSettled, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventChannelWithdraw(el *types.Log) (event *contracts.TokenNetworkChannelWithdraw, err error) {
	event = &contracts.TokenNetworkChannelWithdraw{}
	err = UnpackLog(&tokenNetworkAbi, event, nameChannelWithdraw, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}

func newEventSecretRevealed(el *types.Log) (event *contracts.SecretRegistrySecretRevealed, err error) {
	event = &contracts.SecretRegistrySecretRevealed{}
	err = UnpackLog(&secretRegistryAbi, event, nameSecretRevealed, el)
	if err != nil {
		return
	}
	event.Raw = *el
	return
}
