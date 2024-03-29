package main

import (
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"meow.tf/streamdeck/sdk"

	"github.com/ft-t/apimonkey/pkg/instance"
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

	manager := instance.NewManager()

	lg = zerolog.New(zerolog.MultiLevelWriter(os.Stdout, logFile)).With().Timestamp().Logger()

	sdk.AddHandler(func(event *sdk.WillAppearEvent) {
		if event.Payload == nil {
			return
		}

		instanceRef, err := manager.InitInstance(event.Context)
		if err != nil {
			lg.Err(err).Send()
			return
		}

		if err = instanceRef.SetConfig(event.Payload.Get("settings")); err != nil {
			lg.Err(err).Send()
			return
		}

		if err = manager.StartAsync(event.Context); err != nil {
			lg.Err(err).Send()
			return
		}
	})
	//
	//sdk.AddHandler(func(event *sdk.WillDisappearEvent) {
	//	if event.Payload == nil {
	//		return
	//	}
	//
	//	mut.Lock()
	//	defer mut.Unlock()
	//	instance, ok := instances[event.Context]
	//	if !ok {
	//		return
	//	}
	//
	//	instance.StartAsync()
	//})
	//
	//sdk.AddHandler(func(event *sdk.ReceiveSettingsEvent) {
	//	setSettingsFromPayload(event.Settings, event.Context, instances[event.Context])
	//})
	//
	//sdk.AddHandler(func(event *sdk.KeyDownEvent) {
	//	instance, ok := instances[event.Context]
	//	if !ok {
	//		lg.Warn().Msgf("instance %v not found", event.Context)
	//	}
	//
	//	instance.KeyPressed()
	//})

	err := sdk.Open()
	if err != nil {
		lg.Panic().Err(err).Send()
	}
	sdk.Wait()
}
