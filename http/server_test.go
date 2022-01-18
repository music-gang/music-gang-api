package http_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/app/service"
	apphttp "github.com/music-gang/music-gang-api/http"
	"github.com/music-gang/music-gang-api/mock"
)

func TestServerAPI_Open(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		server := apphttp.NewServerAPI()
		if err := server.Open(); err != nil {
			t.Errorf("error: %v", err)
		}

		if err := server.Close(); err != nil {
			t.Errorf("error: %v", err)
		}
	})

	t.Run("ErrPortAlreadyBinded", func(t *testing.T) {

		server := apphttp.NewServerAPI()
		server.Addr = ":8080"
		defer MustCloseServerAPI(t, server)
		if err := server.Open(); err != nil {
			t.Errorf("error: %v", err)
		}

		if err := server.Open(); err == nil {
			t.Error("error: expected error")
		}
	})
}

func TestBuildInfoHandler(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		s := MustOpenServerAPI(t)
		defer MustCloseServerAPI(t, s)

		app.Commit = "OK"

		req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/build/info", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var info map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			t.Fatal(err)
		}

		if commit, ok := info["commit"]; !ok {
			t.Error("expected commit key in info")
		} else if commit != app.Commit {
			t.Errorf("expected commit %s, got %s", app.Commit, commit)
		}
	})
}

func MustMarshalJSON(tb testing.TB, value interface{}) []byte {

	tb.Helper()

	bytes, err := json.Marshal(value)
	if err != nil {
		tb.Fatal(err)
	}

	return bytes
}

func MustOpenServerAPI(tb testing.TB) *apphttp.ServerAPI {

	tb.Helper()

	server := apphttp.NewServerAPI()

	if err := server.Open(); err != nil {
		tb.Fatal(err)
	}

	initFakeLogger(tb, server)

	return server
}

func MustCloseServerAPI(tb testing.TB, server *apphttp.ServerAPI) {

	tb.Helper()

	if err := server.Close(); err != nil {
		tb.Fatal(err)
	}
}

func initFakeLogger(tb testing.TB, server *apphttp.ServerAPI) {

	tb.Helper()

	server.LogService = &mock.LogService{
		FormatFn: func() string {
			return service.FormatStandard
		},
		LevelFn: func() int {
			return service.LevelAll
		},
		OutputFn: func() io.Writer {
			return io.Discard
		},
		ReportDebugFn:   func(ctx context.Context, msg string) {},
		ReportErrorFn:   func(ctx context.Context, err error) {},
		ReportFatalFn:   func(ctx context.Context, err error) {},
		ReportInfoFn:    func(ctx context.Context, info string) {},
		ReportPanicFn:   func(ctx context.Context, err interface{}) {},
		ReportWarningFn: func(ctx context.Context, warning string) {},
	}
}
