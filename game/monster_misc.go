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
 * Monster movement support functions.
 *
 * =======================================================================
 */
package game

import (
	"goquake2/shared"
	"math"
)

const STEPSIZE = 18
const DI_NODIR = -1

/*
 * Called by monster program code.
 * The move will be adjusted for slopes
 * and stairs, but if the move isn't
 * possible, no move is done, false is
 * returned, and pr_global_struct->trace_normal
 * is set to the normal of the blocking wall
 */
func (G *qGame) svMovestep(ent *edict_t, move []float32, relink bool) bool {
	//  float dz;
	//  vec3_t oldorg, neworg, end;
	//  trace_t trace;
	//  int i;
	//  float stepsize;
	//  vec3_t test;
	//  int contents;

	if ent == nil {
		return false
	}

	/* try the move */
	oldorg := make([]float32, 3)
	copy(oldorg, ent.s.Origin[:])
	neworg := make([]float32, 3)
	shared.VectorAdd(ent.s.Origin[:], move, neworg)

	/* flying monsters don't step up */
	if (ent.flags & (FL_SWIM | FL_FLY)) != 0 {
		/* try one move with vertical motion, then one without */
		//  for (i = 0; i < 2; i++) {
		// 	 VectorAdd(ent->s.origin, move, neworg);

		// 	 if ((i == 0) && ent- {
		// 		 if (!ent->goalentity) {
		// 			 ent->goalentity = ent->enemy;
		// 		 }

		// 		 dz = ent->s.origin[2] - ent->goalentity->s.origin[2];

		// 		 if (ent->goalentity->client) {
		// 			 if (dz > 40) {
		// 				 neworg[2] -= 8;
		// 			 }

		// 			 if (!((ent->flags & FL_SWIM) && (ent->waterlevel < 2))) {
		// 				 if (dz < 30) {
		// 					 neworg[2] += 8;
		// 				 }
		// 			 }
		// 		 } else {
		// 			 if (dz > 8) {
		// 				 neworg[2] -= 8;
		// 			 } else if (dz > 0) {
		// 				 neworg[2] -= dz;
		// 			 } else if (dz < -8) {
		// 				 neworg[2] += 8;
		// 			 } else {
		// 				 neworg[2] += dz;
		// 			 }
		// 		 }
		// 	 }

		// 	 trace = gi.trace(ent->s.origin, ent->mins, ent->maxs,
		// 			 neworg, ent, MASK_MONSTERSOLID);

		// 	 /* fly monsters don't enter water voluntarily */
		// 	 if (ent->flags & FL_FLY) != 0 {
		// 		 if (!ent->waterlevel) {
		// 			 test[0] = trace.endpos[0];
		// 			 test[1] = trace.endpos[1];
		// 			 test[2] = trace.endpos[2] + ent->mins[2] + 1;
		// 			 contents = gi.pointcontents(test);

		// 			 if (contents & MASK_WATER) != 0 {
		// 				 return false;
		// 			 }
		// 		 }
		// 	 }

		// 	 /* swim monsters don't exit water voluntarily */
		// 	 if (ent->flags & FL_SWIM) != 0 {
		// 		 if (ent->waterlevel < 2) {
		// 			 test[0] = trace.endpos[0];
		// 			 test[1] = trace.endpos[1];
		// 			 test[2] = trace.endpos[2] + ent->mins[2] + 1;
		// 			 contents = gi.pointcontents(test);

		// 			 if (!(contents & MASK_WATER)) {
		// 				 return false;
		// 			 }
		// 		 }
		// 	 }

		// 	 if (trace.fraction == 1) {
		// 		 VectorCopy(trace.endpos, ent->s.origin);

		// 		 if (relink) {
		// 			 gi.linkentity(ent);
		// 			 G_TouchTriggers(ent);
		// 		 }

		// 		 return true;
		// 	 }

		// 	 if (!ent->enemy) {
		// 		 break;
		// 	 }
		//  }

		return false
	}

	/* push down from a step height above the wished position */
	var stepsize float32
	if (ent.monsterinfo.aiflags & AI_NOSTEP) == 0 {
		stepsize = STEPSIZE
	} else {
		stepsize = 1
	}

	neworg[2] += stepsize
	end := make([]float32, 3)
	copy(end, neworg)
	end[2] -= stepsize * 2

	trace := G.gi.Trace(neworg, ent.mins[:], ent.maxs[:], end, ent, shared.MASK_MONSTERSOLID)

	if trace.Allsolid {
		return false
	}

	if trace.Startsolid {
		neworg[2] -= stepsize
		trace = G.gi.Trace(neworg, ent.mins[:], ent.maxs[:], end, ent, shared.MASK_MONSTERSOLID)

		if trace.Allsolid || trace.Startsolid {
			return false
		}
	}

	/* don't go in to water */
	if ent.waterlevel == 0 {
		test := []float32{trace.Endpos[0], trace.Endpos[1], trace.Endpos[2] + ent.mins[2] + 1}
		contents := G.gi.Pointcontents(test)

		if (contents & shared.MASK_WATER) != 0 {
			return false
		}
	}

	if trace.Fraction == 1 {
		/* if monster had the ground pulled out, go ahead and fall */
		if (ent.flags & FL_PARTIALGROUND) != 0 {
			shared.VectorAdd(ent.s.Origin[:], move, ent.s.Origin[:])

			if relink {
				G.gi.Linkentity(ent)
				// G_TouchTriggers(ent)
			}

			ent.groundentity = nil
			return true
		}

		return false /* walked off an edge */
	}

	/* check point traces down for dangling corners */
	copy(ent.s.Origin[:], trace.Endpos[:])

	// if !M_CheckBottom(ent) {
	// 	if (ent.flags & FL_PARTIALGROUND) != 0 {
	// 		/* entity had floor mostly pulled out
	// 		from underneath it and is trying to
	// 		correct */
	// 		if relink {
	// 			gi.linkentity(ent)
	// 			G_TouchTriggers(ent)
	// 		}

	// 		return true
	// 	}

	// 	VectorCopy(oldorg, ent.s.origin)
	// 	return false
	// }

	if (ent.flags & FL_PARTIALGROUND) != 0 {
		ent.flags &^= FL_PARTIALGROUND
	}

	ent.groundentity = trace.Ent.(*edict_t)
	ent.groundentity_linkcount = trace.Ent.(*edict_t).linkcount

	/* the move is ok */
	if relink {
		G.gi.Linkentity(ent)
		// G_TouchTriggers(ent)
	}

	return true
}

func (G *qGame) mWalkmove(ent *edict_t, yaw, dist float32) bool {

	if ent == nil {
		return false
	}

	if ent.groundentity == nil && (ent.flags&(FL_FLY|FL_SWIM) == 0) {
		return false
	}

	dyaw := float64(yaw) * math.Pi * 2 / 360

	move := []float32{
		float32(math.Cos(dyaw)) * dist,
		float32(math.Sin(dyaw)) * dist,
		0}

	return G.svMovestep(ent, move, true)
}
