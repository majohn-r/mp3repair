package main

import (
	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/boot"
)

func main() {
	goyek.SetDefault(build)
	boot.Main()
}
