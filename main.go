package main

import "paklari/internal/core"

func main() {
	core.NewExec().ReadEntries("example.json").Run()
}
