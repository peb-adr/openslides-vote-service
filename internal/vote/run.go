package vote

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Run starts the http server.
//
// The server is automaticly closed when ctx is done.
//
// The service is configured by the argument `environment`. It expect strings in
// the format `KEY=VALUE`, like the output from `os.Environmen()`.
//
// log messages are written with the function given in the argument log. It has
// the signature from log.Printf().
func Run(ctx context.Context, environment []string, log func(format string, v ...interface{})) error {
	env := defaultEnv(environment)

	service := &Vote{}

	mux := http.NewServeMux()
	handleStart(mux, service)
	handleStop(mux, service)
	handleVote(mux, service)
	handleHealth(mux)

	listenAddr := ":" + env["VOTE_PORT"]
	srv := &http.Server{Addr: listenAddr, Handler: mux}

	// Shutdown logic in separate goroutine.
	wait := make(chan error)
	go func() {
		// Wait for the context to be closed.
		<-ctx.Done()

		if err := srv.Shutdown(context.Background()); err != nil {
			wait <- fmt.Errorf("HTTP server shutdown: %w", err)
			return
		}
		wait <- nil
	}()

	log("Listen on %s", listenAddr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP Server failed: %v", err)
	}

	return <-wait
}

// defaultEnv parses the environment (output from os.Environ()) and sets specific
// defaut values.
func defaultEnv(environment []string) map[string]string {
	env := map[string]string{
		"VOTE_HOST": "",
		"VOTE_PORT": "9013",
	}

	for _, value := range environment {
		parts := strings.Split(value, "=")
		if len(parts) != 2 {
			panic(fmt.Sprintf("Invalid value from environment(): %s", value))
		}

		env[parts[0]] = parts[1]
	}
	return env
}
