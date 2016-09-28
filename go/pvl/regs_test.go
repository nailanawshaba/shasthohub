// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package pvl

import "testing"

func TestNamedRegsStore(t *testing.T) {
	tests := []struct {
		shouldwork               bool
		op, arg1, arg2, expected string
	}{
		// banning a register works
		{true, "Ban", "banned", "", ""},
		{false, "Ban", "banned", "", ""},
		{false, "Set", "banned", "foo", ""},
		{false, "Get", "banned", "", ""},

		// set and get works
		// cannot ban a set register
		{true, "Set", "x", "foo", ""},
		{true, "Get", "x", "", "foo"},
		{false, "Ban", "x", "", ""},
		{true, "Get", "x", "", "foo"},

		// cannot set twice
		{true, "Set", "y", "bar", ""},
		{true, "Get", "y", "", "bar"},
		{false, "Set", "y", "baz", ""},
		{true, "Get", "y", "", "bar"},

		// cannot use invalid keys
		{false, "Set", "Z", "", ""},
		{false, "Get", "Z", "", ""},
		{false, "Set", "specialchar@", "oosh", ""},
		{false, "Set", "", "oosh", ""},
		{false, "Ban", "#!", "", ""},

		// can use valid keys
		{true, "Set", "tmp1_2", "fuzzle", ""},
		{true, "Get", "tmp1_2", "fuzzle", ""},

		// empty string is an ok value
		{true, "Set", "empty", "", ""},
		{true, "Get", "empty", "", ""},
	}
	regs := *newNamedRegsStore()
	// TODO READ TEST DATJKLDJFKDJ

	shouldBOk(regs.Ban("banned"))
	shouldErr(regs.Ban("banned"))
	shouldErr(regs.Set("banned", "foo"))

	shouldBOk(regs.Set("x", "foo"))
	shouldBOk2(regs.Get("banned"), "foo")
	shouldErr(regs.Set("banned"))

	shouldErr(regs.Get("banned"))

	regs.Get
	regs.Set
	// TODO test that the "" key never works
}
