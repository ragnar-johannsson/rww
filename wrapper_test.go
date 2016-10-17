package rww

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSize(t *testing.T) {
	finalSize := 0
	out := "See you on the other side"

	sizeHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := New(w)
			h.ServeHTTP(ww, r)
			finalSize = ww.Size
		})
	}

	responseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, out)
	})

	ts := httptest.NewServer(sizeHandler(responseHandler))
	defer ts.Close()

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("Error making request: ", err)
	}

	if finalSize != len(out)+1 {
		t.Errorf("Wrong size reported: ", finalSize)
	}
}

func TestStatus(t *testing.T) {
	finalStatus := 0

	statusHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := New(w)
			h.ServeHTTP(ww, r)
			finalStatus = ww.Status
		})
	}

	responseHandler := http.NotFoundHandler()

	ts := httptest.NewServer(statusHandler(responseHandler))
	defer ts.Close()

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("Error making request: ", err)
	}

	if finalStatus != 404 {
		t.Errorf("Wrong status reported: ", finalStatus)
	}
}

func TestAddInterceptWriteFunc(t *testing.T) {
	interceptHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := New(w)
			ww.AddIntercept(
				http.StatusNotFound,
				http.StatusOK,
				func(d []byte) (int, error) {
					w.Write([]byte("Hello\n"))
					return len(d), nil
				},
				nil,
			)

			h.ServeHTTP(ww, r)
		})
	}

	responseHandler := http.NotFoundHandler()

	ts := httptest.NewServer(interceptHandler(responseHandler))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("Error making request: ", err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Incorrect status code received in response: ", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("Error reading response: ", err)
	}

	received := string(body)
	expected := "Hello\n"
	if strings.Compare(received, expected) != 0 {
		t.Errorf("Incorrect body received in response: ", received)
	}

}

func TestAddInterceptWriteFuncWithHeaders(t *testing.T) {
	interceptHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := New(w)
			ww.AddIntercept(
				http.StatusNotFound,
				http.StatusOK,
				func(d []byte) (int, error) {
					w.Write([]byte("\n"))
					return len(d), nil
				},
				map[string]string{
					"X-Header": "X-Content",
				},
			)

			h.ServeHTTP(ww, r)
		})
	}

	responseHandler := http.NotFoundHandler()

	ts := httptest.NewServer(interceptHandler(responseHandler))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("Error making request: ", err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Incorrect status code received in response: ", res.StatusCode)
	}

	received := res.Header.Get("X-Header")
	expected := "X-Content"
	if strings.Compare(received, expected) != 0 {
		t.Errorf("Header not received correctly with response: ", received)
	}

}
