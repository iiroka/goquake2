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
 * This file implements interpolation between two frames. This is used
 * to smooth down network play
 *
 * =======================================================================
 */
package client

import (
	"goquake2/shared"
	"math"
)

func (T *qClient) checkPredictionError() {

	if !T.cl_predict.Bool() ||
		(T.cl.frame.playerstate.Pmove.Pm_flags&shared.PMF_NO_PREDICTION) != 0 {
		return
	}

	/* calculate the last usercmd_t we sent that the server has processed */
	frame := T.cls.netchan.Incoming_acknowledged
	frame &= (CMD_BACKUP - 1)

	/* compare what the server returned with what we had predicted it to be */
	var delta [3]int16
	shared.VectorSubtract16(T.cl.frame.playerstate.Pmove.Origin[:],
		T.cl.predicted_origins[frame][:], delta[:])

	/* save the prediction error for interpolation */
	len := int(math.Abs(float64(delta[0])) + math.Abs(float64(delta[1])) + math.Abs(float64(delta[2])))

	/* 80 world units */
	if len > 640 {
		/* a teleport or something */
		T.cl.prediction_error[0] = 0
		T.cl.prediction_error[1] = 0
		T.cl.prediction_error[2] = 0
	} else {
		// 	 if (cl_showmiss->value && (delta[0] || delta[1] || delta[2])) {
		// 		 Com_Printf("prediction miss on %i: %i\n", cl.frame.serverframe,
		// 				 delta[0] + delta[1] + delta[2]);
		// 	 }

		copy(T.cl.predicted_origins[frame][:], T.cl.frame.playerstate.Pmove.Origin[:])

		/* save for error itnerpolation */
		for i := 0; i < 3; i++ {
			T.cl.prediction_error[i] = float32(delta[i]) * 0.125
		}
	}
}

/*
 * Sets cl.predicted_origin and cl.predicted_angles
 */
func (T *qClient) predictMovement() {
	//  int ack, current;
	//  int frame;
	//  usercmd_t *cmd;
	//  pmove_t pm;
	//  int i;
	//  int step;
	//  vec3_t tmp;

	if T.cls.state != ca_active {
		return
	}

	if T.cl_paused.Bool() {
		return
	}

	if !T.cl_predict.Bool() ||
		(T.cl.frame.playerstate.Pmove.Pm_flags&shared.PMF_NO_PREDICTION) != 0 {
		/* just set angles */
		for i := 0; i < 3; i++ {
			T.cl.predicted_angles[i] = T.cl.viewangles[i] + shared.SHORT2ANGLE(
				int(T.cl.frame.playerstate.Pmove.Delta_angles[i]))
		}

		return
	}

	ack := T.cls.netchan.Incoming_acknowledged
	current := T.cls.netchan.Outgoing_sequence

	/* if we are too far out of date, just freeze */
	if current-ack >= CMD_BACKUP {
		// 	 if (cl_showmiss->value) {
		// 		 Com_Printf("exceeded CMD_BACKUP\n");
		// 	 }

		return
	}

	/* copy current state to pmove */
	//  memset (&pm, 0, sizeof(pm));
	//  pm.trace = CL_PMTrace;
	//  pm.pointcontents = CL_PMpointcontents;
	//  pm_airaccelerate = atof(cl.configstrings[CS_AIRACCEL]);
	//  pm.s = cl.frame.playerstate.pmove;

	/* run frames */
	for ack < current {
		ack++
		// 	 frame = ack & (CMD_BACKUP - 1);
		// 	 cmd = &cl.cmds[frame];

		// 	 // Ignore null entries
		// 	 if (!cmd->msec) {
		// 		 continue;
		// 	 }

		// 	 pm.cmd = *cmd;
		// 	 Pmove(&pm);

		// 	 /* save for debug checking */
		// 	 VectorCopy(pm.s.origin, cl.predicted_origins[frame]);
	}

	//  step = pm.s.origin[2] - (int)(cl.predicted_origin[2] * 8);
	//  VectorCopy(pm.s.velocity, tmp);

	//  if (((step > 126 && step < 130))
	// 	 && !VectorCompare(tmp, vec3_origin)
	// 	 && (pm.s.pm_flags & PMF_ON_GROUND))
	//  {
	// 	 cl.predicted_step = step * 0.125f;
	// 	 cl.predicted_step_time = cls.realtime - (int)(cls.nframetime * 500);
	//  }

	/* copy results out for rendering */
	//  T.cl.predicted_origin[0] = pm.s.origin[0] * 0.125f;
	//  T.cl.predicted_origin[1] = pm.s.origin[1] * 0.125f;
	//  T.cl.predicted_origin[2] = pm.s.origin[2] * 0.125f;

	//  VectorCopy(pm.viewangles, cl.predicted_angles);
}
