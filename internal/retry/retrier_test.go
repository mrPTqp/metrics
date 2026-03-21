package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testClassifier struct {
	ret map[error]ErrorClassification
}

func (c *testClassifier) Classify(err error) ErrorClassification {
	if cls, ok := c.ret[err]; ok {
		return cls
	}
	return NonRetriable
}

func TestDoWithRetry_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before first retry check

	retriableErr := errors.New("retriable")
	classifier := &testClassifier{
		ret: map[error]ErrorClassification{
			retriableErr: Retriable,
		},
	}

	err := DoWithRetry(ctx, classifier, func() error {
		return retriableErr
	}, 5, 10*time.Millisecond)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DoWithRetry() error = %v, want %v", err, context.Canceled)
	}
}

func TestDoWithRetry_Scenarios(t *testing.T) {
	retriableErr := errors.New("retriable")
	nonRetriableErr := errors.New("non-retriable")

	tests := []struct {
		name       string
		maxRetries int
		classifier *testClassifier
		operation  func(calls *int) error
		wantErr    error
		wantCalls  int
	}{
		{
			name:       "success on first try",
			maxRetries: 3,
			classifier: &testClassifier{},
			operation: func(calls *int) error {
				*calls++
				return nil
			},
			wantErr:   nil,
			wantCalls: 1,
		},
		{
			name:       "non-retriable error",
			maxRetries: 5,
			classifier: &testClassifier{
				ret: map[error]ErrorClassification{
					nonRetriableErr: NonRetriable,
				},
			},
			operation: func(calls *int) error {
				*calls++
				return nonRetriableErr
			},
			wantErr:   nonRetriableErr,
			wantCalls: 1,
		},
		{
			name:       "retriable then success",
			maxRetries: 3,
			classifier: &testClassifier{
				ret: map[error]ErrorClassification{
					retriableErr: Retriable,
				},
			},
			operation: func(calls *int) error {
				*calls++
				if *calls == 1 {
					return retriableErr
				}
				return nil
			},
			wantErr:   nil,
			wantCalls: 2,
		},
		{
			name:       "exhaust retries",
			maxRetries: 3,
			classifier: &testClassifier{
				ret: map[error]ErrorClassification{
					retriableErr: Retriable,
				},
			},
			operation: func(calls *int) error {
				*calls++
				return retriableErr
			},
			wantErr:   retriableErr,
			wantCalls: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var calls int

			err := DoWithRetry(ctx, tt.classifier, func() error {
				return tt.operation(&calls)
			}, tt.maxRetries, 0)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("DoWithRetry() error = %v, want nil", err)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DoWithRetry() error = %v, want %v", err, tt.wantErr)
			}

			if calls != tt.wantCalls {
				t.Errorf("operation called %d times, want %d", calls, tt.wantCalls)
			}
		})
	}
}


