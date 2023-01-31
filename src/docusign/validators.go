package docusign

import (
	"github.com/jfcote87/esign/v2/model"
)

type EnvelopeErrors struct {
	Errors []string `json:"errors"`
}

func ValidEnvelopeCreate(envelopeDefinition *model.EnvelopeDefinition) (envelopeErrors EnvelopeErrors, err bool) {

	envelopeErrors = EnvelopeErrors{}
	// if envelopeDefinition.EmailSubject == "" {
	// 	envelopeErrors.Errors = append(envelopeErrors.Errors, "EmailSubject is required")
	// }

	// if envelopeDefinition.Recipients == nil || len(envelopeDefinition.Recipients.Signers) == 0 {
	// 	envelopeErrors.Errors = append(envelopeErrors.Errors, "At least one signer is required")
	// } else {
	// 	for i, signer := range envelopeDefinition.Recipients.Signers {
	// 		if signer.Email == "" {
	// 			envelopeErrors.Errors = append(envelopeErrors.Errors, fmt.Sprintf("Signer (index: %d) email is required", i))
	// 		}
	// 		if signer.Name == "" {
	// 			envelopeErrors.Errors = append(envelopeErrors.Errors, fmt.Sprintf("Signer (index: %d) name is required", i))
	// 		}
	// 		if signer.RecipientID == "" {
	// 			envelopeErrors.Errors = append(envelopeErrors.Errors, fmt.Sprintf("Signer (index: %d) recipient id is required", i))
	// 		}

	// 		// check one email is repeated in Singers
	// 		for j, signer2 := range envelopeDefinition.Recipients.Signers {
	// 			if i != j {
	// 				if signer.Email == signer2.Email {
	// 					envelopeErrors.Errors = append(envelopeErrors.Errors, fmt.Sprintf("Signer (index: %d <-> %d) email is repeated", i, j))
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// if envelopeDefinition.Documents == nil || len(envelopeDefinition.Documents) == 0 {
	// 	envelopeErrors.Errors = append(envelopeErrors.Errors, "At least one document is required")
	// } else {
	// 	for i, document := range envelopeDefinition.Documents {
	// 		if document.DocumentID == "" {
	// 			envelopeErrors.Errors = append(envelopeErrors.Errors, fmt.Sprintf("Document (index: %d) document id is required", i))
	// 		}
	// 		if document.DocumentBase64 == nil {
	// 			envelopeErrors.Errors = append(envelopeErrors.Errors, fmt.Sprintf("Document (index: %d) document base64 is required", i))
	// 		}
	// 	}
	// }

	return envelopeErrors, len(envelopeErrors.Errors) > 0
}
