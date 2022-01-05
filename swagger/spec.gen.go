// Package Swagger provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/pseudo-su/oapi-ui-codegen DO NOT EDIT.
package Swagger

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"strings"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetOpenAPISpec returns the Swagger specification corresponding to the generated code
// in this file.
func GetOpenAPISpec() (*openapi3.T, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewLoader().LoadFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
