// Package http for transport.
package http

import (
	// "github.com/nurdsoft/nurd-commerce-core/shared/pagination"
	"github.com/xuri/excelize/v2"
)

// Default HTTP Response Object
// swagger:model DefaultResponse
type Response struct {
	Error *Error      `json:"error"`
	Data  interface{} `json:"data"`
}

// Default Error Object
// swagger:model DefaultError
type Error struct {
	ErrorCode  string `json:"error_code"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type ResponseWithFilename struct {
	// use this to set fileBytes
	Data interface{}
	// use this to set filename in Content-Disposition header
	Filename string
}

type DownloadFromBytesResponseWithFilename struct {
	// use this to set fileBytes
	Data []byte
	// use this to set filename in Content-Disposition header
	Filename string
}

type ExcelResponseData struct {
	Data     *excelize.File
	Filename string
}
