package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/music-gang/music-gang-api/app"
	"github.com/music-gang/music-gang-api/handler"
	"golang.org/x/crypto/acme/autocert"
)

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

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

	// service handler
	ServiceHandler *handler.ServiceHandler

	// loggin service used by HTTP Server.
	LogService log.Logger
}

// NewServerAPI creates a new API server.
func NewServerAPI() *ServerAPI {

	s := &ServerAPI{
		server:  &http.Server{},
		handler: echo.New(),
	}

	// Set echo as the default HTTP handler.
	s.server.Handler = s.handler

	// Base Middleware
	s.handler.Use(middleware.Secure())
	s.handler.Use(middleware.CORS())
	s.handler.Use(s.HTTPContextMiddleware)
	s.handler.Use(s.RecoverPanicMiddleware)

	s.handler.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Music Gang API")
	})

	// Register routes for the API v1.
	v1Group := s.handler.Group("/v1")
	s.registerRoutes(v1Group)

	return s
}

func (s *ServerAPI) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// Port returns the TCP port for the running server.
// This is useful in tests where we allocate a random port by using ":0".
func (s *ServerAPI) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
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

// URL returns the URL for the server.
// This is useful in tests where we allocate a random port by using ":0".
func (s *ServerAPI) URL() string {

	scheme, port := s.Scheme(), s.Port()

	domain := "localhost"

	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}

	return fmt.Sprintf("%s://%s:%d", scheme, domain, port)
}

// UseTLS returns true if the server is using TLS.
func (s *ServerAPI) UseTLS() bool {
	return s.Domain != ""
}

// registerRoutes registers all routes for the API.
func (s *ServerAPI) registerRoutes(g *echo.Group) {

	buildGroup := g.Group("/build")
	buildGroup.GET("/info", func(c echo.Context) error {
		return SuccessResponseJSON(c, http.StatusOK, map[string]string{
			"commit":          app.Commit,
			"release_version": app.ReleaseVersion,
			"release_date":    app.ReleaseCreatedAt,
		})
	})

	authGroup := g.Group("/auth")
	s.registerAuthRoutes(authGroup)

	userGroup := g.Group("/user", s.JWTVerifyMiddleware)
	s.registerUserRoutes(userGroup)

	vmGroup := g.Group("/vm")
	s.registerVmRoutes(vmGroup)

	contractGroup := g.Group("/contract", s.JWTVerifyMiddleware)
	s.registerContractRoutes(contractGroup)
}

// registerAuthRoutes registers all routes for the API group auth.
func (s *ServerAPI) registerAuthRoutes(g *echo.Group) {
	g.POST("/login", s.AuthLoginHandler)
	g.POST("/register", s.AuthRegisterHandler)
	g.POST("/refresh", s.AuthRefreshHandler)
	g.DELETE("/logout", s.AuthLogoutHandler)

	// oauth2 routes
	g.GET("/oauth2/:source/callback", nil)
}

// registerContractRoutes registers all routes for the API group contract.
func (s *ServerAPI) registerContractRoutes(g *echo.Group) {
	g.POST("", s.ContractCreateHandler)
	g.PUT("/:id", s.ContractUpdateHandler)
	g.GET("/:id", s.ContractHandler)
	g.POST("/:id/revision", s.ContractMakeRevisionHandler)
	g.POST("/:id/call", s.ContractCallHandler)         // latest revision
	g.POST("/:id/call/:rev", s.ContractCallRevHandler) // specific revision
}

// registerUserRoutes register all routes for the API group user.
func (s *ServerAPI) registerUserRoutes(g *echo.Group) {
	g.GET("", s.UserHandler)
	g.PUT("", s.UserUpdateHandler)
}

// registerVmRoutes registers all routes for the API group vm.
func (s *ServerAPI) registerVmRoutes(g *echo.Group) {
	g.GET("/stats", s.VmStatsHandler)
}

// SuccessResponseJSON returns a JSON response with the given status code and data.
func SuccessResponseJSON(c echo.Context, httpCode int, data interface{}) error {
	return c.JSON(httpCode, data)
}

// ListenAndServeTLSRedirect runs an HTTP server on port 80 to redirect users
// to the TLS-enabled port 443 server.
func ListenAndServeTLSRedirect(domain string) error {
	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+domain, http.StatusFound)
	}))
}

// extractJWT from the *http.Request if omitted or wrong formed, empty string is returned
func extractJWT(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
