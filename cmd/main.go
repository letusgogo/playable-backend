package main

import (
	"os"
	"time"

	"github.com/letusgogo/playable-backend/internal"
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
		app.WithEnvPrefix("APP"), // APP_SERVER_ADDRESS → server.address
		app.WithEnvBindings(map[string]string{
			"custom.api.key": "MY_CUSTOM_API_KEY", // Custom mapping example
		}),
	)

	if err := myApp.Start(); err != nil {
		logrus.Fatal(err)
	}
}

func runServer(c *cli.Context, myApp *app.App) error {
	log := logger.GetLogger("server")
	address := myApp.Config().GetString("server.address")

	log.Infof("Starting server on %s", address)

	apiService := internal.NewApiService(internal.ApiServiceConfig{
		Address: address,
	})

	err := apiService.Init()
	if err != nil {
		log.Errorf("Failed to initialize API service: %v", err)
		return err
	}

	err = apiService.Start()
	if err != nil {
		log.Errorf("Failed to start API service: %v", err)
		return err
	}

	// Wait for shutdown signal
	app.WaitForSignal(func(s os.Signal) {
		log.Infof("Received signal %v, shutting down HTTP server gracefully", s)
		err := apiService.StopGracefully(1 * time.Second)
		log.Info("API server stopped, error: ", err)
	})
	return nil
}
