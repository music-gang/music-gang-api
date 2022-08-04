package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/app/entity"
	"github.com/music-gang/music-gang-api/app/service"
)

// ContractCallHandler is the handler for the /contract/:id/call API.
func (s *ServerAPI) ContractCallHandler(c echo.Context) error {

	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid contract id"), nil)
	}

	if res, err := s.ServiceHandler.CallContract(c.Request().Context(), contractID, 0); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"result": res,
		})
	}
}

// ContractCallRevHandler is the handler for the /contract/:id/call/:rev API.
func (s *ServerAPI) ContractCallRevHandler(c echo.Context) error {

	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid contract id"), nil)
	}

	revisionNumber, err := strconv.ParseUint(c.Param("rev"), 10, 32)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid revision number"), nil)
	}

	if res, err := s.ServiceHandler.CallContract(c.Request().Context(), contractID, entity.RevisionNumber(revisionNumber)); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"result": res,
		})
	}
}

// ContractHandler is the handler for the /contract/:id search API.
func (s *ServerAPI) ContractHandler(c echo.Context) error {

	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid contract id"), nil)
	}

	if contract, err := s.ServiceHandler.FindContractByID(c.Request().Context(), contractID); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"contract": contract,
		})
	}
}

// ContractCreateHandler is the handler for the /contract create API.
func (s *ServerAPI) ContractCreateHandler(c echo.Context) error {

	var contractParams entity.Contract
	if err := c.Bind(&contractParams); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	if contract, err := s.ServiceHandler.CreateContract(c.Request().Context(), &contractParams); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"contract": contract,
		})
	}
}

// ContractMakeRevisionHandler is the handler for the /contract/:id/revision create API.
func (s *ServerAPI) ContractMakeRevisionHandler(c echo.Context) error {

	var revision entity.Revision

	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid contract id"), nil)
	}

	formDataRevision := c.FormValue("revision")
	if formDataRevision == "" {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	} else if err := json.Unmarshal([]byte(formDataRevision), &revision); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	file, _, err := c.Request().FormFile("compiled_revision")
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}
	defer file.Close()

	revision.ContractID = contractID

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, file); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "unable to read file"), nil)
	}

	revision.CompiledCode = buf.Bytes()

	if revision, err := s.ServiceHandler.MakeContractRevision(c.Request().Context(), &revision); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"revision": revision,
		})
	}
}

// ContractUpdateHandler is the handler for the /contract/:id update API.
func (s *ServerAPI) ContractUpdateHandler(c echo.Context) error {

	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid contract id"), nil)
	}

	var contractUpdate service.ContractUpdate
	if err := c.Bind(&contractUpdate); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	if contract, err := s.ServiceHandler.UpdateContract(c.Request().Context(), contractID, contractUpdate); err != nil {
		return ErrorResponseJSON(c, err, nil)
	} else {
		return SuccessResponseJSON(c, http.StatusOK, echo.Map{
			"contract": contract,
		})
	}
}
