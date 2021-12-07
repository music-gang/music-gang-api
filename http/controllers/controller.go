package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// MARK: CustomContext, constructor & consts

// Definisce le costanti per i tipi del custom context
const (
	RequestTypeWeb = "web"
	RequestTypeAPI = "api"

	CodeJWTAuthFailed           = 1000
	CodeJWTNeedLogin            = 1001
	CodeBasiAuthFailed          = 2000
	CodeDeviceRegister          = 3000
	CodePeschereccioGuardFailed = 4001
	CodeCheckTVTokenFailed      = 5001
)

// CustomContext - Definisce un custom context per arricchire i dati della request
type CustomContext struct {
	RequestType string
	echo.Context
}

// NewCustomContext - Costruttore del custom context
func NewCustomContext(c echo.Context, requestType string) *CustomContext {
	return &CustomContext{Context: c, RequestType: requestType}
}

// Filters

// PaginateFilter - Definisce la struct per i filtri generici relativi ad un API paginata
type PaginateFilter struct {
	Page  int
	Limit int
	Query string
}

// StandardPaginateFilter - Restituisce una struct di tipo PaginateFilter valorizzata standard
func StandardPaginateFilter() PaginateFilter {
	return PaginateFilter{
		Page:  1,
		Limit: 25,
		Query: "",
	}
}

// MARK: Response interface

// Response - Interface per generalizzare una response api/web
type Response interface {
	GetCode() int
	GetSuccess() bool
	GetMessage() string
	GetContent() echo.Map
}

// NewResponse - Restituisce una nuova response generica in base al context passato
func NewResponse(c echo.Context, success bool, code int, message string, content echo.Map) Response {

	cc := c.(*CustomContext)

	var response Response

	if cc.RequestType == RequestTypeAPI {
		response = NewResponseAPI(success, code, message, content)
	} else {
		response = NewResponseWeb(success, code, message, content)
	}

	return response
}

// MARK: Response API, constructor and implementation

// ResponseData - Definisce la struct della response.data
type ResponseData struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResponseAPI - Define a standard struct response
type ResponseAPI struct {
	Response ResponseData `json:"response"`
	Data     echo.Map     `json:"data,omitempty"`
}

// NewResponseAPI - Restituisce una response
func NewResponseAPI(success bool, code int, message string, content echo.Map) ResponseAPI {
	return ResponseAPI{
		Response: ResponseData{
			Code:    code,
			Success: success,
			Message: message,
		},
		Data: content,
	}
}

// GetCode - Restituisce il codice della response api
func (r ResponseAPI) GetCode() int {
	return r.Response.Code
}

// GetSuccess - Restituisce l'esito della response api
func (r ResponseAPI) GetSuccess() bool {
	return r.Response.Success
}

// GetMessage - Restituisce il message della response api
func (r ResponseAPI) GetMessage() string {
	return r.Response.Message
}

// GetContent - Restituisce l'eventuale content della response api
func (r ResponseAPI) GetContent() echo.Map {
	return r.Data
}

// MARK: Response WEB, constructor and implementation

// ResponseWeb - Definisce la struttura di una response web
type ResponseWeb struct {
	Code    int      `json:"code"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Content echo.Map `json:"content,omitempty"`
}

// NewResponseWeb - Restituisce una nuova response standard
func NewResponseWeb(success bool, code int, message string, content echo.Map) ResponseWeb {
	return ResponseWeb{
		Code:    code,
		Success: success,
		Message: message,
		Content: content,
	}
}

// GetCode - Restituisce il codice della response web
func (r ResponseWeb) GetCode() int {
	return r.Code
}

// GetSuccess - Restituisce l'esito della response web
func (r ResponseWeb) GetSuccess() bool {
	return r.Success
}

// GetMessage - Restituisce il message della response web
func (r ResponseWeb) GetMessage() string {
	return r.Message
}

// GetContent - Restituisce l'eventuale content della response web
func (r ResponseWeb) GetContent() echo.Map {
	return r.Content
}

// MARK Exported funcs

// APIAuthBasicFailedResponse - Restituisce l'errore per api basic auth failed
func APIAuthBasicFailedResponse(c echo.Context) error {
	return c.JSON(http.StatusForbidden, FailedResponse(c, CodeBasiAuthFailed, "Forbidden", nil))
}

// APIJWTAuthFailedResponse - Restituisce l'errore per api jwt auth check failed
func APIJWTAuthFailedResponse(c echo.Context, msg string) error {
	return c.JSON(http.StatusUnauthorized, FailedResponse(c, CodeJWTAuthFailed, msg, nil))
}

// APIJWTNeedLoginFailedResponse - Restituisce l'errore per api jwt need login failed
func APIJWTNeedLoginFailedResponse(c echo.Context, msg string) error {
	return c.JSON(http.StatusNetworkAuthenticationRequired, FailedResponse(c, CodeJWTNeedLogin, msg, nil))
}

// FailedResponse - Restitusce una risposta failed
func FailedResponse(c echo.Context, code int, message string, content echo.Map) Response {
	return NewResponse(c, false, code, message, content)
}

// SuccessResponse - Restituisce una risposta success
func SuccessResponse(c echo.Context, content echo.Map) Response {
	return NewResponse(c, true, 0, "ok!", content)
}
