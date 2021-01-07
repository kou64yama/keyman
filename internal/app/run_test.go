package app_test

import (
	"keyman/internal/app"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRun(t *testing.T) {
	cmd := new(CommandMock)
	cmd.On("Name").Return("test")
	cmd.On("Setup", mock.Anything).Return()
	cmd.On("Run", mock.Anything).Return(nil)

	code := app.Run(cmd)

	assert.Equal(t, 0, code)
	cmd.AssertExpectations(t)
}
