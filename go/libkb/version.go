// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

// defaultVersion is a default version if one isn't specified at build time
// (should be semver compatible).
const defaultVersion = "1.0.19+custom"

// Version should be set at compile time using a build flag such as
//   -X github.com/keybase/client/libkb.Version=1.2.3
// or for prerelease:
//   -X github.com/keybase/client/libkb.Version=1.2.3-400+commit
// CAUTION: Don't change the name of this variable without grepping for
// occurrences in shell scripts!
var Version string

// VersionString returns semantic version string
func VersionString() string {
	if Version != "" {
		return Version
	}
	return defaultVersion
}
