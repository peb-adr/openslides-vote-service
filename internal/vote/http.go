package vote

import (
	"io"
	"net/http"
)

const (
	httpPathInternal = "/internal/vote"
	httpPathExternal = "/system/vote"
)

type starter interface {
	Start(pollID int, voteType PollType, backend BackendID) error
}

func handleStart(mux *http.ServeMux, start starter) {
	mux.HandleFunc(
		httpPathInternal+"/start",
		func(w http.ResponseWriter, r *http.Request) {

		},
	)
}

type stoper interface {
	Stop(pollID int, w io.Writer) error
}

func handleStop(mux *http.ServeMux, stop stoper) {
	mux.HandleFunc(
		httpPathInternal+"/stop",
		func(w http.ResponseWriter, r *http.Request) {

		},
	)
}

type voter interface {
	Vote(pollID int, r io.Reader) error
}

func handleVote(mux *http.ServeMux, vote voter) {
	mux.HandleFunc(
		httpPathExternal,
		func(w http.ResponseWriter, r *http.Request) {

		},
	)
}

func handleHealth(mux *http.ServeMux) {
	mux.HandleFunc(
		httpPathExternal+"/health",
		func(w http.ResponseWriter, r *http.Request) {

		},
	)
}
