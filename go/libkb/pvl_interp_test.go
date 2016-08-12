// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

import (
	"testing"
)

type substituteTest struct {
	a, b string
}

var substituteTests = []substituteTest{
	{a: "%{}", b: "%{}"},
	{a: "%{invalid}", b: "%{invalid}"},
	{a: "%{username_service}", b: "kronk"},
	{a: "x%{username_service}y%{sig_id_short}z", b: "xkronky000z"},
	{a: "http://git(?:hub)?%{username_service}/%20%%{sig_id_short}}{%{}}", b: "http://git(?:hub)?kronk/%20%000}{%{}}"},
	{a: "^%{hostname}/(?:.well-known/keybase.txt|keybase.txt)$", b: "^example\\.com/(?:.well-known/keybase.txt|keybase.txt)$"},
	{a: "^.*%{sig_id_short}.*$", b: "^.*000.*$"},
	{a: "^keybase-site-verification=%{sig_id_short}$", b: "^keybase-site-verification=000$"},
	{a: "^%{sig_id_medium}$", b: "^sig%\\{sig_id_medium\\}\\.\\*\\$\\(\\^\\)\\\\/$"},
}

var sampleVars = PvlScriptVariables{
	UsernameService:  "kronk",
	UsernameKeybase:  "kronk_on_kb",
	Sig:              []byte{1, 2, 3, 4, 5},
	SigIDMedium:      "sig%{sig_id_medium}.*$(^)\\/",
	SigIDShort:       "000",
	Hostname: "example.com",
}

func TestSubstitute(t *testing.T) {
	for _, test := range substituteTests {
		res, err := pvlSubstitute(test.a, sampleVars, nil)
		if err != nil {
			t.Errorf("subsitute returned an error: %v", err)
		}
		if res != test.b {
			t.Logf("lens: %v %v", len(res), len(test.b))
			t.Errorf("wrong substitute result\n%v\n%v\n%v",
				test.a, res, test.b)
		}
	}
}
