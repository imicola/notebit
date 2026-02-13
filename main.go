package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"notebit/pkg/logger"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize Logger
	err := logger.Initialize(logger.Config{
		Level:         logger.DEBUG,
		LogDir:        "logs",
		FileName:      "notebit.log",
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		MaxBackups:    5,
		ConsoleOutput: true,
	})
	if err != nil {
		println("Failed to initialize logger:", err.Error())
	}
	defer logger.GetDefault().Close()

	logger.Info("Starting Notebit application...")

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "notebit",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
