## ResponseWriter Wrapper [![GoDoc](https://godoc.org/github.com/ragnar-johannsson/rww?status.png)](https://godoc.org/github.com/ragnar-johannsson/rww)

`rww` is a wrapper for `http.ResponseWriter` enabling response manipulation based on status codes specified downstream in the middleware chain, as well as getting some basic information about the response being sent out.

### Example usage

Logging handler using rww to report final status and response content size:

```go
logger := func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ww := rww.New(w)

        h.ServeHTTP(ww, r)

        log.Printf(
            "- %s - \"%s %s %s\" %d %d",
            r.RemoteAddr,
            r.Method,
            r.URL.RequestURI(),
            r.Proto,
            ww.Status, // Get both the response status and the
            ww.Size,   // content length from the wrapper
        )
    })
}

helloWorld := func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello world!")
}

http.HandleFunc("/", helloWorld)
http.ListenAndServe(":8080", logger(http.DefaultServeMux))
```


Redirect handler using rww to handle response based on downstream status:

```go
redirector := func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ww := rww.New(w)
        u, _ := url.Parse("http://other.host/")
        u.Path = r.URL.EscapedPath()

        ww.AddIntercept(
            // Expected status code
            http.StatusNotFound,
            // Intended status code
            http.StatusTemporaryRedirect,
            // Injected http.ResponseWriter.Write() func
            nil,
            // Headers to add to the response
            map[string]string{
                "Location": u.String(),
            },
        )
    })
}

staticFileHandler := http.FileServer(http.Dir("/path/to/files"))
http.ListenAndServe(":8080", redirector(staticFileHandler))
```

### License

BSD 2-Clause. See the LICENSE file for details.
