package v1

import (
	"fmt"

	"net/http"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ethereum/go-ethereum/common"
)

/*
Address is api of /api/1/address
*/
func Address(w rest.ResponseWriter, r *rest.Request) {
	data := make(map[string]interface{})
	data["our_address"] = API.Atmosphere.NodeAddress.String()
	err := w.WriteJson(data)
	if err != nil {
		log.Warn(fmt.Sprintf("writejson err %s", err))
	}
}

/*
GetBalanceByTokenAddress : get account's balance and locked account on token
*/
func GetBalanceByTokenAddress(w rest.ResponseWriter, r *rest.Request) {
	tokenAddressStr := r.PathParam("tokenaddress")
	var tokenAddress common.Address
	if tokenAddressStr == "" {
		tokenAddress = utils.EmptyAddress
	} else {
		tokenAddress = common.HexToAddress(tokenAddressStr)
	}
	resp, err := API.GetBalanceByTokenAddress(tokenAddress)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = w.WriteJson(resp)
	if err != nil {
		log.Warn(fmt.Sprintf("writejson err %s", err))
	}
}
