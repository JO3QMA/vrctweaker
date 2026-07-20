package usecase

import "encoding/json"

func jsonMarshalStd(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
