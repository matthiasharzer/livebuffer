package marshalutil

import (
	"encoding/json"
)

func UnmarshalAny[T any](data any, target T) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}
