package models

import (
	"fmt"
	"testing"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var token common.Address

func TestModelDB_NewNonParticipantChannel(t *testing.T) {
	model := setupDb(t)
	defer func() {
		model.CloseDB()
	}()
	token = utils.NewRandomAddress()
	p1 := utils.NewRandomAddress()
	p2 := utils.NewRandomAddress()
	channel := utils.Sha3(p1[:], p2[:], token[:])
	err := model.NewNonParticipantChannel(token, channel, p1, p2)
	if err != nil {
		t.Error(err)
		return
	}
	p3 := utils.NewRandomAddress()
	channel2 := utils.Sha3(p1[:], p3[:], token[:])
	err = model.NewNonParticipantChannel(token, channel2, p1, p3)
	if err != nil {
		t.Error(err)
		return
	}
	edges, err := model.GetAllNonParticipantChannelByToken(token)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, len(edges), 4)
	err = model.RemoveNonParticipantChannel(utils.NewRandomHash())
	assert.EqualValues(t, err != nil, true)
	err = model.RemoveNonParticipantChannel(channel2)
	assert.EqualValues(t, err == nil, true)
	edges, err = model.GetAllNonParticipantChannelByToken(token)
	assert.EqualValues(t, len(edges), 2)
}
func TestReadDbAgain(t *testing.T) {
	TestModelDB_NewNonParticipantChannel(t)
	model, err := OpenDb(dbPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer model.CloseDB()
	edges, err := model.GetAllNonParticipantChannelByToken(token)
	if err != nil {
		t.Error(err)
		return
	}
	log.Trace(fmt.Sprintf("len edges=%d", len(edges)))
}
