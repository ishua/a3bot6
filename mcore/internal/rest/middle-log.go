package rest

import (
	"bytes"
	"github.com/ishua/a3bot6/mcore/pkg/logger"
	"io"
	"net/http"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func middleLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger.Debugf("req method: %s, path: %s", r.Method, r.URL.EscapedPath())

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Debugf("Error reading body: %v\n", err)
			return
		}

		logger.Debugf("Request Body: %s", string(bodyBytes))

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(lrw, r)

		logger.Debugf("Response Status: %d", lrw.statusCode)
		logger.Debugf("Response Body: %s", lrw.body.String())
	})
}
