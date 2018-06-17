package subcommands

import "encoding/json"
import "io/ioutil"
import "os"
import "path"

type Identity struct {
	Name         string `json:"name"`
	Jurisdiction string `json:"jurisdiction"`
	EMail        string `json:"email"`
}

func configPath(home string) string {
	return path.Join(home, ".config", "licensezero")
}

func identityPath(home string) string {
	return path.Join(configPath(home), "identity.json")
}

func readIdentity(home string) (*Identity, error) {
	path := identityPath(home)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var identity Identity
	json.Unmarshal(data, &identity)
	return &identity, nil
}

func writeIdentity(home string, identity *Identity) error {
	data, jsonError := json.Marshal(identity)
	if jsonError != nil {
		return jsonError
	}
	directoryError := makeConfigDirectory(home)
	if directoryError != nil {
		return directoryError
	}
	return ioutil.WriteFile(identityPath(home), data, 0744)
}

func makeConfigDirectory(home string) error {
	path := configPath(home)
	return os.MkdirAll(path, 0744)
}