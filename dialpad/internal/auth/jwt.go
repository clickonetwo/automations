/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/clickonetwo/automations/dialpad/internal/middleware"
)

func ValidateDialpadJwt(c *gin.Context, signed, secret string) (json.RawMessage, error) {
	validator := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// notest
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}
	token, err1 := jwt.Parse(signed, validator, jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}))
	if err1 != nil {
		middleware.CtxLogS(c).Errorf("Invalid dialpad token: %v", err1)
		return nil, err1
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		middleware.CtxLogS(c).Errorf("Invalid dialpad token claims: %#v", token.Claims)
	}
	bytes, err2 := json.Marshal(claims)
	if err2 != nil {
		middleware.CtxLogS(c).Errorf("Token claims cannot be marshaled: %v", err2)
		return nil, err2
	}
	return bytes, nil
}

func MakeNonce() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("Could not generate nonce: %v", err))
	}
	return hex.EncodeToString(b)
}
