package v1

import (
	"net/http"

	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere"
	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/params"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
	"github.com/ant0ine/go-json-rest/rest"
)

/*
API is the interface of Atmosphere network
should be set before start restful server
*/
var API *atmosphere.API

/*
Config is the configuration of Atmosphere network
should be set before start restful server
*/
var Config *params.Config

/*
Start the restful server
*/
func Start() {

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(

		/*
			prepare update
		*/
		rest.Post("/api/1/prepare-update", PrepareUpdate),
		/*
			transfers
		*/
		rest.Get("/api/1/querysenttransfer", GetSentTransfers),
		rest.Get("/api/1/queryreceivedtransfer", GetReceivedTransfers),
		rest.Post("/api/1/transfers/:token/:target", Transfers),
		rest.Get("/api/1/transferstatus/:token/:locksecrethash", GetTransferStatus),
		rest.Post("/api/1/transfercancel/:token/:locksecrethash", CancelTransfer),
		/*
			transfer with specified secret
		*/
		rest.Post("/api/1/transfers/allowrevealsecret", AllowRevealSecret),
		rest.Get("/api/1/getunfinishedreceivedtransfer/:tokenaddress/:locksecrethash", GetUnfinishedReceivedTransfer),
		rest.Post("/api/1/registersecret", RegisterSecret),
		/*
			token swap
		*/
		rest.Put("/api/1/token_swaps/:target/:locksecrethash", TokenSwap),
		/*
			accounts
		*/
		rest.Get("/api/1/address", Address),
		rest.Get("/api/1/balance", GetBalanceByTokenAddress),
		rest.Get("/api/1/balance/", GetBalanceByTokenAddress),
		rest.Get("/api/1/balance/:tokenaddress", GetBalanceByTokenAddress),
		/*
			channels
		*/
		rest.Get("/api/1/channels/:channel", SpecifiedChannel),
		rest.Get("/api/1/channels", GetChannelList),
		rest.Put("/api/1/channels", OpenChannel),
		rest.Patch("/api/1/channels/:channel", CloseSettleDepositChannel),
		rest.Get("/api/1/thirdparty/:channel/:3rd", ChannelFor3rdParty),
		rest.Get("/api/1/pfs/:channel", BalanceUpdateForPFS),
		/*
			tokens
		*/
		rest.Get("/api/1/tokens", Tokens),
		rest.Get("/api/1/tokens/:token/partners", TokenPartners),
		/*
			utils
		*/
		rest.Get("/api/1/path/:target_address/:token/:amount", FindPath),
		rest.Get("/api/1/secret", GetRandomSecret), // api to provide random secret and lockSecretHash pair
		rest.Get("/api/1/fee_policy", GetFeePolicy),
		rest.Post("/api/1/fee_policy", SetFeePolicy),
		rest.Get("/api/1/fee", GetAllFeeChargeRecord),

		/*
			test
		*/
		rest.Get("/api/1/stop", Stop),
		rest.Get("/api/1/switch/:mesh", SwitchNetwork),
		rest.Post("/api/1/updatenodes", UpdateMeshNetworkNodes),

		/*
			others TODO
		*/
		/*
			1. withdraw
			{ "amount":3333,}
			2. prepare for withdraw:
			{"op":"preparewithdraw",}
			3. cancel prepare:
			{"op": "cancelprepare"}
		*/
		rest.Put("/api/1/withdraw/:channel", withdraw),
		/*
			1. prepare for withdraw:
			{"op":"preparesettle",}
			3. cancel prepare:
			{"op": "cancelprepare"}
		*/
		rest.Put("/api/1/settle/:channel", nil),
		/*
			events
		*/
		//rest.Get("/api/1/events/network", EventNetwork),
		//rest.Get("/api/1/events/tokens/:token", EventTokens),
		//rest.Get("/api/1/events/channels/:channel", EventChannels),
		/*
			for debug only
		*/
		rest.Get("/api/1/debug/balance/:token/:addr", Balance),
		rest.Get("/api/1/debug/transfer/:token/:addr/:value", TransferToken),
		rest.Get("/api/1/debug/ethbalance/:addr", EthBalance),
		rest.Get("/api/1/debug/ethstatus", EthereumStatus),
		rest.Get("/api/1/debug/force-unlock/:channel/:locksecrethash/:secrethash", ForceUnlock),
		rest.Post("/api/1/debug/notify_network_down", NotifyNetworkDown), // notify atmosphere network down
		rest.Get("/api/1/debug/shutdown", func(writer rest.ResponseWriter, request *rest.Request) {
			API.Atmosphere.Stop()
			utils.SystemExit(0)
		}),
	)
	if err != nil {
		log.Crit(fmt.Sprintf("maker router :%s", err))
	}
	api.SetApp(router)
	listen := fmt.Sprintf("%s:%d", Config.APIHost, Config.APIPort)
	log.Crit(fmt.Sprintf("http listen and serve :%s", http.ListenAndServe(listen, api.MakeHandler())))
}

/*
Stop for app user, call this api before quit.
*/
func Stop(w rest.ResponseWriter, r *rest.Request) {
	//test only
	API.Stop()
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.(http.ResponseWriter).Write([]byte("ok"))
	if err != nil {
		log.Warn(fmt.Sprintf("writejson err %s", err))
	}
}
