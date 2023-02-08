package response_callback

import (
	"app/database/models"
	"log"
	"os"
	"strings"

	"github.com/jfcote87/esign/v2/model"
)

type ResponseCallbackInterface interface {
	SetParams(Params []models.ResponseMetaData)
	SetFile(File *os.File)
	SetOriginalData(OrignalData models.GoCusignEnvelopeDefinition)
	SetEnvolepeData(EnvelopeData model.EnvelopeDefinition)
	SendCallBack() error
}

func NewResponseCallback(Type string) ResponseCallbackInterface {

	switch strings.ToLower(Type) {
	case "cap":
		return &CapCallBack{}
	default:
		return nil
	}
}
