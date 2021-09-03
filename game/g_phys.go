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

import (
	"goquake2/shared"
)

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

type pushed_t struct {
	ent      *edict_t
	origin   [3]float32
	angles   [3]float32
	deltayaw float32
}

/*
 * Objects need to be moved back on a failed push,
 * otherwise riders would continue to slide.
 */
func (G *qGame) svPush(pusher *edict_t, move, amove []float32) bool {
	//  int i, e;
	//  edict_t *check, *block;
	//  pushed_t *p;
	//  vec3_t org, org2, move2, forward, right, up;
	//  vec3_t realmins, realmaxs;

	if pusher == nil {
		return false
	}

	/* clamp the move to 1/8 units, so the position will
	be accurate for client side prediction */
	for i := 0; i < 3; i++ {
		temp := move[i] * 8.0

		if temp > 0.0 {
			temp += 0.5
		} else {
			temp -= 0.5
		}

		move[i] = 0.125 * float32(int(temp))
	}

	/* we need this for pushing things later */
	org := make([]float32, 3)
	shared.VectorSubtract([]float32{0, 0, 0}, amove, org)
	forward := make([]float32, 3)
	right := make([]float32, 3)
	up := make([]float32, 3)
	shared.AngleVectors(org, forward, right, up)

	/* save the pusher's original position */
	G.pushed[G.pushed_i].ent = pusher
	copy(G.pushed[G.pushed_i].origin[:], pusher.s.Origin[:])
	copy(G.pushed[G.pushed_i].angles[:], pusher.s.Angles[:])

	if pusher.client != nil {
		G.pushed[G.pushed_i].deltayaw = float32(pusher.client.ps.Pmove.Delta_angles[shared.YAW])
	}

	G.pushed_i++

	/* move the pusher to it's final position */
	shared.VectorAdd(pusher.s.Origin[:], move, pusher.s.Origin[:])
	shared.VectorAdd(pusher.s.Angles[:], amove, pusher.s.Angles[:])
	G.gi.Linkentity(pusher)

	/* Create a real bounding box for
	rotating brush models. */
	//  RealBoundingBox(pusher,realmins,realmaxs);

	/* see if any solid entities
	are inside the final position */
	//  check = g_edicts + 1;

	for e := 1; e < G.num_edicts; e++ {
		check := &G.g_edicts[e]
		if !check.inuse {
			continue
		}

		if (check.movetype == MOVETYPE_PUSH) ||
			(check.movetype == MOVETYPE_STOP) ||
			(check.movetype == MOVETYPE_NONE) ||
			(check.movetype == MOVETYPE_NOCLIP) {
			continue
		}

		if check.area.Prev == nil {
			continue /* not linked in anywhere */
		}

		/* if the entity is standing on the pusher,
		it will definitely be moved */
		// 	 if (check->groundentity != pusher)
		// 	 {
		// 		 /* see if the ent needs to be tested */
		// 		 if ((check->absmin[0] >= realmaxs[0]) ||
		// 			 (check->absmin[1] >= realmaxs[1]) ||
		// 			 (check->absmin[2] >= realmaxs[2]) ||
		// 			 (check->absmax[0] <= realmins[0]) ||
		// 			 (check->absmax[1] <= realmins[1]) ||
		// 			 (check->absmax[2] <= realmins[2]))
		// 		 {
		// 			 continue;
		// 		 }

		// 		 /* see if the ent's bbox is inside
		// 			the pusher's final position */
		// 		 if (!SV_TestEntityPosition(check))
		// 		 {
		// 			 continue;
		// 		 }
		// 	 }

		// 	 if ((pusher->movetype == MOVETYPE_PUSH) ||
		// 		 (check->groundentity == pusher))
		// 	 {
		// 		 /* move this entity */
		// 		 pushed_p->ent = check;
		// 		 VectorCopy(check->s.origin, pushed_p->origin);
		// 		 VectorCopy(check->s.angles, pushed_p->angles);
		// 		 pushed_p++;

		// 		 /* try moving the contacted entity */
		// 		 VectorAdd(check->s.origin, move, check->s.origin);

		// 		 if (check->client)
		// 		 {
		// 			 check->client->ps.pmove.delta_angles[YAW] += amove[YAW];
		// 		 }

		// 		 /* figure movement due to the pusher's amove */
		// 		 VectorSubtract(check->s.origin, pusher->s.origin, org);
		// 		 org2[0] = DotProduct(org, forward);
		// 		 org2[1] = -DotProduct(org, right);
		// 		 org2[2] = DotProduct(org, up);
		// 		 VectorSubtract(org2, org, move2);
		// 		 VectorAdd(check->s.origin, move2, check->s.origin);

		// 		 /* may have pushed them off an edge */
		// 		 if (check->groundentity != pusher)
		// 		 {
		// 			 check->groundentity = NULL;
		// 		 }

		// 		 block = SV_TestEntityPosition(check);

		// 		 if (!block)

		// 		 {   /* pushed ok */
		// 			 gi.linkentity(check);
		// 			 continue;
		// 		 }

		// 		 /* if it is ok to leave in the old position, do it
		// 			this is only relevent for riding entities, not
		// 			pushed */
		// 		 VectorSubtract(check->s.origin, move, check->s.origin);
		// 		 block = SV_TestEntityPosition(check);

		// 		 if (!block)
		// 		 {
		// 			 pushed_p--;
		// 			 continue;
		// 		 }
		// 	 }

		/* save off the obstacle so we can
		call the block function */
		G.obstacle = check

		// 	 /* move back any entities we already moved
		// 		go backwards, so if the same entity was pushed
		// 		twice, it goes back to the original position */
		// 	 for (p = pushed_p - 1; p >= pushed; p--)
		// 	 {
		// 		 VectorCopy(p->origin, p->ent->s.origin);
		// 		 VectorCopy(p->angles, p->ent->s.angles);

		// 		 if (p->ent->client)
		// 		 {
		// 			 p->ent->client->ps.pmove.delta_angles[YAW] = p->deltayaw;
		// 		 }

		// 		 gi.linkentity(p->ent);
		// 	 }

		return false
	}

	//  /* see if anything we moved has touched a trigger */
	//  for (p = pushed_p - 1; p >= pushed; p--)
	//  {
	// 	 G_TouchTriggers(p->ent);
	//  }

	return true
}

/*
 * Bmodel objects don't interact with each
 * other, but push all box objects
 */
func (G *qGame) svPhysics_Pusher(ent *edict_t) {

	if ent == nil {
		return
	}

	/* if not a team captain, so movement
	will be handled elsewhere */
	if (ent.flags & FL_TEAMSLAVE) != 0 {
		return
	}

	/* make sure all team slaves can move before commiting
	any moves or calling any think functions if the move
	is blocked, all moved objects will be backed out */
	G.pushed_i = 0

	var part *edict_t
	for part = ent; part != nil; part = part.teamchain {
		if part.velocity[0] != 0 || part.velocity[1] != 0 || part.velocity[2] != 0 ||
			part.avelocity[0] != 0 || part.avelocity[1] != 0 || part.avelocity[2] != 0 {
			/* object is moving */
			move := make([]float32, 3)
			amove := make([]float32, 3)
			shared.VectorScale(part.velocity[:], FRAMETIME, move)
			shared.VectorScale(part.avelocity[:], FRAMETIME, amove)

			if !G.svPush(part, move, amove) {
				break /* move was blocked */
			}
		}
	}

	//  if (pushed_p > &pushed[MAX_EDICTS -1 ])
	//  {
	// 	 gi.error("pushed_p > &pushed[MAX_EDICTS - 1], memory corrupted");
	//  }

	if part != nil {
		/* the move failed, bump all nextthink
		times and back out moves */
		for mv := ent; mv != nil; mv = mv.teamchain {
			if mv.nextthink > 0 {
				mv.nextthink += FRAMETIME
			}
		}

		// 	 /* if the pusher has a "blocked" function, call it
		// 		otherwise, just stay in place until the obstacle
		// 		is gone */
		// 	 if (part->blocked)
		// 	 {
		// 		 part->blocked(part, obstacle);
		// 	 }
	} else {
		/* the move succeeded, so call all think functions */
		for part := ent; part != nil; part = part.teamchain {
			G.svRunThink(part)
		}
	}
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
		G.gTouchTriggers(ent)

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
	case MOVETYPE_PUSH,
		MOVETYPE_STOP:
		G.svPhysics_Pusher(ent)
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
