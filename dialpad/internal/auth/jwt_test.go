/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package auth

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang-jwt/jwt/v5"

	"github.com/clickonetwo/automations/dialpad/internal/middleware"
)

func TestValidateDialpadJwt(t *testing.T) {
	secret := MakeNonce()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"date_started": 123456})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	c, _ := middleware.CreateTestContext()
	if claims, err := ValidateDialpadJwt(c, signed, secret); err != nil {
		t.Fatal(err)
	} else {
		if expected, err := json.Marshal(map[string]interface{}{"date_started": any(123456)}); err != nil {
			t.Fatal(err)
		} else if diff := deep.Equal([]byte(claims), expected); diff != nil {
			t.Fatal(diff)
		}
	}
	if _, err := ValidateDialpadJwt(c, signed, MakeNonce()); err == nil {
		t.Errorf("validated dialpad jwt with wrong secret")
	}
}
