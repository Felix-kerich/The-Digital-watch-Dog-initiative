package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSON is a custom type for handling JSON data in GORM
type JSON map[string]interface{}

// Value implements the driver.Valuer interface for database serialization
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}

	val, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

// Scan implements the sql.Scanner interface for database deserialization
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSON)
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(data, j)
}
