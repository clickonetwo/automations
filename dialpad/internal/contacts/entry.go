/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package contacts

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/go-test/deep"
)

type Entry struct {
	FullId    string   `json:"id,omitempty"`
	Uid       string   `json:"uid"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phones    []string `json:"phones"`
	Emails    []string `json:"emails"`
}

type Anomaly struct {
	Entry
	Diff []string
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

func CompareById(left, right []Entry) (both, leftOnly, rightOnly []Entry, anomalies []Anomaly) {
	leftMap := make(map[string]Entry, len(left))
	for _, e := range left {
		leftMap[e.FullId] = e
	}
	rightMap := make(map[string]Entry, len(right))
	for _, e := range right {
		rightMap[e.FullId] = e
	}
	for _, l := range left {
		if r, ok := rightMap[l.FullId]; !ok {
			leftOnly = append(leftOnly, l)
		} else {
			both = append(both, l)
			if diff := deep.Equal(l, r); diff != nil {
				anomalies = append(anomalies, Anomaly{r, diff})
				log.Printf("Uid %s diff: %v", l.Uid, diff)
			}
		}
	}
	for _, r := range right {
		if _, ok := leftMap[r.FullId]; !ok {
			rightOnly = append(rightOnly, r)
		}
		// both case already handled above
	}
	return
}

func FindOffsetDuplicates(left, right []Entry, offset int64) (leftDupes, rightDupes [][]string) {
	// we sort the inputs so self-comparisons always find left < right
	telMap := make(map[string][]Entry, len(left))
	for _, e := range left {
		telMap[e.Phones[0]] = append(telMap[e.Phones[0]], e)
	}
	for _, r := range right {
		for _, l := range telMap[r.Phones[0]] {
			if l.FirstName != r.FirstName || l.LastName != r.LastName {
				continue
			}
			_, li := ExtractUid(l.FullId)
			_, ri := ExtractUid(r.FullId)
			if li+offset == ri {
				rightDupes = append(rightDupes, []string{l.FullId, r.FullId, l.Phones[0], l.FirstName, l.LastName})
				break
			} else if ri+offset == li {
				leftDupes = append(leftDupes, []string{l.FullId, r.FullId, l.Phones[0], l.FirstName, l.LastName})
				break
			}
		}
	}
	return
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

func ExtractUid(uid string) (string, int64) {
	reBig := regexp.MustCompile(`\A.*_uid_([0-9]+)\z`)
	bigMatch := reBig.FindStringSubmatch(uid)
	if len(bigMatch) > 0 {
		i, _ := strconv.ParseInt(bigMatch[1], 10, 64)
		return bigMatch[1], i
	}
	reDigits := regexp.MustCompile(`\A([0-9]+)\z`)
	digitMatch := reDigits.FindStringSubmatch(uid)
	if len(digitMatch) > 0 {
		i, _ := strconv.ParseInt(digitMatch[1], 10, 64)
		return digitMatch[1], i
	}
	panic(fmt.Errorf("invalid UID found in contact: %s", uid))
}
