package main

import (
	"os"
	"time"

	"github.com/letusgogo/playable-backend/internal/anbox"
	"github.com/letusgogo/playable-backend/internal/api"
	"github.com/letusgogo/playable-backend/internal/game"
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

	// anbox gateway client
	var anboxConfig anbox.AnboxConfig
	err := myApp.Config().UnmarshalKey("anbox", &anboxConfig)
	if err != nil {
		log.Errorf("Failed to unmarshal anbox gateway config: %v", err)
		return err
	}

	anboxClient, err := anbox.NewClient(anboxConfig)
	if err != nil {
		log.Errorf("Failed to create anbox client: %v", err)
		return err
	}

	// game manager
	var gamesList []*game.Game
	err = myApp.Config().UnmarshalKey("games", &gamesList)
	if err != nil {
		log.Errorf("Failed to unmarshal game config: %v", err)
		return err
	}

	gameManager := game.NewManager(gamesList, anboxClient)
	gameManager.Init(c.Context)
	gameManager.Start(c.Context)
	defer func() {
		gameManager.Stop(c.Context)
	}()

	apiService := api.NewApiService(api.ApiServiceConfig{
		Address: address,
	}, gameManager)

	err = apiService.Init()
	if err != nil {
		log.Errorf("Failed to initialize API service: %v", err)
		return err
	}

	err = apiService.Start()
	if err != nil {
		log.Errorf("Failed to start API service: %v", err)
		return err
	} else {
		log.Infof("Starting server on %s", address)
	}

	// Wait for shutdown signal
	app.WaitForSignal(func(s os.Signal) {
		log.Infof("Received signal %v, shutting down HTTP server gracefully", s)
		err := apiService.StopGracefully(1 * time.Second)
		log.Info("API server stopped, error: ", err)
	})

	return nil
}
