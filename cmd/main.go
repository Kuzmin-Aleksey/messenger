package main

import (
	"messanger/app"
	"os"
)

var config = "config/config.yaml"

func main() {
	if len(os.Args) > 1 {
		config = os.Args[1]
	}
	app.Run(config)
}
