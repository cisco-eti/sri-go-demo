package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Counter counts operations of a specified type
	DeviceCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eti_apps_helloworld_device_counter",
			Help: "This is Device counter",
		},
		[]string{"device"},
	)

	// Counter counts operations of a specified type
	PetFamilyCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eti_apps_helloworld_pet_family_counter",
			Help: "This is Pet Family counter",
		},
		[]string{"petfamily"},
	)

	// Counter counts operations of a specified type
	PetTypeCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eti_apps_helloworld_pet_type_counter",
			Help: "This is Pet Type counter",
		},
		[]string{"pettype"},
	)
)

func init() {
	prometheus.MustRegister(DeviceCounter)
	prometheus.MustRegister(PetFamilyCounter)
	prometheus.MustRegister(PetTypeCounter)
}
