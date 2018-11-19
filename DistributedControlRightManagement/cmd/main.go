package main

import (
	"github.com/DistributedControlRightManagement/kgcenter"
	"math/big"
	"fmt"
)

func main() {
	x:=new(big.Int).Mod(big.NewInt(29),big.NewInt(23))
	fmt.Println(x)
	StartMain()
}

func StartMain() {
	kgcenter.LockIn()
	kgcenter.LockOut()
}