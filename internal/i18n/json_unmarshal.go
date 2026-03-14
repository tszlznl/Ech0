package i18n

import "encoding/json"

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
