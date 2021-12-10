package api

import (
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/music-gang/music-gang-api/app/service"
	"golang.org/x/crypto/acme/autocert"
)

// ServerAPI is the main server for the API
type ServerAPI struct {
	ln net.Listener
	// server is the main server for the API
	server *http.Server

	// handler is the main handler for the API
	handler *echo.Echo

	// Addr Bind address for the server.
	Addr string
	// Domain name to use for the server.
	// If specified, server is run on TLS using acme/autocert.
	Domain string

	// JWTSecret is the secret used to sign JWT tokens.
	JWTSecret string

	// Services used by HTTP handler.
	AuthService service.AuthService
	UserService service.UserService
}

// NewAPISerer creates a new API server.
func NewAPISerer() *ServerAPI {

	addr := ":8080"
	domain := ""

	jwtSecret := "secret"

	s := &ServerAPI{
		server:    &http.Server{},
		handler:   echo.New(),
		Addr:      addr,
		Domain:    domain,
		JWTSecret: jwtSecret,
	}

	// Set echo as the default HTTP handler.
	s.server.Handler = s.handler

	// Base Middleware
	s.handler.Use(middleware.Recover())

	// Register routes
	s.registerRoutes()

	return s
}

// Open validates the server options and start it on the bind address.
func (s *ServerAPI) Open() (err error) {

	if s.Domain != "" {
		s.ln = autocert.NewListener(s.Domain)
	} else {
		if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
			return err
		}
	}

	go s.server.Serve(s.ln)

	return nil
}

// Scheme returns the scheme used by the server.
func (s *ServerAPI) Scheme() string {
	if s.Domain != "" {
		return "https"
	}
	return "http"
}

// UseTLS returns true if the server is using TLS.
func (s *ServerAPI) UseTLS() bool {
	return s.Domain != ""
}

// registerRoutes registers all routes for the API.
func (s *ServerAPI) registerRoutes() {}
