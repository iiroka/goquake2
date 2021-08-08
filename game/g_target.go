/*
 * Copyright (C) 1997-2001 Id Software, Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 * See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA
 * 02111-1307, USA.
 *
 * =======================================================================
 *
 * Targets.
 *
 * =======================================================================
 */
package game

import (
	"fmt"
	"strings"
)

func spTargetSpeaker(ent *edict_t, G *qGame) error {

	if ent == nil {
		return nil
	}

	if len(G.st.Noise) == 0 {
		G.gi.Dprintf("target_speaker with no noise set at %s\n", vtos(ent.s.Origin[:]))
		return nil
	}

	var buffer string
	if !strings.Contains(G.st.Noise, ".wav") {
		buffer = fmt.Sprintf("%s.wav", G.st.Noise)
	} else {
		buffer = G.st.Noise
	}

	ent.noise_index = G.gi.Soundindex(buffer)

	if ent.Volume == 0 {
		ent.Volume = 1.0
	}

	if ent.Attenuation == 0 {
		ent.Attenuation = 1.0
	} else if ent.Attenuation == -1 { /* use -1 so 0 defaults to 1 */
		ent.Attenuation = 0
	}

	/* check for prestarted looping sound */
	if (ent.Spawnflags & 1) != 0 {
		ent.s.Sound = ent.noise_index
	}

	// ent.use = Use_Target_Speaker

	/* must link the entity so we get areas and clusters so
	   the server can determine who to send updates to */
	G.gi.Linkentity(ent)
	return nil
}
