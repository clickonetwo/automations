/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"slices"
)

type SmsEvent struct {
	Date       int64 // UnixMicro
	MessageId  string
	Name       string
	Email      string
	TargetType string
	TargetId   int64
	SenderId   int64
	Direction  string
	ToPhones   []string
	FromPhone  string
	Text       string
	MmsUrl     string
}

func SelectThreadByEmailPhone(email, phone string, events []SmsEvent) []SmsEvent {
	var es []SmsEvent
	for _, event := range events {
		if event.Email == email {
			if slices.Contains(event.ToPhones, phone) || event.FromPhone == phone {
				es = append(es, event)
			}
		}
	}
	return es
}
