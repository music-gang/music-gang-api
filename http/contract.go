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

	return handleContractCall(c, s, contractID, 0)
}

// ContractCallRevHandler is the handler for the /contract/:id/call/:rev API.
func (s *ServerAPI) ContractCallRevHandler(c echo.Context) error {

	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid contract id"), nil)
	}

	revisionNumber, err := strconv.ParseUint(c.Param("rev"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid revision number"), nil)
	}

	return handleContractCall(c, s, contractID, entity.RevisionNumber(revisionNumber))
}

// handleContractCall handles the /contract/:id/call and /contract/:id/call/:rev business logic.
func handleContractCall(c echo.Context, s *ServerAPI, contractID int64, revisionNumber entity.RevisionNumber) (err error) {

	var revision *entity.Revision

	if revisionNumber != 0 {
		revision, err = s.ContractSearchService.FindRevisionByContractAndRev(c.Request().Context(), contractID, revisionNumber)
		if err != nil {
			s.LogService.ReportError(c.Request().Context(), err)
			return ErrorResponseJSON(c, err, nil)
		}
	} else {

		contract, err := s.ContractSearchService.FindContractByID(c.Request().Context(), contractID)
		if err != nil {
			s.LogService.ReportError(c.Request().Context(), err)
			return ErrorResponseJSON(c, err, nil)
		}

		revision = contract.LastRevision
	}

	result, err := s.VmCallableService.ExecContract(c.Request().Context(), revision)
	if err != nil {
		s.LogService.ReportError(c.Request().Context(), err)
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"result": result,
	})
}

// ContractHandler is the handler for the /contract/:id search API.
func (s *ServerAPI) ContractHandler(c echo.Context) error {

	contractID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid contract id"), nil)
	}

	return handleContract(c, s, contractID)
}

// handleContract handles the /contract/:id search business logic.
func handleContract(c echo.Context, s *ServerAPI, contractID int64) error {

	contract, err := s.ContractSearchService.FindContractByID(c.Request().Context(), contractID)
	if err != nil {
		s.LogService.ReportError(c.Request().Context(), err)
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"contract": contract,
	})
}

// ContractCreateHandler is the handler for the /contract create API.
func (s *ServerAPI) ContractCreateHandler(c echo.Context) error {

	var contractParams entity.Contract
	if err := c.Bind(&contractParams); err != nil {
		return ErrorResponseJSON(c, apperr.Errorf(apperr.EINVALID, "invalid request"), nil)
	}

	return handleContractCreate(c, s, &contractParams)
}

// handleContractCreate handles the /contract create business logic.
func handleContractCreate(c echo.Context, s *ServerAPI, contract *entity.Contract) error {

	if err := s.VmCallableService.CreateContract(c.Request().Context(), contract); err != nil {
		s.LogService.ReportError(c.Request().Context(), err)
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"contract": contract,
	})
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

	return handleContractMakeRevision(c, s, &revision)
}

// handleContractMakeRevision handles the /contract/:id/revision business logic.
func handleContractMakeRevision(c echo.Context, s *ServerAPI, revision *entity.Revision) error {

	if err := s.VmCallableService.MakeRevision(c.Request().Context(), revision); err != nil {
		s.LogService.ReportError(c.Request().Context(), err)
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"revision": revision,
	})
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

	return handleContractUpdate(c, s, contractID, contractUpdate)
}

// handleContractUpdate handles the /contract/:id update business logic.
func handleContractUpdate(c echo.Context, s *ServerAPI, contractID int64, contractUpdate service.ContractUpdate) error {

	contract, err := s.VmCallableService.UpdateContract(c.Request().Context(), contractID, contractUpdate)
	if err != nil {
		s.LogService.ReportError(c.Request().Context(), err)
		return ErrorResponseJSON(c, err, nil)
	}

	return SuccessResponseJSON(c, http.StatusOK, echo.Map{
		"contract": contract,
	})
}
