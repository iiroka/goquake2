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

func think_CalcMoveSpeed(ent *edict_t, G *qGame) {
	// edict_t *ent;
	// float min;
	// float time;
	// float newspeed;
	// float ratio;
	// float dist;

	// if (!self)
	// {
	// 	return;
	// }

	// if (self->flags & FL_TEAMSLAVE)
	// {
	// 	return; /* only the team master does this */
	// }

	// /* find the smallest distance any member of the team will be moving */
	// min = fabs(self->moveinfo.distance);

	// for (ent = self->teamchain; ent; ent = ent->teamchain)
	// {
	// 	dist = fabs(ent->moveinfo.distance);

	// 	if (dist < min)
	// 	{
	// 		min = dist;
	// 	}
	// }

	// time = min / self->moveinfo.speed;

	// /* adjust speeds so they will all complete at the same time */
	// for (ent = self; ent; ent = ent->teamchain)
	// {
	// 	newspeed = fabs(ent->moveinfo.distance) / time;
	// 	ratio = newspeed / ent->moveinfo.speed;

	// 	if (ent->moveinfo.accel == ent->moveinfo.speed)
	// 	{
	// 		ent->moveinfo.accel = newspeed;
	// 	}
	// 	else
	// 	{
	// 		ent->moveinfo.accel *= ratio;
	// 	}

	// 	if (ent->moveinfo.decel == ent->moveinfo.speed)
	// 	{
	// 		ent->moveinfo.decel = newspeed;
	// 	}
	// 	else
	// 	{
	// 		ent->moveinfo.decel *= ratio;
	// 	}

	// 	ent->moveinfo.speed = newspeed;
	// }
}

func think_SpawnDoorTrigger(ent *edict_t, G *qGame) {
	// edict_t *other;
	// vec3_t mins, maxs;

	if ent == nil || G == nil {
		return
	}

	if (ent.flags & FL_TEAMSLAVE) != 0 {
		return /* only the team leader spawns a trigger */
	}

	// VectorCopy(ent->absmin, mins);
	// VectorCopy(ent->absmax, maxs);

	// for (other = ent->teamchain; other; other = other->teamchain)
	// {
	// 	AddPointToBounds(other->absmin, mins, maxs);
	// 	AddPointToBounds(other->absmax, mins, maxs);
	// }

	// /* expand */
	// mins[0] -= 60;
	// mins[1] -= 60;
	// maxs[0] += 60;
	// maxs[1] += 60;

	// other, _ := G.gSpawn()
	// VectorCopy(mins, other->mins);
	// VectorCopy(maxs, other->maxs);
	// other.owner = ent
	// other.solid = shared.SOLID_TRIGGER
	// other.movetype = MOVETYPE_NONE
	// other->touch = Touch_DoorTrigger;
	// G.gi.Linkentity(other)

	// if (ent.Spawnflags & DOOR_START_OPEN) != 0 {
	// 	door_use_areaportals(ent, true);
	// }

	think_CalcMoveSpeed(ent, G)
}

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

func spFuncDoor(ent *edict_t, G *qGame) error {

	if ent == nil || G == nil {
		return nil
	}

	// if (ent.sounds != 1)
	// {
	// 	ent->moveinfo.sound_start = gi.soundindex("doors/dr1_strt.wav");
	// 	ent->moveinfo.sound_middle = gi.soundindex("doors/dr1_mid.wav");
	// 	ent->moveinfo.sound_end = gi.soundindex("doors/dr1_end.wav");
	// }

	gSetMovedir(ent.s.Angles[:], ent.movedir[:])
	ent.movetype = MOVETYPE_PUSH
	ent.solid = shared.SOLID_BSP
	G.gi.Setmodel(ent, ent.Model)

	// ent->blocked = door_blocked;
	// ent->use = door_use;

	if ent.Speed == 0 {
		ent.Speed = 100
	}

	// if (deathmatch->value)
	// {
	// 	ent->speed *= 2;
	// }

	if ent.Accel == 0 {
		ent.Accel = ent.Speed
	}

	if ent.Decel == 0 {
		ent.Decel = ent.Speed
	}

	if ent.Wait == 0 {
		ent.Wait = 3
	}

	if G.st.Lip == 0 {
		G.st.Lip = 8
	}

	if ent.Dmg == 0 {
		ent.Dmg = 2
	}

	/* calculate second position */
	copy(ent.pos1[:], ent.s.Origin[:])
	// abs_movedir[0] = fabs(ent->movedir[0]);
	// abs_movedir[1] = fabs(ent->movedir[1]);
	// abs_movedir[2] = fabs(ent->movedir[2]);
	// ent->moveinfo.distance = abs_movedir[0] * ent->size[0] + abs_movedir[1] *
	// 						 ent->size[1] + abs_movedir[2] * ent->size[2] -
	// 						 st.lip;
	// shared.VectorMA(ent.pos1[:], ent.moveinfo.distance, ent.movedir[:], ent.pos2[:])

	// /* if it starts open, switch the positions */
	// if (ent->spawnflags & DOOR_START_OPEN) != 0 {
	// 	VectorCopy(ent->pos2, ent->s.origin);
	// 	VectorCopy(ent->pos1, ent->pos2);
	// 	VectorCopy(ent->s.origin, ent->pos1);
	// }

	// ent.moveinfo.state = STATE_BOTTOM;

	if ent.Health != 0 {
		// 	ent->takedamage = DAMAGE_YES;
		// 	ent->die = door_killed;
		ent.max_health = ent.Health
		// } else if (ent->targetname && ent->message) {
		// 	gi.soundindex("misc/talk.wav");
		// 	ent->touch = door_touch;
	}

	// ent->moveinfo.speed = ent->speed;
	// ent->moveinfo.accel = ent->accel;
	// ent->moveinfo.decel = ent->decel;
	// ent->moveinfo.wait = ent->wait;
	// VectorCopy(ent->pos1, ent->moveinfo.start_origin);
	// VectorCopy(ent->s.angles, ent->moveinfo.start_angles);
	// VectorCopy(ent->pos2, ent->moveinfo.end_origin);
	// VectorCopy(ent->s.angles, ent->moveinfo.end_angles);

	if (ent.Spawnflags & 16) != 0 {
		ent.s.Effects |= shared.EF_ANIM_ALL
	}

	if (ent.Spawnflags & 64) != 0 {
		ent.s.Effects |= shared.EF_ANIM_ALLFAST
	}

	// /* to simplify logic elsewhere, make non-teamed doors into a team of one */
	// if (!ent->team)
	// {
	// 	ent->teammaster = ent;
	// }

	G.gi.Linkentity(ent)

	ent.nextthink = G.level.time + FRAMETIME

	if ent.Health != 0 || len(ent.Targetname) > 0 {
		ent.think = think_CalcMoveSpeed
	} else {
		ent.think = think_SpawnDoorTrigger
	}

	// /* Map quirk for waste3 (to make that secret armor behind
	//  * the secret wall - this func_door - count, #182) */
	// if (Q_stricmp(level.mapname, "waste3") == 0 && Q_stricmp(ent->model, "*12") == 0)
	// {
	// 	ent->target = "t117";
	// }
	return nil
}

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
