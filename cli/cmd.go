package main

import (
	"os"

	"github.com/yzimhao/utilgo/pack"
)

func main() {
	new_tag(os.Args[1])
}

func new_tag(version string) {
	pack.RunCommand("git", "tag", version)
	pack.RunCommand("git", "push", "tag:"+version)
}

func releases(version string) {

}
