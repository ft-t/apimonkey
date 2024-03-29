package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"meow.tf/streamdeck/sdk"

	"github.com/ft-t/apimonkey/pkg/executor"
	"github.com/ft-t/apimonkey/pkg/instance"
	"github.com/ft-t/apimonkey/pkg/scripts"
	sdk2 "github.com/ft-t/apimonkey/pkg/sdk"
)

var lg zerolog.Logger

func recoverPanic() {
	if rec := recover(); rec != nil {
		switch v := rec.(type) {
		case error:
			lg.Err(v).Send()
		default:
			lg.Error().Msgf("%v", v)
		}
	}
}

func main() {
	logFile := &lumberjack.Logger{
		Filename:   "logs/log.log",
		MaxSize:    30,
		MaxBackups: 3,
		MaxAge:     10,
		Compress:   false,
	}

	lg = zerolog.New(zerolog.MultiLevelWriter(os.Stdout, logFile)).With().Timestamp().Logger()
	log.Logger = lg

	manager := instance.NewManager(
		instance.NewDefaultFactory(
			sdk2.NewSDK(),
			executor.NewExecutor(
				scripts.NewLua(),
			),
		),
	)
	
	defer recoverPanic()

	sdk.AddHandler(func(event *sdk.WillAppearEvent) {
		defer recoverPanic()

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

	lg.Info().Msgf("Starting StreamDeck plugin. args %v", os.Args)
	err := sdk.Open()
	if err != nil {
		lg.Panic().Err(err).Send()
	}

	sdk.Wait()
}
