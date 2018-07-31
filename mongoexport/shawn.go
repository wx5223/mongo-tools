// Copyright (C) MongoDB, Inc. 2014-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package mongoexport

import (
	"fmt"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	"github.com/mongodb/mongo-tools/common/json"
	"gopkg.in/mgo.v2/bson"
	"io"
	"strings"
	"reflect"
)

// JSONExportOutput is an implementation of ExportOutput that writes documents
// to the output in JSON format.
type ShawnExportOutput struct {

	Fields       []string
	NoHeaderLine bool
	Out          io.Writer
	NumExported  int64
	FieldSplit   string `used for shawn's own mode split fields`
	LineSplit    string `used for shawn's own mode split lines`
}

// NewJSONExportOutput creates a new JSONExportOutput in array mode if specified,
// configured to write data to the given io.Writer.
func NewShawnExportOutput(fields []string, noHeaderLine bool, out io.Writer, fieldSplit string, lineSplit string) *ShawnExportOutput {
	if fieldSplit == "" {
		fieldSplit = "\001"
	}
	if lineSplit == "" {
		lineSplit = "\n"
	}
	if lineSplit == "005" {
		lineSplit = "\005"
	}
	return &ShawnExportOutput{
		fields,
		noHeaderLine,
		out,
		0,
		fieldSplit,
		lineSplit,
	}
}

// WriteHeader writes the opening square bracket if in array mode, otherwise it
// behaves as a no-op.
func (exporter *ShawnExportOutput) WriteHeader() error {
	if !exporter.NoHeaderLine {
		result := strings.Join(exporter.Fields, exporter.FieldSplit)
		_, err := exporter.Out.Write([]byte(result+ exporter.LineSplit))
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteFooter writes the closing square bracket if in array mode, otherwise it
// behaves as a no-op.
func (exporter *ShawnExportOutput) WriteFooter() error {
	return nil
}

// Flush is a no-op for JSON export formats.
func (exporter *ShawnExportOutput) Flush() error {
	return nil
}

func (exporter *ShawnExportOutput) ExportDocument(document bson.D) error {
	rowOut := make([]string, 0, len(exporter.Fields))
	extendedDoc, err := bsonutil.ConvertBSONValueToJSON(document)
	if err != nil {
		return err
	}

	for _, fieldName := range exporter.Fields {
		fieldVal := extractFieldByName(fieldName, extendedDoc)
		if fieldVal == nil {
			rowOut = append(rowOut, "")
		} else if reflect.TypeOf(fieldVal) == reflect.TypeOf(bson.M{}) ||
			reflect.TypeOf(fieldVal) == reflect.TypeOf(bson.D{}) ||
			reflect.TypeOf(fieldVal) == marshalDType ||
			reflect.TypeOf(fieldVal) == reflect.TypeOf([]interface{}{}) {
			buf, err := json.Marshal(fieldVal)
			if err != nil {
				rowOut = append(rowOut, "")
			} else {
				rowOut = append(rowOut, string(buf))
			}
		} else {
			rowOut = append(rowOut, fmt.Sprintf("%v", fieldVal))
		}
	}

	result := strings.Join(rowOut, exporter.FieldSplit)
	_, err1 := exporter.Out.Write([]byte(result+ exporter.LineSplit))
	if err1 != nil {
		return err1
	}

	exporter.NumExported++
	return nil
}

