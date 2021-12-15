package http_test

import (
	"testing"

	"github.com/music-gang/music-gang-api/http"
)

func TestServerAPI_Open(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		server := http.NewServerAPI()
		if err := server.Open(); err != nil {
			t.Errorf("error: %v", err)
		}

		if err := server.Close(); err != nil {
			t.Errorf("error: %v", err)
		}
	})

	t.Run("ErrPortAlreadyBinded", func(t *testing.T) {

		server := http.NewServerAPI()
		defer MustCloseServerAPI(t, server)
		if err := server.Open(); err != nil {
			t.Errorf("error: %v", err)
		}

		if err := server.Open(); err == nil {
			t.Error("error: expected error")
		}
	})
}

func MustOpenServerAPI(tb testing.TB) *http.ServerAPI {

	tb.Helper()

	server := http.NewServerAPI()

	if err := server.Open(); err != nil {
		tb.Fatal(err)
	}

	return server
}

func MustCloseServerAPI(tb testing.TB, server *http.ServerAPI) {

	tb.Helper()

	if err := server.Close(); err != nil {
		tb.Fatal(err)
	}
}
