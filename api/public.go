package api

import "bytes"
import "encoding/json"
import "errors"
import "github.com/licensezero/cli/data"
import "io/ioutil"
import "net/http"

type publicRequest struct {
	Action     string `json:"action"`
	LicensorID string `json:"licensorID"`
	Token      string `json:"token"`
	ProjectID  string `json:"projectID"`
	Terms      string `json:"terms"`
}

// PublicResponse contains API instructions for adding public licensing information.
type PublicResponse struct {
	Error    interface{} `json:"error"`
	Metadata struct {
		License     string      `json:"license"`
		LicenseZero interface{} `json:"licensezero"`
	} `json:"metadata"`
	License struct {
		Document          string `json:"document"`
		LicensorSignature string `json:"licensorSignature"`
		AgentSignature    string `json:"agentSignature"`
	} `json:"license"`
}

// Public sends a public API request.
func Public(licensor *data.Licensor, projectID string, terms string) (*PublicResponse, error) {
	bodyData := publicRequest{
		Action:     "public",
		ProjectID:  projectID,
		Terms:      terms,
		LicensorID: licensor.LicensorID,
		Token:      licensor.Token,
	}
	body, err := json.Marshal(bodyData)
	if err != nil {
		return nil, err
	}
	response, err := http.Post("https://licensezero.com/api/v0", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.New("error sending request")
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var parsed PublicResponse
	err = json.Unmarshal(responseBody, &parsed)
	if err != nil {
		return nil, err
	}
	if message, ok := parsed.Error.(string); ok {
		return nil, errors.New(message)
	}
	return &parsed, nil
}
