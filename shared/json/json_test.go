package json

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJSONScan(t *testing.T) {
	var j JSON
	err := j.Scan([]byte(`{"key": "value"}`))
	if err != nil {
		t.Errorf("Scan() error = %v, wantErr %v", err, false)
	}

	expected := JSON(json.RawMessage(`{"key": "value"}`))
	if string(j) != string(expected) {
		t.Errorf("Scan() = %v, want %v", j, expected)
	}
}

func TestJSONScanInvalid(t *testing.T) {
	var j JSON
	err := j.Scan("invalid")
	if err == nil {
		t.Errorf("Scan() error = %v, wantErr %v", err, true)
	}
}

func TestJSONScanNil(t *testing.T) {
	var j JSON
	err := j.Scan(nil)
	if err != nil {
		t.Errorf("Scan() error = %v, wantErr %v", err, false)
	}

	if j != nil {
		t.Errorf("Scan() = %v, want %v", j, nil)
	}
}

func TestJSONScanEmpty(t *testing.T) {
	var j JSON
	err := j.Scan([]byte(``))
	if err == nil {
		t.Errorf("Scan() error = %v, wantErr %v", err, true)
	}
}

func TestJSONScanInvalidJSON(t *testing.T) {
	var j JSON
	err := j.Scan([]byte(`{invalid json}`))
	if err == nil {
		t.Errorf("Scan() error = %v, wantErr %v", err, true)
	}
}

func TestJSONValue(t *testing.T) {
	j := JSON(json.RawMessage(`{"key": "value"}`))
	val, err := j.Value()
	if err != nil {
		t.Errorf("Value() error = %v, wantErr %v", err, false)
	}

	expected := []byte(`{"key": "value"}`)
	if string(val.([]byte)) != string(expected) {
		t.Errorf("Value() = %v, want %v", val, expected)
	}
}

func TestJSONValueEmpty(t *testing.T) {
	var j JSON
	val, err := j.Value()
	if err != nil {
		t.Errorf("Value() error = %v, wantErr %v", err, false)
	}

	if val != nil {
		t.Errorf("Value() = %v, want %v", val, nil)
	}
}

func TestJSONValueInvalid(t *testing.T) {
	j := JSON(json.RawMessage(`invalid`))
	val, err := j.Value()
	if err == nil {
		t.Errorf("Value() error = %v, wantErr %v", err, true)
	}

	if val != nil {
		t.Errorf("Value() = %v, want %v", val, nil)
	}
}

func TestJSONMarshalJSON(t *testing.T) {
	j := JSON(json.RawMessage(`{"key": "value"}`))
	bytes, err := j.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v, wantErr %v", err, false)
	}

	expected := []byte(`{"key": "value"}`)
	if string(bytes) != string(expected) {
		t.Errorf("MarshalJSON() = %v, want %v", string(bytes), string(expected))
	}
}

func TestJSONMarshalJSONEmpty(t *testing.T) {
	var j JSON
	bytes, err := j.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v, wantErr %v", err, false)
	}

	if len(bytes) != 0 {
		t.Errorf("MarshalJSON() = %v, want %v", string(bytes), "[]")
	}
}

func TestUnmarshalJSON_AssignsValueCorrectly(t *testing.T) {
	j := JSON(`{"name":"Oscar","age":30}`)
	input := []byte(`{"name":"Oscar","age":30}`)

	err := j.UnmarshalJSON(input)

	require.NoError(t, err)
	require.JSONEq(t, string(input), string(j))
}

