package response

import (
	"encoding/json"
	"net/http"
)

func WriteResponse(content any, status int, writer http.ResponseWriter) error {
	writer.Header().Set("Content-Type", "application/json")

	raw, err := json.Marshal(content)
	if err != nil {
		return err
	}

	writer.WriteHeader(status)
	_, err = writer.Write(raw)
	if err != nil {
		return err
	}

	return nil
}

func WriteOkResponse(content any, writer http.ResponseWriter) error {
	return WriteResponse(content, http.StatusOK, writer)
}

func WriteBadRequestResponse(content any, writer http.ResponseWriter) error {
	return WriteResponse(content, http.StatusBadRequest, writer)
}

func WriteNotFoundResponse(content any, writer http.ResponseWriter) error {
	return WriteResponse(content, http.StatusNotFound, writer)
}

func WriteForbiddenResponse(content any, writer http.ResponseWriter) error {
	return WriteResponse(content, http.StatusForbidden, writer)
}

func WriteUnauthorizedResponse(content any, writer http.ResponseWriter) error {
	return WriteResponse(content, http.StatusUnauthorized, writer)
}

func WriteInternalErrorResponse(content any, writer http.ResponseWriter) error {
	return WriteResponse(content, http.StatusInternalServerError, writer)
}
