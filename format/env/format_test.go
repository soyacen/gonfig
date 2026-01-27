package env

import (
	"reflect"
	"testing"
)

// TestParse_Success tests successful parsing of valid environment variables format.
func TestParse_Success(t *testing.T) {
	data := []byte("NAME=Alice\nAGE=30\nIS_STUDENT=true")
	parser := Env{}
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}

	expectedMap := map[string]interface{}{
		"NAME":       "Alice",
		"AGE":        "30",
		"IS_STUDENT": "true",
	}

	if !reflect.DeepEqual(expectedMap, result.AsMap()) {
		t.Errorf("Expected map %v, got %v", expectedMap, result.AsMap())
	}
}

// TestParse_EmptyData tests parsing of empty data.
func TestParse_EmptyData(t *testing.T) {
	data := []byte("")
	parser := Env{}
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}

	expectedMap := map[string]interface{}{}

	if !reflect.DeepEqual(expectedMap, result.AsMap()) {
		t.Errorf("Expected empty map, got %v", result.AsMap())
	}
}

// TestParse_SingleVariable tests parsing of a single environment variable.
func TestParse_SingleVariable(t *testing.T) {
	data := []byte("KEY=value")
	parser := Env{}
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}

	expectedMap := map[string]interface{}{
		"KEY": "value",
	}

	if !reflect.DeepEqual(expectedMap, result.AsMap()) {
		t.Errorf("Expected map %v, got %v", expectedMap, result.AsMap())
	}
}

// TestParse_ValuesWithSpecialCharacters tests parsing of values containing special characters.
func TestParse_ValuesWithSpecialCharacters(t *testing.T) {
	data := []byte("URL=https://example.com:8080\nPASSWORD=p@ss!w0rd#123")
	parser := Env{}
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}

	expectedMap := map[string]interface{}{
		"URL":      "https://example.com:8080",
		"PASSWORD": "p@ss!w0rd#123",
	}

	if !reflect.DeepEqual(expectedMap, result.AsMap()) {
		t.Errorf("Expected map %v, got %v", expectedMap, result.AsMap())
	}
}

// TestParse_ValuesWithEquals tests parsing of values containing equals sign.
func TestParse_ValuesWithEquals(t *testing.T) {
	data := []byte("FORMULA=a=b+c\nCONNECTION_STRING=host=localhost;port=5432")
	parser := Env{}
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}

	expectedMap := map[string]interface{}{
		"FORMULA":           "a=b+c",
		"CONNECTION_STRING": "host=localhost;port=5432",
	}

	if !reflect.DeepEqual(expectedMap, result.AsMap()) {
		t.Errorf("Expected map %v, got %v", expectedMap, result.AsMap())
	}
}

// TestParse_QuotedValues tests parsing of quoted values.
func TestParse_QuotedValues(t *testing.T) {
	data := []byte("NAME=\"John Doe\"\nDESCRIPTION='Multi word value'")
	parser := Env{}
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}

	// godotenv removes quotes from values
	resultMap := result.AsMap()
	if resultMap["NAME"] != "John Doe" {
		t.Errorf("Expected 'John Doe', got %v", resultMap["NAME"])
	}
	if resultMap["DESCRIPTION"] != "Multi word value" {
		t.Errorf("Expected 'Multi word value', got %v", resultMap["DESCRIPTION"])
	}
}

// TestParse_EmptyValue tests parsing of variables with empty values.
func TestParse_EmptyValue(t *testing.T) {
	data := []byte("EMPTY=\nKEY=value")
	parser := Env{}
	result, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}

	expectedMap := map[string]interface{}{
		"EMPTY": "",
		"KEY":   "value",
	}

	if !reflect.DeepEqual(expectedMap, result.AsMap()) {
		t.Errorf("Expected map %v, got %v", expectedMap, result.AsMap())
	}
}
