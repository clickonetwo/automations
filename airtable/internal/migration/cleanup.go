/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package migration

import (
	"fmt"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

var (
	nonDigitsOnly        = regexp.MustCompile(`\D+`)
	spacesDashesOnly     = regexp.MustCompile(`[-—‑\s]+`)
	zone1NumberFront     = regexp.MustCompile(`^\s*(\(\d{3}\)\s*\d{3}(?:\s*-?\s*)\d{4})(?:\D|$)`)
	zone1Plus1           = regexp.MustCompile(`^\(\+?0?0?1\)(\(\d{3}\)\d{3}\d{4})(?:\D|$)`)
	zone1AreaFront       = regexp.MustCompile(`^(\(\d{3}\)\(\d{3}\)\d{4})(\D|$)`)
	zone1DoubleArea      = regexp.MustCompile(`^\((\d{3})\)(\((\d{3})\)\d{3}\d{4})(?:\D|$)`)
	zone1AreaZipArea     = regexp.MustCompile(`^\((\d{3})\)\(\d{5}\)(?:\+?0?0?1)?(\(?(\d{3})\)?\d{3}\d{4})(?:\D|$)`)
	zone1ZipArea         = regexp.MustCompile(`^\(\+?0?0?1\)\(\d{5}\)(?:\+?0?0?1)?(\(?\d{3}\)?\d{3}\d{4})(?:\D|$)`)
	zone1ZipFront        = regexp.MustCompile(`^\(\d{5}\)(\(\d{3}\)\d{3}\d{4})(?:\D|$)`)
	zone1ZipDouble       = regexp.MustCompile(`^\(\d{5}\)\(\d{5}\)(?:\+?0?0?1)?(\(?\d{3}\)?\d{3}\d{4})(?:\D|$)`)
	zone1DigitsOnly      = regexp.MustCompile(`^(?:\+?1)?(\(?\d{3}\)?\d{7})(?:\D|$)`)
	zone1AreaTenDigits   = regexp.MustCompile(`^\([0+]?0?1?\)\((?:1|\d{3})\)(\d{10})(?:\D|$)`)
	intlPlus             = regexp.MustCompile(`(^\(?\+0?0?[2-9]\d\d?\)?\(?\d+\)?\d+)(?:\D|$)`)
	intlNonPlus          = regexp.MustCompile(`(^\([2-9]\d\d?\)\(?\d+\)?\d+)(?:\D|$)`)
	twoDigitCountryCodes = mapset.NewSet(
		"20", "27",
		"30", "31", "32", "33", "34", "36", "39",
		"40", "41", "43", "44", "45", "46", "47", "48", "49",
		"51", "52", "53", "54", "55", "56", "57", "58",
		"60", "61", "62", "63", "64", "65", "66",
		"70", "71", "72", "73", "74", "75", "76", "77", "78", "79",
		"81", "82", "84", "86", "87", "88",
		"90", "91", "92", "93", "94", "95", "98",
	)
)

func cleanField(fromCol, val string, toRow map[string]string) error {
	switch fromCol {
	case "Outside of US Phone Number":
		if e164, ok := toRow["E.164 number"]; !ok {
			return fmt.Errorf("%q processed before %q", fromCol, "Phone Number")
		} else if e164 != "" {
			// already have a US phone number, ignore this one
			break
		}
		fallthrough
	case "Phone Number":
		phone := findValidPhone(val)
		if phone == "" {
			// can't be cleaned
			toRow["E.164 number"] = ""
			toRow["Phone"] = val
		} else {
			toRow["E.164 number"] = phone
			f, err := formatE164(phone)
			if err != nil {
				return err
			}
			toRow["Phone"] = f
		}
	case "State / Province":
		toRow["Migrated State"] = val
	case "Preferred Language":
		language := preferredLanguage(val)
		if language == "" && val != "" {
			return fmt.Errorf("preferred language not mapped: %q", val)
		}
		toRow["Preferred Language"] = language
	case "If other language, please specify:":
		if pl, ok := toRow["Preferred Language"]; !ok {
			return fmt.Errorf("%q processed before %q", fromCol, "Preferred Language")
		} else if pl != "" {
			if o := preferredLanguage(val); o == pl {
				// their other preferred language is already their preferred language
				toRow["Preferred Language (Other)"] = ""
				break
			}
		}
		toRow["Preferred Language (Other)"] = val
	case "What type of legal assistance are you looking for?":
		toRow["Requested Legal Assistance"] = assistanceType(val)
	case `Do you have a case in immigration court ("removal proceedings")?`:
		c := removalCase(val)
		if c == "" && val != "" {
			return fmt.Errorf("removal case not mapped: %q", val)
		}
		toRow["In Removal Proceedings"] = c
	default:
		return fmt.Errorf("don't know how to clean fromField %q", fromCol)
	}
	return nil
}

func findValidPhone(s string) string {
	if match := zone1NumberFront.FindStringSubmatch(s); match != nil {
		return makeE164(spacesDashesOnly.ReplaceAllString(match[1], ""), true)
	}
	ns := spacesDashesOnly.ReplaceAllString(s, "")
	if ns == "" {
		return ""
	}
	if match := zone1Plus1.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], true)
	}
	if match := zone1AreaFront.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], true)
	}
	if match := zone1DoubleArea.FindStringSubmatch(ns); match != nil && match[1] == match[3] {
		return makeE164(match[2], true)
	}
	if match := zone1AreaZipArea.FindStringSubmatch(ns); match != nil && match[1] == match[3] {
		return makeE164(match[2], true)
	}
	if match := zone1ZipArea.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], true)
	}
	if match := zone1ZipFront.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], true)
	}
	if match := zone1ZipDouble.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], true)
	}
	if match := zone1DigitsOnly.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], true)
	}
	if match := zone1AreaTenDigits.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], true)
	}
	if match := intlPlus.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], false)
	}
	if match := intlNonPlus.FindStringSubmatch(ns); match != nil {
		return makeE164(match[1], false)
	}
	return ""
}

func makeE164(s string, z1 bool) string {
	do := nonDigitsOnly.ReplaceAllString(s, "")
	if z1 {
		if strings.HasPrefix(s, "1") {
			return "+" + do
		}
		return "+1" + do
	}
	// international, find country code
	cc, rest := do[:2], do[2:]
	if !twoDigitCountryCodes.Contains(cc) {
		cc, rest = do[:3], do[3:]
	}
	// strip leading 0*1*
	for rest[0] == '0' || rest[0] == '1' {
		rest = rest[1:]
	}
	// remove duplicated country code
	if strings.HasPrefix(rest, cc) {
		rest = rest[len(cc):]
	}
	if len(rest) < 5 {
		return ""
	}
	return "+" + cc + rest
}

func formatE164(phone string) (string, error) {
	if strings.HasPrefix(phone, "+1") {
		// Zone 1
		if len(phone) != 12 {
			return "", fmt.Errorf("invalid zone 1 phone (length %d): %q", len(phone), phone)
		}
		return fmt.Sprintf("(%s) %s-%s", phone[2:5], phone[5:8], phone[8:12]), nil
	} else if strings.HasPrefix(phone, "+") {
		if len(phone) < 10 || len(phone) > 16 {
			return "", fmt.Errorf("invalid international phone (length %d): %q", len(phone), phone)
		}
		// international, separate the country code
		prefix, suffix := phone[1:3], phone[3:]
		if !twoDigitCountryCodes.Contains(prefix) {
			prefix, suffix = phone[1:4], phone[4:]
		}
		var suffixSuffix string
		if len(suffix)%3 == 1 {
			// put last 4 together
			suffix, suffixSuffix = suffix[:len(suffix)-4], "-"+suffix[len(suffix)-4:]
		}
		for i := len(suffix); i > 3; i = i - 3 {
			suffix = suffix[0:i-3] + "-" + suffix[i-3:]
		}
		return fmt.Sprintf("(+%s) %s%s", prefix, suffix, suffixSuffix), nil
	}
	return "", fmt.Errorf("invalid E164 phone (no +): %q", phone)
}

func preferredLanguage(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(s, "en") || strings.HasPrefix(s, "in") {
		return "English"
	}
	if strings.HasPrefix(s, "sp") || strings.HasPrefix(s, "es") {
		return "Spanish"
	}
	if strings.HasPrefix(s, "po") {
		return "Portuguese"
	}
	if strings.HasPrefix(s, "o") {
		return "Other"
	}
	return ""
}

func assistanceType(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(s, "as") {
		return "Asylum"
	}
	if strings.HasPrefix(s, "re") {
		return "Residency"
	}
	if strings.HasPrefix(s, "na") {
		return "Naturalization"
	}
	if strings.Contains(s, "vawa") {
		return "VAWA petition"
	}
	if strings.HasPrefix(s, "pe") {
		return "Petition for a family member"
	}
	return "I'm not sure/Other"
}

func removalCase(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(s, "n") {
		return "No"
	}
	if strings.HasPrefix(s, "y") || strings.HasPrefix(s, "s") {
		return "Yes"
	}
	if strings.HasPrefix(s, "i") {
		return "I'm not sure"
	}
	return ""
}
