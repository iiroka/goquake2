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

	// 	/* airborn monsters should always check for ground */
	// 	if (!ent->groundentity)
	// 	{
	// 		M_CheckGround(ent);
	// 	}

	// 	groundentity = ent->groundentity;

	// 	SV_CheckVelocity(ent);

	// 	if (groundentity) {
	// 		wasonground = true;
	// 	} else {
	// 		wasonground = false;
	// 	}

	// 	if (ent->avelocity[0] || ent->avelocity[1] || ent->avelocity[2]) {
	// 		SV_AddRotationalFriction(ent);
	// 	}

	/* add gravity except:
	   flying monsters
	   swimming monsters who are in the water */
	// 	if (!wasonground)
	// 	{
	// 		if (!(ent->flags & FL_FLY))
	// 		{
	// 			if (!((ent->flags & FL_SWIM) && (ent->waterlevel > 2)))
	// 			{
	// 				if (ent->velocity[2] < sv_gravity->value * -0.1)
	// 				{
	// 					hitsound = true;
	// 				}

	// 				if (ent->waterlevel == 0)
	// 				{
	// 					SV_AddGravity(ent);
	// 				}
	// 			}
	// 		}
	// 	}

	/* friction for flying monsters that have been given vertical velocity */
	// 	if ((ent->flags & FL_FLY) && (ent->velocity[2] != 0))
	// 	{
	// 		speed = fabs(ent->velocity[2]);
	// 		control = speed < STOPSPEED ? STOPSPEED : speed;
	// 		friction = FRICTION / 3;
	// 		newspeed = speed - (FRAMETIME * control * friction);

	// 		if (newspeed < 0)
	// 		{
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

	// 	if (ent->velocity[2] || ent->velocity[1] || ent->velocity[0])
	// 	{
	// 		/* apply friction: let dead monsters who
	// 		   aren't completely onground slide */
	// 		if ((wasonground) || (ent->flags & (FL_SWIM | FL_FLY)))
	// 		{
	// 			if (!((ent->health <= 0.0) && !M_CheckBottom(ent)))
	// 			{
	// 				vel = ent->velocity;
	// 				speed = sqrt(vel[0] * vel[0] + vel[1] * vel[1]);

	// 				if (speed)
	// 				{
	// 					friction = FRICTION;

	// 					control = speed < STOPSPEED ? STOPSPEED : speed;
	// 					newspeed = speed - FRAMETIME * control * friction;

	// 					if (newspeed < 0)
	// 					{
	// 						newspeed = 0;
	// 					}

	// 					newspeed /= speed;

	// 					vel[0] *= newspeed;
	// 					vel[1] *= newspeed;
	// 				}
	// 			}
	// 		}

	// 		if (ent->svflags & SVF_MONSTER)
	// 		{
	// 			mask = MASK_MONSTERSOLID;
	// 		}
	// 		else
	// 		{
	// 			mask = MASK_SOLID;
	// 		}

	// 		VectorCopy(ent->s.origin, oldorig);
	// 		SV_FlyMove(ent, FRAMETIME, mask);

	// 		/* Evil hack to work around dead parasites (and maybe other monster)
	// 		   falling through the worldmodel into the void. We copy the current
	// 		   origin (see above) and after the SV_FlyMove() was performend we
	// 		   checl if we're stuck in the world model. If yes we're undoing the
	// 		   move. */
	// 		if (!VectorCompare(ent->s.origin, oldorig))
	// 		{
	// 			tr = gi.trace(ent->s.origin, ent->mins, ent->maxs, ent->s.origin, ent, mask);

	// 			if (tr.startsolid)
	// 			{
	// 				VectorCopy(oldorig, ent->s.origin);
	// 			}
	// 		}

	// 		gi.linkentity(ent);
	// 		G_TouchTriggers(ent);

	// 		if (!ent->inuse)
	// 		{
	// 			return;
	// 		}

	// 		if (ent->groundentity)
	// 		{
	// 			if (!wasonground)
	// 			{
	// 				if (hitsound)
	// 				{
	// 					gi.sound(ent, 0, gi.soundindex("world/land.wav"), 1, 1, 0);
	// 				}
	// 			}
	// 		}
	// 	}

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
	case MOVETYPE_STEP:
		G.svPhysics_Step(ent)
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
