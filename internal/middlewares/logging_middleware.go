package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NorskHelsenett/ror/pkg/rlog"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rlog.Info(fmt.Sprintf("Started %s %s", r.Method, r.URL.Path))

		// Call the next handler
		next.ServeHTTP(w, r)

		rlog.Info(fmt.Sprintf("Completed %s in %v", r.URL.Path, time.Since(start)))
	})
}
