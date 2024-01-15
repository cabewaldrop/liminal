package capture

import "net/http"

type captureResponseWriter struct {
	rw   http.ResponseWriter
	code int
}

func NewResponseWriter(rw http.ResponseWriter) *captureResponseWriter {
	return &captureResponseWriter{
		rw: rw,
	}
}

func (crw *captureResponseWriter) WriteHeader(code int) {
	crw.code = code
	crw.rw.WriteHeader(code)
}

func (crw *captureResponseWriter) Header() http.Header {
	return crw.rw.Header()
}

func (crw *captureResponseWriter) Write(bytes []byte) (int, error) {
	return crw.rw.Write(bytes)
}

func (crw *captureResponseWriter) Status() int {
	return crw.code
}
