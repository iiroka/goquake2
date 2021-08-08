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
 * .sp2 sprites
 *
 * =======================================================================
 */
package gl3

import "goquake2/shared"

func (T *qGl3) loadSP2(mod *gl3model_t, buffer []byte) error {

	sprout := shared.Dsprite(buffer)

	if sprout.Version != shared.SPRITE_VERSION {
		T.ri.Sys_Error(shared.ERR_DROP, "%s has wrong version number (%v should be %v)",
			mod.name, sprout.Version, shared.SPRITE_VERSION)
	}

	if sprout.Numframes > shared.MAX_MD2SKINS {
		T.ri.Sys_Error(shared.ERR_DROP, "%s has too many frames (%v > %v)",
			mod.name, sprout.Numframes, shared.MAX_MD2SKINS)
	}

	mod.skins = make([]*gl3image_t, sprout.Numframes)
	for i := range sprout.Frames {
		mod.skins[i] = T.findImage(sprout.Frames[i].Name, it_sprite)
	}

	mod.extradata = sprout
	mod.mtype = mod_sprite
	return nil
}
