package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/vitistack/common/pkg/loggers/vlog"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		vlog.Info(fmt.Sprintf("Started %s %s", r.Method, r.URL.Path))

		// Call the next handler
		next.ServeHTTP(w, r)

		vlog.Info(fmt.Sprintf("Completed %s in %v", r.URL.Path, time.Since(start)))
	})
}
