package device

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/cisco-eti/sre-go-helloworld/pkg/metrics"
	"github.com/cisco-eti/sre-go-helloworld/pkg/utils"
)

// Get godoc
// @Summary Get Device Zone
// @Description get device zone by device Id
// @ID get-device-zone
// @Tags DeviceZone
// @Accept json
// @Produce json
// @Param deviceId path string true "Device ID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /deviceZone/{deviceId} [get]
func (d *Device) GetDeviceHandler(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceID")

	d.log.Info("DeviceZoneHandler deviceId:" + deviceID)
	metrics.DeviceCounter.WithLabelValues(deviceID).Add(1)

	var err error
	switch deviceID {
	case "A":
		err = utils.OKResponse(w, "Plumbing")
	case "B":
		err = utils.OKResponse(w, "Gardening")
	case "C":
		err = utils.OKResponse(w, "Lighting")
	case "D":
		err = utils.OKResponse(w, "FrontDesk")
	default:
		err = utils.NotFoundResponse(w, "Unknown Device Specified")
	}

	if err != nil {
		d.log.Warn("error: %s", err)
	}
}
