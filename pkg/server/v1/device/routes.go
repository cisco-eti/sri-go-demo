package device

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (d *Device) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Method("GET", "/{deviceID}", http.HandlerFunc(d.GetDeviceHandler))

	return r
}
