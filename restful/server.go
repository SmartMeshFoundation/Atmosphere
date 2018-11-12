package restful

import (
	photon "github.com/SmartMeshFoundation/Atmosphere"
	"github.com/SmartMeshFoundation/Atmosphere/params"
	"github.com/SmartMeshFoundation/Atmosphere/restful/v1"
)

func init() {

}

/*
Start restful server
PhotonAPI is the interface of atmosphere network
config is the configuration of atmosphere network
*/
func Start(API *photon.API, config *params.Config) {
	v1.API = API
	v1.Config = config
	v1.Start()
}
