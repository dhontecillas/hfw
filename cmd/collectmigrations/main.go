package main

import (
	"flag"
	"fmt"

	"github.com/dhontecillas/hfw/pkg/bundler"
	"github.com/dhontecillas/hfw/pkg/obs/logs"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		fmt.Printf("-> %#v <-\n", args)
		return
	}
	targetDir := args[0]
	scanDirs := args[1:]

	l := logs.NewLogrus(nil)
	err := bundler.UpdateMigrations(targetDir, scanDirs, l)
	if err != nil {
		print("cannot collect migrations: %s\n", err.Error())
	}
}
