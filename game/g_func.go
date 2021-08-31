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
 * Level functions. Platforms, buttons, dooors and so on.
 *
 * =======================================================================
 */
package game

import "goquake2/shared"

/*
 * =========================================================
 *
 * PLATS
 *
 * movement options:
 *
 * linear
 * smooth start, hard stop
 * smooth start, smooth stop
 *
 * start
 * end
 * acceleration
 * speed
 * deceleration
 * begin sound
 * end sound
 * target fired when reaching end
 * wait at end
 *
 * object characteristics that use move segments
 * ---------------------------------------------
 * movetype_push, or movetype_stop
 * action when touched
 * action when blocked
 * action when used
 *  disabled?
 * auto trigger spawning
 *
 *
 * =========================================================
 */

/* ==================================================================== */

/*
 * QUAKED func_timer (0.3 0.1 0.6) (-8 -8 -8) (8 8 8) START_ON
 *
 * "wait"	base time between triggering all targets, default is 1
 * "random"	wait variance, default is 0
 *
 * so, the basic time between firing is a random time
 * between (wait - random) and (wait + random)
 *
 * "delay"			delay before first firing when turned on, default is 0
 * "pausetime"		additional delay used only the very first time
 *                  and only if spawned with START_ON
 *
 * These can used but not touched.
 */
func func_timer_think(self *edict_t, G *qGame) {
	if self == nil || G == nil {
		return
	}

	G.gUseTargets(self, self.activator)
	self.nextthink = G.level.time + self.Wait + shared.Crandk()*self.Random
}

func spFuncTimer(self *edict_t, G *qGame) error {
	if self == nil || G == nil {
		return nil
	}

	if self.Wait == 0 {
		self.Wait = 1.0
	}

	// self.use = func_timer_use
	self.think = func_timer_think

	if self.Random >= self.Wait {
		self.Random = self.Wait - FRAMETIME
		G.gi.Dprintf("func_timer at %s has random >= wait\n", vtos(self.s.Origin[:]))
	}

	if (self.Spawnflags & 1) != 0 {
		self.nextthink = G.level.time + 1.0 + G.st.pausetime + self.Delay +
			self.Wait + shared.Crandk()*self.Random
		self.activator = self
	}

	self.svflags = shared.SVF_NOCLIENT
	return nil
}
