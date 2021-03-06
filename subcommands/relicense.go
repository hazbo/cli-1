package subcommands

import "bytes"
import "encoding/json"
import "errors"
import "flag"
import "github.com/licensezero/cli/api"
import "github.com/licensezero/cli/data"
import "github.com/licensezero/cli/inventory"
import "io/ioutil"
import "os"
import "path"

const relicenseDescription = "Relicense on Charity terms."

// Relicense updates LICENSE and metadata with Charity terms.
var Relicense = &Subcommand{
	Tag:         "seller",
	Description: relicenseDescription,
	Handler: func(args []string, paths Paths) {
		flagSet := flag.NewFlagSet("relicense", flag.ExitOnError)
		projectID := projectIDFlag(flagSet)
		id := idFlag(flagSet)
		silent := silentFlag(flagSet)
		flagSet.SetOutput(ioutil.Discard)
		flagSet.Usage = relicenseUsage
		flagSet.Parse(args)
		if *projectID == "" && *id == "" {
			relicenseUsage()
		}
		if *projectID != "" && *id != "" {
			relicenseUsage()
		}
		if *projectID != "" {
			*id = *projectID
		}
		licensor, err := data.ReadLicensor(paths.Home)
		if err != nil {
			Fail(licensorHint)
		}
		response, err := api.Public(licensor, *id, "charity")
		if err != nil {
			Fail("Error sending license information request: " + err.Error())
		}
		// Add metadata to licensezero.json.
		licensezeroJSON := path.Join(paths.CWD, "licensezero.json")
		data, err := ioutil.ReadFile(licensezeroJSON)
		if err != nil {
			Fail("Could not read licensezero.json.")
		}
		var existingJSON interface{}
		var existingMetadata inventory.LicenseZeroJSONFile
		err = json.Unmarshal(data, &existingJSON)
		if err != nil {
			Fail("Error parsing licensezero.json.")
		}
		err = json.Unmarshal(data, &existingMetadata)
		if err != nil {
			Fail("Error parsing licensezero.json.")
		}
		newEntries := []inventory.ProjectManifestEnvelope{}
		for _, entry := range existingMetadata.Envelopes {
			if entry.Manifest.ProjectID != *id {
				newEntries = append(newEntries, entry)
			}
		}
		itemsMap := existingJSON.(map[string]interface{})
		itemsMap["licensezero"] = newEntries
		serialized := new(bytes.Buffer)
		encoder := json.NewEncoder(serialized)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(existingJSON)
		if err != nil {
			Fail("Error serializing new JSON.")
		}
		err = ioutil.WriteFile(licensezeroJSON, serialized.Bytes(), 0644)
		if !*silent {
			os.Stdout.WriteString("Added metadata to licensezero.json.\n")
		}
		// Overwrite LICENSE.
		err = overwriteLICENSE(response)
		if err != nil {
			Fail(err.Error())
		}
		if !*silent {
			os.Stdout.WriteString("Appended terms to LICENSE.\n")
		}
		os.Exit(0)
	},
}

func overwriteLICENSE(response *api.PublicResponse) error {
	toWrite := "" +
		response.License.Document + "\n\n" +
		"---\n\n" +
		"Licensor Signature (Ed25519):\n\n" +
		signatureLines(response.License.LicensorSignature) + "\n\n" +
		"---\n\n" +
		"Agent Signature (Ed25519):\n\n" +
		signatureLines(response.License.AgentSignature) + "\n"
	err := ioutil.WriteFile("LICENSE", []byte(toWrite), 0644)
	if err != nil {
		return errors.New("Error writing LICENSE")
	}
	return nil
}

func relicenseUsage() {
	usage := relicenseDescription + "\n\n" +
		"Usage:\n" +
		"  licensezero relicense --id ID [--stack]\n\n" +
		"Options:\n" +
		flagsList(map[string]string{
			"id ID":  idLine,
			"silent": silentLine,
			"stack":  "Stack multiple project metadata entries.",
		})
	Fail(usage)
}
