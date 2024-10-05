package pet

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (p *Pet) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Method("GET", "/", http.HandlerFunc(p.GetAllPets))
	r.Method("POST", "/", http.HandlerFunc(p.PostAllPets))
	r.Method("GET", "/{petID}", http.HandlerFunc(p.GetPetByID))
	r.Method("POST", "/{petID}", http.HandlerFunc(p.PostPetByID))
	r.Method("DELETE", "/{petID}", http.HandlerFunc(p.DeletePetByID))

	return r
}
