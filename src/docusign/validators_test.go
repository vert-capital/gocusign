package docusign

import (
	"testing"

	"github.com/jfcote87/esign/v2/model"
	"github.com/stretchr/testify/assert"
)

func TestValidEnvelopeCreateBlank(t *testing.T) {
	var envelope model.EnvelopeDefinition

	_, errBool := ValidEnvelopeCreate(&envelope)

	if errBool {
		assert.Equal(t, true, errBool, "Expected error")
	}
}

func TestValidEnvelopeCreateDuplicatedMail(t *testing.T) {
	envelope := model.EnvelopeDefinition{
		EmailSubject: "Test",
		Recipients: &model.Recipients{
			Signers: []model.Signer{
				{
					Email:       "test@test.com",
					Name:        "Test",
					RecipientID: "1",
				},
				{
					Email:       "test@test.com",
					Name:        "Test 2",
					RecipientID: "2",
				},
			},
		},
	}

	envelopeErrors, err := ValidEnvelopeCreate(&envelope)
	assert.Equal(t, true, err)
	assert.Contains(t, envelopeErrors.Errors, "Signer (index: 0 <-> 1) email is repeated")
}

func TestValidEnvelopeCreateWithoutDocument(t *testing.T) {
	envelope := model.EnvelopeDefinition{
		EmailSubject: "Test",
	}

	envelopeErrors, err := ValidEnvelopeCreate(&envelope)
	assert.Equal(t, true, err)
	assert.Contains(t, envelopeErrors.Errors, "At least one document is required")
}

func TestValidEnvelopeCreateWithoutBase64(t *testing.T) {
	envelope := model.EnvelopeDefinition{
		EmailSubject: "Test",
		Documents: []model.Document{
			{
				DocumentID: "",
			},
		},
	}

	envelopeErrors, err := ValidEnvelopeCreate(&envelope)
	assert.Equal(t, true, err)
	assert.Contains(t, envelopeErrors.Errors, "Document (index: 0) document base64 is required")
	assert.Contains(t, envelopeErrors.Errors, "Document (index: 0) document id is required")
}
