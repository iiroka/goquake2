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
 * Monster utility functions.
 *
 * =======================================================================
 */
package game

import "goquake2/shared"

/* ================================================================== */

func (G *qGame) monster_start(self *edict_t) bool {
	if self == nil {
		return false
	}

	// if (deathmatch.value) {
	// 	G_FreeEdict(self);
	// 	return false;
	// }

	// if ((self.spawnflags & 4) && !(self.monsterinfo.aiflags & AI_GOOD_GUY)) {
	// 	self->spawnflags &= ~4;
	// 	self->spawnflags |= 1;
	// }

	// if ((self->spawnflags & 2) && !self->targetname) {
	// 	if (g_fix_triggered->value) {
	// 		self->spawnflags &= ~2;
	// 	}

	// 	gi.dprintf ("triggered %s at %s has no targetname\n", self->classname, vtos (self->s.origin));
	// }

	// if (!(self->monsterinfo.aiflags & AI_GOOD_GUY)) {
	// 	level.total_monsters++;
	// }

	self.nextthink = G.level.time + FRAMETIME
	self.svflags |= shared.SVF_MONSTER
	self.s.Renderfx |= shared.RF_FRAMELERP
	// self.takedamage = DAMAGE_AIM
	// self.air_finished = level.time + 12
	// self.use = monster_use

	// if(!self->max_health) {
	// 	self->max_health = self->health;
	// }

	// self.clipmask = MASK_MONSTERSOLID

	self.s.Skinnum = 0
	self.deadflag = DEAD_NO
	// self->svflags &= ~SVF_DEADMONSTER;

	// if (!self->monsterinfo.checkattack)
	// {
	// 	self->monsterinfo.checkattack = M_CheckAttack;
	// }

	// VectorCopy(self->s.origin, self->s.old_origin);

	// if (st.item)
	// {
	// 	self->item = FindItemByClassname(st.item);

	// 	if (!self->item)
	// 	{
	// 		gi.dprintf("%s at %s has bad item: %s\n", self->classname,
	// 				vtos(self->s.origin), st.item);
	// 	}
	// }

	// /* randomize what frame they start on */
	// if (self->monsterinfo.currentmove)
	// {
	// 	self->s.frame = self->monsterinfo.currentmove->firstframe +
	// 		(randk() % (self->monsterinfo.currentmove->lastframe -
	// 				   self->monsterinfo.currentmove->firstframe + 1));
	// }

	return true
}

func (G *qGame) walkmonster_start(self *edict_t) {
	if self == nil {
		return
	}

	// self.think = walkmonster_start_go
	G.monster_start(self)
}
