package json

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type JSON json.RawMessage

// Scan implements the sql.Scanner interface.
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	var result json.RawMessage
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = JSON(result)
	return nil
}

// Value implements the driver.Valuer interface.
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	if !json.Valid(j) {
		return nil, errors.New("invalid JSON value")
	}

	return json.RawMessage(j).MarshalJSON()
}

// MarshalJSON implements the json.Marshaler interface.
func (j JSON) MarshalJSON() ([]byte, error) {
	return j, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (j *JSON) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return errors.New("invalid JSON value")
	}
	*j = append((*j)[0:0], data...)
	return nil
}
