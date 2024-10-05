package pet

import "github.com/cisco-eti/sre-go-helloworld/pkg/metrics"

func PetTypeCounter(PetType string) {
	metrics.PetTypeCounter.WithLabelValues(PetType).Add(1)
}

func PetFamilyCounter(PetFamily string) {
	metrics.PetFamilyCounter.WithLabelValues(PetFamily).Add(1)
}
