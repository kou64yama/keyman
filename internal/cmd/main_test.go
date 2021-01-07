package cmd_test

import (
	"bytes"
	"errors"
	"flag"
	"keyman/internal/app"
	"keyman/internal/cmd"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CommandMock struct {
	app.Command
	mock.Mock
}

func (m *CommandMock) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *CommandMock) Synopsis() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *CommandMock) Setup(f *flag.FlagSet) {
	m.Called(f)
}

func (m *CommandMock) Run(args []string) error {
	return m.Called(args).Error(0)
}

type MainCommandTestSuite struct {
	suite.Suite
}

func (s *MainCommandTestSuite) TestName() {
	c := new(cmd.MainCommand)
	s.NotEmpty(c.Name())
}

func (s *MainCommandTestSuite) TestSynopsis() {
	act := new(CommandMock)
	act.On("Synopsis").Return([]string{"foo", "bar"})

	c := new(cmd.MainCommand)
	c.Actions = append(c.Actions, act)
	s.Len(c.Synopsis(), 3)

	act.AssertExpectations(s.T())
}

func (s *MainCommandTestSuite) TestSetup() {
	f := flag.NewFlagSet("test", flag.ContinueOnError)

	c := new(cmd.MainCommand)
	c.Setup(f)

	s.Equal(os.Stdout, c.Stdout)
	s.Equal(os.Stderr, c.Stderr)
	s.Empty(c.Actions)
}

func (s *MainCommandTestSuite) TestRun() {
	tests := []struct {
		name   string
		args   []string
		stdout []byte
		stderr []byte
	}{
		{
			name:   "keyman test",
			args:   []string{"test"},
			stdout: []byte{},
			stderr: []byte{},
		},
		{
			name:   "keyman -V",
			args:   []string{"-V"},
			stdout: []byte("Keyman 0.0.0\n"),
			stderr: []byte{},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			f := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := bytes.NewBuffer([]byte{})
			stderr := bytes.NewBuffer([]byte{})

			act := new(CommandMock)
			act.On("Name").Return("test")
			act.On("Setup", mock.Anything)
			act.On("Run", mock.Anything).Return(nil)

			c := new(cmd.MainCommand)
			c.Setup(f)
			c.Stdout = stdout
			c.Stderr = stderr
			c.Actions = []app.Command{act}

			if !s.NoError(f.Parse(tt.args)) {
				return
			}

			err := c.Run(f.Args())
			if !s.NoError(err) {
				return
			}
			s.Equal(tt.stdout, stdout.Bytes())
			s.Equal(tt.stderr, stderr.Bytes())
		})
	}
}

func (s *MainCommandTestSuite) TestRunErr() {
	tests := []struct {
		name   string
		args   []string
		stdout []byte
		stderr []byte
	}{
		{
			name:   "keyman",
			args:   []string{},
			stdout: []byte{},
			stderr: []byte{},
		},
		{
			name:   "keyman test",
			args:   []string{"test"},
			stdout: []byte{},
			stderr: []byte{},
		},
		{
			name:   "keyman error",
			args:   []string{"error"},
			stdout: []byte{},
			stderr: []byte{},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			f := flag.NewFlagSet("test", flag.ContinueOnError)
			stdout := bytes.NewBuffer([]byte{})
			stderr := bytes.NewBuffer([]byte{})

			act := new(CommandMock)
			act.On("Name").Return("test")
			act.On("Setup", mock.Anything)
			act.On("Run", mock.Anything).Return(errors.New("test error"))

			c := new(cmd.MainCommand)
			c.Setup(f)
			c.Stdout = stdout
			c.Stderr = stderr
			c.Actions = []app.Command{act}

			if !s.NoError(f.Parse(tt.args)) {
				return
			}

			err := c.Run(f.Args())
			s.Error(err)
		})
	}
}

func TestMainCommandTestSuite(t *testing.T) {
	suite.Run(t, new(MainCommandTestSuite))
}
