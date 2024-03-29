package sdk

import "meow.tf/streamdeck/sdk"

type SDK struct {
}

func NewSDK() *SDK {
	return &SDK{}
}

func (s *SDK) ShowAlert(ctxID string) {
	sdk.ShowAlert(ctxID)
}

func (s *SDK) ShowOk(ctxID string) {
	sdk.ShowOk(ctxID)
}

func (s *SDK) SetTitle(ctxID string, title string, target int) {
	sdk.SetTitle(ctxID, title, target)
}

func (s *SDK) SetImage(ctxID string, imageData string, target int) {
	sdk.SetImage(ctxID, imageData, target)
}
