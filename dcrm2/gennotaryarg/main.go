package main

import (
	"crypto/rand"

	"flag"

	"encoding/json"
	"fmt"

	"io/ioutil"
	"os"

	"github.com/SmartMeshFoundation/Atmosphere/DistributedControlRightManagement/configs"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/commitments"
	"github.com/SmartMeshFoundation/Atmosphere/dcrm/zkp"
)

func main() {
	thresholdNum := flag.Int("num", 7, "threshold number")
	file := flag.String("file", "", "save arg to file")
	flag.Parse()
	var PaillierPrivateKey, _ = zkp.GenerateKey(rand.Reader, 1023)
	var zkPublicParams = zkp.GenerateParams(configs.G, 256, 512, &PaillierPrivateKey.PublicKey)
	var masterPK = commitments.GenerateNMMasterPublicKey()
	arg := &dcrm.NotaryShareArg{
		PaillierPrivateKey: PaillierPrivateKey,
		ZkPublicParams:     zkPublicParams,
		MasterPK:           masterPK,
		ThresholdNum:       *thresholdNum,
	}
	data, err := json.MarshalIndent(arg, "", "\t")
	if err != nil {
		panic(err)
	}
	if len(*file) == 0 {
		fmt.Printf("%s", string(data))
	} else {
		ioutil.WriteFile(*file, data, os.ModePerm)
	}

}
