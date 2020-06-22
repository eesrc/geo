package serializing

import (
	"database/sql/driver"
	"encoding/json"
)

// ValueJSON is a generic db driver Value implementation using JSON
func ValueJSON(obj interface{}) (driver.Value, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ScanJSON is a generic db driver Scan implementation using JSON
func ScanJSON(obj, src interface{}) error {
	data := getBytesFromSrc(src)

	return json.Unmarshal(data, obj)
}

func getBytesFromSrc(src interface{}) []byte {
	var data []byte
	if b, ok := src.([]byte); ok {
		data = b
	} else if s, ok := src.(string); ok {
		data = []byte(s)
	}

	return data
}
