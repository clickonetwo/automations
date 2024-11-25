/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package contacts

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

var (
	nonGeographicAreaCodes = mapset.NewSet(
		"500", "511",
		"521", "522", "523", "524", "525", "526", "527", "528", "529",
		"532", "533", "535", "538",
		"542", "543", "544", "545", "546", "547", "549",
		"550", "552", "553", "554", "555", "556", "558",
		"566", "569",
		"576", "577", "578", "588", "589",
	)
	twoDigitCountryCodes = mapset.NewSet(
		"20", "27", "30", "31", "32", "33", "34", "36", "39",
		"40", "41", "43", "44", "45", "46", "47", "48", "49",
		"51", "52", "53", "54", "55", "56", "57", "58",
		"60", "61", "62", "63", "64", "65", "66",
		"70", "71", "72", "73", "74", "75", "76", "77", "78", "79",
		"81", "82", "84", "86", "87", "88",
		"90", "91", "92", "93", "94", "95", "98",
	)
	NoContent     = errors.New("no content")
	nonDigitsOnly = regexp.MustCompile(`\D+`)
)

// CanonicalizePhoneNumber validates a phone number and returns it in E.164 format.
//
// The canonicalization is intended to make the number acceptable to Dialpad, so
// it's based on experience with what Dialpad will accept.  Here are the steps:
//
//  1. Non-numeric characters other than a leading '+' are removed.
//  2. An initial '011' is replaced with '+'.
//  3. An initial "+1" classifies a number as local (North American).
//     Any other initial "+ digit" prefix means it's international.
//  4. Numbers with no '+' and more than 10 digits must be international, so a '+' is prepended.
//  5. Numbers with no '+' and 10 or fewer digits are assumed to be local, so a '+1' is prepended.
//  6. If a local number has fewer than 10 digits after the +1, it's rejected.
//  7. If a local number, after the area code, has a prefix starting with "0" or "1" (such as `510-123-4567`),
//     it's rejected because North American prefixes can't start with either of those numbers.
//     (This typically means it's an international number that has not been prefixed correctly with a '+'.)
//  8. If a local number has one of the "non-geographic" (aka 5XX) area codes
//     that are reserved for machine-to-machine communication,
//     Dialpad will not accept it, so it is rejected.
func CanonicalizePhoneNumber(phoneNumber string) (string, error) {
	number := strings.TrimSpace(phoneNumber)
	digits := nonDigitsOnly.ReplaceAllString(number, "")
	if number[0] == '+' {
		number = "+" + digits
	} else if strings.HasPrefix(digits, "011") {
		digits = digits[3:]
		number = "+" + digits
	} else {
		number = digits
	}
	if digits == "" {
		return "", NoContent
	}
	class := "unknown"
	prefix := ""
	if strings.HasPrefix(number, "+1") {
		class = "local"
		prefix = "+1"
		digits = digits[1:]
	} else if strings.HasPrefix(number, "+") {
		class = "international"
		prefix = "+"
	} else if len(digits) > 10 {
		class = "international"
		prefix = "+"
	} else {
		class = "local"
		prefix = "+1"
	}
	if class == "local" {
		if len(digits) < 10 {
			return "", fmt.Errorf("too few digits: %q", phoneNumber)
		}
		if digits[3] == '0' || digits[3] == '1' {
			return "", fmt.Errorf("invalid prefix (starts with 0 or 1): %q", phoneNumber)
		}
		if nonGeographicAreaCodes.Contains(digits[0:3]) {
			return "", fmt.Errorf("non-geographic area code: %q", phoneNumber)
		}
	}
	return prefix + digits, nil
}

// ParsePhones takes a sequence of phone numbers (separated by ',', ';', '/', or '|')
// and returns a slice of the valid ones in canonical form.
//
// It also returns a slice of errors, one for each of the non-valid phone numbers.
func ParsePhones(phones string) (results []string, errs []error) {
	phones = strings.TrimSpace(phones)
	if phones == "" {
		return nil, nil
	}
	re := regexp.MustCompile(`[,;/|]`)
	candidates := re.Split(phones, -1)
	for _, c := range candidates {
		if strings.TrimSpace(c) == "" {
			continue
		}
		if result, err := CanonicalizePhoneNumber(c); err != nil {
			if !errors.Is(err, NoContent) {
				errs = append(errs, err)
			}
		} else {
			results = append(results, result)
		}
	}
	return
}

// CanonicalizeEmail validates an email and returns it in canonical form.
func CanonicalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", NoContent
	}
	a, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("%v: %q", err, email)
	}
	return a.Address, nil
}

// ParseEmails takes a sequence of emails (separated by ',', ';', or '|')
// and returns a slice of the valid ones in canonical form.
//
// It also returns a slice of errors, one for each of the non-valid emails.
func ParseEmails(emails string) (results []string, errs []error) {
	if strings.TrimSpace(emails) == "" {
		return nil, nil
	}
	re := regexp.MustCompile(`[,;|]`)
	candidates := re.Split(emails, -1)
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if result, err := CanonicalizeEmail(c); err != nil {
			if !errors.Is(err, NoContent) {
				errs = append(errs, err)
			}
		} else {
			results = append(results, result)
		}
	}
	return
}

// CanonicalizeName validates a name and returns it in canonical form
//
// Dialpad accepts all Unicode characters in names
func CanonicalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", NoContent
	}
	return name, nil
}

// ParseNames takes a first and last name and returns them in canonical form.
//
// Because Dialpad requires both first and last name, we will use the
// same name for both if only one is provided.
func ParseNames(first, last string) (f string, l string, e error) {
	if strings.TrimSpace(first) == "" && strings.TrimSpace(last) == "" {
		return
	}
	var ef, el error
	f, ef = CanonicalizeName(first)
	l, el = CanonicalizeName(last)
	if ef != nil && el != nil {
		e = fmt.Errorf("invalid first (%v) and last (%v) name", ef, el)
		return
	} else if ef != nil {
		f = l
	} else if el != nil {
		l = f
	}
	return
}

// CanonicalizeDate takes a creation date and returns it as a Unix time
func CanonicalizeDate(ts string) (string, error) {
	ts = strings.TrimSpace(ts)
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
	t, err := time.ParseInLocation("01/02/2006 03:04:05 PM", ts, loc)
	if err != nil {
		return "", fmt.Errorf("%v: %q", err, ts)
	}
	return fmt.Sprintf("%d", t.Unix()), nil
}

// ParseDate takes a date and returns it as a Unix time
func ParseDate(ts string) (string, error) {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return "", nil
	}
	return CanonicalizeDate(ts)
}
