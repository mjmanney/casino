package main

import (
	"casino/libs/fsm"
	"fmt"
)

func main() {
	fmt.Println("game up", fsm.State("Created"))
}
