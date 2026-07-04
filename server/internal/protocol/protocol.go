package protocol

import "encoding/json"

// Marshal encodes v into JSON bytes for packet payloads.
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal decodes JSON payload into v.
func Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
