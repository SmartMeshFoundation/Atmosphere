package main

import (
	"fmt"

	"github.com/SmartMeshFoundation/Atmosphere/cmd/atmosphere/mainimpl"
)

func main() {
	if _, err := mainimpl.StartMain(); err != nil {
		fmt.Printf("quit with err %s\n", err)
	}
}
