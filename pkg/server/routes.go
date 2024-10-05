package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	slokmiddleware "github.com/slok/go-http-metrics/middleware"
	slokstd "github.com/slok/go-http-metrics/middleware/std"

	etimiddleware "github.com/cisco-eti/sre-go-helloworld/pkg/middleware"
)

// Router for helloworld server
func (s *Server) Router(
	extraMiddleware ...func(http.Handler) http.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type",
			"Content-Length", "Accept-Encoding", "X-CSRF-Token",
			etimiddleware.SharedAccessKeyHeader},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	}))

	for _, mw := range extraMiddleware {
		r.Use(mw)
	}

	r.Method("GET", "/", http.HandlerFunc(s.RootHandler))
	r.Method("GET", "/metrics", http.HandlerFunc(s.MetricsHandler))
	r.Method("GET", "/ping", http.HandlerFunc(s.PingHandler))
	r.Method("GET", "/s3", http.HandlerFunc(s.S3Handler))
	r.Method("GET", "/gci", http.HandlerFunc(s.GciHandler))
	r.Method("GET", "/docs", http.HandlerFunc(s.DocsHandler))
	r.Mount("/auth", s.v1auth.Router())

	authedV1Router := chi.NewRouter()
	authedV1Router.Use(etimiddleware.OAuthMiddleware(s.log))
	authedV1Router.Mount("/device", s.v1device.Router())
	authedV1Router.Mount("/pet", s.v1pet.Router())
	r.Mount("/v1", authedV1Router)

	return r
}

func MetricMiddleware() func(http.Handler) http.Handler {
	mdlw := slokmiddleware.New(slokmiddleware.Config{
		Recorder: prometheus.NewRecorder(prometheus.Config{}),
	})
	return slokstd.HandlerProvider("", mdlw)
}
