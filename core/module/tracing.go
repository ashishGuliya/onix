package module

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing the header
func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// tracingWrapper wraps http.handler execution with tracing
func tracingWrapper(name string, tp string, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer(tp)
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()

		startTime := time.Now()

		// Wrap ResponseWriter to capture status code
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next middleware/handler
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Measure response time
		duration := time.Since(startTime)

		// Add tracing attributes
		span.SetAttributes(
			attribute.String(tp+".name", name),
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.Path), // Avoid sensitive query params
			attribute.Int("http.status_code", rw.statusCode),
			attribute.String("http.status_text", http.StatusText(rw.statusCode)),
			attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
		)
	})
}
