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
 * This file implements all static entities at client site.
 *
 * =======================================================================
 */
package client

import (
	"goquake2/shared"
	"math"
)

/*
 * Sets cl.refdef view values
 */
func (T *qClient) calcViewValues() {
	//  int i;
	//  float lerp, backlerp, ifov;
	//  frame_t *oldframe;
	//  player_state_t *ps, *ops;

	/* find the previous frame to interpolate from */
	ps := &T.cl.frame.playerstate
	i := (T.cl.frame.serverframe - 1) & shared.UPDATE_MASK
	oldframe := &T.cl.frames[i]

	if (oldframe.serverframe != T.cl.frame.serverframe-1) || !oldframe.valid {
		oldframe = &T.cl.frame /* previous frame was dropped or invalid */
	}

	ops := &oldframe.playerstate

	/* see if the player entity was teleported this frame */
	if (math.Abs(float64(ops.Pmove.Origin[0]-ps.Pmove.Origin[0])) > 256*8) ||
		(math.Abs(float64(ops.Pmove.Origin[1]-ps.Pmove.Origin[1])) > 256*8) ||
		(math.Abs(float64(ops.Pmove.Origin[2]-ps.Pmove.Origin[2])) > 256*8) {
		ops = ps /* don't interpolate */
	}

	//  if(cl_paused->value){
	// 	 lerp = 1.0f;
	//  }
	//  else
	//  {
	lerp := T.cl.lerpfrac
	//  }

	/* calculate the origin */
	if (T.cl_predict.Bool()) && (T.cl.frame.playerstate.Pmove.Pm_flags&shared.PMF_NO_PREDICTION) == 0 {
		// 	 /* use predicted values */
		// 	 unsigned delta;

		backlerp := 1.0 - lerp

		for i := 0; i < 3; i++ {
			T.cl.refdef.Vieworg[i] = T.cl.predicted_origin[i] + ops.Viewoffset[i] +
				T.cl.lerpfrac*(ps.Viewoffset[i]-ops.Viewoffset[i]) -
				backlerp*T.cl.prediction_error[i]
		}

		/* smooth out stair climbing */
		// delta := T.cls.realtime - T.cl.predicted_step_time

		// if delta < 100 {
		// 	T.cl.refdef.vieworg[2] -= T.cl.predicted_step * (100 - delta) * 0.01
		// }
	} else {
		/* just use interpolated values */
		for i := 0; i < 3; i++ {
			T.cl.refdef.Vieworg[i] = float32(ops.Pmove.Origin[i])*0.125 +
				ops.Viewoffset[i] + lerp*(float32(ps.Pmove.Origin[i])*0.125+
				ps.Viewoffset[i]-(float32(ops.Pmove.Origin[i])*0.125+
				ops.Viewoffset[i]))
		}
	}

	/* if not running a demo or on a locked frame, add the local angle movement */
	if T.cl.frame.playerstate.Pmove.Pm_type < shared.PM_DEAD {
		/* use predicted values */
		for i := 0; i < 3; i++ {
			T.cl.refdef.Viewangles[i] = T.cl.predicted_angles[i]
		}
	} else {
		/* just use interpolated values */
		for i := 0; i < 3; i++ {
			T.cl.refdef.Viewangles[i] = shared.LerpAngle(ops.Viewangles[i], ps.Viewangles[i], lerp)
		}
	}

	//  if (cl_kickangles->value)
	//  {
	// 	 for (i = 0; i < 3; i++) {
	// 		 cl.refdef.viewangles[i] += LerpAngle(ops->kick_angles[i],
	// 				 ps->kick_angles[i], lerp);
	// 	 }
	//  }

	shared.AngleVectors(T.cl.refdef.Viewangles[:], T.cl.v_forward[:], T.cl.v_right[:], T.cl.v_up[:])

	//  /* interpolate field of view */
	//  ifov = ops->fov + lerp * (ps->fov - ops->fov);
	//  if (horplus->value) {
	// 	 cl.refdef.fov_x = AdaptFov(ifov, cl.refdef.width, cl.refdef.height);
	//  } else {
	// 	 cl.refdef.fov_x = ifov;
	//  }

	/* don't interpolate blend color */
	for i := 0; i < 4; i++ {
		T.cl.refdef.Blend[i] = ps.Blend[i]
	}

	//  /* add the weapon */
	//  CL_AddViewWeapon(ps, ops);
}

/*
 * Emits all entities, particles, and lights to the refresh
 */
func (T *qClient) addEntities() {
	if T.cls.state != ca_active {
		return
	}

	if T.cl.time > T.cl.frame.servertime {
		if T.cl_showclamp.Bool() {
			T.common.Com_Printf("high clamp %v\n", T.cl.time-T.cl.frame.servertime)
		}

		T.cl.time = T.cl.frame.servertime
		T.cl.lerpfrac = 1.0
	} else if T.cl.time < T.cl.frame.servertime-100 {
		if T.cl_showclamp.Bool() {
			T.common.Com_Printf("low clamp %v\n", T.cl.frame.servertime-100-T.cl.time)
		}

		T.cl.time = T.cl.frame.servertime - 100
		T.cl.lerpfrac = 0
	} else {
		T.cl.lerpfrac = 1.0 - float32(T.cl.frame.servertime-T.cl.time)*0.01
	}

	// if (T.cl_timedemo.Bool()) {
	// 	T.cl.lerpfrac = 1.0;
	// }

	T.calcViewValues()
	// CL_AddPacketEntities(&cl.frame);
	// CL_AddTEnts();
	// CL_AddParticles();
	// CL_AddDLights();
	// CL_AddLightStyles();
}
