// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// Export-Import for RPC for Teams

package teams

import (
	keybase1 "github.com/keybase/client/go/protocol/keybase1"
)

func (t *Team) ExportToTeamPlusAllKeys(idTime keybase1.Time) keybase1.TeamPlusAllKeys {
	var perTeamKeys = make(map[int]keybase1.PerTeamKey)
	var err error
	var i = 0
	for err == nil {
		perTeamKeys[i], err = t.Chain.GetPerTeamKeyAtGeneration(i)
		i++
	}
	// hack
	// for i := 0; i < t.Chain.GetLatestSeqno(); i++ {
	// 	if err != nil {
	// 		t.G().Log.Error("error getting perTeamKey: %s", err)
	// 	}
	// }
	var members keybase1.TeamMembers
	members, err = t.Members()
	if err != nil {
		t.G().Log.Error("error getting members: %s", err)
	}
	ret := keybase1.TeamPlusAllKeys{
		Id:          t.Chain.GetID(),
		Name:        t.Chain.GetName(),
		PerTeamKeys: perTeamKeys,
		Writers:     members.Writers,
		Readers:     members.Readers, //OPT, remove writers
	}

	return ret
}
