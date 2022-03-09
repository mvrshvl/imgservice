package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestWrite(t *testing.T) {
	const (
		testErr = "test"
		testFmt = "level=%s msg=%s"
	)

	testCases := []struct {
		name        string
		write       func(ctx context.Context, args ...interface{})
		expectedMsg string
	}{
		{
			name:        "error",
			write:       Error,
			expectedMsg: fmt.Sprintf(testFmt, "error", testErr),
		},
		{
			name:        "info",
			write:       Info,
			expectedMsg: fmt.Sprintf(testFmt, "info", testErr),
		},
		{
			name:        "debug",
			write:       Debug,
			expectedMsg: fmt.Sprintf(testFmt, "debug", testErr),
		},
		{
			name: "warnf",
			write: func(ctx context.Context, args ...interface{}) {
				Warnf(ctx, "%s", args[0])
			},
			expectedMsg: fmt.Sprintf(testFmt, "warning", testErr),
		},
		{
			name: "errorf",
			write: func(ctx context.Context, args ...interface{}) {
				Errorf(ctx, "%s", args[0])
			},
			expectedMsg: fmt.Sprintf(testFmt, "error", testErr),
		},
		{
			name: "infof",
			write: func(ctx context.Context, args ...interface{}) {
				Infof(ctx, "%s", args[0])
			},
			expectedMsg: fmt.Sprintf(testFmt, "info", testErr),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			buff := new(bytes.Buffer)

			logger := New(logrus.DebugLevel, buff)

			ctx := context.Background()
			ctx = WithCtx(ctx, logger)

			testCase.write(ctx, testErr)

			resultLog, err := io.ReadAll(buff)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(resultLog), testCase.expectedMsg) {
				t.Fatalf("expected log %v, result %v", testCase.expectedMsg, string(resultLog))
			}
		})
	}
}
