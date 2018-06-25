package api

import "bytes"
import "encoding/json"
import "errors"
import "github.com/licensezero/cli/data"
import "io/ioutil"
import "net/http"
import "strconv"

type ResetRequest struct {
	Action     string `json:"action"`
	LicensorID string `json:"licensorID"`
	EMail      string `json:"email"`
}

type ResetResponse struct {
	Error interface{} `json:"error"`
}

func Reset(identity *data.Identity, licensor *data.Licensor) error {
	bodyData := ResetRequest{
		Action:     "reset",
		LicensorID: licensor.LicensorID,
		EMail:      identity.EMail,
	}
	body, err := json.Marshal(bodyData)
	if err != nil {
		return errors.New("could not construct reset request")
	}
	response, err := http.Post("https://licensezero.com/api/v0", "application/json", bytes.NewBuffer(body))
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("Server responded " + strconv.Itoa(response.StatusCode))
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	var parsed ResetResponse
	err = json.Unmarshal(responseBody, &parsed)
	if err != nil {
		return err
	}
	if message, ok := parsed.Error.(string); ok {
		return errors.New(message)
	}
	return nil
}
