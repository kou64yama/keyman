package app_test

import (
	"bytes"
	"flag"
	"keyman/internal/app"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsage(t *testing.T) {
	tests := []struct {
		name     string
		synopsis []string
	}{
		{
			name:     "single",
			synopsis: []string{"test [option]"},
		},
		{
			name: "multiple",
			synopsis: []string{
				"test [option]",
				"test foo",
				"test bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			cmd := new(CommandMock)
			cmd.On("Synopsis").Return(tt.synopsis)

			w := bytes.NewBuffer([]byte{})
			f := flag.NewFlagSet("test", flag.ContinueOnError)
			f.SetOutput(w)
			app.Usage(cmd, f)()

			for _, s := range tt.synopsis {
				assert.Contains(t, w.String(), s)
			}
			cmd.AssertExpectations(t)
		})
	}
}
