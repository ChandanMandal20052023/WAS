package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"was/backend/internal/utils"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[PANIC RECOVERED] %v\n%s", rec, debug.Stack())
				utils.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

