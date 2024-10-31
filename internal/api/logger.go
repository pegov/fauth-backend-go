package api

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	RequestBodyMaxSize  = 2 * 1024 * 1024 // 2MB
	ResponseBodyMaxSize = 2 * 1024 * 1024 // 2MB
)

func NewSlogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			query := r.URL.RawQuery

			// dump request body
			br := newBodyReader(r.Body, RequestBodyMaxSize, true)
			r.Body = br

			// dump response body
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			bw := newBodyWriter(ResponseBodyMaxSize)
			ww.Tee(bw)

			defer func() {
				params := map[string]string{}
				for i, k := range chi.RouteContext(r.Context()).URLParams.Keys {
					params[k] = chi.RouteContext(r.Context()).URLParams.Values[i]
				}

				status := ww.Status()
				method := r.Method
				host := r.Host
				route := chi.RouteContext(r.Context()).RoutePattern()
				end := time.Now()
				latency := end.Sub(start)

				attrs := []slog.Attr{
					{
						Key: "req",
						Value: slog.GroupValue([]slog.Attr{
							slog.Time("ts", start),
							slog.String("method", method),
							slog.String("host", host),
							slog.String("path", path),
							slog.String("query", query),
							// slog.Any("params", params),
							slog.String("route", route),
						}...),
					},
					{
						Key: "res",
						Value: slog.GroupValue([]slog.Attr{
							slog.Time("ts", end),
							slog.Duration("latency", latency),
							slog.Int("status", status),
						}...),
					},
				}

				logger.LogAttrs(
					r.Context(),
					slog.LevelInfo,
					strconv.Itoa(status),
					attrs...,
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

type bodyWriter struct {
	body    *bytes.Buffer
	maxSize int
}

// implements io.Writer
func (w *bodyWriter) Write(b []byte) (int, error) {
	if w.body.Len()+len(b) > w.maxSize {
		return w.body.Write(b[:w.maxSize-w.body.Len()])
	}
	return w.body.Write(b)
}

func newBodyWriter(maxSize int) *bodyWriter {
	return &bodyWriter{
		body:    bytes.NewBufferString(""),
		maxSize: maxSize,
	}
}

type bodyReader struct {
	io.ReadCloser
	body    *bytes.Buffer
	maxSize int
	bytes   int
}

// implements io.Reader
func (r *bodyReader) Read(b []byte) (int, error) {
	n, err := r.ReadCloser.Read(b)
	if r.body != nil {
		if r.body.Len()+n > r.maxSize {
			r.body.Write(b[:r.maxSize-r.body.Len()])
		} else {
			r.body.Write(b[:n])
		}
	}
	r.bytes += n
	return n, err
}

func newBodyReader(reader io.ReadCloser, maxSize int, recordBody bool) *bodyReader {
	var body *bytes.Buffer
	if recordBody {
		body = bytes.NewBufferString("")
	}
	return &bodyReader{
		ReadCloser: reader,
		body:       body,
		maxSize:    maxSize,
		bytes:      0,
	}
}
