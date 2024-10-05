package models

import "gorm.io/gorm"

type Pet struct {
	gorm.Model
	Name   string
	Type   string
	Family string
}
