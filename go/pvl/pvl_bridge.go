// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package pvl

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/keybase/client/go/logger"
	keybase1 "github.com/keybase/client/go/protocol"
	jsonw "github.com/keybase/go-jsonw"
)

type PowerBox interface {
	// Data
	HintURL() string
	SigBody() []byte
	SigID() keybase1.SigID
	UsernameKeybase() string
	UsernameService() string
	Hostname() string

	// Fetchers
	GetText(url string) (string, ProofError)
	GetHTML(url string) (*goquery.Document, ProofError)
	GetJSON(url string) (*jsonw.Wrapper, ProofError)

	// Logging
	Log() logger.Logger

	// Manipulations
	FindBase64Block(haystack string, needle []byte) bool
	WhitespaceNormalize(string) string
}

type ProofError interface {
	error
	GetProofStatus() keybase1.ProofStatus
	GetDesc() string
}

type ProofErrorImpl struct {
	Status keybase1.ProofStatus
	Desc   string
}

func NewProofError(s keybase1.ProofStatus, d string, a ...interface{}) *ProofErrorImpl {
	return &ProofErrorImpl{s, fmt.Sprintf(d, a...)}
}

func (e *ProofErrorImpl) Error() string {
	return fmt.Sprintf("%s (code=%d)", e.Desc, int(e.Status))
}

func (e *ProofErrorImpl) GetProofStatus() keybase1.ProofStatus { return e.Status }
func (e *ProofErrorImpl) GetDesc() string                      { return e.Desc }
