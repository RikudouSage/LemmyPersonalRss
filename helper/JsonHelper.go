package helper

import "encoding/json"

type RawJson []byte

func MapJson[TOutput any](raw RawJson, output *TOutput) error {
	return json.Unmarshal(raw, output)
}
