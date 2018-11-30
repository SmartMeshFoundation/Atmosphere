package main

import (
	"github.com/SmartMeshFoundation/Atmosphere/DistributedControlRightManagement/kgcenter"
)

func main() {
	/*x := new(big.Int).Mod(big.NewInt(29), big.NewInt(23))
	fmt.Println(x)*/
	StartMain()
}

func StartMain() {
	kgcenter.LockIn()
	kgcenter.LockOut()
	//初始化一个list

}
