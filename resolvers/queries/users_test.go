package queries_test

import (
	"context"
	"testing"
)

func TestLocalUsers(t *testing.T) {
	tests := []struct {
		scenario string
		test     func(ctx context.Context, t *testing.T)
	}{}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {

		})
	}
}
