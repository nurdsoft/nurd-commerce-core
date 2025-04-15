// Package http for conx.
package encode

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"
	"io"
	"log"
	"net/http"
	"strconv"
	// authError "github.com/nurdsoft/nurd-commerce-core/internal/auth/errors"
	httpInternal "github.com/nurdsoft/nurd-commerce-core/internal/transport/http"
	appError "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	// "github.com/nurdsoft/nurd-commerce-core/shared/pagination"
	"github.com/gabriel-vasile/mimetype"
)

const (
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
)

func CSVResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	ioi := response.([]byte)

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("sample_campaign_users.csv"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("content-length", fmt.Sprintf("%d", binary.Size(ioi)))
	_, err := io.Copy(w, bytes.NewReader(ioi))
	if err != nil {
		return err
	}
	return nil
}

// func ResponseWithPagination(ctx context.Context, w http.ResponseWriter, r interface{}) error {
// 	res, _ := r.(*pagination.ResponseWithPagination) //nolint:errcheck

// 	resp := &httpInternal.Response{Error: &httpInternal.Error{}, Pagination: res.Pagination, Data: res.Data}

// 	w.Header().Set(contentTypeHeader, jsonContentType)

// 	return encodeJSONToWriter(w, resp)
// }

func Response(ctx context.Context, w http.ResponseWriter, r interface{}) error {

	resp := &httpInternal.Response{Error: &httpInternal.Error{}, Data: r}

	w.Header().Set(contentTypeHeader, jsonContentType)

	return encodeJSONToWriter(w, resp)
}

func ViewFileResponse(_ context.Context, w http.ResponseWriter, r interface{}) error {
	data, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", r))
	if err != nil {
		return err
	}

	w.Header().Set(contentTypeHeader, "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	//data := []byte("this is some data stored as a byte slice in Go Lang!")

	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)

	_, err = io.Copy(w, reader)

	return err
}

func DownloadPdfFileResponse(_ context.Context, w http.ResponseWriter, r interface{}) error {
	resp := r.(*httpInternal.DownloadFromBytesResponseWithFilename)

	w.Header().Set(contentTypeHeader, "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprint(len(resp.Data)))
	if resp.Filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%s%s`, resp.Filename, ".pdf"))
	} else {
		w.Header().Set("Content-Disposition", "attachment")
	}

	_, err := w.Write(resp.Data)
	return err
}

func ViewImageResponse(_ context.Context, w http.ResponseWriter, r interface{}) error {
	data, ok := r.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert into byte array while rendering image")
	}

	// set the default MIME type to send, use mimetype library which supports a much greater range of file types.
	mime := mimetype.Detect(data)

	// Generate the server headers
	w.Header().Set("Content-Type", mime.String())
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))

	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)

	_, err := io.Copy(w, reader)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return err
}

func DownloadFileResponse(_ context.Context, w http.ResponseWriter, r interface{}) error {
	resp := r.(*httpInternal.ResponseWithFilename)
	data, ok := resp.Data.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert into byte array while downloading file")
	}

	// set the default MIME type to send, use mimetype library which supports a much greater range of file types.
	mime := mimetype.Detect(data)
	fileExtension := mime.Extension()

	// Set the Content-Disposition header with the modified filename
	w.Header().Set("Content-Type", mime.String())

	if resp.Filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%s%s`, resp.Filename, fileExtension))
	} else {
		w.Header().Set("Content-Disposition", "attachment")
	}
	// Generate the server headers
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

	//TODO: the below header can be added later if needed or it should be removed
	//w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

	// convert byte slice to io.Reader
	reader := bytes.NewReader(data)

	_, err := io.Copy(w, reader)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return err
}

func ExcelResponse(_ context.Context, w http.ResponseWriter, r interface{}) error {
	resp := r.(*httpInternal.ExcelResponseData)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.Filename))

	err := resp.Data.Write(w)
	if err != nil {
		return err
	}

	return err
}

func Error(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set(contentTypeHeader, jsonContentType)

	var errCode int
	var errMsg string

	resp := httpInternal.Response{
		Data:  &struct{}{},
		Error: &httpInternal.Error{StatusCode: errCode, Message: errMsg},
	}

	// Handle the error if it's from the database
	convertedErr := dbErrors.HandleDbError(err)

	var apiErr *appError.APIError
	if apiError, ok := appError.IsAPIError(convertedErr); ok {
		apiErr = apiError
		errCode = apiErr.StatusCode
		resp = httpInternal.Response{
			Data:  &struct{}{},
			Error: &httpInternal.Error{StatusCode: apiErr.StatusCode, ErrorCode: "BAPI_" + apiErr.ErrorCode, Message: apiErr.Message},
		}
	} else {
		// Log full details
		log.Printf("Unknown error: %s", err.Error())
		errCode = http.StatusInternalServerError
		resp = httpInternal.Response{
			Data:  &struct{}{},
			Error: &httpInternal.Error{StatusCode: http.StatusInternalServerError, ErrorCode: "BAPI_INTERNAL_ERROR", Message: "An internal error occurred."},
		}
	}

	w.WriteHeader(errCode)

	_ = encodeJSONToWriter(w, resp) // nolint: errcheck
}

func encodeJSONToWriter(w io.Writer, message interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(message)
}
