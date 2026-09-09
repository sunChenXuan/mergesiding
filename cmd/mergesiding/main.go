package main

import (
	"os"

	"github.com/sunChenXuan/mergesiding/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:]))
}
