package middleware

import "net/http"

// CustomExampleMiddleware is a placeholder for your own custom middleware.
// It logs a message before and after handling a request.
func CustomExampleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("INFO: Custom middleware - before request")
		next.ServeHTTP(w, r)
		// fmt.Println("INFO: Custom middleware - after request")
	})
}
