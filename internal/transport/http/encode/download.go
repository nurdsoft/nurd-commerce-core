package encode

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
)

type FileObject struct {
	Data     []byte
	FileName string
}

func DownloadFileResponseFromFileObject(_ context.Context, w http.ResponseWriter, r interface{}) error {
	fileObject, ok := r.(*FileObject)
	if !ok {
		return fmt.Errorf("cannot convert into byte array while downloading file")
	}

	// set the default MIME type to send, use mimetype library which supports a much greater range of file types.
	mime := mimetype.Detect(fileObject.Data)

	fileSize := len(string(fileObject.Data))

	// Generate the server headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", mime.String())
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileObject.FileName))
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Content-Length", strconv.Itoa(fileSize))
	w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

	// convert byte slice to io.Reader
	reader := bytes.NewReader(fileObject.Data)

	_, err := io.Copy(w, reader)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return err
}
