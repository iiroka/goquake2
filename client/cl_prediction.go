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
	"strconv"
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

func (T *qClient) clipMoveToEntities(start, mins, maxs, end []float32, tr *shared.Trace_t) {

	var headnode int
	var angles []float32
	var bmins [3]float32
	var bmaxs [3]float32
	for i := 0; i < T.cl.frame.num_entities; i++ {
		num := (T.cl.frame.parse_entities + i) & (MAX_PARSE_ENTITIES - 1)
		ent := &T.cl_parse_entities[num]

		if ent.Solid == 0 {
			continue
		}

		if ent.Number == T.cl.playernum+1 {
			continue
		}

		if ent.Solid == 31 {
			/* special value for bmodel */
			cmodel := T.cl.model_clip[ent.Modelindex]

			if cmodel == nil {
				continue
			}

			headnode = cmodel.Headnode
			angles = ent.Angles[:]
		} else {
			/* encoded bbox */
			x := 8 * (ent.Solid & 31)
			zd := 8 * ((ent.Solid >> 5) & 31)
			zu := 8*((ent.Solid>>10)&63) - 32

			bmins[0] = float32(-x)
			bmins[1] = float32(-x)
			bmaxs[0] = float32(x)
			bmaxs[1] = float32(x)
			bmins[2] = float32(-zd)
			bmaxs[2] = float32(zu)

			headnode = T.common.CMHeadnodeForBox(bmins[:], bmaxs[:])
			angles = []float32{0, 0, 0} /* boxes don't rotate */
		}

		if tr.Allsolid {
			return
		}

		trace := T.common.CMTransformedBoxTrace(start, end,
			mins, maxs, headnode, shared.MASK_PLAYERSOLID,
			ent.Origin[:], angles)

		if trace.Allsolid || trace.Startsolid ||
			(trace.Fraction < tr.Fraction) {
			trace.Ent = ent

			if tr.Startsolid {
				tr.Copy(trace)
				tr.Startsolid = true
			} else {
				tr.Copy(trace)
			}
		}
	}
}

func clPMTrace(start, mins, maxs, end []float32, a interface{}) shared.Trace_t {

	T := a.(*qClient)

	/* check against world */
	t := T.common.CMBoxTrace(start, end, mins, maxs, 0, shared.MASK_PLAYERSOLID)
	if t.Fraction < 1.0 {
		// t.ent = (struct edict_s *)1;
	}

	/* check all other solid models */
	T.clipMoveToEntities(start, mins, maxs, end, &t)

	return t
}

/*
 * Sets cl.predicted_origin and cl.predicted_angles
 */
func (T *qClient) predictMovement() {

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
	pm := shared.Pmove_t{}
	pm.TraceArg = T
	pm.Trace = clPMTrace
	//  pm.pointcontents = CL_PMpointcontents;
	aa, _ := strconv.ParseFloat(T.cl.configstrings[shared.CS_AIRACCEL], 32)
	T.common.SetAirAccelerate(float32(aa))
	pm.S = T.cl.frame.playerstate.Pmove

	/* run frames */
	for ack < current {
		ack++
		frame := ack & (CMD_BACKUP - 1)
		cmd := &T.cl.cmds[frame]

		// Ignore null entries
		if cmd.Msec == 0 {
			continue
		}

		pm.Cmd = *cmd
		T.common.Pmove(&pm)

		/* save for debug checking */
		copy(T.cl.predicted_origins[frame][:], pm.S.Origin[:])
	}

	step := int(pm.S.Origin[2]) - int(T.cl.predicted_origin[2]*8)

	if (step > 126 && step < 130) &&
		(pm.S.Velocity[0] != 0 || pm.S.Velocity[1] != 0 || pm.S.Velocity[2] != 0) &&
		(pm.S.Pm_flags&shared.PMF_ON_GROUND) != 0 {
		T.cl.predicted_step = float32(step) * 0.125
		T.cl.predicted_step_time = uint(T.cls.realtime - int(T.cls.nframetime*500))
	}

	/* copy results out for rendering */
	T.cl.predicted_origin[0] = float32(pm.S.Origin[0]) * 0.125
	T.cl.predicted_origin[1] = float32(pm.S.Origin[1]) * 0.125
	T.cl.predicted_origin[2] = float32(pm.S.Origin[2]) * 0.125

	copy(T.cl.predicted_angles[:], pm.Viewangles[:])
}
