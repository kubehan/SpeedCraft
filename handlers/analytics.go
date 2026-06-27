package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"speedcraft/models"
)

// statusRecorder captures status code from inner handler.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// analyticsBuffer buffers stats writes to avoid hammering DB on every request.
type analyticsEntry struct {
	path       string
	durationMs int64
	status     int
}

var (
	analyticsCh   = make(chan analyticsEntry, 1024)
	analyticsOnce sync.Once
)

// startAnalyticsWorker launches a single goroutine that drains the channel and writes DB.
// Buffers are non-blocking: if channel is full, the entry is dropped (preferable to blocking request).
func startAnalyticsWorker() {
	analyticsOnce.Do(func() {
		go func() {
			for e := range analyticsCh {
				_ = models.RecordPageView(e.path, e.durationMs, e.status)
			}
		}()
	})
}

// shouldTrackPath returns true if the path is a public page we want to track.
func shouldTrackPath(path string) bool {
	// Skip admin, API, static, ad clicks
	if strings.HasPrefix(path, "/admin") ||
		strings.HasPrefix(path, "/api") ||
		strings.HasPrefix(path, "/static") ||
		strings.HasPrefix(path, "/ad/") ||
		path == "/favicon.ico" ||
		path == "/robots.txt" ||
		path == "/sitemap.xml" ||
		path == "/version" {
		return false
	}
	return true
}

// AnalyticsMiddleware wraps a handler to record view count + response time.
func AnalyticsMiddleware(next http.Handler) http.Handler {
	startAnalyticsWorker()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip non-trackable paths
		if !shouldTrackPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		rec := &statusRecorder{ResponseWriter: w, statusCode: 200}
		start := time.Now()
		next.ServeHTTP(rec, r)
		duration := time.Since(start).Milliseconds()

		// Normalize path: collapse /blog/xxx → /blog/* and /page/xxx → /page/*
		path := r.URL.Path
		if strings.HasPrefix(path, "/blog/") && len(path) > 6 {
			path = "/blog/*"
		} else if strings.HasPrefix(path, "/page/") && len(path) > 6 {
			// avoid /page/.../raw being merged separately
			if strings.HasSuffix(path, "/raw") {
				path = "/page/*/raw"
			} else {
				path = "/page/*"
			}
		}

		// Non-blocking send
		select {
		case analyticsCh <- analyticsEntry{path: path, durationMs: duration, status: rec.statusCode}:
		default:
			// dropped
		}
	})
}
