package docusign

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"app/database"
	dbModels "app/database/models"

	"github.com/jfcote87/esign"
	"github.com/jfcote87/esign/v2/envelopes"
	"github.com/jfcote87/esign/v2/model"
)

type DocusignConfigType struct {
	UserId             string
	APIAccountId       string
	AccountBaseUri     string
	IntegratorKey      string
	KeyparisId         string
	ConsentRedirectUri string
	Environment        string
	IsProduction       bool
	cfg                esign.JWTConfig
}

var DocusignConfig DocusignConfigType

func ReadDocusignConfig() {

	DocusignConfig.UserId = os.Getenv("USER_ID")
	DocusignConfig.APIAccountId = os.Getenv("API_ACCOUNT_ID")
	DocusignConfig.AccountBaseUri = os.Getenv("ACCOUNT_BASE_URI")
	DocusignConfig.IntegratorKey = os.Getenv("INTEGRATION_KEY")
	DocusignConfig.KeyparisId = os.Getenv("KEYPAIRS_ID")
	DocusignConfig.ConsentRedirectUri = os.Getenv("CONSENT_REDIRECT_URI")
	DocusignConfig.Environment = os.Getenv("ENVIROMENT")
	DocusignConfig.IsProduction = false

	if DocusignConfig.Environment == "production" {
		DocusignConfig.IsProduction = true
	}

	DocusignConfig.cfg = esign.JWTConfig{
		AccountID:     DocusignConfig.APIAccountId,
		IntegratorKey: DocusignConfig.IntegratorKey,
		KeyPairID:     DocusignConfig.KeyparisId,
		IsDemo:        !DocusignConfig.IsProduction,
		PrivateKey:    strings.Replace(os.Getenv("CERTIFICATE"), "\\n", "\n", -1),
	}
}

func (ds *DocusignConfigType) GetConsentURI() string {

	consentURL := ds.cfg.UserConsentURL(ds.ConsentRedirectUri)

	consentURL, _ = url.QueryUnescape(consentURL)

	return consentURL
}

func (ds *DocusignConfigType) getCredentials() (*esign.OAuth2Credential, error) {

	return ds.cfg.Credential(ds.UserId, nil, nil)
}

func (ds *DocusignConfigType) GetEnvelope(envelopeID string) (*model.Envelope, error) {
	cred, _ := ds.getCredentials()
	sv := envelopes.New(cred)
	envSummary, err := sv.Get(envelopeID).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return envSummary, nil
}

func (ds *DocusignConfigType) DownloadEnvelopeSigned(envelopeID string, args ...string) (file *os.File, err error) {
	fileName := fmt.Sprintf("%s.pdf", envelopeID)

	cred, _ := ds.getCredentials()
	sv := envelopes.New(cred)
	dn, err := sv.DocumentsGet("combined", envelopeID).
		Certificate().
		Watermark().
		Do(context.TODO())

	if err != nil {
		return nil, err
	}

	defer dn.Close()

	file, err = ioutil.TempFile("", fileName)

	if err != nil {
		return nil, err
	}

	io.Copy(file, dn)

	return file, nil
}

func (ds *DocusignConfigType) CreateEnvelop(envelopeDefinition *model.EnvelopeDefinition) (envSummary *model.EnvelopeSummary, err error) {

	ctx := context.TODO()

	cred, err := ds.getCredentials()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	sv := envelopes.New(cred)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	envSummary, err = sv.Create(envelopeDefinition).Do(ctx)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if database.DB != nil {

		envelopeJson, _ := json.Marshal(envelopeDefinition)

		database.DB.Create(&dbModels.Envelope{
			ID:       envSummary.EnvelopeID,
			Status:   envSummary.Status,
			JsonData: string(envelopeJson),
		})
	}

	return envSummary, nil
}

func (ds *DocusignConfigType) ViewsCreateRecipient(envelopeId string, recipient model.RecipientViewRequest) (urlReturn *model.ViewURL, err error) {

	ctx := context.TODO()

	cred, err := ds.getCredentials()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	sv := envelopes.New(cred)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	var envelopeRecipients = recipient
	urlReturn, err = sv.ViewsCreateRecipient(envelopeId, &envelopeRecipients).Do(ctx)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return urlReturn, nil
}

func (ds *DocusignConfigType) EnvelopeRecipients(envelopeId string) (recipientsReturn *model.Recipients, err error) {

	ctx := context.TODO()

	cred, err := ds.getCredentials()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	sv := envelopes.New(cred)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	recipientsReturn, err = sv.RecipientsList(envelopeId).Do(ctx)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return recipientsReturn, nil
}
