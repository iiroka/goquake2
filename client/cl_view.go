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
 *  =======================================================================
 *
 * This file implements the camera, e.g the player's view
 *
 * =======================================================================
 */
package client

import (
	"goquake2/shared"
	"math"
)

/*
 * Specifies the model that will be used as the world
 */
func (T *qClient) clearScene() {
	//  r_numdlights = 0;
	//  r_numparticles = 0;
	T.r_entities = make([]shared.Entity_t, 0)
}

func (T *qClient) addEntity(ent shared.Entity_t) {
	if len(T.r_entities) >= shared.MAX_ENTITIES {
		return
	}

	T.r_entities = append(T.r_entities, ent)
}

/*
 * Call before entering a new level, or after changing dlls
 */
func (T *qClient) prepRefresh() error {
	//  char mapname[MAX_QPATH];
	//  int i;
	//  char name[MAX_QPATH];
	//  float rotate;
	//  vec3_t axis;

	if len(T.cl.configstrings[shared.CS_MODELS+1]) == 0 {
		return nil
	}

	//  SCR_AddDirtyPoint(0, 0);
	//  SCR_AddDirtyPoint(viddef.width - 1, viddef.height - 1);

	/* let the refresher load the map */
	mapname := T.cl.configstrings[shared.CS_MODELS+1][5:]
	mapname = mapname[0 : len(mapname)-4] /* cut off ".bsp" */

	/* register models, pics, and skins */
	T.common.Com_Printf("Map: %s\r", mapname)
	T.scrUpdateScreen()
	err := T.R_BeginRegistration(mapname)
	if err != nil {
		return err
	}
	T.common.Com_Printf("                                     \r")

	/* precache status bar pics */
	T.common.Com_Printf("pics\r")
	T.scrUpdateScreen()
	//  SCR_TouchPics();
	T.common.Com_Printf("                                     \r")

	//  CL_RegisterTEntModels();

	//  num_cl_weaponmodels = 1;
	//  strcpy(cl_weaponmodels[0], "weapon.md2");

	for i := 1; i < shared.MAX_MODELS && len(T.cl.configstrings[shared.CS_MODELS+i]) > 0; i++ {
		name := T.cl.configstrings[shared.CS_MODELS+i]

		if name[0] != '*' {
			T.common.Com_Printf("%s\r", name)
		}

		T.scrUpdateScreen()
		T.input.Update()

		if name[0] == '#' {
			// 		 /* special player weapon model */
			// 		 if (num_cl_weaponmodels < MAX_CLIENTWEAPONMODELS)
			// 		 {
			// 			 Q_strlcpy(cl_weaponmodels[num_cl_weaponmodels],
			// 					 cl.configstrings[CS_MODELS + i] + 1,
			// 					 sizeof(cl_weaponmodels[num_cl_weaponmodels]));
			// 			 num_cl_weaponmodels++;
			// 		 }
		} else {
			T.cl.model_draw[i], err = T.R_RegisterModel(T.cl.configstrings[shared.CS_MODELS+i])
			if err != nil {
				return err
			}

			// 		 if (name[0] == '*') {
			// 			 cl.model_clip[i] = CM_InlineModel(cl.configstrings[CS_MODELS + i]);
			// 		 } else {
			// 			 cl.model_clip[i] = NULL;
			// 		 }
		}

		if name[0] != '*' {
			T.common.Com_Printf("                                     \r")
		}
	}

	T.common.Com_Printf("images\r")
	T.scrUpdateScreen()

	//  for (i = 1; i < MAX_IMAGES && cl.configstrings[CS_IMAGES + i][0]; i++) {
	// 	 cl.image_precache[i] = Draw_FindPic(cl.configstrings[CS_IMAGES + i]);
	// 	 IN_Update();
	//  }

	T.common.Com_Printf("                                     \r")

	for i := 0; i < shared.MAX_CLIENTS; i++ {
		if len(T.cl.configstrings[shared.CS_PLAYERSKINS+i]) == 0 {
			continue
		}

		T.common.Com_Printf("client %i\r", i)
		T.scrUpdateScreen()
		T.input.Update()
		T.parseClientinfo(i)
		T.common.Com_Printf("                                     \r")
	}

	T.loadClientinfo(&T.cl.baseclientinfo, "unnamed\\male/grunt")

	/* set sky textures and speed */
	T.common.Com_Printf("sky\r")
	T.scrUpdateScreen()
	//  rotate = (float)strtod(cl.configstrings[CS_SKYROTATE], (char **)NULL);
	//  sscanf(cl.configstrings[CS_SKYAXIS], "%f %f %f", &axis[0], &axis[1], &axis[2]);
	//  R_SetSky(cl.configstrings[CS_SKY], rotate, axis);
	T.common.Com_Printf("                                     \r")

	//  /* the renderer can now free unneeded stuff */
	//  R_EndRegistration();

	//  /* clear any lines of console text */
	//  Con_ClearNotify();

	T.scrUpdateScreen()
	T.cl.refresh_prepped = true
	T.cl.force_refdef = true /* make sure we have a valid refdef */

	//  /* start the cd track */
	//  int track = (int)strtol(cl.configstrings[CS_CDTRACK], (char **)NULL, 10);

	//  OGG_PlayTrack(track);
	return nil
}

func (T *qClient) renderView(stereo_separation float32) error {
	if T.cls.state != ca_active {
		// R_EndWorldRenderpass();
		return nil
	}

	if !T.cl.refresh_prepped {
		// R_EndWorldRenderpass();
		return nil // still loading
	}

	// if (cl_timedemo->value) {
	// 	if (!cl.timedemo_start) {
	// 		cl.timedemo_start = Sys_Milliseconds();
	// 	}

	// 	cl.timedemo_frames++;
	// }

	/* an invalid frame will just use the exact previous refdef
	   we can't use the old frame if the video mode has changed, though... */
	if T.cl.frame.valid && (T.cl.force_refdef || !T.cl_paused.Bool()) {
		T.cl.force_refdef = false

		T.clearScene()

		/* build a refresh entity list and calc cl.sim*
		   this also calls CL_CalcViewValues which loads
		   v_forward, etc. */
		T.addEntities()

		// 	// before changing viewport we should trace the crosshair position
		// 	V_Render3dCrosshair();

		// 	if (cl_testparticles->value)
		// 	{
		// 		V_TestParticles();
		// 	}

		// 	if (cl_testentities->value)
		// 	{
		// 		V_TestEntities();
		// 	}

		// 	if (cl_testlights->value)
		// 	{
		// 		V_TestLights();
		// 	}

		// 	if (cl_testblend->value)
		// 	{
		// 		cl.refdef.blend[0] = 1;
		// 		cl.refdef.blend[1] = 0.5;
		// 		cl.refdef.blend[2] = 0.25;
		// 		cl.refdef.blend[3] = 0.5;
		// 	}

		// 	/* offset vieworg appropriately if
		// 	   we're doing stereo separation */

		// 	if (stereo_separation != 0)
		// 	{
		// 		vec3_t tmp;

		// 		VectorScale(cl.v_right, stereo_separation, tmp);
		// 		VectorAdd(cl.refdef.vieworg, tmp, cl.refdef.vieworg);
		// 	}

		/* never let it sit exactly on a node line, because a water plane can
		   dissapear when viewed with the eye exactly on it. the server protocol
		   only specifies to 1/8 pixel, so add 1/16 in each axis */
		T.cl.refdef.Vieworg[0] += 1.0 / 16
		T.cl.refdef.Vieworg[1] += 1.0 / 16
		T.cl.refdef.Vieworg[2] += 1.0 / 16

		T.cl.refdef.Time = float32(T.cl.time) * 0.001

		T.cl.refdef.Areabits = T.cl.frame.areabits[:]

		// 	if (!cl_add_entities->value) {
		// 		r_numentities = 0;
		// 	}

		// 	if (!cl_add_particles->value) {
		// 		r_numparticles = 0;
		// 	}

		// 	if (!cl_add_lights->value) {
		// 		r_numdlights = 0;
		// 	}

		// 	if (!cl_add_blend->value) {
		// 		VectorClear(cl.refdef.blend);
		// 	}

		// 	cl.refdef.num_entities = r_numentities;
		T.cl.refdef.Entities = T.r_entities
		// 	cl.refdef.num_particles = r_numparticles;
		// 	cl.refdef.particles = r_particles;
		// 	cl.refdef.num_dlights = r_numdlights;
		// 	cl.refdef.dlights = r_dlights;
		// 	cl.refdef.lightstyles = r_lightstyles;

		T.cl.refdef.Rdflags = T.cl.frame.playerstate.Rdflags

		// 	/* sort entities for better cache locality */
		// 	qsort(cl.refdef.entities, cl.refdef.num_entities,
		// 			sizeof(cl.refdef.entities[0]), (int (*)(const void *, const void *))
		// 			entitycmpfnc);
	} else if T.cl.frame.valid && T.cl_paused.Bool() && T.gl1_stereo.Bool() {
		// We need to adjust the refdef in stereo mode when paused.
		// 	vec3_t tmp;
		T.calcViewValues()
		// 	VectorScale( cl.v_right, stereo_separation, tmp );
		// 	VectorAdd( cl.refdef.vieworg, tmp, cl.refdef.vieworg );

		// 	cl.refdef.vieworg[0] += 1.0/16;
		// 	cl.refdef.vieworg[1] += 1.0/16;
		// 	cl.refdef.vieworg[2] += 1.0/16;

		T.cl.refdef.Time = float32(T.cl.time) * 0.001
	}

	T.cl.refdef.X = T.scr_vrect.x
	T.cl.refdef.Y = T.scr_vrect.y
	T.cl.refdef.Width = T.scr_vrect.width
	T.cl.refdef.Height = T.scr_vrect.height
	fov, err := T.calcFov(T.cl.refdef.Fov_x, float32(T.cl.refdef.Width), float32(T.cl.refdef.Height))
	if err != nil {
		return err
	}
	T.cl.refdef.Fov_y = fov

	if err := T.R_RenderFrame(T.cl.refdef); err != nil {
		return err
	}

	// if (T.cl_stats)
	// {
	T.common.Com_Printf("ent:%v  lt:%v  part:%v\n", len(T.r_entities), 0, 0)
	// r_numdlights, r_numparticles)
	// }

	// if (log_stats->value && (log_stats_file != 0))
	// {
	// 	fprintf(log_stats_file, "%i,%i,%i,", r_numentities,
	// 			r_numdlights, r_numparticles);
	// }

	T.scrAddDirtyPoint(T.scr_vrect.x, T.scr_vrect.y)
	T.scrAddDirtyPoint(T.scr_vrect.x+T.scr_vrect.width-1,
		T.scr_vrect.y+T.scr_vrect.height-1)

	// SCR_DrawCrosshair();
	return nil
}

func (T *qClient) calcFov(fov_x, width, height float32) (float32, error) {

	if (fov_x < 1) || (fov_x > 179) {
		return 0, T.common.Com_Error(shared.ERR_DROP, "Bad fov: %f", fov_x)
	}

	x := float64(width) / math.Tan(float64(fov_x)/360*math.Pi)

	a := math.Atan(float64(height) / x)

	return float32(a * 360 / math.Pi), nil
}
