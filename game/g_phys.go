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
 * Quake IIs legendary physic engine.
 *
 * =======================================================================
 */
package game

/*
 * Runs thinking code for
 * this frame if necessary
 */
func (G *qGame) svRunThink(ent *edict_t) bool {

	if ent == nil {
		return false
	}

	thinktime := ent.nextthink

	if thinktime <= 0 {
		return true
	}

	if thinktime > G.level.time+0.001 {
		return true
	}

	ent.nextthink = 0

	//  if (!ent->think)
	//  {
	// 	 gi.error("NULL ent->think");
	//  }

	//  ent->think(ent);

	return false
}

/* ================================================================== */

/*
 * Non moving objects can only think
 */
func (G *qGame) svPhysics_None(ent *edict_t) {
	if ent == nil {
		return
	}

	/* regular thinking */
	G.svRunThink(ent)
}

/* ================================================================== */

func (G *qGame) runEntity(ent *edict_t) error {
	if ent == nil {
		return nil
	}

	// if (ent->prethink)
	// {
	// 	ent->prethink(ent);
	// }

	switch ent.movetype {
	// case MOVETYPE_PUSH:
	// 	case MOVETYPE_STOP:
	// 		SV_Physics_Pusher(ent);
	// 		break;
	case MOVETYPE_NONE:
		G.svPhysics_None(ent)
	// 	case MOVETYPE_NOCLIP:
	// 		SV_Physics_Noclip(ent);
	// 		break;
	// 	case MOVETYPE_STEP:
	// 		SV_Physics_Step(ent);
	// 		break;
	// 	case MOVETYPE_TOSS:
	// 	case MOVETYPE_BOUNCE:
	// 	case MOVETYPE_FLY:
	// 	case MOVETYPE_FLYMISSILE:
	// 		SV_Physics_Toss(ent);
	// 		break;
	default:
		return G.gi.Error("SV_Physics: bad movetype %v", ent.movetype)
	}
	return nil
}
