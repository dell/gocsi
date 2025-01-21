package logging

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestWithRequestLogging(t *testing.T) {
	w := &bytes.Buffer{}
	tests := []struct {
		name  string
		input io.Writer
		want  opts
	}{
		{
			name: "non nil input",
			want: opts{
				reqw:             w,
				repw:             nil,
				disableLogVolCtx: false,
			},
			input: w,
		},
		{
			name: "nil input",
			want: opts{
				reqw:             os.Stdout,
				repw:             nil,
				disableLogVolCtx: false,
			},
			input: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithRequestLogging(tt.input)

			// Use interceptor type to check the expected opts
			i := newLoggingInterceptor(opt)

			// Check interceptor type opts to verify WithRequestLogging set them as expected
			if i.opts != tt.want {
				t.Errorf("WithRequestLogging() returned function with parameters = %v, want function with parameters %v", i.opts, tt.want)
			}
		})
	}
}

func TestWithResponseLogging(t *testing.T) {
	w := &bytes.Buffer{}
	tests := []struct {
		name  string
		input io.Writer
		want  opts
	}{
		{
			name: "non nil input",
			want: opts{
				reqw:             nil,
				repw:             w,
				disableLogVolCtx: false,
			},
			input: w,
		},
		{
			name: "nil input",
			want: opts{
				reqw:             nil,
				repw:             os.Stdout,
				disableLogVolCtx: false,
			},
			input: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithResponseLogging(tt.input)

			// Use interceptor type to check the expected opts
			i := newLoggingInterceptor(opt)

			// Check interceptor type opts to verify WithResponseLogging set them as expected
			if i.opts != tt.want {
				t.Errorf("WithResponseLogging() returned function with parameters = %v, want function with parameters %v", i.opts, tt.want)
			}
		})
	}
}

func TestWithDisableLogVolumeContext(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "switch to true",
			want: true,
		},
		{
			name: "test default",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use interceptor type to check if WithDisableLogVolumeContext is able to switch value from false to true
			i := newLoggingInterceptor()

			// if want = true, use WithDisableLogVolumeContext to switch disableLogVolCtx from false to true
			if tt.want {
				i = newLoggingInterceptor(WithDisableLogVolumeContext())
			}
			// Check interceptor type opts to verify WithDisableLogVolumeContext switched or didn't switch
			if i.opts.disableLogVolCtx != tt.want {
				t.Errorf("WithResponseLogging() returned function with parameters = %v, want function with parameters %v", i.opts, tt.want)
			}
		})
	}
}
