package retry

import (
	"context"
	"errors"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestOnErr(t *testing.T) {
	_ = logger.Initialize("debug")
	t.Run(
		"function returns an error in given interval",
		func(t *testing.T) {
			ctx := context.Background()
			targetErr := errors.New("some error")
			targetErrs := []error{targetErr}
			retryFn := func(args ...any) error { return targetErr }
			rIntervals := []time.Duration{
				100 * time.Millisecond,
				200 * time.Millisecond,
				400 * time.Millisecond,
			}

			err := OnErr(ctx, targetErrs, rIntervals, retryFn)
			require.Eventually(
				t,
				func() bool { return errors.Is(err, targetErr) },
				750*time.Millisecond,
				250*time.Millisecond,
			)
		},
	)

	t.Run(
		"function returns no error at the last retry",
		func(t *testing.T) {
			ctx := context.Background()
			targetErr := errors.New("custom error")
			targetErrs := []error{targetErr}
			funcRetryCounter := 0
			retryFn := func(a, b int) error {
				funcRetryCounter++
				if funcRetryCounter > 3 {
					return nil
				}
				return targetErr
			}
			rIntervals := []time.Duration{
				100 * time.Millisecond,
				100 * time.Millisecond,
				100 * time.Millisecond,
				100 * time.Millisecond,
			}

			err := OnErr(ctx, targetErrs, rIntervals,
				func(args ...any) error {
					a := args[0].(int)
					b := args[1].(int)
					return retryFn(a, b)
				}, 1, 2,
			)
			require.NoError(t, err)
		},
	)

	t.Run(
		"function returns second error from the given slice",
		func(t *testing.T) {
			ctx := context.Background()
			targetErrs := []error{errors.New("custom error"), errors.New("super custom error")}
			funcRetryCounter := 0
			retryFn := func(a, b int) error {
				funcRetryCounter++
				if funcRetryCounter > 3 {
					return nil
				}
				return targetErrs[1]
			}
			rIntervals := []time.Duration{
				100 * time.Millisecond,
				100 * time.Millisecond,
				100 * time.Millisecond,
			}

			err := OnErr(ctx, targetErrs, rIntervals,
				func(args ...any) error {
					a := args[0].(int)
					b := args[1].(int)
					return retryFn(a, b)
				}, 1, 2,
			)
			require.NoError(t, err)
		},
	)
}
