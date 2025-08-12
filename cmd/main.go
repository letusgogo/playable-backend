package main

import (
	"os"
	"time"

	"github.com/letusgogo/quick/app"
	"github.com/letusgogo/quick/logger"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	myApp := app.NewApp("playable", "Playable backend")
	myApp.SetVersion("1.0.0")

	myApp.Init(
		app.WithCommands([]*cli.Command{
			{
				Name:  "serve",
				Usage: "Start the server",
				Action: func(c *cli.Context) error {
					return runServer(c, myApp)
				},
			},
		}),
		// Set environment variable prefix
		app.WithEnvPrefix("APP"), // APP_SERVER_PORT → server.port
	)

	if err := myApp.Start(); err != nil {
		logrus.Fatal(err)
	}
}

func runServer(c *cli.Context, myApp *app.App) error {
	log := logger.GetLogger("server")
	log.Info("Starting server")
	// Wait for shutdown signal
	app.WaitForSignal(func(s os.Signal) {
		log.Infof("Received signal %v, shutting down HTTP server gracefully", s)
		// Here you would normally stop your HTTP server
		time.Sleep(1 * time.Second) // Simulate graceful shutdown
		log.Info("HTTP server stopped")
	})
	return nil
}
