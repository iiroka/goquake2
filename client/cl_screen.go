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
 * This file implements the 2D stuff. For example the HUD and the
 * networkgraph.
 *
 * =======================================================================
 */
package client

import "goquake2/shared"

func (T *qClient) scrInit() {
	T.scr_viewsize = T.common.Cvar_Get("viewsize", "100", shared.CVAR_ARCHIVE)
	T.scr_conspeed = T.common.Cvar_Get("scr_conspeed", "3", 0)
	T.scr_centertime = T.common.Cvar_Get("scr_centertime", "2.5", 0)
	T.scr_showturtle = T.common.Cvar_Get("scr_showturtle", "0", 0)
	T.scr_showpause = T.common.Cvar_Get("scr_showpause", "1", 0)
	T.scr_netgraph = T.common.Cvar_Get("netgraph", "0", 0)
	T.scr_timegraph = T.common.Cvar_Get("timegraph", "0", 0)
	T.scr_debuggraph = T.common.Cvar_Get("debuggraph", "0", 0)
	T.scr_graphheight = T.common.Cvar_Get("graphheight", "32", 0)
	T.scr_graphscale = T.common.Cvar_Get("graphscale", "1", 0)
	T.scr_graphshift = T.common.Cvar_Get("graphshift", "0", 0)
	T.scr_drawall = T.common.Cvar_Get("scr_drawall", "0", 0)
	T.r_hudscale = T.common.Cvar_Get("r_hudscale", "-1", shared.CVAR_ARCHIVE)
	T.r_consolescale = T.common.Cvar_Get("r_consolescale", "-1", shared.CVAR_ARCHIVE)
	T.r_menuscale = T.common.Cvar_Get("r_menuscale", "-1", shared.CVAR_ARCHIVE)

	/* register our commands */
	// Cmd_AddCommand("timerefresh", SCR_TimeRefresh_f);
	// Cmd_AddCommand("loading", SCR_Loading_f);
	// Cmd_AddCommand("sizeup", SCR_SizeUp_f);
	// Cmd_AddCommand("sizedown", SCR_SizeDown_f);
	// Cmd_AddCommand("sky", SCR_Sky_f);

	T.scr_initialized = true
}

/*
 * Sets scr_vrect, the coordinates of the rendered window
 */
func (T *qClient) scrCalcVrect() {

	/* bound viewsize */
	if T.scr_viewsize.Int() < 40 {
		T.common.Cvar_Set("viewsize", "40")
	}

	if T.scr_viewsize.Int() > 100 {
		T.common.Cvar_Set("viewsize", "100")
	}

	size := T.scr_viewsize.Int()

	T.scr_vrect.width = T.viddef.width * size / 100
	T.scr_vrect.height = T.viddef.height * size / 100

	T.scr_vrect.x = (T.viddef.width - T.scr_vrect.width) / 2
	T.scr_vrect.y = (T.viddef.height - T.scr_vrect.height) / 2
}

func (T *qClient) scrDrawConsole() {
	T.conCheckResize()

	if (T.cls.state == ca_disconnected) || (T.cls.state == ca_connecting) {
		/* forced full screen console */
		T.conDrawConsole(1.0)
		return
	}

	if (T.cls.state != ca_active) || !T.cl.refresh_prepped {
		/* connected, but can't render */
		T.conDrawConsole(0.5)
		// Draw_Fill(0, viddef.height/2, viddef.width, viddef.height/2, 0)
		return
	}

	if T.scr_con_current > 0 {
		T.conDrawConsole(T.scr_con_current)
	} else {
		// if (cls.key_dest == key_game) || (cls.key_dest == key_message) {
		// 	Con_DrawNotify() /* only draw notify in game */
		// }
	}
}

// ----
/*
 * This is called every frame, and can also be called
 * explicitly to flush text to the screen.
 */
func (T *qClient) scrUpdateScreen() error {
	//  int numframes;
	//  int i;
	//  float separation[2] = {0, 0};
	//  float scale = SCR_GetMenuScale();

	//  /* if the screen is disabled (loading plaque is
	// 	up, or vid mode changing) do nothing at all */
	//  if (cls.disable_screen)
	//  {
	// 	 if (Sys_Milliseconds() - cls.disable_screen > 120000)
	// 	 {
	// 		 cls.disable_screen = 0;
	// 		 Com_Printf("Loading plaque timed out.\n");
	// 	 }

	// 	 return;
	//  }

	if !T.scr_initialized || !T.con.initialized {
		return nil /* not initialized yet */
	}

	numframes := 1
	separation := []float32{0, 0}
	//  if ( gl1_stereo->value )
	//  {
	// 	 numframes = 2;
	// 	 separation[0] = -gl1_stereo_separation->value / 2;
	// 	 separation[1] = +gl1_stereo_separation->value / 2;
	//  }
	//  else
	//  {
	// 	 separation[0] = 0;
	// 	 separation[1] = 0;
	// 	 numframes = 1;
	//  }

	for i := 0; i < numframes; i++ {
		if err := T.R_BeginFrame(separation[i]); err != nil {
			return err
		}

		// 	 if (scr_draw_loading == 2)
		// 	 {
		// 		 /* loading plaque over black screen */
		// 		 int w, h;

		// 		 R_EndWorldRenderpass();
		// 		 if(i == 0){
		// 			 R_SetPalette(NULL);
		// 		 }

		// 		 if(i == numframes - 1){
		// 			 scr_draw_loading = false;
		// 		 }

		// 		 Draw_GetPicSize(&w, &h, "loading");
		// 		 Draw_PicScaled((viddef.width - w * scale) / 2, (viddef.height - h * scale) / 2, "loading", scale);
		// 	 }

		// 	 /* if a cinematic is supposed to be running,
		// 		handle menus and console specially */
		// 	 else if (cl.cinematictime > 0)
		// 	 {
		// 		 if (cls.key_dest == key_menu)
		// 		 {
		// 			 if (cl.cinematicpalette_active)
		// 			 {
		// 				 R_SetPalette(NULL);
		// 				 cl.cinematicpalette_active = false;
		// 			 }

		// 			 R_EndWorldRenderpass();
		// 			 M_Draw();
		// 		 }
		// 		 else if (cls.key_dest == key_console)
		// 		 {
		// 			 if (cl.cinematicpalette_active)
		// 			 {
		// 				 R_SetPalette(NULL);
		// 				 cl.cinematicpalette_active = false;
		// 			 }

		// 			 R_EndWorldRenderpass();
		// 			 SCR_DrawConsole();
		// 		 }
		// 		 else
		// 		 {
		// 			 R_EndWorldRenderpass();
		// 			 SCR_DrawCinematic();
		// 		 }
		// 	 }
		// 	 else
		// 	 {
		// 		 /* make sure the game palette is active */
		// 		 if (cl.cinematicpalette_active)
		// 		 {
		// 			 R_SetPalette(NULL);
		// 			 cl.cinematicpalette_active = false;
		// 		 }

		/* do 3D refresh drawing, and then update the screen */
		T.scrCalcVrect()

		// 		 /* clear any dirty part of the background */
		// 		 SCR_TileClear();

		if err := T.renderView(separation[i]); err != nil {
			return err
		}

		// 		 SCR_DrawStats();

		// 		 if (cl.frame.playerstate.stats[STAT_LAYOUTS] & 1)
		// 		 {
		// 			 SCR_DrawLayout();
		// 		 }

		// 		 if (cl.frame.playerstate.stats[STAT_LAYOUTS] & 2)
		// 		 {
		// 			 CL_DrawInventory();
		// 		 }

		// 		 SCR_DrawNet();
		// 		 SCR_CheckDrawCenterString();

		// 		 if (scr_timegraph->value)
		// 		 {
		// 			 SCR_DebugGraph(cls.rframetime * 300, 0);
		// 		 }

		// 		 if (scr_debuggraph->value || scr_timegraph->value ||
		// 			 scr_netgraph->value)
		// 		 {
		// 			 SCR_DrawDebugGraph();
		// 		 }

		// 		 SCR_DrawPause();

		T.scrDrawConsole()

		// 		 M_Draw();

		// 		 SCR_DrawLoading();
		// 	 }
	}

	//  SCR_Framecounter();
	T.R_EndFrame()
	return nil
}

func (T *qClient) scrClampScale(scale float32) float32 {

	f := float32(T.viddef.width) / 320.0
	if scale > f {
		scale = f
	}

	f = float32(T.viddef.height) / 240.0
	if scale > f {
		scale = f
	}

	if scale < 1 {
		scale = 1
	}

	return scale
}

func (T *qClient) scrGetDefaultScale() float32 {
	i := T.viddef.width / 640
	j := T.viddef.height / 240

	if i > j {
		i = j
	}
	if i < 1 {
		i = 1
	}

	return float32(i)
}

// func (T *qClient) scrDrawCrosshair(void)
// {
// 	float scale;

// 	if (!crosshair->value)
// 	{
// 		return;
// 	}

// 	if (crosshair->modified)
// 	{
// 		crosshair->modified = false;
// 		SCR_TouchPics();
// 	}

// 	if (!crosshair_pic[0])
// 	{
// 		return;
// 	}

// 	if (crosshair_scale->value < 0)
// 	{
// 		scale = SCR_GetDefaultScale();
// 	}
// 	else
// 	{
// 		scale = SCR_ClampScale(crosshair_scale->value);
// 	}

// 	Draw_PicScaled(scr_vrect.x + (scr_vrect.width - crosshair_width * scale) / 2,
// 			scr_vrect.y + (scr_vrect.height - crosshair_height * scale) / 2,
// 			crosshair_pic, scale);
// }

func (T *qClient) scrGetHUDScale() float32 {

	if !T.scr_initialized {
		return 1.0
	} else if T.r_hudscale.Float() < 0 {
		return T.scrGetDefaultScale()
	} else if T.r_hudscale.Float() == 0 { /* HACK: allow scale 0 to hide the HUD */
		return 0
	} else {
		return T.scrClampScale(T.r_hudscale.Float())
	}
}

func (T *qClient) scrGetConsoleScale() float32 {

	if !T.scr_initialized {
		return 1.0
	} else if T.r_consolescale.Float() < 0 {
		return T.scrGetDefaultScale()
	} else {
		return T.scrClampScale(T.r_consolescale.Float())
	}
}

func (T *qClient) scrGetMenuScale() float32 {

	if !T.scr_initialized {
		return 1.0
	} else if T.r_menuscale.Float() < 0 {
		return T.scrGetDefaultScale()
	} else {
		return T.scrClampScale(T.r_menuscale.Float())
	}
}
