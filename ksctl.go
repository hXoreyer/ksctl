package main

import (
	"flag"
	"ksctl/ctl"
)

func main() {
	path := flag.String("f", "./ksctl.yaml", "配置文件")
	flag.Parse()

	yl := ctl.NewYaml(*path)
	yl.Run()
}
