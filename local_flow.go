package doauth

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// LocalFlow handles the local web-based OAuth2 flow.
type LocalFlow struct {
	// Port is the port number for the local callback server (e.g., 8080).
	Port int
	// Timeout is how long to wait for the callback before giving up.
	Timeout time.Duration
	// CallbackPath is the path for the redirect URI (e.g., "/callback").
	CallbackPath string
	// logger is a custom slog logger for the LocalFlow.
	logger *slog.Logger
}

// Result represents the data received from the callback server.
type Result struct {
	Code  string
	State string
	Err   error
}

// LocalFlowOption defines a functional option for configuring the LocalFlow.
type LocalFlowOption func(*LocalFlow)

// WithPort sets the port for the local callback server.
func WithPort(port int) LocalFlowOption {
	return func(f *LocalFlow) {
		f.Port = port
	}
}

// WithTimeout sets the timeout for waiting for the callback.
func WithTimeout(timeout time.Duration) LocalFlowOption {
	return func(f *LocalFlow) {
		f.Timeout = timeout
	}
}

// WithCallbackPath sets the URL path for the callback.
func WithCallbackPath(path string) LocalFlowOption {
	return func(f *LocalFlow) {
		f.CallbackPath = path
	}
}

// WithLocalFlowLogger sets a custom slog logger for the LocalFlow.
func WithLocalFlowLogger(logger *slog.Logger) LocalFlowOption {
	return func(f *LocalFlow) {
		f.logger = logger
	}
}

// NewLocalFlow creates a new LocalFlow with the provided options and sensible defaults.
func NewLocalFlow(opts ...LocalFlowOption) *LocalFlow {
	f := &LocalFlow{
		Port:         DefaultPort,
		Timeout:      2 * time.Minute,
		CallbackPath: DefaultCallbackPath,
		logger:       slog.Default(),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// OpenBrowser opens the specified URL in the default system browser.
func (f *LocalFlow) OpenBrowser(url string) error {
	f.logger.Debug("opening browser", "url", url)
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = CmdWindows
		args = []string{ArgsWindows, url}
	case "darwin":
		cmd = CmdDarwin
		args = []string{url}
	default: // linux, bsd, etc.
		cmd = CmdLinux
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

// WaitForCode starts a local HTTP server and blocks until the authorization code is received or the timeout is reached.
func (f *LocalFlow) WaitForCode(ctx context.Context) (*Result, error) {
	if f.Port == 0 {
		f.Port = DefaultPort
	}
	if f.Timeout == 0 {
		f.Timeout = 2 * time.Minute
	}
	if f.CallbackPath == "" {
		f.CallbackPath = DefaultCallbackPath
	}

	resultChan := make(chan *Result, 1)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", f.Port),
		Handler: mux,
	}

	mux.HandleFunc(f.CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get(ParamCode)
		state := r.URL.Query().Get(ParamState)
		errStr := r.URL.Query().Get(ParamError)

		f.logger.Debug("callback received", "code", code != "", "state", state, "error", errStr)

		if errStr != "" {
			resultChan <- &Result{Err: fmt.Errorf("auth server returned error: %s", errStr)}
			fmt.Fprintln(w, "Authentication failed. You can close this window.")
		} else if code == "" {
			resultChan <- &Result{Err: fmt.Errorf("no code received in callback")}
			fmt.Fprintln(w, "Invalid callback received. You can close this window.")
		} else {
			resultChan <- &Result{Code: code, State: state}
			fmt.Fprintln(w, "Authentication successful! You can now close this window.")
		}
	})

	f.logger.Debug("starting local callback server", "addr", server.Addr)

	// Start server in a goroutine
	go func() {
		// Create a listener manually to catch "address already in use" early if needed
		ln, err := net.Listen("tcp", server.Addr)
		if err != nil {
			resultChan <- &Result{Err: fmt.Errorf("failed to start listener: %w", err)}
			return
		}
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			resultChan <- &Result{Err: fmt.Errorf("server error: %w", err)}
		}
	}()

	// Cleanup and wait
	defer func() {
		f.logger.Debug("shutting down local callback server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	select {
	case res := <-resultChan:
		if res.Err != nil {
			f.logger.Debug("wait for code failed", "error", res.Err)
		} else {
			f.logger.Debug("wait for code successful")
		}
		return res, res.Err
	case <-time.After(f.Timeout):
		f.logger.Debug("wait for code timed out")
		return nil, fmt.Errorf("timeout waiting for callback")
	case <-ctx.Done():
		f.logger.Debug("wait for code context cancelled")
		return nil, ctx.Err()
	}
}
