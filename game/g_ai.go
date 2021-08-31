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

import "goquake2/shared"

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

	if (self.monsterinfo.aiflags & AI_STAND_GROUND) != 0 {
		// 	if (self->enemy) {
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

		return
	}

	if G.findTarget(self) {
		return
	}

	if G.level.time > self.monsterinfo.pausetime {
		self.monsterinfo.walk(self, G)
		return
	}

	if (self.Spawnflags&1) == 0 && (self.monsterinfo.idle != nil) &&
		(G.level.time > self.monsterinfo.idle_time) {
		if self.monsterinfo.idle_time > 0 {
			self.monsterinfo.idle(self, G)
			self.monsterinfo.idle_time = G.level.time + 15 + shared.Frandk()*15
		} else {
			self.monsterinfo.idle_time = G.level.time + shared.Frandk()*15
		}
	}
}

/*
 * The monster is walking it's beat
 */
func ai_walk(self *edict_t, dist float32, G *qGame) {

	if self == nil || G == nil {
		return
	}

	G.mMoveToGoal(self, dist)

	/* check for noticing a player */
	if G.findTarget(self) {
		return
	}

	// if (self.monsterinfo.search) && (level.time > self.monsterinfo.idle_time) {
	// 	if self.monsterinfo.idle_time {
	// 		self.monsterinfo.search(self)
	// 		self.monsterinfo.idle_time = level.time + 15 + random()*15
	// 	} else {
	// 		self.monsterinfo.idle_time = level.time + random()*15
	// 	}
	// }
}

/*
 * Self is currently not attacking anything,
 * so try to find a target
 *
 * Returns TRUE if an enemy was sighted
 *
 * When a player fires a missile, the point
 * of impact becomes a fakeplayer so that
 * monsters that see the impact will respond
 * as if they had seen the player.
 *
 * To avoid spending too much time, only
 * a single client (or fakeclient) is
 * checked each frame. This means multi
 * player games will have slightly
 * slower noticing monsters.
 */
func (G *qGame) findTarget(self *edict_t) bool {
	//  edict_t *client;
	//  qboolean heardit;
	//  int r;

	if self == nil {
		return false
	}

	if (self.monsterinfo.aiflags & AI_GOOD_GUY) != 0 {
		return false
	}

	/* if we're going to a combat point, just proceed */
	if (self.monsterinfo.aiflags & AI_COMBAT_POINT) != 0 {
		return false
	}

	/* if the first spawnflag bit is set, the monster
	will only wake up on really seeing the player,
	not another monster getting angry or hearing
	something */

	//  heardit = false;
	var client *edict_t

	if (G.level.sight_entity_framenum >= (G.level.framenum - 1)) &&
		(self.Spawnflags&1) == 0 {
		client = G.level.sight_entity

		if client.enemy == self.enemy {
			return false
		}
	} else if G.level.sound_entity_framenum >= (G.level.framenum - 1) {
		// 	 client = level.sound_entity;
		// 	 heardit = true;
		//  } else if (!(self->enemy) &&
		// 		  (level.sound2_entity_framenum >= (level.framenum - 1)) &&
		// 		  !(self->spawnflags & 1))
		//  {
		// 	 client = level.sound2_entity;
		// 	 heardit = true;
	} else {
		client = G.level.sight_client
		if client == nil {
			return false /* no clients to get mad at */
		}
	}

	/* if the entity went away, forget it */
	if !client.inuse {
		return false
	}

	if client == self.enemy {
		return true
	}

	if client.client != nil {
		if (client.flags & FL_NOTARGET) != 0 {
			return false
		}
	} else if (client.svflags & shared.SVF_MONSTER) != 0 {
		if client.enemy == nil {
			return false
		}

		if (client.enemy.flags & FL_NOTARGET) != 0 {
			return false
		}
		//  } else if (heardit) {
		// 	 if (client->owner->flags & FL_NOTARGET) {
		// 		 return false;
		// 	 }
	} else {
		return false
	}

	//  if (!heardit)
	//  {
	// 	 r = range(self, client);

	// 	 if (r == RANGE_FAR)
	// 	 {
	// 		 return false;
	// 	 }

	// 	 /* is client in an spot too dark to be seen? */
	// 	 if (client->light_level <= 5)
	// 	 {
	// 		 return false;
	// 	 }

	// 	 if (!visible(self, client))
	// 	 {
	// 		 return false;
	// 	 }

	// 	 if (r == RANGE_NEAR)
	// 	 {
	// 		 if ((client->show_hostile < level.time) && !infront(self, client))
	// 		 {
	// 			 return false;
	// 		 }
	// 	 }
	// 	 else if (r == RANGE_MID)
	// 	 {
	// 		 if (!infront(self, client))
	// 		 {
	// 			 return false;
	// 		 }
	// 	 }

	// 	 self->enemy = client;

	// 	 if (strcmp(self->enemy->classname, "player_noise") != 0)
	// 	 {
	// 		 self->monsterinfo.aiflags &= ~AI_SOUND_TARGET;

	// 		 if (!self->enemy->client)
	// 		 {
	// 			 self->enemy = self->enemy->enemy;

	// 			 if (!self->enemy->client)
	// 			 {
	// 				 self->enemy = NULL;
	// 				 return false;
	// 			 }
	// 		 }
	// 	 }
	//  }
	//  else /* heardit */
	//  {
	// 	 vec3_t temp;

	// 	 if (self->spawnflags & 1)
	// 	 {
	// 		 if (!visible(self, client))
	// 		 {
	// 			 return false;
	// 		 }
	// 	 }
	// 	 else
	// 	 {
	// 		 if (!gi.inPHS(self->s.origin, client->s.origin))
	// 		 {
	// 			 return false;
	// 		 }
	// 	 }

	// 	 VectorSubtract(client->s.origin, self->s.origin, temp);

	// 	 if (VectorLength(temp) > 1000) /* too far to hear */
	// 	 {
	// 		 return false;
	// 	 }

	// 	 /* check area portals - if they are different
	// 		and not connected then we can't hear it */
	// 	 if (client->areanum != self->areanum)
	// 	 {
	// 		 if (!gi.AreasConnected(self->areanum, client->areanum))
	// 		 {
	return false
	// 		 }
	// 	 }

	// 	 self->ideal_yaw = vectoyaw(temp);
	// 	 M_ChangeYaw(self);

	// 	 /* hunt the sound for a bit; hopefully find the real player */
	// 	 self->monsterinfo.aiflags |= AI_SOUND_TARGET;
	// 	 self->enemy = client;
	// }

	//  FoundTarget(self);

	//  if (!(self->monsterinfo.aiflags & AI_SOUND_TARGET) &&
	// 	 (self->monsterinfo.sight))
	//  {
	// 	 self->monsterinfo.sight(self, self->enemy);
	//  }

	// return true
}
