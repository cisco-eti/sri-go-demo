package middleware

import (
	"net/http"

	"github.com/cisco-eti/sre-go-helloworld/pkg/utils"
	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

const (
	SharedAccessKeyHeader = "X-API-ACCESS-KEY"
	sharedAccessKeySecret = "c94bcd16-5e7c-4d41-95b0-70c9610e5663"
)

// OAuthMiddleware adds Mux middleware operations to authenticate requests
func SharedKeyMiddleware(log *etilogger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(SharedAccessKeyHeader) != sharedAccessKeySecret {
				log.Info("Shared key not authorized")
				utils.UnauthorizedResponse(w)
				return
			}

			log.Info("Authorized request for " + r.RequestURI)
			next.ServeHTTP(w, r)
		})
	}
}
