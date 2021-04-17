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

type creater interface {
	Create(ctx context.Context, pollID int, config io.Reader) error
}

func handleCreate(mux *http.ServeMux, create creater) {
	mux.HandleFunc(
		httpPathInternal+"/create",
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
	Vote(ctx context.Context, pollID, requestUser int, r io.Reader) error
}

type authenticater interface {
	Authenticate(http.ResponseWriter, *http.Request) (context.Context, error)
	FromContext(context.Context) int
}

func handleVote(mux *http.ServeMux, vote voter, auth authenticater) {
	// TODO: Get user-id
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
