package rww

import (
	"net/http"
)

type WriteFunc func([]byte) (int, error)

type inject struct {
	statusCode int
	headers    map[string]string
	writeFunc  WriteFunc
}

type intercepts map[int]inject

// ResponseWriterWrapper is a wrapper for instances of http.ResponseWrapper,
// allowing injection of Write functions and headers based on status codes
// declared by downstream http.Handlers.
type ResponseWriterWrapper struct {
	Status int
	Size   int
	http.ResponseWriter

	intercept     intercepts
	interceptFunc WriteFunc
}

func (ww *ResponseWriterWrapper) Header() http.Header {
	return ww.ResponseWriter.Header()
}

func (ww *ResponseWriterWrapper) Write(data []byte) (w int, err error) {
	if ww.interceptFunc != nil {
		w, err = ww.interceptFunc(data)
	} else {
		w, err = ww.ResponseWriter.Write(data)
	}
	ww.Size += w

	return
}

func (ww *ResponseWriterWrapper) WriteHeader(statusCode int) {
	if inject, exists := ww.intercept[statusCode]; exists {
		statusCode = inject.statusCode
		ww.interceptFunc = inject.writeFunc

		if inject.headers != nil {
			for k, v := range inject.headers {
				ww.Header().Add(k, v)
			}
		}
	}

	ww.Status = statusCode
	ww.ResponseWriter.WriteHeader(statusCode)
}

// AddIntercept defines an intercept for status code expectedStatus. If received,
// intendedStatus is returned instead. If WriteFunc w is specified, it is called
// instead of the wrapped ResponseWriter.Write. Any headers provided are also
// added to the response.
func (ww *ResponseWriterWrapper) AddIntercept(expectedStatus, intendedStatus int,
	w WriteFunc, headers map[string]string) {

	ww.intercept[expectedStatus] = inject{intendedStatus, headers, w}
}

// New returns a new wrapper for the provided http.ResponseWriter
func New(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{200, 0, w, make(intercepts), nil}
}
