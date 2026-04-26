package doauth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalFlow_WaitForCode(t *testing.T) {
	flow := NewLocalFlow(
		WithPort(8888),
		WithTimeout(2*time.Second),
		WithCallbackPath("/test-callback"),
	)

	ctx := context.Background()

	// Start waiting in a goroutine
	resChan := make(chan *Result, 1)
	go func() {
		res, err := flow.WaitForCode(ctx)
		if err != nil {
			resChan <- &Result{Err: err}
			return
		}
		resChan <- res
	}()

	// Give the server a moment to start
	time.Sleep(200 * time.Millisecond)

	// Simulate the browser callback
	resp, err := http.Get("http://localhost:8888/test-callback?code=secret-code&state=expected-state")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Wait for the result
	select {
	case res := <-resChan:
		assert.NoError(t, res.Err)
		assert.Equal(t, "secret-code", res.Code)
		assert.Equal(t, "expected-state", res.State)
	case <-time.After(3 * time.Second):
		t.Fatal("test timed out waiting for result")
	}
}

func TestLocalFlow_Timeout(t *testing.T) {
	flow := NewLocalFlow(
		WithPort(8889),
		WithTimeout(500*time.Millisecond),
	)

	_, err := flow.WaitForCode(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "timeout waiting for callback", err.Error())
}
