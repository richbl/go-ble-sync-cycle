package video

import (
	"context"
	"testing"
	"time"

	config "ble-sync-cycle/internal/configuration"
	speed "ble-sync-cycle/internal/speed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMpv struct {
	mock.Mock
}

func (m *MockMpv) Initialize() error {
	return m.Called().Error(0)
}

func (m *MockMpv) TerminateDestroy() {
	m.Called()
}

func (m *MockMpv) SetOption(option string, format int, value interface{}) error {
	return m.Called(option, format, value).Error(0)
}

func (m *MockMpv) Command(cmd []string) error {
	return m.Called(cmd).Error(0)
}

func (m *MockMpv) SetProperty(property string, format int, value interface{}) error {
	return m.Called(property, format, value).Error(0)
}

func (m *MockMpv) GetProperty(property string, format int) (interface{}, error) {
	args := m.Called(property, format)
	return args.Get(0), args.Error(1)
}

func TestPlaybackController_Start(t *testing.T) {
	mockMpv := new(MockMpv)

	mockMpv.On("Initialize").Return(nil)
	mockMpv.On("TerminateDestroy").Return()
	mockMpv.On("SetOption", "window-scale", 0, 0.5).Return(nil)
	mockMpv.On("Command", []string{"loadfile", "test_video.mp4"}).Return(nil)
	mockMpv.On("SetProperty", "pause", 0, false).Return(nil)
	mockMpv.On("GetProperty", "pause", 0).Return(false, nil)
	mockMpv.On("SetProperty", "speed", 0, mock.Anything).Return(nil)

	videoConfig := config.VideoConfig{
		FilePath:          "test_video.mp4",
		UpdateIntervalSec: 1,
		SpeedMultiplier:   2.0,
	}

	speedConfig := config.SpeedConfig{
		SpeedThreshold: 0.5,
	}

	controller, _ := NewPlaybackController(videoConfig, speedConfig)
	speedController := speed.NewSpeedController(3)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		speedController.UpdateSpeed(5.0)
		time.Sleep(time.Second)
		speedController.UpdateSpeed(6.0)
	}()

	err := controller.Start(ctx, speedController)
	assert.NoError(t, err)

	mockMpv.AssertExpectations(t)
}
