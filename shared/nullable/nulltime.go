package nullable

import (
	"database/sql"
	"strings"
	"time"
)

type NullTime struct {
	sql.NullTime
}

func NewNullTime(time time.Time) NullTime {
	return NullTime{sql.NullTime{
		Time:  time,
		Valid: true,
	}}
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	nt.Valid = false
	s := strings.Trim(string(data), "\"")
	if s == "null" {
		return nil
	}
	if s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return err
		}
		if !t.IsZero() {
			nt.Valid = true
			nt.Time = t
		}
	}

	return nil
}
