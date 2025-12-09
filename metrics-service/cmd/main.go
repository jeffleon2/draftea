package main

import (
	"fmt"
	"os"

	"github.com/jeffleon2/draftea-metric-service/config"
	"github.com/jeffleon2/draftea-metric-service/internal/app"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		fmt.Println("Error reading config file", err)
		os.Exit(1)
	}
	myApp := &app.App{}
	myApp.Initialize(cfg)
	myApp.Run()
}
