/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

type Target struct {
	Name string
	Type string
	Id   int64
}

var (
	OasisMain = Target{
		Name: "Oasis Main Office",
		Type: "office",
		Id:   5527348325810176,
	}
	OasisEnglish = Target{
		Name: "Oasis English Main Line",
		Type: "department",
		Id:   5655719774945280,
	}
	OasisSpanish = Target{
		Name: "Oasis Spanish Main Line",
		Type: "department",
		Id:   6592322884239360,
	}
	OasisPortuguese = Target{
		Name: "Oasis Portuguese Main Line",
		Type: "department",
		Id:   6152389518327808,
	}
)
