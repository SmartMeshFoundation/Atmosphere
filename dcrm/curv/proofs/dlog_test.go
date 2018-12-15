package proofs

import (
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/SmartMeshFoundation/Atmosphere/log"
	"github.com/SmartMeshFoundation/Atmosphere/utils"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, utils.MyStreamHandler(os.Stdout)))
}

func TestProve(t *testing.T) {
	witness := big.NewInt(30)
	proof := Prove(witness)
	log.Trace(fmt.Sprintf("proof=%s", utils.StringInterface(proof, 7)))
	if !Verify(proof) {
		t.Error("should pass")
	}

}
