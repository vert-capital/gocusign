package models

import (
	"github.com/jfcote87/esign/v2/model"
	"gorm.io/gorm"
)

type ResponseMetaData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResponseCallback struct {
	ResponseType     string             `json:"response_type"`
	ResponseMetaData []ResponseMetaData `json:"response_meta_data"`
}

type GoCusignEnvelopeDefinition struct {
	model.EnvelopeDefinition
	ResponseCallback ResponseCallback `json:"response_callback"`
}

type Envelope struct {
	gorm.Model
	ID                  string `sql:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Status              string `sql:"type:varchar(25);not null"`
	OriginalRequestData string `sql:"type:jsonb;not null"`
	JsonData            string `sql:"type:jsonb;not null"`
}
