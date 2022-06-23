package main

import (
	"os"

	"github.com/yzimhao/utilgo/pack"
)

func main() {
	version := os.Args[1]
	pack.Build("../cmd/", "../dist/trading_engine_{{.OS}}_{{.Arch}}", version)
}
