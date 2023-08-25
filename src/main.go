package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"app/database"
	"app/response_callback"

	dbModels "app/database/models"

	"github.com/gorilla/mux"
	"github.com/jfcote87/esign/v2/model"

	"app/docusign"
)

func CreateEnvelopeHandler(w http.ResponseWriter, r *http.Request) {

	var envelope dbModels.GoCusignEnvelopeDefinition

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	for key := range envelope.Documents {
		if envelope.Documents[key].FileExtension == "" &&
			strings.Contains(envelope.Documents[key].FileExtension, "docx") &&
			len(envelope.Documents[key].FileExtension) > 4 {
			envelope.Documents[key].FileExtension = "docx"
		}
	}

	err = json.Unmarshal(body, &envelope)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON"))
		return
	}

	structureErrors, errBool := docusign.ValidEnvelopeCreate(&envelope.EnvelopeDefinition)

	if errBool {
		jsonStructureErrors, _ := json.Marshal(structureErrors)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(jsonStructureErrors))
		return
	}

	if envelope.Status == "" {
		envelope.Status = "sent"
	}

	envSummary, err := docusign.DocusignConfig.CreateEnvelop(&envelope)
	if err != nil {
		json_resp, _ := json.Marshal(map[string]interface{}{"error_type": "docusign_request", "error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(json_resp)
		return
	}

	resp, err := json.Marshal(envSummary)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	var envelope model.EnvelopeDefinition
	var envelopeData dbModels.Envelope
	var envelopeDataGoCUsign dbModels.GoCusignEnvelopeDefinition

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	json.Unmarshal(body, &envelope)

	log.Println("Callback document: " + envelope.EnvelopeID)

	if database.DB != nil {

		log.Println("has database")

		envelopeJson, _ := json.Marshal(envelope)

		database.DB.Model(&dbModels.Envelope{}).Where("id = ?", envelope.EnvelopeID).Updates(dbModels.Envelope{
			Status:   envelope.Status,
			JsonData: string(envelopeJson),
		})

		database.DB.Model(&dbModels.Envelope{}).Where("id = ?", envelope.EnvelopeID).First(&envelopeData)

		json.Unmarshal([]byte(envelopeData.OriginalRequestData), &envelopeDataGoCUsign)

		log.Println("status: " + envelope.Status)
		log.Println("envelopeDataGoCUsign.ResponseCallback.ResponseType" + envelopeDataGoCUsign.ResponseCallback.ResponseType)

		if envelopeDataGoCUsign.ResponseCallback.ResponseType != "" && envelope.Status == "completed" {

			fmt.Println("inside callback")

			respCallback := response_callback.NewResponseCallback(envelopeDataGoCUsign.ResponseCallback.ResponseType)

			log.Println("downloading envelope signed")

			file, err := docusign.DocusignConfig.DownloadEnvelopeSigned(envelope.EnvelopeID)

			if err != nil {
				log.Println(err)
				return
			}

			defer file.Close()

			log.Println("respCallback: ", respCallback)

			if respCallback != nil {

				log.Println("sending to response")

				respCallback.SetParams(envelopeDataGoCUsign.ResponseCallback.ResponseMetaData)
				respCallback.SetOriginalData(envelopeDataGoCUsign)
				respCallback.SetEnvolepeData(envelope)
				respCallback.SetFile(file)
				err = respCallback.SendCallBack()

				if err != nil {
					log.Println("error: ", err)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Is running!\n"))
}

func JsonCreateEnvelope(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(&model.EnvelopeDefinition{
		EmailSubject: "Test",
		Documents: []model.Document{
			{
				DocumentBase64: []byte("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ago8PCAvVGl0bGUgKE1pY3Jvc29mdCBXb3JkIC0gRG9jdW1lbnRvMSkgL1Byb2R1Y2VyICj+/1wwMDBtXDAwMGFcMDAwY1wwMDBPXDAwMFNcMDAwIFwwMDBWXDAwMGVcMDAwclwwMDBzXDAwMONcMDAwb1wwMDAgXDAwMDFcMDAwMlwwMDAuXDAwMDZcMDAwIFwwMDBcKFwwMDBDXDAwMG9cMDAwbVwwMDBwXDAwMGlcMDAwbFwwMDBhXDAwMOdcMDAw41wwMDBvXDAwMCBcMDAwMlwwMDAxXDAwMEdcMDAwMVwwMDAxXDAwMDVcMDAwXClcMDAwIFwwMDBRXDAwMHVcMDAwYVwwMDByXDAwMHRcMDAwelwwMDAgXDAwMFBcMDAwRFwwMDBGXDAwMENcMDAwb1wwMDBuXDAwMHRcMDAwZVwwMDB4XDAwMHQpCi9DcmVhdG9yIChXb3JkKSAvQ3JlYXRpb25EYXRlIChEOjIwMjIxMDA2MTIxMjU4WjAwJzAwJykgL01vZERhdGUgKEQ6MjAyMjEwMDYxMjEyNThaMDAnMDAnKQo+PgplbmRvYmoKeHJlZgowIDE0CjAwMDAwMDAwMDAgNjU1MzUgZiAKMDAwMDAwMDE4NyAwMDAwMCBuIAowMDAwMDAzMTM1IDAwMDAwIG4gCjAwMDAwMDAwMjIgMDAwMDAgbiAKMDAwMDAwMDI5MSAwMDAwMCBuIAowMDAwMDAzMTAwIDAwMDAwIG4gCjAwMDAwMDAwMDAgMDAwMDAgbiAKMDAwMDAwMzI2NyAwMDAwMCBuIAowMDAwMDAwMzg4IDAwMDAwIG4gCjAwMDAwMDMyMTggMDAwMDAgbiAKMDAwMDAwMzcyNSAwMDAwMCBuIAowMDAwMDAzNDI5IDAwMDAwIG4gCjAwMDAwMDM5NjEgMDAwMDAgbiAKMDAwMDAxMDc5MyAwMDAwMCBuIAp0cmFpbGVyCjw8IC9TaXplIDE0IC9Sb290IDkgMCBSIC9JbmZvIDEzIDAgUiAvSUQgWyA8YzE3YzRlZjhlZWUxNjIxYWViZWM4OTA2NWQ2YmVmYWE+CjxjMTdjNGVmOGVlZTE2MjFhZWJlYzg5MDY1ZDZiZWZhYT4gXSA+PgpzdGFydHhyZWYKMTEyMzUKJSVFT0YK"),
				DocumentID:     "1",
			},
		},
		Recipients: &model.Recipients{
			Signers: []model.Signer{
				{
					Email:       "test@test.com",
					Name:        "Test",
					RecipientID: "1",
					Tabs: &model.Tabs{
						SignHereTabs: []model.SignHere{
							{
								TabBase: model.TabBase{
									DocumentID:  "1",
									RecipientID: "1",
								},
								TabPosition: model.TabPosition{
									AnchorString:  "This is a small",
									AnchorXOffset: "10",
									AnchorYOffset: "10",
									AnchorUnits:   "pixels",
									TabLabel:      "SignHereTab",
								},
							},
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func ConsentHandler(w http.ResponseWriter, r *http.Request) {
	url := docusign.DocusignConfig.GetConsentURI()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GetEnvelopeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	envelopeID := vars["envelopeID"]
	var dbEnvelope dbModels.Envelope
	var err error
	var resp *model.Envelope

	if database.DB == nil {
		resp, err = docusign.DocusignConfig.GetEnvelope(envelopeID)
	} else {
		database.DB.Where("id = ?", envelopeID).First(&dbEnvelope)
		json.Unmarshal([]byte(dbEnvelope.JsonData), &resp)
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// jsonResponse, _ := json.Marshal(resp)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func DownloadEnvelopeSigned(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	envelopeID := vars["envelopeID"]
	fileName := fmt.Sprintf("%s.pdf", envelopeID)

	var err error

	file, err := docusign.DocusignConfig.DownloadEnvelopeSigned(envelopeID)

	if err != nil {
		json_resp, _ := json.Marshal(map[string]interface{}{"error_type": "docusign_request", "error": err.Error()})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(json_resp)
		return
	}

	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))
	w.WriteHeader(http.StatusOK)
	http.ServeFile(w, r, file.Name())
}

func ViewsCreateRecipient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	envelopeID := vars["envelopeID"]

	var recipient model.RecipientViewRequest

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	err = json.Unmarshal(body, &recipient)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON"))
		return
	}

	urlReturn, err := docusign.DocusignConfig.ViewsCreateRecipient(envelopeID, recipient)
	if err != nil {
		json_resp, _ := json.Marshal(map[string]interface{}{"error_type": "docusign_request", "error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(json_resp)
		return
	}

	resp, err := json.Marshal(urlReturn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}

func EnvelopeRecipients(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	envelopeID := vars["envelopeID"]

	recipientsReturn, err := docusign.DocusignConfig.EnvelopeRecipients(envelopeID)
	if err != nil {
		json_resp, _ := json.Marshal(map[string]interface{}{"error_type": "docusign_request", "error": err.Error()})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(json_resp)
		return
	}

	resp, err := json.Marshal(recipientsReturn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}

func main() {

	database.InitDatabase()

	docusign.ReadDocusignConfig()

	r := mux.NewRouter()

	// Routes consist of a path and a handler function.
	r.HandleFunc("/", RootHandler)
	r.HandleFunc("/consent", ConsentHandler).Methods("GET")
	r.HandleFunc("/envelope/example_create", JsonCreateEnvelope).Methods("GET")
	r.HandleFunc("/envelope/{envelopeID}/download", DownloadEnvelopeSigned).Methods("GET")
	r.HandleFunc("/envelope/{envelopeID}", GetEnvelopeHandler).Methods("GET")
	r.HandleFunc("/callback", CallbackHandler).Methods("POST")
	r.HandleFunc("/envelope", CreateEnvelopeHandler).Methods("POST")
	r.HandleFunc("/envelope/{envelopeID}/views/recipient", ViewsCreateRecipient).Methods("POST")
	r.HandleFunc("/envelope/{envelopeID}/recipients", EnvelopeRecipients).Methods("GET")
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8080", r))
}
