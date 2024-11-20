/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package contacts

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/go-test/deep"
)

type Entry struct {
	Uid       string   `json:"uid"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phones    []string `json:"phones"`
	Emails    []string `json:"emails"`
}

func DiffEntries(dialpad, local []Entry) (update []Entry, create []Entry) {
	oldMap := make(map[string]Entry, len(dialpad))
	for _, o := range dialpad {
		oldMap[o.Uid] = o
	}
	for _, n := range local {
		o, ok := oldMap[n.Uid]
		if !ok {
			create = append(create, n)
		} else if diff := deep.Equal(o, n); diff != nil {
			update = append(update, n)
		}
	}
	return
}

func FindOffsetDuplicates(entries []Entry, offset int64) []Entry {
	keyMap := make(map[string][]Entry, len(entries))
	for _, e := range entries {
		if len(e.Phones) > 0 {
			keyMap[e.Phones[0]] = append(keyMap[e.Phones[0]], e)
		} else if len(e.Emails) > 0 {
			keyMap[e.Emails[0]] = append(keyMap[e.Emails[0]], e)
		} else {
			key := e.FirstName + "|" + e.LastName
			keyMap[key] = append(keyMap[key], e)
		}
	}
	results := make([]Entry, 0)
	for _, v := range keyMap {
		if len(v) != 2 {
			continue
		}
		e1, e2 := v[0], v[1]
		if e1.FirstName != e2.FirstName || e1.LastName != e2.LastName {
			continue
		}
		if diff := deep.Equal(e1.Emails, e2.Emails); diff != nil {
			continue
		}
		t1, err1 := ExtractUid(e1.Uid)
		t2, err2 := ExtractUid(e2.Uid)
		if err1 != nil || err2 != nil {
			continue
		}
		if t1+offset == t2 {
			results = append(results, e2)
		} else if t2+offset == t1 {
			results = append(results, e1)
		}
	}
	return results
}

func FindWithoutPhones(entries []Entry) []Entry {
	results := make([]Entry, 0)
	for _, entry := range entries {
		if len(entry.Phones) == 0 {
			results = append(results, entry)
		}
	}
	return results
}

func ExtractUid(uid string) (int64, error) {
	reBig := regexp.MustCompile(`\A.*_uid_([0-9]+)\z`)
	bigMatch := reBig.FindStringSubmatch(uid)
	if len(bigMatch) > 0 {
		return strconv.ParseInt(bigMatch[1], 10, 64)
	}
	reDigits := regexp.MustCompile(`\A[0-9]+\z`)
	digitMatch := reDigits.FindStringSubmatch(uid)
	if len(digitMatch) > 0 {
		return strconv.ParseInt(digitMatch[1], 10, 64)
	}
	return 0, errors.New("invalid uid")
}
