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
 * Miscellaneos entities, functs and functions.
 *
 * =======================================================================
 */
package game

import "goquake2/shared"

const START_OFF = 1

func spLight(self *edict_t, G *qGame) error {
	if self == nil {
		return nil
	}

	/* no targeted lights in deathmatch, because they cause global messages */
	// if (!self->targetname || deathmatch->value) {
	// 	G_FreeEdict(self);
	// 	return;
	// }

	if self.Style >= 32 {
		// self.use = light_use;

		if (self.Spawnflags & START_OFF) != 0 {
			return G.gi.Configstring(shared.CS_LIGHTS+self.Style, "a")
		} else {
			return G.gi.Configstring(shared.CS_LIGHTS+self.Style, "m")
		}
	}
	return nil
}
