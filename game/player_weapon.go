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
 * Player weapons.
 *
 * =======================================================================
 */
package game

import (
	"goquake2/game/misc"
	"goquake2/shared"
)

/*
 * The old weapon has been dropped all
 * the way, so make the new one current
 */
func (G *qGame) changeWeapon(ent *edict_t) {

	if ent == nil {
		return
	}

	//  if (ent.client.grenade_time) {
	// 	 ent->client->grenade_time = level.time;
	// 	 ent->client->weapon_sound = 0;
	// 	 weapon_grenade_fire(ent, false);
	// 	 ent->client->grenade_time = 0;
	//  }

	ent.client.pers.lastweapon = ent.client.pers.weapon
	ent.client.pers.weapon = ent.client.newweapon
	ent.client.newweapon = nil
	ent.client.machinegun_shots = 0

	/* set visible model */
	if ent.s.Modelindex == 255 {
		var i int
		if ent.client.pers.weapon != nil {
			i = ((ent.client.pers.weapon.weapmodel & 0xff) << 8)
		} else {
			i = 0
		}

		ent.s.Skinnum = (ent.index - 1) | i
	}

	if ent.client.pers.weapon != nil && len(ent.client.pers.weapon.ammo) > 0 {
		ent.client.ammo_index = G.findItemIndex(ent.client.pers.weapon.ammo)
	} else {
		ent.client.ammo_index = 0
	}

	if ent.client.pers.weapon == nil {
		/* dead */
		ent.client.ps.Gunindex = 0
		return
	}

	// ent.client.weaponstate = WEAPON_ACTIVATING
	ent.client.ps.Gunframe = 0
	ent.client.ps.Gunindex = G.gi.Modelindex(ent.client.pers.weapon.view_model)

	ent.client.anim_priority = ANIM_PAIN

	if (ent.client.ps.Pmove.Pm_flags & shared.PMF_DUCKED) != 0 {
		ent.s.Frame = misc.FRAME_crpain1
		ent.client.anim_end = misc.FRAME_crpain4
	} else {
		ent.s.Frame = misc.FRAME_pain301
		ent.client.anim_end = misc.FRAME_pain304
	}
}

/*
 * Called by ClientBeginServerFrame and ClientThink
 */
func (G *qGame) thinkWeapon(ent *edict_t) {
	if ent == nil {
		return
	}

	/* if just died, put the weapon away */
	if ent.Health < 1 {
		ent.client.newweapon = nil
		G.changeWeapon(ent)
	}

	/* call active weapon think routine */
	//  if (ent.client.pers.weapon && ent.client.pers.weapon.weaponthink) {
	// 	 is_quad = (ent->client->quad_framenum > level.framenum);

	// 	 if (ent->client->silencer_shots) {
	// 		 is_silenced = MZ_SILENCED;
	// 	 }
	// 	 else
	// 	 {
	// 		 is_silenced = 0;
	// 	 }

	// 	 ent->client->pers.weapon->weaponthink(ent);
	//  }
}
