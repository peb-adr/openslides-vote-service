package vote

import (
	"context"
	"io"
	"net/http"
)

const (
	httpPathInternal = "/internal/vote"
	httpPathExternal = "/system/vote"
)

type starter interface {
	Start(ctx context.Context, pollID int, config io.Reader) error
}

func handleStart(mux *http.ServeMux, start starter) {
	mux.HandleFunc(
		httpPathInternal+"/start",
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "TODO", 500)
		},
	)
}

type stoper interface {
	Stop(ctx context.Context, pollID int, w io.Writer) error
}

func handleStop(mux *http.ServeMux, stop stoper) {
	mux.HandleFunc(
		httpPathInternal+"/stop",
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "TODO", 500)
		},
	)
}

type clearer interface {
	Clear(ctx context.Context, pollID int) error
}

func handleClear(mux *http.ServeMux, clear clearer) {
	mux.HandleFunc(
		httpPathInternal+"/clear",
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "TODO", 500)
		},
	)
}

type voter interface {
	Vote(ctx context.Context, pollID int, r io.Reader) error
}

func handleVote(mux *http.ServeMux, vote voter) {
	mux.HandleFunc(
		httpPathExternal,
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "TODO", 500)
		},
	)
}

func handleHealth(mux *http.ServeMux) {
	mux.HandleFunc(
		httpPathExternal+"/health",
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "TODO", 500)
		},
	)
}
