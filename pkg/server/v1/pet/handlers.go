package pet

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/cisco-eti/sre-go-helloworld/pkg/models"
	"github.com/cisco-eti/sre-go-helloworld/pkg/utils"
)

// Get godoc
// @Summary Get All Pets
// @Description Get all pets in the pet store
// @Tags getallpets
// @Produce json
// @Success 200 {object} models.Pets
// @Failure 404 {object} models.Error
// @Router /pet [get]
func (p *Pet) GetAllPets(w http.ResponseWriter, r *http.Request) {
	var pets []models.Pet
	p.db.Find(&pets)
	utils.OKResponse(w, pets)
}

func (p *Pet) PostAllPets(w http.ResponseWriter, r *http.Request) {
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	var newPet models.Pet
	err := json.NewDecoder(r.Body).Decode(&newPet)
	if err != nil {
		utils.BadRequestResponse(w, err.Error())
		return
	}

	p.db.Create(&models.Pet{
		Name:   newPet.Name,
		Family: newPet.Family,
		Type:   newPet.Type,
	})

	PetFamilyCounter(newPet.Family)
	PetTypeCounter(newPet.Type)

	utils.CreatedResponse(w, newPet)
}

// Get godoc
// @Summary Get Pet by ID
// @Description Get one Pet by ID
// @Tags getpetid
// @Produce json
// @Success 200 {object} models.Pet
// @Failure 404 {object} models.Error
// @Router /pet [get]
func (p *Pet) GetPetByID(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	p.log.Info("GetPetByID PetID:" + petID)

	var pet models.Pet
	p.db.Find(&pet, petID)
	utils.OKResponse(w, pet)
}

// Get godoc
// @Summary Update Pet by ID
// @Description Update one Pet by ID
// @Tags updatepetid
// @Success 200 {object}
// @Failure 404 {object} models.Error
// @Router /pet [get]
func (p *Pet) PostPetByID(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	p.log.Info("PostPetByID PetID:" + petID)

	// Declare a new Pet struct.
	var newPet models.Pet
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		utils.BadRequestResponse(w, err.Error())
		return
	}

	p.db.Save(&newPet)

	PetFamilyCounter(newPet.Family)
	PetTypeCounter(newPet.Type)

	utils.CreatedResponse(w, newPet)
}

// Get godoc
// @Summary Delete Pet by ID
// @Description Delete Pet by ID
// @Tags deletepetid
// @Success 200 {object}
// @Failure 404 {object} models.Error
// @Router /pet [delete]
func (p *Pet) DeletePetByID(w http.ResponseWriter, r *http.Request) {
	petID := chi.URLParam(r, "petID")
	p.log.Info("DeletePetByID PetID:" + petID)

	p.db.Delete(&models.Pet{}, petID)

	utils.OKResponse(w, fmt.Sprintf("Pet %s successfully updated", petID))
}
