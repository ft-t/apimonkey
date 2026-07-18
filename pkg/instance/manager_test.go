package instance_test

import (
	"context"
	"errors"
	"testing"

	"github.com/valyala/fastjson"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ft-t/apimonkey/pkg/instance"
)

func TestNewManager(t *testing.T) {
	factory := NewMockFactory(gomock.NewController(t))
	mgr := instance.NewManager(factory)

	ctxID := "1231231"
	mockInstance := NewMockInstance(gomock.NewController(t))
	factory.EXPECT().Create(ctxID).Return(mockInstance)

	ctx := context.Background()
	mockInstance.EXPECT().StartAsync(ctx)
	mockInstance.EXPECT().Stop()
	mockInstance.EXPECT().KeyPressed(ctx)

	_, err := mgr.InitInstance(ctxID)
	assert.Nil(t, err)

	assert.NoError(t, mgr.StartAsync(ctx, ctxID))
	assert.NoError(t, mgr.KeyPressed(ctx, ctxID))

	js := &fastjson.Value{}
	mockInstance.EXPECT().SetConfig(js).Return(nil)

	assert.NoError(t, mgr.SetInstanceConfig(ctxID, js))
	assert.NoError(t, mgr.Stop(ctxID))
}

func TestNotExist(t *testing.T) {
	factory := NewMockFactory(gomock.NewController(t))
	mgr := instance.NewManager(factory)

	ctxID := "1231231"
	ctx := context.Background()

	assert.ErrorContains(t, mgr.StartAsync(ctx, ctxID), "instance not found")
	assert.ErrorContains(t, mgr.KeyPressed(ctx, ctxID), "instance not found")
	assert.ErrorContains(t, mgr.SetInstanceConfig(ctxID, nil), "instance not found")
	assert.NoError(t, mgr.Stop(ctxID))
}

func TestFailConfig(t *testing.T) {
	factory := NewMockFactory(gomock.NewController(t))
	mgr := instance.NewManager(factory)

	ctxID := "1231231"
	mockInstance := NewMockInstance(gomock.NewController(t))
	factory.EXPECT().Create(ctxID).Return(mockInstance)

	_, err := mgr.InitInstance(ctxID)
	assert.NoError(t, err)

	mockInstance.EXPECT().SetConfig(gomock.Any()).Return(errors.New("cfg error"))

	assert.ErrorContains(t, mgr.SetInstanceConfig(ctxID, nil), "cfg error")
}

func TestFailKeyPressed(t *testing.T) {
	factory := NewMockFactory(gomock.NewController(t))
	mgr := instance.NewManager(factory)

	ctxID := "1231231"
	mockInstance := NewMockInstance(gomock.NewController(t))
	factory.EXPECT().Create(ctxID).Return(mockInstance)

	_, err := mgr.InitInstance(ctxID)
	assert.NoError(t, err)

	ctx := context.Background()
	mockInstance.EXPECT().KeyPressed(ctx).Return(errors.New("key error"))

	assert.ErrorContains(t, mgr.KeyPressed(ctx, ctxID), "key error")
}
