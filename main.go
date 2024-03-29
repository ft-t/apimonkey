package main

import (
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"meow.tf/streamdeck/sdk"

	"github.com/ft-t/apimonkey/pkg/executor"
	"github.com/ft-t/apimonkey/pkg/instance"
	"github.com/ft-t/apimonkey/pkg/scripts"
	sdk2 "github.com/ft-t/apimonkey/pkg/sdk"
)

var lg zerolog.Logger

func main() {
	logFile := &lumberjack.Logger{
		Filename:   "logs/log.log",
		MaxSize:    30,
		MaxBackups: 3,
		MaxAge:     10,
		Compress:   false,
	}

	manager := instance.NewManager(
		instance.NewDefaultFactory(
			sdk2.NewSDK(),
			executor.NewExecutor(
				scripts.NewLua(),
			),
		),
	)

	lg = zerolog.New(zerolog.MultiLevelWriter(os.Stdout, logFile)).With().Timestamp().Logger()

	sdk.AddHandler(func(event *sdk.WillAppearEvent) {
		if event.Payload == nil {
			return
		}

		_, err := manager.InitInstance(event.Context)
		if err != nil {
			lg.Err(err).Send()
			return
		}

		if err = manager.SetInstanceConfig(event.Context, event.Payload.Get("settings")); err != nil {
			lg.Err(err).Send()
			return
		}

		if err = manager.StartAsync(event.Context); err != nil {
			lg.Err(err).Send()
			return
		}
	})
	sdk.AddHandler(func(event *sdk.WillDisappearEvent) {
		if event.Payload == nil {
			return
		}

		if err := manager.Stop(event.Context); err != nil {
			lg.Err(err).Send()
		}
	})

	sdk.AddHandler(func(event *sdk.ReceiveSettingsEvent) {
		if err := manager.SetInstanceConfig(event.Context, event.Settings); err != nil {
			lg.Err(err).Send()
			return
		}
	})

	sdk.AddHandler(func(event *sdk.KeyDownEvent) {
		if err := manager.KeyPressed(event.Context); err != nil {
			lg.Err(err).Send()
			return
		}
	})

	err := sdk.Open()
	if err != nil {
		lg.Panic().Err(err).Send()
	}

	sdk.Wait()
}
