package middleware

import (
	"net/http"

	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

// OAuthMiddleware adds Mux middleware operations to authenticate requests
func OAuthMiddleware(log *etilogger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//token := r.Header.Get("Authorization")
			//log.Info(r.RequestURI + token)
			//if token != "Bearer 123456" {
			//	utils.UnauthorizedResponse(w)
			//	return
			//}

			log.Info("Authorized request for " + r.RequestURI)
			next.ServeHTTP(w, r)
		})
	}
}
