package response_callback

import (
	"app/database/models"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/jfcote87/esign/v2/model"
)

type CapCallBack struct {
	Params        []models.ResponseMetaData
	File          *os.File
	OriginalData  models.GoCusignEnvelopeDefinition
	EnvelopeData  model.EnvelopeDefinition
	EnvelopeModel models.Envelope
	RequestID     string
	TaskTitle     string
	FieldName     string
	Outcome       string
	ResponseURL   string
	FileBase64    string
	FileName      string
}

func (cap *CapCallBack) SetParams(Params []models.ResponseMetaData) {
	cap.Params = Params
}

func (cap *CapCallBack) SetFile(File *os.File) {
	cap.File = File
}

func (cap *CapCallBack) SetOriginalData(OrignalData models.GoCusignEnvelopeDefinition) {
	cap.OriginalData = OrignalData
}

func (cap *CapCallBack) SetEnvolepeData(EnvelopeData model.EnvelopeDefinition) {
	cap.EnvelopeData = EnvelopeData
}

func (cap *CapCallBack) makeRequest() error {
	client := &http.Client{}

	bodyData := cap.GenerateXML()
	stringBody := strings.NewReader(bodyData)

	req, err := http.NewRequest("POST", cap.ResponseURL, stringBody)

	if err != nil {
		log.Println(err)
		return err
	}

	req.Header.Add("SOAPAction", "http://iteris.cap.webservices/CompleteTask")
	req.Header.Add("Content-Type", "text/xml")

	req.SetBasicAuth(
		os.Getenv("CAP_USERNAME"),
		os.Getenv("CAP_PASSWORD"),
	)

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return err
	}

	defer resp.Body.Close()

	return nil

}

func (cap *CapCallBack) SendCallBack() error {

	cap.parseMetaData()

	err := cap.validate()

	if err != nil {
		log.Println(err)
		return err
	}

	cap.FileName = fmt.Sprintf("%s.pdf", cap.EnvelopeData.EnvelopeID)

	bytes, err := os.ReadFile(cap.File.Name())

	if err != nil {
		log.Println(err)
		return err
	}

	cap.FileBase64 = toBase64(bytes)

	err = cap.makeRequest()

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (cap *CapCallBack) GenerateXML() string {

	var msg bytes.Buffer

	templateStr := `
		<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:xsd="http://www.w3.org/2001/XMLSchema"
		xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
		<soap:Body>
			<CompleteTask xmlns="http://iteris.cap.webservices">
				<requestId>{{.TaskID}}</requestId>
				<taskTitle>'{{.TaskTitle}}'</taskTitle>
				<metadataValues>
					<MetadataValue>
						<Name>{{.FiledName}}</Name>
						<Attachments>
							<MetadataAttachment>
								<FileName>{{.FileName}}</FileName>
								<ContentType>application/pdf</ContentType>
								<Base64Content>{{.Base64}}</Base64Content>
							</MetadataAttachment>
						</Attachments>
					</MetadataValue>
				</metadataValues>
				<outcome>{{.Outcome}}</outcome>
				<comments></comments>
			</CompleteTask>
		</soap:Body>
	</soap:Envelope>
	`

	tmpl, _ := template.New("cap").Parse(templateStr)
	tmpl.Execute(&msg, map[string]string{
		"TaskID":    cap.RequestID,
		"TaskTitle": cap.TaskTitle,
		"FiledName": cap.FieldName,
		"FileName":  cap.FileName,
		"Base64":    cap.FileBase64,
		"Outcome":   cap.Outcome,
	})

	return msg.String()

}

func (cap *CapCallBack) parseMetaData() {

	for _, param := range cap.Params {
		switch param.Name {
		case "request_id":
			cap.RequestID = param.Value
		case "task_title":
			cap.TaskTitle = param.Value
		case "field_name":
			cap.FieldName = param.Value
		case "outcome":
			cap.Outcome = param.Value
		case "response_url":
			cap.ResponseURL = param.Value
		}
	}
}

func (cap *CapCallBack) validate() error {
	if cap.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}

	if cap.TaskTitle == "" {
		return fmt.Errorf("task_title is required")
	}

	if cap.FieldName == "" {
		return fmt.Errorf("field_name is required")
	}

	if cap.Outcome == "" {
		return fmt.Errorf("outcome is required")
	}

	if cap.ResponseURL == "" {
		return fmt.Errorf("response_url is required")
	}

	return nil
}
