package main

import (
	"embed"

	"notebit/pkg/logger"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	hideConsoleWindow()

	// Initialize Logger
	err := logger.Initialize(logger.LoadConfigFromEnv(logger.Config{
		Level:         logger.INFO,
		LogDir:        "logs",
		FileName:      "notebit.log",
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		MaxBackups:    15,                // 15 days
		ConsoleOutput: true,
		ConsoleColor:  true,
		BatchSize:     10,
		FlushInterval: 100, // 100ms
	}))
	if err != nil {
		logger.Fatal("Failed to initialize logger: %v", err)
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
		logger.Fatal("Error starting application: %v", err)
	}
}
