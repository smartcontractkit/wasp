package main

import (
	"github.com/smartcontractkit/wasp"
)

func main() {
	if _, err := wasp.NewDashboard().Deploy(nil); err != nil {
		panic(err)
	}
}
