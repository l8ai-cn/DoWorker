package sessionapi

import (
	"fmt"
	"net/http"
	"time"
)

const sessionStreamWriteTimeout = 60 * time.Second

func prepareSessionStreamWriter(writer http.ResponseWriter) error {
	return http.NewResponseController(writer).
		SetWriteDeadline(time.Now().Add(sessionStreamWriteTimeout))
}

func writeSessionStreamFrame(writer http.ResponseWriter, frame string) error {
	if err := prepareSessionStreamWriter(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprint(writer, frame); err != nil {
		return err
	}
	return http.NewResponseController(writer).Flush()
}
