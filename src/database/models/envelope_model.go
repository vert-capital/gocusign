package models

import "gorm.io/gorm"

type Envelope struct {
	gorm.Model
	ID       string `sql:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Status   string `sql:"type:varchar(25);not null"`
	JsonData string `sql:"type:jsonb;not null"`
}
