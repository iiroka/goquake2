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

import "goquake2/shared"

func (G *qGame) svCheckVelocity(ent *edict_t) {
	if ent == nil {
		return
	}

	if shared.VectorLength(ent.velocity[:]) > G.sv_maxvelocity.Float() {
		shared.VectorNormalize(ent.velocity[:])
		shared.VectorScale(ent.velocity[:], G.sv_maxvelocity.Float(), ent.velocity[:])
	}
}

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

	if ent.think == nil {
		G.gi.Error("NULL ent->think %v", ent.Classname)
	}

	ent.think(ent, G)

	return false
}

/* ================================================================== */

/* PUSHMOVE */

/*
 * Does not change the entities velocity at all
 */

func (G *qGame) svPushEntity(ent *edict_t, push []float32) shared.Trace_t {

	start := make([]float32, 3)
	end := make([]float32, 3)
	copy(start, ent.s.Origin[:])
	shared.VectorAdd(start, push, end)

	retry := true
	var trace shared.Trace_t
	for retry {
		retry = false

		mask := shared.MASK_SOLID
		// if ent.clipmask {
		// 	mask = ent.clipmask
		// }

		trace = G.gi.Trace(start, ent.mins[:], ent.maxs[:], end, ent, mask)

		if trace.Startsolid || trace.Allsolid {
			mask = mask ^ shared.CONTENTS_DEADMONSTER
			trace = G.gi.Trace(start, ent.mins[:], ent.maxs[:], end, ent, mask)
		}

		copy(ent.s.Origin[:], trace.Endpos[:])
		G.gi.Linkentity(ent)

		/* Push slightly away from non-horizontal surfaces,
		prevent origin stuck in the plane which causes
		the entity to be rendered in full black. */
		if trace.Plane.Type != 2 {
			/* Limit the fix to gibs, debris and dead monsters.
			Everything else may break existing maps. Items
			may slide to unreachable locations, monsters may
			get stuck, etc. */
			// 		 if (((strncmp(ent->classname, "monster_", 8) == 0) && ent->health < 1) ||
			// 				 (strcmp(ent->classname, "debris") == 0) || (ent->s.effects & EF_GIB))
			// 		 {
			// 			 VectorAdd(ent->s.origin, trace.plane.normal, ent->s.origin);
			// 		 }
		}

		if trace.Fraction != 1.0 {
			// 		 SV_Impact(ent, &trace);

			// 		 /* if the pushed entity went away
			// 			and the pusher is still there */
			// 		 if (!trace.ent->inuse && ent->inuse) {
			// 			 /* move the pusher back and try again */
			// 			 VectorCopy(start, ent->s.origin);
			// 			 gi.linkentity(ent);
			// 			 retry = true
			// 		 }
		}
	}

	// 	 if (ent.inuse) {
	// 		 G_TouchTriggers(ent);
	// 	 }

	return trace
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

/* TOSS / BOUNCE */

/*
 * Toss, bounce, and fly movement.
 * When onground, do nothing.
 */
func (G *qGame) svPhysics_Toss(ent *edict_t) {
	//  trace_t trace;
	//  vec3_t move;
	//  float backoff;
	//  edict_t *slave;
	//  qboolean wasinwater;
	//  qboolean isinwater;
	//  vec3_t old_origin;

	if ent == nil {
		return
	}

	/* regular thinking */
	G.svRunThink(ent)

	/* if not a team captain, so movement
	will be handled elsewhere */
	if (ent.flags & FL_TEAMSLAVE) != 0 {
		return
	}

	if ent.velocity[2] > 0 {
		ent.groundentity = nil
	}

	/* check for the groundentity going away */
	if ent.groundentity != nil {
		if !ent.groundentity.inuse {
			ent.groundentity = nil
		}
	}

	/* if onground, return without moving */
	if ent.groundentity != nil {
		return
	}

	//  VectorCopy(ent->s.origin, old_origin);

	G.svCheckVelocity(ent)

	/* add gravity */
	//  if ((ent.movetype != MOVETYPE_FLY) &&
	// 	 (ent.movetype != MOVETYPE_FLYMISSILE)) {
	// 	 SV_AddGravity(ent);
	//  }

	/* move angles */
	shared.VectorMA(ent.s.Angles[:], FRAMETIME, ent.avelocity[:], ent.s.Angles[:])

	/* move origin */
	move := make([]float32, 3)
	shared.VectorScale(ent.velocity[:], FRAMETIME, move)
	trace := G.svPushEntity(ent, move)

	if !ent.inuse {
		return
	}

	if trace.Fraction < 1 {
		// 	 if (ent.movetype == MOVETYPE_BOUNCE) {
		// 		 backoff = 1.5;
		// 	 } else {
		// 		 backoff = 1;
		// 	 }

		// 	 ClipVelocity(ent->velocity, trace.plane.normal, ent->velocity, backoff);

		// 	 /* stop if on ground */
		// 	 if (trace.plane.normal[2] > 0.7)
		// 	 {
		// 		 if ((ent->velocity[2] < 60) || (ent->movetype != MOVETYPE_BOUNCE))
		// 		 {
		// 			 ent->groundentity = trace.ent;
		// 			 ent->groundentity_linkcount = trace.ent->linkcount;
		// 			 VectorCopy(vec3_origin, ent->velocity);
		// 			 VectorCopy(vec3_origin, ent->avelocity);
		// 		 }
		// 	 }
	}

	//  /* check for water transition */
	//  wasinwater = (ent->watertype & MASK_WATER);
	//  ent->watertype = gi.pointcontents(ent->s.origin);
	//  isinwater = ent->watertype & MASK_WATER;

	//  if (isinwater)
	//  {
	// 	 ent->waterlevel = 1;
	//  }
	//  else
	//  {
	ent.waterlevel = 0
	//  }

	//  if (!wasinwater && isinwater)
	//  {
	// 	 gi.positioned_sound(old_origin, g_edicts, CHAN_AUTO,
	// 			 gi.soundindex("misc/h2ohit1.wav"), 1, 1, 0);
	//  }
	//  else if (wasinwater && !isinwater)
	//  {
	// 	 gi.positioned_sound(ent->s.origin, g_edicts, CHAN_AUTO,
	// 			 gi.soundindex("misc/h2ohit1.wav"), 1, 1, 0);
	//  }

	//  /* move teamslaves */
	//  for (slave = ent->teamchain; slave; slave = slave->teamchain)
	//  {
	// 	 VectorCopy(ent->s.origin, slave->s.origin);
	// 	 gi.linkentity(slave);
	//  }
}

func (G *qGame) svPhysics_Step(ent *edict_t) {
	// 	qboolean wasonground;
	// 	qboolean hitsound = false;
	// 	float *vel;
	// 	float speed, newspeed, control;
	// 	float friction;
	// 	edict_t *groundentity;
	// 	int mask;
	// 	vec3_t oldorig;
	// 	trace_t tr;

	if ent == nil {
		return
	}

	/* airborn monsters should always check for ground */
	// 	if (ent.groundentity == nil) {
	// 		M_CheckGround(ent);
	// 	}

	groundentity := ent.groundentity

	G.svCheckVelocity(ent)

	wasonground := groundentity != nil

	// 	if (ent.avelocity[0] != 0 || ent.avelocity[1] != 0 || ent.avelocity[2] != 0) {
	// 		SV_AddRotationalFriction(ent);
	// 	}

	/* add gravity except:
	   flying monsters
	   swimming monsters who are in the water */
	if !wasonground {
		if (ent.flags & FL_FLY) == 0 {
			// 			if (!((ent.flags & FL_SWIM) && (ent.waterlevel > 2))) {
			// 				if (ent.velocity[2] < sv_gravity.value * -0.1) {
			// 					hitsound = true;
			// 				}

			// 				if (ent.waterlevel == 0) {
			// 					SV_AddGravity(ent);
			// 				}
			// 			}
		}
	}

	/* friction for flying monsters that have been given vertical velocity */
	// 	if ((ent->flags & FL_FLY) && (ent->velocity[2] != 0))
	// 	{
	// 		speed = fabs(ent->velocity[2]);
	// 		control = speed < STOPSPEED ? STOPSPEED : speed;
	// 		friction = FRICTION / 3;
	// 		newspeed = speed - (FRAMETIME * control * friction);

	// 		if (newspeed < 0) {
	// 			newspeed = 0;
	// 		}

	// 		newspeed /= speed;
	// 		ent->velocity[2] *= newspeed;
	// 	}

	/* friction for flying monsters that have been given vertical velocity */
	// 	if ((ent->flags & FL_SWIM) && (ent->velocity[2] != 0))
	// 	{
	// 		speed = fabs(ent->velocity[2]);
	// 		control = speed < STOPSPEED ? STOPSPEED : speed;
	// 		newspeed = speed - (FRAMETIME * control * WATERFRICTION * ent->waterlevel);

	// 		if (newspeed < 0)
	// 		{
	// 			newspeed = 0;
	// 		}

	// 		newspeed /= speed;
	// 		ent->velocity[2] *= newspeed;
	// 	}

	if ent.velocity[2] != 0 || ent.velocity[1] != 0 || ent.velocity[0] != 0 {
		/* apply friction: let dead monsters who
		   aren't completely onground slide */
		if (wasonground) || (ent.flags&(FL_SWIM|FL_FLY)) != 0 {
			// 			if (!((ent->health <= 0.0) && !M_CheckBottom(ent)))
			// 			{
			// 				vel = ent->velocity;
			// 				speed = sqrt(vel[0] * vel[0] + vel[1] * vel[1]);

			// 				if (speed != 0) {
			// 					friction = FRICTION;

			// 					control = speed < STOPSPEED ? STOPSPEED : speed;
			// 					newspeed = speed - FRAMETIME * control * friction;

			// 					if (newspeed < 0) {
			// 						newspeed = 0;
			// 					}

			// 					newspeed /= speed;

			// 					vel[0] *= newspeed;
			// 					vel[1] *= newspeed;
			// 				}
			// 			}
		}

		// 		if (ent.svflags & SVF_MONSTER) != 0 {
		// 			mask = MASK_MONSTERSOLID;
		// 		} else {
		// 			mask = MASK_SOLID;
		// 		}

		// 		VectorCopy(ent->s.origin, oldorig);
		// 		SV_FlyMove(ent, FRAMETIME, mask);

		// 		/* Evil hack to work around dead parasites (and maybe other monster)
		// 		   falling through the worldmodel into the void. We copy the current
		// 		   origin (see above) and after the SV_FlyMove() was performend we
		// 		   checl if we're stuck in the world model. If yes we're undoing the
		// 		   move. */
		// 		if (!VectorCompare(ent->s.origin, oldorig)) {
		// 			tr = gi.trace(ent->s.origin, ent->mins, ent->maxs, ent->s.origin, ent, mask);

		// 			if (tr.startsolid) {
		// 				VectorCopy(oldorig, ent->s.origin);
		// 			}
		// 		}

		G.gi.Linkentity(ent)
		// 		G_TouchTriggers(ent);

		if !ent.inuse {
			return
		}

		// 		if (ent.groundentity != nil) {
		// 			if (!wasonground) {
		// 				if (hitsound) {
		// 					gi.sound(ent, 0, gi.soundindex("world/land.wav"), 1, 1, 0);
		// 				}
		// 			}
		// 		}
	}

	/* regular thinking */
	G.svRunThink(ent)
}

/* ================================================================== */

func (G *qGame) runEntity(ent *edict_t) error {
	if ent == nil {
		return nil
	}

	// if (ent.prethink != nil) {
	// 	ent.prethink(ent, G);
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
	case MOVETYPE_STEP:
		G.svPhysics_Step(ent)
	case MOVETYPE_TOSS,
		MOVETYPE_BOUNCE,
		MOVETYPE_FLY,
		MOVETYPE_FLYMISSILE:
		G.svPhysics_Toss(ent)
	default:
		return G.gi.Error("SV_Physics: bad movetype %v", ent.movetype)
	}
	return nil
}
