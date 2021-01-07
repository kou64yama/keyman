package app_test

import (
	"bytes"
	"errors"
	"flag"
	"keyman/internal/app"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CommandMock struct {
	mock.Mock
}

func (m *CommandMock) Name() string {
	return m.Called().String(0)
}

func (m *CommandMock) Synopsis() []string {
	return m.Called().Get(0).([]string)
}

func (m *CommandMock) Setup(f *flag.FlagSet) {
	m.Called(f)
}

func (m *CommandMock) Run(args []string) error {
	return m.Called(args).Error(0)
}

type ExecutorTestSuite struct {
	suite.Suite
}

func (s *ExecutorTestSuite) TestRun() {
	cmd := new(CommandMock)
	cmd.On("Name").Return("test")
	cmd.On("Setup", mock.Anything)
	cmd.On("Run", mock.Anything).Return(nil)

	stderr := bytes.NewBuffer([]byte{})

	exec := &app.Executor{Stderr: stderr}
	err := exec.Run(cmd)

	s.NoError(err)
	cmd.AssertExpectations(s.T())
}

func (s *ExecutorTestSuite) TestRunHelp() {
	cmd := new(CommandMock)
	cmd.On("Name").Return("test")
	cmd.On("Synopsis").Return([]string{"test [option]"})
	cmd.On("Setup", mock.Anything)

	stderr := bytes.NewBuffer([]byte{})
	exec := &app.Executor{Stderr: stderr}

	err := exec.Run(cmd, "-h")

	s.Equal(flag.ErrHelp, err)
	s.Contains(stderr.String(), "Usage:")
	cmd.AssertExpectations(s.T())
}

func (s *ExecutorTestSuite) TestHandleError() {
	tests := []struct {
		name   string
		err    error
		want   int
		stderr []byte
	}{
		{name: "nil", stderr: []byte{}},
		{name: "flag.ErrHelp", err: flag.ErrHelp, want: 2, stderr: []byte{}},
		{name: "others", err: errors.New("test error"), want: 1, stderr: []byte("test error\n")},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			stderr := bytes.NewBuffer([]byte{})
			exec := &app.Executor{Stderr: stderr}

			got := exec.HandleError(tt.err)

			s.Equal(tt.want, got)
			s.Equal(tt.stderr, stderr.Bytes())
		})
	}
}

func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}
