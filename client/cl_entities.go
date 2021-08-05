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

func (T *qClient) addPacketEntities(frame *frame_t) {
	// 	entity_t ent = {0};
	// 	entity_state_t *s1;
	// 	float autorotate;
	// 	int i;
	// 	int pnum;
	// 	centity_t *cent;
	// 	int autoanim;
	// 	clientinfo_t *ci;
	// 	unsigned int effects, renderfx;

	/* To distinguish baseq2, xatrix and rogue. */
	// 	cvar_t *game = Cvar_Get("game",  "", CVAR_LATCH | CVAR_SERVERINFO);

	// 	/* bonus items rotate at a fixed rate */
	// 	autorotate = anglemod(cl.time * 0.1f);

	/* brush models can auto animate their frames */
	autoanim := 2 * T.cl.time / 1000

	var ent shared.Entity_t
	for pnum := 0; pnum < frame.num_entities; pnum++ {
		s1 := &T.cl_parse_entities[(frame.parse_entities+pnum)&(MAX_PARSE_ENTITIES-1)]

		cent := &T.cl_entities[s1.Number]

		effects := s1.Effects
		renderfx := s1.Renderfx

		/* set frame */
		if (effects & shared.EF_ANIM01) != 0 {
			ent.Frame = autoanim & 1
		} else if (effects & shared.EF_ANIM23) != 0 {
			ent.Frame = 2 + (autoanim & 1)
		} else if (effects & shared.EF_ANIM_ALL) != 0 {
			ent.Frame = autoanim
		} else if (effects & shared.EF_ANIM_ALLFAST) != 0 {
			ent.Frame = T.cl.time / 100
		} else {
			ent.Frame = s1.Frame
		}

		/* quad and pent can do different things on client */
		if (effects & shared.EF_PENT) != 0 {
			effects &^= shared.EF_PENT
			effects |= shared.EF_COLOR_SHELL
			renderfx |= shared.RF_SHELL_RED
		}

		if (effects & shared.EF_QUAD) != 0 {
			effects &^= shared.EF_QUAD
			effects |= shared.EF_COLOR_SHELL
			renderfx |= shared.RF_SHELL_BLUE
		}

		if (effects & shared.EF_DOUBLE) != 0 {
			effects &^= shared.EF_DOUBLE
			effects |= shared.EF_COLOR_SHELL
			renderfx |= shared.RF_SHELL_DOUBLE
		}

		if (effects & shared.EF_HALF_DAMAGE) != 0 {
			effects &^= shared.EF_HALF_DAMAGE
			effects |= shared.EF_COLOR_SHELL
			renderfx |= shared.RF_SHELL_HALF_DAM
		}

		ent.Oldframe = cent.prev.Frame
		ent.Backlerp = 1.0 - T.cl.lerpfrac

		// 		if (renderfx & (RF_FRAMELERP | RF_BEAM)) != 0 {
		// 			/* step origin discretely, because the
		// 			   frames do the animation properly */
		// 			VectorCopy(cent->current.origin, ent.origin);
		// 			VectorCopy(cent->current.old_origin, ent.oldorigin);
		// 		}
		// 		else
		// 		{
		/* interpolate origin */
		for i := 0; i < 3; i++ {
			ent.Oldorigin[i] = cent.prev.Origin[i] + T.cl.lerpfrac*
				(cent.current.Origin[i]-cent.prev.Origin[i])
			ent.Origin[i] = ent.Oldorigin[i]
		}
		// 		}

		/* tweak the color of beams */
		if (renderfx & shared.RF_BEAM) != 0 {
			/* the four beam colors are encoded in 32 bits of skinnum (hack) */
			ent.Alpha = 0.30
			ent.Skinnum = (s1.Skinnum >> ((shared.Randk() % 4) * 8)) & 0xff
			ent.Model = nil
		} else {
			/* set skin */
			if s1.Modelindex == 255 {
				/* use custom player skin */
				ent.Skinnum = 0
				ci := &T.cl.clientinfo[s1.Skinnum&0xff]
				ent.Skin = ci.skin
				ent.Model = ci.model

				if ent.Skin == nil || ent.Model == nil {
					ent.Skin = T.cl.baseclientinfo.skin
					ent.Model = T.cl.baseclientinfo.model
				}

				// 				if (renderfx & RF_USE_DISGUISE)
				// 				{
				// 					if (ent.skin != NULL)
				// 					{
				// 						if (!strncmp((char *)ent.skin, "players/male", 12))
				// 						{
				// 							ent.skin = R_RegisterSkin("players/male/disguise.pcx");
				// 							ent.model = R_RegisterModel("players/male/tris.md2");
				// 						}
				// 						else if (!strncmp((char *)ent.skin, "players/female", 14))
				// 						{
				// 							ent.skin = R_RegisterSkin("players/female/disguise.pcx");
				// 							ent.model = R_RegisterModel("players/female/tris.md2");
				// 						}
				// 						else if (!strncmp((char *)ent.skin, "players/cyborg", 14))
				// 						{
				// 							ent.skin = R_RegisterSkin("players/cyborg/disguise.pcx");
				// 							ent.model = R_RegisterModel("players/cyborg/tris.md2");
				// 						}
				// 					}
				// 				}
			} else {
				ent.Skinnum = s1.Skinnum
				ent.Skin = nil
				ent.Model = T.cl.model_draw[s1.Modelindex]
			}
		}

		/* only used for black hole model right now */
		if (renderfx&shared.RF_TRANSLUCENT) != 0 && (renderfx&shared.RF_BEAM) == 0 {
			ent.Alpha = 0.70
		}

		/* render effects (fullbright, translucent, etc) */
		if (effects & shared.EF_COLOR_SHELL) != 0 {
			ent.Flags = 0 /* renderfx go on color shell entity */
		} else {
			ent.Flags = renderfx
		}

		// 		/* calculate angles */
		// 		if (effects & EF_ROTATE) != 0 {
		// 			/* some bonus items auto-rotate */
		// 			ent.angles[0] = 0;
		// 			ent.angles[1] = autorotate;
		// 			ent.angles[2] = 0;
		// 		} else if (effects & EF_SPINNINGLIGHTS) != 0 {
		// 			ent.angles[0] = 0;
		// 			ent.angles[1] = anglemod(cl.time / 2) + s1->angles[1];
		// 			ent.angles[2] = 180;
		// 			{
		// 				vec3_t forward;
		// 				vec3_t start;

		// 				AngleVectors(ent.angles, forward, NULL, NULL);
		// 				VectorMA(ent.origin, 64, forward, start);
		// 				V_AddLight(start, 100, 1, 0, 0);
		// 			}
		// 		} else {
		// 			/* interpolate angles */
		for i := 0; i < 3; i++ {
			a1 := cent.current.Angles[i]
			a2 := cent.prev.Angles[i]
			ent.Angles[i] = shared.LerpAngle(a2, a1, T.cl.lerpfrac)
		}
		// 		}

		if s1.Number == T.cl.playernum+1 {
			ent.Flags |= shared.RF_VIEWERMODEL

			// 			if (effects & EF_FLAG1) != 0 {
			// 				V_AddLight(ent.origin, 225, 1.0f, 0.1f, 0.1f);
			// 			}

			// 			else if (effects & EF_FLAG2) != 0 {
			// 				V_AddLight(ent.origin, 225, 0.1f, 0.1f, 1.0f);
			// 			}

			// 			else if (effects & EF_TAGTRAIL) != 0 {
			// 				V_AddLight(ent.origin, 225, 1.0f, 1.0f, 0.0f);
			// 			}

			// 			else if (effects & EF_TRACKERTRAIL) != 0 {
			// 				V_AddLight(ent.origin, 225, -1.0f, -1.0f, -1.0f);
			// 			}

			continue
		}

		/* if set to invisible, skip */
		if s1.Modelindex == 0 {
			continue
		}

		if (effects & shared.EF_BFG) != 0 {
			ent.Flags |= shared.RF_TRANSLUCENT
			ent.Alpha = 0.30
		}

		if (effects & shared.EF_PLASMA) != 0 {
			ent.Flags |= shared.RF_TRANSLUCENT
			ent.Alpha = 0.6
		}

		// 		if (effects & EF_SPHERETRANS) != 0 {
		// 			ent.flags |= RF_TRANSLUCENT;

		// 			if (effects & EF_TRACKERTRAIL) != 0 {
		// 				ent.alpha = 0.6f;
		// 			}

		// 			else
		// 			{
		// 				ent.alpha = 0.3f;
		// 			}
		// 		}

		/* add to refresh list */
		T.addEntity(ent)

		// 		/* color shells generate a seperate entity for the main model */
		// 		if (effects & EF_COLOR_SHELL)
		// 		{
		// 			/* all of the solo colors are fine.  we need to catch any of
		// 			   the combinations that look bad (double & half) and turn
		// 			   them into the appropriate color, and make double/quad
		// 			   something special */
		// 			if (renderfx & RF_SHELL_HALF_DAM)
		// 			{
		// 				if (strcmp(game->string, "rogue") == 0)
		// 				{
		// 					/* ditch the half damage shell if any of red, blue, or double are on */
		// 					if (renderfx & (RF_SHELL_RED | RF_SHELL_BLUE | RF_SHELL_DOUBLE))
		// 					{
		// 						renderfx &= ~RF_SHELL_HALF_DAM;
		// 					}
		// 				}
		// 			}

		// 			if (renderfx & RF_SHELL_DOUBLE)
		// 			{
		// 				if (strcmp(game->string, "rogue") == 0)
		// 				{
		// 					/* lose the yellow shell if we have a red, blue, or green shell */
		// 					if (renderfx & (RF_SHELL_RED | RF_SHELL_BLUE | RF_SHELL_GREEN))
		// 					{
		// 						renderfx &= ~RF_SHELL_DOUBLE;
		// 					}

		// 					/* if we have a red shell, turn it to purple by adding blue */
		// 					if (renderfx & RF_SHELL_RED)
		// 					{
		// 						renderfx |= RF_SHELL_BLUE;
		// 					}

		// 					/* if we have a blue shell (and not a red shell),
		// 					   turn it to cyan by adding green */
		// 					else if (renderfx & RF_SHELL_BLUE)
		// 					{
		// 						/* go to green if it's on already,
		// 						   otherwise do cyan (flash green) */
		// 						if (renderfx & RF_SHELL_GREEN)
		// 						{
		// 							renderfx &= ~RF_SHELL_BLUE;
		// 						}

		// 						else
		// 						{
		// 							renderfx |= RF_SHELL_GREEN;
		// 						}
		// 					}
		// 				}
		// 			}

		// 			ent.flags = renderfx | RF_TRANSLUCENT;
		// 			ent.alpha = 0.30f;
		// 			V_AddEntity(&ent);
		// 		}

		ent.Skin = nil /* never use a custom skin on others */
		ent.Skinnum = 0
		ent.Flags = 0
		ent.Alpha = 0

		/* duplicate for linked models */
		if s1.Modelindex2 != 0 {
			if s1.Modelindex2 == 255 {
				/* custom weapon */
				ci := &T.cl.clientinfo[s1.Skinnum&0xff]
				i := (s1.Skinnum >> 8) /* 0 is default weapon model */

				if !T.cl_vwep.Bool() || (i > MAX_CLIENTWEAPONMODELS-1) {
					i = 0
				}

				ent.Model = ci.weaponmodel[i]

				if ent.Model == nil {
					if i != 0 {
						ent.Model = ci.weaponmodel[0]
					}

					if ent.Model == nil {
						ent.Model = T.cl.baseclientinfo.weaponmodel[0]
					}
				}
			} else {
				ent.Model = T.cl.model_draw[s1.Modelindex2]
			}

			/* check for the defender sphere shell and make it translucent */
			// if !Q_strcasecmp(cl.configstrings[CS_MODELS+(s1.modelindex2)],
			// 	"models/items/shell/tris.md2") {
			// 	ent.alpha = 0.32
			// 	ent.flags = RF_TRANSLUCENT
			// }

			T.addEntity(ent)

			ent.Flags = 0
			ent.Alpha = 0
		}

		if s1.Modelindex3 != 0 {
			ent.Model = T.cl.model_draw[s1.Modelindex3]
			T.addEntity(ent)
		}

		if s1.Modelindex4 != 0 {
			ent.Model = T.cl.model_draw[s1.Modelindex4]
			T.addEntity(ent)
		}

		// 		if (effects & EF_POWERSCREEN) != 0 {
		// 			ent.model = cl_mod_powerscreen;
		// 			ent.oldframe = 0;
		// 			ent.frame = 0;
		// 			ent.flags |= (RF_TRANSLUCENT | RF_SHELL_GREEN);
		// 			ent.alpha = 0.30f;
		// 			V_AddEntity(&ent);
		// 		}

		// 		/* add automatic particle trails */
		// 		if ((effects & ~EF_ROTATE))
		// 		{
		// 			if (effects & EF_ROCKET)
		// 			{
		// 				CL_RocketTrail(cent->lerp_origin, ent.origin, cent);

		// 				if (cl_r1q2_lightstyle->value)
		// 				{
		// 					V_AddLight(ent.origin, 200, 1, 0.23f, 0);
		// 				}
		// 				else
		// 				{
		// 					V_AddLight(ent.origin, 200, 1, 1, 0);
		// 				}
		// 			}

		// 			/* Do not reorder EF_BLASTER and EF_HYPERBLASTER.
		// 			   EF_BLASTER | EF_TRACKER is a special case for
		// 			   EF_BLASTER2 */
		// 			else if (effects & EF_BLASTER)
		// 			{
		// 				if (effects & EF_TRACKER)
		// 				{
		// 					CL_BlasterTrail2(cent->lerp_origin, ent.origin);
		// 					V_AddLight(ent.origin, 200, 0, 1, 0);
		// 				}
		// 				else
		// 				{
		// 					CL_BlasterTrail(cent->lerp_origin, ent.origin);
		// 					V_AddLight(ent.origin, 200, 1, 1, 0);
		// 				}
		// 			}
		// 			else if (effects & EF_HYPERBLASTER)
		// 			{
		// 				if (effects & EF_TRACKER)
		// 				{
		// 					V_AddLight(ent.origin, 200, 0, 1, 0);
		// 				}
		// 				else
		// 				{
		// 					V_AddLight(ent.origin, 200, 1, 1, 0);
		// 				}
		// 			}
		// 			else if (effects & EF_GIB)
		// 			{
		// 				CL_DiminishingTrail(cent->lerp_origin, ent.origin,
		// 						cent, effects);
		// 			}
		// 			else if (effects & EF_GRENADE)
		// 			{
		// 				CL_DiminishingTrail(cent->lerp_origin, ent.origin,
		// 						cent, effects);
		// 			}
		// 			else if (effects & EF_FLIES)
		// 			{
		// 				CL_FlyEffect(cent, ent.origin);
		// 			}
		// 			else if (effects & EF_BFG)
		// 			{
		// 				static int bfg_lightramp[6] = {300, 400, 600, 300, 150, 75};

		// 				if (effects & EF_ANIM_ALLFAST)
		// 				{
		// 					CL_BfgParticles(&ent);
		// 					i = 200;
		// 				}
		// 				else
		// 				{
		// 					i = bfg_lightramp[s1->frame];
		// 				}

		// 				V_AddLight(ent.origin, i, 0, 1, 0);
		// 			}
		// 			else if (effects & EF_TRAP)
		// 			{
		// 				ent.origin[2] += 32;
		// 				CL_TrapParticles(&ent);
		// 				i = (randk() % 100) + 100;
		// 				V_AddLight(ent.origin, i, 1, 0.8f, 0.1f);
		// 			}
		// 			else if (effects & EF_FLAG1)
		// 			{
		// 				CL_FlagTrail(cent->lerp_origin, ent.origin, 242);
		// 				V_AddLight(ent.origin, 225, 1, 0.1f, 0.1f);
		// 			}
		// 			else if (effects & EF_FLAG2)
		// 			{
		// 				CL_FlagTrail(cent->lerp_origin, ent.origin, 115);
		// 				V_AddLight(ent.origin, 225, 0.1f, 0.1f, 1);
		// 			}
		// 			else if (effects & EF_TAGTRAIL)
		// 			{
		// 				CL_TagTrail(cent->lerp_origin, ent.origin, 220);
		// 				V_AddLight(ent.origin, 225, 1.0, 1.0, 0.0);
		// 			}
		// 			else if (effects & EF_TRACKERTRAIL)
		// 			{
		// 				if (effects & EF_TRACKER)
		// 				{
		// 					float intensity;

		// 					intensity = 50 + (500 * ((float)sin(cl.time / 500.0f) + 1.0f));
		// 					V_AddLight(ent.origin, intensity, -1.0, -1.0, -1.0);
		// 				}
		// 				else
		// 				{
		// 					CL_Tracker_Shell(cent->lerp_origin);
		// 					V_AddLight(ent.origin, 155, -1.0, -1.0, -1.0);
		// 				}
		// 			}
		// 			else if (effects & EF_TRACKER)
		// 			{
		// 				CL_TrackerTrail(cent->lerp_origin, ent.origin, 0);
		// 				V_AddLight(ent.origin, 200, -1, -1, -1);
		// 			}
		// 			else if (effects & EF_IONRIPPER)
		// 			{
		// 				CL_IonripperTrail(cent->lerp_origin, ent.origin);
		// 				V_AddLight(ent.origin, 100, 1, 0.5, 0.5);
		// 			}
		// 			else if (effects & EF_BLUEHYPERBLASTER)
		// 			{
		// 				V_AddLight(ent.origin, 200, 0, 0, 1);
		// 			}
		// 			else if (effects & EF_PLASMA)
		// 			{
		// 				if (effects & EF_ANIM_ALLFAST)
		// 				{
		// 					CL_BlasterTrail(cent->lerp_origin, ent.origin);
		// 				}

		// 				V_AddLight(ent.origin, 130, 1, 0.5, 0.5);
		// 			}
		// 		}

		copy(cent.lerp_origin, ent.Origin[:])
	}
}

/*
 * Adapts a 4:3 aspect FOV to the current aspect (Hor+)
 */
func adaptFov(fov, w, h float32) float32 {

	if w <= 0 || h <= 0 {
		return fov
	}

	/*
	 * Formula:
	 *
	 * fov = 2.0 * atan(width / height * 3.0 / 4.0 * tan(fov43 / 2.0))
	 *
	 * The code below is equivalent but precalculates a few values and
	 * converts between degrees and radians when needed.
	 */
	return float32(math.Atan(math.Tan(float64(fov)/360.0*math.Pi)*float64(w/h*0.75)) / math.Pi * 360.0)
}

/*
 * Sets cl.refdef view values
 */
func (T *qClient) calcViewValues() {

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

	var lerp float32
	if T.cl_paused.Bool() {
		lerp = 1.0
	} else {
		lerp = T.cl.lerpfrac
	}

	/* calculate the origin */
	if (T.cl_predict.Bool()) && (T.cl.frame.playerstate.Pmove.Pm_flags&shared.PMF_NO_PREDICTION) == 0 {
		/* use predicted values */

		backlerp := 1.0 - lerp

		for i := 0; i < 3; i++ {
			T.cl.refdef.Vieworg[i] = T.cl.predicted_origin[i] + ops.Viewoffset[i] +
				T.cl.lerpfrac*(ps.Viewoffset[i]-ops.Viewoffset[i]) -
				backlerp*T.cl.prediction_error[i]
		}

		/* smooth out stair climbing */
		delta := T.cls.realtime - int(T.cl.predicted_step_time)

		if delta < 100 {
			T.cl.refdef.Vieworg[2] -= T.cl.predicted_step * float32(100-delta) * 0.01
		}
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

	if T.cl_kickangles.Bool() {
		for i := 0; i < 3; i++ {
			T.cl.refdef.Viewangles[i] += shared.LerpAngle(ops.Kick_angles[i], ps.Kick_angles[i], lerp)
		}
	}

	shared.AngleVectors(T.cl.refdef.Viewangles[:], T.cl.v_forward[:], T.cl.v_right[:], T.cl.v_up[:])

	/* interpolate field of view */
	ifov := ops.Fov + lerp*(ps.Fov-ops.Fov)
	if T.horplus.Bool() {
		T.cl.refdef.Fov_x = adaptFov(ifov, float32(T.cl.refdef.Width), float32(T.cl.refdef.Height))
	} else {
		T.cl.refdef.Fov_x = ifov
	}

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
	T.addPacketEntities(&T.cl.frame)
	T.addTEnts()
	T.addParticles()
	T.addDLights()
	T.addLightStyles()
}
