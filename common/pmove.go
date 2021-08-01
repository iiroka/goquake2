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
 * Player movement code. This is the core of Quake IIs legendary physics
 * engine
 *
 * =======================================================================
 */
package common

import "goquake2/shared"

/* all of the locals will be zeroed before each
 * pmove, just to make damn sure we don't have
 * any differences when running on client or server */

type pml_t struct {
	origin   [3]float32 /* full float precision */
	velocity [3]float32 /* full float precision */

	forward, right, up [3]float32
	frametime          float32

	//  csurface_t *groundsurface;
	groundplane    shared.Cplane_t
	groundcontents int

	previous_origin [3]float32
	ladder          bool
}

/*
 * On exit, the origin will have a value that is pre-quantized to the 0.125
 * precision of the network channel and in a valid position.
 */
func PM_SnapPosition(pm *shared.Pmove_t, pml *pml_t) {
	println("PM_SnapPosition")
	//  int sign[3];
	//  int i, j, bits;
	//  short base[3];
	//  /* try all single bits first */
	//  static int jitterbits[8] = {0, 4, 1, 2, 3, 5, 6, 7};

	//  /* snap velocity to eigths */
	//  for (i = 0; i < 3; i++)
	//  {
	// 	 pm->s.velocity[i] = (int)(pml.velocity[i] * 8);
	//  }

	//  for (i = 0; i < 3; i++)
	//  {
	// 	 if (pml.origin[i] >= 0)
	// 	 {
	// 		 sign[i] = 1;
	// 	 }
	// 	 else
	// 	 {
	// 		 sign[i] = -1;
	// 	 }

	// 	 pm->s.origin[i] = (int)(pml.origin[i] * 8);

	// 	 if (pm->s.origin[i] * 0.125f == pml.origin[i])
	// 	 {
	// 		 sign[i] = 0;
	// 	 }
	//  }

	//  VectorCopy(pm->s.origin, base);

	//  /* try all combinations */
	//  for (j = 0; j < 8; j++)
	//  {
	// 	 bits = jitterbits[j];
	// 	 VectorCopy(base, pm->s.origin);

	// 	 for (i = 0; i < 3; i++)
	// 	 {
	// 		 if (bits & (1 << i))
	// 		 {
	// 			 pm->s.origin[i] += sign[i];
	// 		 }
	// 	 }

	// 	 if (PM_GoodPosition())
	// 	 {
	// 		 return;
	// 	 }
	//  }

	//  /* go back to the last position */
	//  VectorCopy(pml.previous_origin, pm->s.origin);
}

func PM_CalculateViewHeightForDemo(pm *shared.Pmove_t) {
	if pm.S.Pm_type == shared.PM_GIB {
		pm.Viewheight = 8
	} else {
		if (pm.S.Pm_flags & shared.PMF_DUCKED) != 0 {
			pm.Viewheight = -2
		} else {
			pm.Viewheight = 22
		}
	}
}

/*
 * Can be called by either the server or the client
 */
func (T *qCommon) Pmove(pm *shared.Pmove_t) {
	/* clear results */
	pm.Numtouch = 0
	pm.Viewangles[0] = 0
	pm.Viewangles[1] = 0
	pm.Viewangles[2] = 0
	pm.Viewheight = 0
	// pm.Groundentity = 0
	pm.Watertype = 0
	pm.Waterlevel = 0

	/* clear all pmove local vars */
	pml := pml_t{}

	/* convert origin and velocity to float values */
	pml.origin[0] = float32(pm.S.Origin[0]) * 0.125
	pml.origin[1] = float32(pm.S.Origin[1]) * 0.125
	pml.origin[2] = float32(pm.S.Origin[2]) * 0.125

	pml.velocity[0] = float32(pm.S.Velocity[0]) * 0.125
	pml.velocity[1] = float32(pm.S.Velocity[1]) * 0.125
	pml.velocity[2] = float32(pm.S.Velocity[2]) * 0.125

	/* save old org in case we get stuck */
	for i := range pm.S.Origin {
		pml.previous_origin[i] = float32(pm.S.Origin[i])
	}

	pml.frametime = float32(pm.Cmd.Msec) * 0.001

	// 	PM_ClampAngles();

	if pm.S.Pm_type == shared.PM_SPECTATOR {
		// 		PM_FlyMove(false);
		PM_SnapPosition(pm, &pml)
		return
	}

	if pm.S.Pm_type >= shared.PM_DEAD {
		pm.Cmd.Forwardmove = 0
		pm.Cmd.Sidemove = 0
		pm.Cmd.Upmove = 0
	}

	if pm.S.Pm_type == shared.PM_FREEZE {
		if T.client.IsAttractloop() {
			PM_CalculateViewHeightForDemo(pm)
			// 			PM_CalculateWaterLevelForDemo();
			// 			PM_UpdateUnderwaterSfx();
		}
		return /* no movement at all */
	}

	// 	/* set mins, maxs, and viewheight */
	// 	PM_CheckDuck();

	// 	if (pm->snapinitial) {
	// 		PM_InitialSnapPosition();
	// 	}

	// 	/* set groundentity, watertype, and waterlevel */
	// 	PM_CatagorizePosition();

	// 	if (pm->s.pm_type == PM_DEAD) {
	// 		PM_DeadMove();
	// 	}

	// 	PM_CheckSpecialMovement();

	// 	/* drop timing counter */
	// 	if (pm->s.pm_time) {
	// 		int msec;

	// 		msec = pm->cmd.msec >> 3;

	// 		if (!msec) {
	// 			msec = 1;
	// 		}

	// 		if (msec >= pm->s.pm_time)
	// 		{
	// 			pm->s.pm_flags &= ~(PMF_TIME_WATERJUMP | PMF_TIME_LAND | PMF_TIME_TELEPORT);
	// 			pm->s.pm_time = 0;
	// 		}
	// 		else
	// 		{
	// 			pm->s.pm_time -= msec;
	// 		}
	// 	}

	// 	if (pm->s.pm_flags & PMF_TIME_TELEPORT)
	// 	{
	// 		/* teleport pause stays exactly in place */
	// 	}
	// 	else if (pm->s.pm_flags & PMF_TIME_WATERJUMP)
	// 	{
	// 		/* waterjump has no control, but falls */
	// 		pml.velocity[2] -= pm->s.gravity * pml.frametime;

	// 		if (pml.velocity[2] < 0)
	// 		{
	// 			/* cancel as soon as we are falling down again */
	// 			pm->s.pm_flags &= ~(PMF_TIME_WATERJUMP | PMF_TIME_LAND | PMF_TIME_TELEPORT);
	// 			pm->s.pm_time = 0;
	// 		}

	// 		PM_StepSlideMove();
	// 	}
	// 	else
	// 	{
	// 		PM_CheckJump();

	// 		PM_Friction();

	// 		if (pm->waterlevel >= 2)
	// 		{
	// 			PM_WaterMove();
	// 		}
	// 		else
	// 		{
	// 			vec3_t angles;

	// 			VectorCopy(pm->viewangles, angles);

	// 			if (angles[PITCH] > 180)
	// 			{
	// 				angles[PITCH] = angles[PITCH] - 360;
	// 			}

	// 			angles[PITCH] /= 3;

	// 			AngleVectors(angles, pml.forward, pml.right, pml.up);

	// 			PM_AirMove();
	// 		}
	// 	}

	// 	/* set groundentity, watertype, and waterlevel for final spot */
	// 	PM_CatagorizePosition();

	// #if !defined(DEDICATED_ONLY)
	//     PM_UpdateUnderwaterSfx();
	// #endif

	// 	PM_SnapPosition();
}
