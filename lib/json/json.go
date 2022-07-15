package json_strict

import (
	"encoding/json"
	"os"
)

func UnmarshalFile(file string, output interface{}) error {
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(output)
}
