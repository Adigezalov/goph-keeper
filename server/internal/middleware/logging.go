package middleware

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/logger"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
	hijacker   http.Hijacker
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     0,
	}
	if hijacker, ok := w.(http.Hijacker); ok {
		rw.hijacker = hijacker
	}
	return rw
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusCode == 0 {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.hijacker != nil {
		return rw.hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		route := mux.CurrentRoute(r)
		path := r.URL.Path
		if route != nil {
			if template, err := route.GetPathTemplate(); err == nil && template != "" {
				path = template
			}
		}

		userID, _ := GetUserIDFromContext(r.Context())
		sessionID, _ := GetSessionIDFromContext(r.Context())

		logger.Log.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       path,
			"remote":     r.RemoteAddr,
			"user_agent": r.UserAgent(),
			"user_id":    userID,
			"session_id": sessionID,
		}).Info("Incoming request")

		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		fields := logrus.Fields{
			"method":     r.Method,
			"path":       path,
			"status":     rw.statusCode,
			"duration":   duration.String(),
			"size":       rw.size,
			"user_id":    userID,
			"session_id": sessionID,
		}

		if rw.statusCode >= 500 {
			logger.Log.WithFields(fields).Error("Request completed with server error")
		} else if rw.statusCode >= 400 {
			logger.Log.WithFields(fields).Warn("Request completed with client error")
		} else {
			logger.Log.WithFields(fields).Info("Request completed successfully")
		}
	})
}
