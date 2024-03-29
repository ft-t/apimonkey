package sdk_test

import (
	"testing"

	sdk2 "github.com/ft-t/apimonkey/pkg/sdk"
)

func TestSdk(t *testing.T) { // just coverage
	sdk := sdk2.NewSDK()

	defer func() {
		_ = recover()
	}()

	sdk.ShowAlert("ctxID")
	sdk.ShowOk("ctxID")
	sdk.SetTitle("ctxID", "title", 0)
	sdk.SetImage("ctxID", "imageData", 0)
}
