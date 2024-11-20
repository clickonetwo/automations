/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package contacts

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-test/deep"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var (
	ImportColumnNames = []string{"Creation Date", "First_Name", "Last_Name", "Phones", "Email"}
	ExportColumnNames = []string{"Dialpad UID", "First Name", "Last Name", "Phones", "Emails"}
)

func ImportContacts(path string, showErrors bool) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := BOMAwareCSVReader(f)
	record, err := reader.Read()
	if diff := deep.Equal(record, ImportColumnNames); diff != nil {
		return nil, fmt.Errorf("unexpected column names in input: %v", diff)
	}
	return parseRecords(reader, showErrors)
}

func ExportContacts(entries []Entry, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	defer writer.Flush()
	if err = writer.Write(ExportColumnNames); err != nil {
		log.Panicf("error writing record to csv: %v", err)
	}
	for _, entry := range entries {
		phones := strings.Join(entry.Phones, ";")
		emails := strings.Join(entry.Emails, ";")
		if err = writer.Write([]string{entry.Uid, entry.FirstName, entry.LastName, phones, emails}); err != nil {
			log.Panicf("error writing record to csv: %v", err)
		}
	}
	return nil
}

func parseRecords(reader *csv.Reader, showErrors bool) ([]Entry, error) {
	var result []Entry
	var errs []error
	var row = 1
	for {
		row++
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Panicf("error reading record from csv: %v", err)
		}
		if len(record) != len(ImportColumnNames) {
			if showErrors {
				log.Printf("Skipping row %d: too few fields (%d)", row, len(record))
			}
			continue
		}
		var entry Entry
		if entry.Uid, err = ParseDate(record[0]); err != nil {
			if showErrors {
				log.Printf("Skipping row %d: invalid date: %v", row, err)
			}
			continue
		}
		if entry.Uid == "" {
			// silently skip blanks
			continue
		}
		if entry.FirstName, entry.LastName, err = ParseNames(record[1], record[2]); err != nil {
			if showErrors {
				log.Printf("Skipping row %d: invalid name: %v", row, err)
			}
			continue
		}
		if entry.FirstName == "" && entry.LastName == "" {
			// silently skip blanks
			continue
		}
		entry.Phones, errs = ParsePhones(record[3])
		if len(errs) > 0 {
			if showErrors {
				for i, e := range errs {
					log.Printf(
						"Ignoring %sphone in row %d (%s %s): %v",
						ordinal(i+1, len(errs)), row, entry.FirstName, entry.LastName, e,
					)
				}
			}
		}
		if len(entry.Phones) == 0 {
			// silently skip entries with no valid phones
			continue
		}
		if entry.Emails, errs = ParseEmails(record[4]); len(errs) > 0 {
			if showErrors {
				// No one cares about email errors
			}
		}
		result = append(result, entry)
	}
	return result, nil
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

func ordinal(i int, max int) string {
	if max == i {
		if i == 1 {
			return ""
		}
		return "last "
	}
	switch i {
	case 1:
		return "first "
	case 2:
		return "second "
	case 3:
		return "third "
	default:
		return fmt.Sprintf("%dth ", i)
	}
}
