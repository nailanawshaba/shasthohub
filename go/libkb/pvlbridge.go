// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/keybase/client/go/logger"
	keybase1 "github.com/keybase/client/go/protocol"
	pvl "github.com/keybase/client/go/pvl"
	jsonw "github.com/keybase/go-jsonw"
)

type PvlBridge struct {
	g       *GlobalContext
	rc      RemoteProofChainLink
	h       SigHint
	sigbody []byte
	sigid   keybase1.SigID
}

var _ pvl.PowerBox = (*PvlBridge)(nil)

func PvlCheckProof(g *GlobalContext, pvlchunk *jsonw.Wrapper, service keybase1.ProofType, rc RemoteProofChainLink, h SigHint) ProofError {
	sigBody, sigID, err := OpenSig(rc.GetArmoredSig())
	if err != nil {
		return NewProofError(keybase1.ProofStatus_BAD_SIGNATURE,
			"Bad signature: %v", err)
	}
	bridge := PvlBridge{g, rc, h, sigBody, sigID}

	return pvl.CheckProof(&bridge, pvlchunk, service)
}

func (p *PvlBridge) HintURL() string {
	return p.h.apiURL
}

func (p *PvlBridge) SigBody() []byte {
	return p.sigbody
}

func (p *PvlBridge) SigID() keybase1.SigID {
	return p.sigid
}

func (p *PvlBridge) UsernameKeybase() string {
	return p.rc.GetUsername()
}

func (p *PvlBridge) UsernameService() string {
	return p.rc.GetRemoteUsername()
}

func (p *PvlBridge) Hostname() string {
	return p.rc.GetHostname()
}

func (p *PvlBridge) GetText(url string) (string, pvl.ProofError) {
	res, err := p.g.XAPI.GetText(NewAPIArg(p.g, url))
	if err != nil {
		return "", XapiError(err, url)
	}
	return res.Body, nil
}

func (p *PvlBridge) GetHTML(url string) (*goquery.Document, pvl.ProofError) {
	res, err := p.g.XAPI.GetHTML(NewAPIArg(p.g, url))
	if err != nil {
		return nil, XapiError(err, url)
	}
	return res.GoQuery, nil
}

func (p *PvlBridge) GetJSON(url string) (*jsonw.Wrapper, pvl.ProofError) {
	res, err := p.g.XAPI.Get(NewAPIArg(p.g, url))
	if err != nil {
		return nil, XapiError(err, url)
	}
	return res.Body, nil
}

func (p *PvlBridge) Log() logger.Logger {
	return p.g.Log
}

func (p *PvlBridge) FindBase64Block(haystack string, needle []byte) bool {
	return FindBase64Block(haystack, needle, false)
}

func (p *PvlBridge) WhitespaceNormalize(x string) string {
	return WhitespaceNormalize(x)
}
