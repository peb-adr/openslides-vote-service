package vote

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type CreaterStub struct {
	pid    int
	config string
	err    error
}

func (c *CreaterStub) Create(ctx context.Context, pollID int, config io.Reader) error {
	c.pid = pollID

	cfg, err := io.ReadAll(config)
	if err != nil {
		return err
	}
	c.config = string(cfg)
	return c.err
}

func TestHandleCreate(t *testing.T) {
	creater := &CreaterStub{}

	url := "/internal/vote/create"
	mux := http.NewServeMux()
	handleCreate(mux, creater)

	t.Run("Get request", func(t *testing.T) {
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 405 {
			t.Errorf("Got status %s, expected 405 - Method not allowed", resp.Result().Status)
		}
	})

	t.Run("No pid", func(t *testing.T) {
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, httptest.NewRequest("POST", url, nil))

		if resp.Result().StatusCode != 400 {
			t.Errorf("Got status %s, expected 400 - Bad Request", resp.Result().Status)
		}
	})

	t.Run("Invalid pid", func(t *testing.T) {
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, httptest.NewRequest("POST", url+"?pid=value", nil))

		if resp.Result().StatusCode != 400 {
			t.Errorf("Got status %s, expected 400 - Bad Request", resp.Result().Status)
		}
	})

	t.Run("Valid", func(t *testing.T) {
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, httptest.NewRequest("POST", url+"?pid=1", strings.NewReader("request body")))

		if resp.Result().StatusCode != 200 {
			t.Errorf("Got status %s, expected 200 - OK", resp.Result().Status)
		}

		if creater.pid != 1 {
			t.Errorf("Creater was called with pid %d, expected 1", creater.pid)
		}

		if creater.config != "request body" {
			t.Errorf("Creater was called with body `%s` expected `request body`", creater.config)
		}
	})

	t.Run("Invalid error", func(t *testing.T) {
		creater.err = ErrInvalid

		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, httptest.NewRequest("POST", url+"?pid=1", strings.NewReader("request body")))

		if resp.Result().StatusCode != 400 {
			t.Errorf("Got status %s, expected 400", resp.Result().Status)
		}

		var body struct {
			Error string `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decoding resp body: %v", err)
		}

		if body.Error != "invalid" {
			t.Errorf("Got error `%s`, expected `invalid`", body.Error)
		}
	})

	t.Run("Internal error", func(t *testing.T) {
		creater.err = errors.New("foobar")

		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, httptest.NewRequest("POST", url+"?pid=1", strings.NewReader("request body")))

		if resp.Result().StatusCode != 500 {
			t.Errorf("Got status %s, expected 400", resp.Result().Status)
		}

		var body struct {
			Error string `json:"error"`
			MSG   string `json:"msg"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decoding resp body: %v", err)
		}

		if body.Error != "internal" {
			t.Errorf("Got error `%s`, expected `internal`", body.Error)
		}

		if body.MSG != "foobar" {
			t.Errorf("Got error message `%s`, expected `foobar`", body.MSG)
		}
	})
}
