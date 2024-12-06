/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package migration

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/go-test/deep"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var (
	fromFieldNames = []string{
		"Name",
		"Submission Date",
		"Type of Inquiry",
		"Phone Number",
		"Outside of US Phone Number",
		"Form Language Filled Out In",
		"Phone Number Corrected",
		"Email",
		"City",
		"State (Corrected)",
		"State / Province",
		"Preferred Language",
		"If other language, please specify:",
		"Identify as LGBTQ+?",
		"What type of legal assistance are you looking for?",
		`Do you have a case in immigration court ("removal proceedings")?`,
		"Message (tell us a little bit about the type of assistance you want or need):",
		"County (if in CA):",
		"Date of 1st Follow Up Attempt",
		"1st Attempt Result",
		"Date of Connection",
		"Connection Result",
		"Ineligible",
		"Notes",
		"Res/Natz Date of Connection",
		"Residency/Natz Result",
		"Res/Natz ineligible",
		"Res/Natz Notes",
		"Res/Natz Case Accepted?",
		"Case Opened",
		"∆ contact/opening",
		"Case Filed",
		"∆ opening/filed",
		"Phone Number Digits",
		"Phone Number Formatted",
		"Record ID",
		"Created By",
		"Month of Inquiry",
		"Year",
		"Asylum Screening Questionnaire",
		"Asylum Screening Questionnaire Responses",
		"Creation Date",
		"Record ID (for Migration) (from Asylum Screening Questionnaire)",
	}
	fromToCopyMap = map[string]string{
		"Name":                               "Name",
		"Submission Date":                    "Submission Date",
		"Type of Inquiry":                    "*ignore*",
		"Phone Number":                       "*clean*",
		"Outside of US Phone Number":         "*clean*",
		"Form Language Filled Out In":        "Form Language Filled Out",
		"Phone Number Corrected":             "*ignore*",
		"Email":                              "Email",
		"City":                               "City",
		"State (Corrected)":                  "*ignore*",
		"State / Province":                   "*clean*",
		"Preferred Language":                 "*clean*",
		"If other language, please specify:": "*clean*",
		"Identify as LGBTQ+?":                "LGBTQ+?",
		"What type of legal assistance are you looking for?":                            "*clean*",
		`Do you have a case in immigration court ("removal proceedings")?`:              "*clean*",
		"Message (tell us a little bit about the type of assistance you want or need):": "Service Request Information",
		"County (if in CA):":                       "CA County",
		"Date of 1st Follow Up Attempt":            "*ignore*",
		"1st Attempt Result":                       "*ignore*",
		"Date of Connection":                       "Date of Connection",
		"Connection Result":                        "Connection Result",
		"Ineligible":                               "Ineligible Reason",
		"Notes":                                    "Front Desk Notes",
		"Res/Natz Date of Connection":              "Res/Natz Date of Connection",
		"Residency/Natz Result":                    "Res/Natz Result",
		"Res/Natz ineligible":                      "Res/Natz Ineligible",
		"Res/Natz Notes":                           "Res/Natz Notes",
		"Res/Natz Case Accepted?":                  "Res/Natz Case Accepted",
		"Case Opened":                              "Res/Natz Case Opened",
		"∆ contact/opening":                        "∆ Res/Natz contact to opening",
		"Case Filed":                               "Res/Natz Case Filed",
		"∆ opening/filed":                          "∆ Res/Natz opening to filed",
		"Phone Number Digits":                      "*ignore*",
		"Phone Number Formatted":                   "*ignore*",
		"Record ID":                                "*ignore*",
		"Created By":                               "*ignore*",
		"Month of Inquiry":                         "*ignore*",
		"Year":                                     "*ignore*",
		"Asylum Screening Questionnaire":           "*ignore*",
		"Asylum Screening Questionnaire Responses": "*ignore*",
		"Creation Date":                            "*ignore*",
		"Record ID (for Migration) (from Asylum Screening Questionnaire)": "Asylum Screening Record ID for Migration",
	}
	toFieldNames = []string{
		"Name",
		"Phone",
		"E.164 number",
		"Submission Date",
		"Form Language Filled Out",
		"Email",
		"City",
		"Migrated State",
		"State",
		"LGBTQ+?",
		"Preferred Language",
		"Preferred Language (Other)",
		"Requested Legal Assistance",
		"In Removal Proceedings",
		"Service Request Information",
		"CA County",
		"Date of Connection",
		"Connection Result",
		"Ineligible Reason",
		"Front Desk Notes",
		"Res/Natz Date of Connection",
		"Res/Natz Result",
		"Res/Natz Ineligible",
		"Res/Natz Notes",
		"Res/Natz Case Accepted",
		"Res/Natz Case Opened",
		"∆ Res/Natz contact to opening",
		"Res/Natz Case Filed",
		"∆ Res/Natz opening to filed",
		"Asylum Screening Questionnaire",
		"Asylum Screening Record ID for Migration",
	}
)

func ImportAndClean(r *csv.Reader, w *csv.Writer) error {
	header, err := r.Read()
	if err != nil {
		return err
	}
	if diff := deep.Equal(header, fromFieldNames); diff != nil {
		return fmt.Errorf("input headers are wrong: %v", diff)
	}
	if err = w.Write(toFieldNames); err != nil {
		return err
	}
	for {
		fromRow, err := nextRowFrom(r)
		toRow := make(map[string]string)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		for i, fromCol := range fromFieldNames {
			val, ok := fromRow[fromCol]
			if !ok {
				return fmt.Errorf("column %q does not exist in row %d: %q", fromCol, i+1, fromRow)
			}
			toCol := fromToCopyMap[fromCol]
			switch toCol {
			case "*ignore*":
				// ignore this field
			case "*clean*":
				// clean the value into the target
				if err = cleanField(fromCol, val, toRow); err != nil {
					return fmt.Errorf("can't clean fromCol %q: %v", fromCol, err)
				}
			default:
				toRow[toCol] = val
			}
		}
		if err = nextRowTo(toRow, w); err != nil {
			return err
		}
	}
	return nil
}

func nextRowFrom(r *csv.Reader) (map[string]string, error) {
	vals, err := r.Read()
	if err != nil {
		return nil, err
	}
	if len(vals) != len(fromFieldNames) {
		return nil, fmt.Errorf("read %d fields, expected %d: %q", len(vals), len(fromFieldNames), vals)
	}
	row := make(map[string]string)
	for i, key := range fromFieldNames {
		row[key] = vals[i]
	}
	return row, nil
}

func nextRowTo(row map[string]string, w *csv.Writer) error {
	var cols []string
	for _, col := range toFieldNames {
		val := row[col]
		cols = append(cols, val)
	}
	if err := w.Write(cols); err != nil {
		return err
	}
	return nil
}

// BOMAwareCSVReader will detect a UTF BOM (Byte Order Mark) at the
// start of the data and transform to UTF8 accordingly.
// If there is no BOM, it will read the data without any transformation.
//
// This code taken from [this StackOverflow answer](https://stackoverflow.com/a/76023436/558006).
func BOMAwareCSVReader(reader io.Reader) *csv.Reader {
	var transformer = unicode.BOMOverride(encoding.Nop.NewDecoder())
	return csv.NewReader(transform.NewReader(reader, transformer))
}
