package auth

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (a *Auth) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Method("GET", "/login", http.HandlerFunc(a.LoginRedirectHandler))
	r.Method("GET", "/login/token", http.HandlerFunc(a.LoginExchangeHandler))
	r.Method("GET", "/logout", http.HandlerFunc(a.LogoutHandler))

	return r
}
