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
 * The basic AI functions like enemy detection, attacking and so on.
 *
 * =======================================================================
 */
package game

/*
 *
 * Used for standing around and looking
 * for players Distance is for slight
 * position adjustments needed by the
 * animations
 */
func ai_stand(self *edict_t, dist float32, G *qGame) {

	if self == nil || G == nil {
		return
	}

	// if (dist) {
	// 	M_walkmove(self, self->s.angles[YAW], dist);
	// }

	// if (self->monsterinfo.aiflags & AI_STAND_GROUND)
	// {
	// 	if (self->enemy)
	// 	{
	// 		VectorSubtract(self->enemy->s.origin, self->s.origin, v);
	// 		self->ideal_yaw = vectoyaw(v);

	// 		if ((self->s.angles[YAW] != self->ideal_yaw) &&
	// 			self->monsterinfo.aiflags & AI_TEMP_STAND_GROUND)
	// 		{
	// 			self->monsterinfo.aiflags &=
	// 				~(AI_STAND_GROUND | AI_TEMP_STAND_GROUND);
	// 			self->monsterinfo.run(self);
	// 		}

	// 		M_ChangeYaw(self);
	// 		ai_checkattack(self);
	// 	}
	// 	else
	// 	{
	// 		FindTarget(self);
	// 	}

	// 	return;
	// }

	// if (FindTarget(self))
	// {
	// 	return;
	// }

	// if (level.time > self->monsterinfo.pausetime)
	// {
	// 	self->monsterinfo.walk(self);
	// 	return;
	// }

	// if (!(self->spawnflags & 1) && (self->monsterinfo.idle) &&
	// 	(level.time > self->monsterinfo.idle_time))
	// {
	// 	if (self->monsterinfo.idle_time)
	// 	{
	// 		self->monsterinfo.idle(self);
	// 		self->monsterinfo.idle_time = level.time + 15 + random() * 15;
	// 	}
	// 	else
	// 	{
	// 		self->monsterinfo.idle_time = level.time + random() * 15;
	// 	}
	// }
}
