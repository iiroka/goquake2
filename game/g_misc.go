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

func spPathCorner(self *edict_t, G *qGame) error {
	if self == nil || G == nil {
		return nil
	}

	if len(self.Targetname) == 0 {
		G.gi.Dprintf("path_corner with no targetname at %s\n",
			vtos(self.s.Origin[:]))
		G.gFreeEdict(self)
		return nil
	}

	self.solid = shared.SOLID_TRIGGER
	// self.touch = path_corner_touch
	self.mins = [3]float32{-8, -8, -8}
	self.maxs = [3]float32{8, 8, 8}
	self.svflags |= shared.SVF_NOCLIENT
	G.gi.Linkentity(self)
	return nil
}

/* ===================================================== */

/*
 * QUAKED point_combat (0.5 0.3 0) (-8 -8 -8) (8 8 8) Hold
 *
 * Makes this the target of a monster and it will head here
 * when first activated before going after the activator.  If
 * hold is selected, it will stay here.
 */
//  void
//  point_combat_touch(edict_t *self, edict_t *other, cplane_t *plane /* unused */,
// 		 csurface_t *surf /* unused */)
//  {
// 	 edict_t *activator;

// 	 if (!self || !other)
// 	 {
// 		 return;
// 	 }

// 	 if (other->movetarget != self)
// 	 {
// 		 return;
// 	 }

// 	 if (self->target)
// 	 {
// 		 other->target = self->target;
// 		 other->goalentity = other->movetarget = G_PickTarget(other->target);

// 		 if (!other->goalentity)
// 		 {
// 			 gi.dprintf("%s at %s target %s does not exist\n",
// 					 self->classname,
// 					 vtos(self->s.origin),
// 					 self->target);
// 			 other->movetarget = self;
// 		 }

// 		 self->target = NULL;
// 	 }
// 	 else if ((self->spawnflags & 1) && !(other->flags & (FL_SWIM | FL_FLY)))
// 	 {
// 		 other->monsterinfo.pausetime = level.time + 100000000;
// 		 other->monsterinfo.aiflags |= AI_STAND_GROUND;
// 		 other->monsterinfo.stand(other);
// 	 }

// 	 if (other->movetarget == self)
// 	 {
// 		 other->target = NULL;
// 		 other->movetarget = NULL;
// 		 other->goalentity = other->enemy;
// 		 other->monsterinfo.aiflags &= ~AI_COMBAT_POINT;
// 	 }

// 	 if (self->pathtarget)
// 	 {
// 		 char *savetarget;

// 		 savetarget = self->target;
// 		 self->target = self->pathtarget;

// 		 if (other->enemy && other->enemy->client)
// 		 {
// 			 activator = other->enemy;
// 		 }
// 		 else if (other->oldenemy && other->oldenemy->client)
// 		 {
// 			 activator = other->oldenemy;
// 		 }
// 		 else if (other->activator && other->activator->client)
// 		 {
// 			 activator = other->activator;
// 		 }
// 		 else
// 		 {
// 			 activator = other;
// 		 }

// 		 G_UseTargets(self, activator);
// 		 self->target = savetarget;
// 	 }
//  }

func spPointCombat(self *edict_t, G *qGame) error {
	if self == nil || G == nil {
		return nil
	}

	if G.deathmatch.Bool() {
		G.gFreeEdict(self)
		return nil
	}

	self.solid = shared.SOLID_TRIGGER
	// self.touch = point_combat_touch
	self.mins = [3]float32{-8, -8, -16}
	self.maxs = [3]float32{8, 8, 16}
	self.svflags = shared.SVF_NOCLIENT
	G.gi.Linkentity(self)
	return nil
}

const START_OFF = 1

func spLight(self *edict_t, G *qGame) error {
	if self == nil {
		return nil
	}

	/* no targeted lights in deathmatch, because they cause global messages */
	if len(self.Targetname) == 0 || G.deathmatch.Bool() {
		G.gFreeEdict(self)
		return nil
	}

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
