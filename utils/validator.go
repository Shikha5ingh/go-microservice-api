package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/xeipuuv/gojsonschema"
)

// Validator wraps the gojsonschema.Schema
type Validator struct {
	schema *gojsonschema.Schema
}

// NewValidator loads and compiles the JSON schema
func NewValidator(schemaPath string) (*Validator, error) {
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %v", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %v", err)
	}

	return &Validator{schema: schema}, nil
}

// Validate checks the JSON data against the schema
func (v *Validator) Validate(jsonData []byte) error {
	log.Println("Validating JSON data against schema")
	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		log.Println("Schema Validation Error:", err)
		return fmt.Errorf("schema validation error: %v", err)
	} else {
		log.Println("Schema Validation Result:", result)
	}

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
	return nil
}
