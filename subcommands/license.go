package subcommands

import "bytes"
import "encoding/json"
import "errors"
import "flag"
import "github.com/licensezero/cli/api"
import "github.com/licensezero/cli/data"
import "io/ioutil"
import "os"
import "path"

var License = Subcommand{
	Description: "Write a public license file and metadata.",
	Handler: func(args []string, paths Paths) {
		flagSet := flag.NewFlagSet("license", flag.ExitOnError)
		noncommercial := flagSet.Bool("noncommercial", false, "Use noncommercial public license.")
		reciprocal := flagSet.Bool("reciprocal", false, "Use reciprocal public license.")
		stack := flagSet.Bool("stack", false, "Stack licensing metadata.")
		if *noncommercial && *reciprocal {
			os.Stderr.WriteString("specify either --reciprocal OR --noncommercial")
			os.Exit(1)
		}
		flagSet.Parse(args)
		if flagSet.NArg() != 1 {
			licenseUsage()
		} else {
			projectID := flagSet.Args()[0]
			if *noncommercial && *reciprocal {
				licenseUsage()
			}
			licensor, err := data.ReadLicensor(paths.Home)
			if err != nil {
				os.Stderr.WriteString("create a licensor identity with `licensezero register` or `licensezero set-licensor-id`.")
				os.Exit(1)
			}
			var terms string
			if *noncommercial {
				terms = "noncommercial"
			}
			if *reciprocal {
				terms = "reciprocal"
			}
			response, err := api.License(licensor, projectID, terms)
			if err != nil {
				os.Stderr.WriteString("error sending license information request")
				os.Exit(1)
			}
			// Add metadata to package.json.
			newEntry := response.Metadata
			package_json := path.Join(paths.CWD, "package.json")
			data, err := ioutil.ReadFile(package_json)
			if err != nil {
				os.Stderr.WriteString("could not read package.json")
				os.Exit(1)
			}
			var existingMetadata interface{}
			err = json.Unmarshal(data, &existingMetadata)
			if err != nil {
				os.Stderr.WriteString("error parsing package.json")
				os.Exit(1)
			}
			itemsMap := existingMetadata.(map[string]interface{})
			var entries []interface{}
			if _, ok := itemsMap["licensezero"]; ok {
				if entries, ok := itemsMap["licensezero"].([]interface{}); ok {
					if *stack {
						entries = append(entries, newEntry)
					} else {
						os.Stderr.WriteString("package.json already has License Zero metadata. Use --stack to stack metadata.")
						os.Exit(1)
					}
				} else {
					os.Stderr.WriteString("package.json has an invalid licensezero property.")
					os.Exit(1)
				}
			} else {
				if *stack {
					os.Stderr.WriteString("Cannot stack License Zero metadata. There is no preexisting metadata.")
					os.Exit(1)
				} else {
					entries = []interface{}{newEntry}
				}
			}
			itemsMap["licensezero"] = entries
			serialized, err := json.Marshal(existingMetadata)
			if err != nil {
				os.Stderr.WriteString("error serializing new JSON")
				os.Exit(1)
			}
			indented := bytes.NewBuffer([]byte{})
			err = json.Indent(indented, serialized, "", "  ")
			if err != nil {
				os.Stderr.WriteString("error indenting new JSON")
				os.Exit(1)
			}
			err = ioutil.WriteFile(package_json, indented.Bytes(), 0644)
			// Append to LICENSE.
			err = writeLICENSE(&response)
			if err != nil {
				os.Stderr.WriteString(err.Error())
				os.Exit(1)
			}
			os.Exit(0)
		}
	},
}

func writeLICENSE(response *api.LicenseResponse) error {
	var toWrite string
	existing, err := ioutil.ReadFile("LICENSE")
	if err != nil {
		if os.IsNotExist(err) {
			toWrite = ""
		} else {
			return errors.New("Could not open LICENSE.")
		}
	} else {
		toWrite = string(existing)
	}
	if len(toWrite) != 0 {
		toWrite = toWrite + "\n\n"
	}
	toWrite = toWrite +
		response.License.Document + "\n\n" +
		"---\n\n" +
		"Licensor Signature (Ed25519):\n\n" +
		signatureLines(response.License.LicensorSignature) + "\n\n" +
		"---\n\n" +
		"Agent Signature (Ed25519):\n\n" +
		signatureLines(response.License.AgentSignature)
	err = ioutil.WriteFile("LICENSE", []byte(toWrite), 0644)
	if err != nil {
		return errors.New("Error writing LICENSE")
	}
	return nil
}

func signatureLines(signature string) string {
	return "" +
		signature[0:32] + "\n" +
		signature[32:64] + "\n" +
		signature[64:96] + "\n" +
		signature[96:]
}

func licenseUsage() {
	os.Stderr.WriteString(`Usage:
	 <project id> (--noncommercial | --reciprocal)
`)
	os.Exit(1)
}
