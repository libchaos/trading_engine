package main

import (
	"os"

	"github.com/yzimhao/utilgo/pack"
)

func main() {
	version := ""
	if len(os.Args) > 1 {
		version = os.Args[1]
	}
	new_tag(version)
}

func new_tag(version string) {
	pack.NewTag(version)
}

func releases(version string) {

}
