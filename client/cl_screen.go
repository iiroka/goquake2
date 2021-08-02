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

import (
	"fmt"
	"goquake2/shared"
	"strconv"
)

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
		T.Draw_Fill(0, T.viddef.height/2, T.viddef.width, T.viddef.height/2, 0)
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

func (T *qClient) scrAddDirtyPoint(x, y int) {
	if x < T.scr_dirty.x1 {
		T.scr_dirty.x1 = x
	}

	if x > T.scr_dirty.x2 {
		T.scr_dirty.x2 = x
	}

	if y < T.scr_dirty.y1 {
		T.scr_dirty.y1 = y
	}

	if y > T.scr_dirty.y2 {
		T.scr_dirty.y2 = y
	}
}

func (T *qClient) scrDirtyScreen() {
	T.scrAddDirtyPoint(0, 0)
	T.scrAddDirtyPoint(T.viddef.width-1, T.viddef.height-1)
}

/*
 * Clear any parts of the tiled background that were drawn on last frame
 */
func (T *qClient) scrTileClear() {
	// int i;
	// int top, bottom, left, right;
	// dirty_t clear;

	if T.scr_con_current == 1.0 {
		return /* full screen console */
	}

	if T.scr_viewsize.Int() == 100 {
		return /* full screen rendering */
	}

	// if T.cl.cinematictime > 0 {
	// 	return /* full screen cinematic */
	// }

	/* erase rect will be the union of the past three
	   frames so tripple buffering works properly */
	clear := dirty_t{}
	clear = T.scr_dirty

	for i := 0; i < 2; i++ {
		if T.scr_old_dirty[i].x1 < clear.x1 {
			clear.x1 = T.scr_old_dirty[i].x1
		}

		if T.scr_old_dirty[i].x2 > clear.x2 {
			clear.x2 = T.scr_old_dirty[i].x2
		}

		if T.scr_old_dirty[i].y1 < clear.y1 {
			clear.y1 = T.scr_old_dirty[i].y1
		}

		if T.scr_old_dirty[i].y2 > clear.y2 {
			clear.y2 = T.scr_old_dirty[i].y2
		}
	}

	T.scr_old_dirty[1] = T.scr_old_dirty[0]
	T.scr_old_dirty[0] = T.scr_dirty

	T.scr_dirty.x1 = 9999
	T.scr_dirty.x2 = -9999
	T.scr_dirty.y1 = 9999
	T.scr_dirty.y2 = -9999

	/* don't bother with anything convered by the console */
	top := int(T.scr_con_current * float32(T.viddef.height))

	if top >= clear.y1 {
		clear.y1 = top
	}

	if clear.y2 <= clear.y1 {
		return /* nothing disturbed */
	}

	top = T.scr_vrect.y
	bottom := top + T.scr_vrect.height - 1
	left := T.scr_vrect.x
	right := left + T.scr_vrect.width - 1

	if clear.y1 < top {
		/* clear above view screen */
		i := top - 1
		if clear.y2 < top-1 {
			i = clear.y2
		}
		T.Draw_TileClear(clear.x1, clear.y1,
			clear.x2-clear.x1+1, i-clear.y1+1, "backtile")
		clear.y1 = top
	}

	if clear.y2 > bottom {
		/* clear below view screen */
		i := bottom + 1
		if clear.y1 > bottom+1 {
			i = clear.y1
		}
		T.Draw_TileClear(clear.x1, i,
			clear.x2-clear.x1+1, clear.y2-i+1, "backtile")
		clear.y2 = bottom
	}

	if clear.x1 < left {
		/* clear left of view screen */
		i := left - 1
		if clear.x2 < left-1 {
			i = clear.x2
		}
		T.Draw_TileClear(clear.x1, clear.y1,
			i-clear.x1+1, clear.y2-clear.y1+1, "backtile")
		clear.x1 = left
	}

	if clear.x2 > right {
		/* clear left of view screen */
		i := right + 1
		if clear.x1 > right+1 {
			i = clear.x1
		}
		T.Draw_TileClear(i, clear.y1,
			clear.x2-i+1, clear.y2-clear.y1+1, "backtile")
		clear.x2 = right
	}
}

const STAT_MINUS = 10

var sb_nums = [2][11]string{
	{
		"num_0", "num_1", "num_2", "num_3", "num_4", "num_5",
		"num_6", "num_7", "num_8", "num_9", "num_minus",
	},
	{
		"anum_0", "anum_1", "anum_2", "anum_3", "anum_4", "anum_5",
		"anum_6", "anum_7", "anum_8", "anum_9", "anum_minus",
	},
}

const ICON_WIDTH = 24
const ICON_HEIGHT = 24
const CHAR_WIDTH = 16
const ICON_SPACE = 8

func (T *qClient) scrDrawFieldScaled(x, y, color, width, value int, factor float32) {
	// char num[16], *ptr;
	// int l;
	// int frame;

	if width < 1 {
		return
	}

	/* draw number string */
	if width > 5 {
		width = 5
	}

	T.scrAddDirtyPoint(x, y)
	T.scrAddDirtyPoint(x+int(float32(width*CHAR_WIDTH+2)*factor), y+int(factor*24))

	num := fmt.Sprintf("%v", value)
	l := len(num)

	if l > width {
		l = width
	}

	x += int(float32(2+CHAR_WIDTH*(width-l)) * factor)

	for index := 0; index < l; index++ {
		var frame int
		if num[index] == '-' {
			frame = STAT_MINUS
		} else {
			frame = int(num[index] - '0')
		}

		T.Draw_PicScaled(x, y, sb_nums[color][frame], factor)
		x += int(CHAR_WIDTH * factor)
	}
}

func (T *qClient) scrDrawField(x, y, color, width, value int) {
	T.scrDrawFieldScaled(x, y, color, width, value, 1.0)
}

func (T *qClient) scrExecuteLayoutString(s string) {
	// int x, y;
	// int value;
	// char *token;
	// int width;
	// int index;
	// clientinfo_t *ci;

	scale := T.scrGetHUDScale()

	if (T.cls.state != ca_active) || !T.cl.refresh_prepped {
		return
	}

	if len(s) == 0 {
		return
	}

	x := 0
	y := 0

	index := 0
	for index >= 0 && index < len(s) {
		token, indx := shared.COM_Parse(s, index)
		if indx < 0 {
			break
		}
		index = indx

		if token == "xl" {
			token, index = shared.COM_Parse(s, index)
			xi, _ := strconv.ParseInt(token, 10, 32)
			x = int(scale * float32(xi))
			continue
		}

		if token == "xr" {
			token, index = shared.COM_Parse(s, index)
			xi, _ := strconv.ParseInt(token, 10, 32)
			x = T.viddef.width + int(scale*float32(xi))
			continue
		}

		if token == "xv" {
			token, index = shared.COM_Parse(s, index)
			xi, _ := strconv.ParseInt(token, 10, 32)
			x = T.viddef.width/2 - int(scale*160) + int(scale*float32(xi))
			continue
		}

		if token == "yt" {
			token, index = shared.COM_Parse(s, index)
			yi, _ := strconv.ParseInt(token, 10, 32)
			y = int(scale * float32(yi))
			continue
		}

		if token == "yb" {
			token, index = shared.COM_Parse(s, index)
			yi, _ := strconv.ParseInt(token, 10, 32)
			y = T.viddef.height + int(scale*float32(yi))
			continue
		}

		if token == "yv" {
			token, index = shared.COM_Parse(s, index)
			yi, _ := strconv.ParseInt(token, 10, 32)
			y = T.viddef.height/2 - int(scale*120) + int(scale*float32(yi))
			continue
		}

		if token == "pic" {
			/* draw a pic from a stat number */
			token, index = shared.COM_Parse(s, index)
			idx, _ := strconv.ParseInt(token, 10, 32)

			// 		if ((idx < 0) || (idx >= sizeof(cl.frame.playerstate.stats))) {
			// 			Com_Error(ERR_DROP, "bad stats index %d (0x%x)", index, index);
			// 		}

			value := T.cl.frame.playerstate.Stats[idx]

			// 		if (value >= MAX_IMAGES) {
			// 			Com_Error(ERR_DROP, "Pic >= MAX_IMAGES");
			// 		}

			if len(T.cl.configstrings[shared.CS_IMAGES+value]) > 0 {
				T.scrAddDirtyPoint(x, y)
				T.scrAddDirtyPoint(x+int(23*scale), y+int(23*scale))
				T.Draw_PicScaled(x, y, T.cl.configstrings[shared.CS_IMAGES+value], scale)
			}

			continue
		}

		// 	if (!strcmp(token, "client"))
		// 	{
		// 		/* draw a deathmatch client block */
		// 		int score, ping, time;

		// 		token = COM_Parse(&s);
		// 		x = viddef.width / 2 - scale*160 + scale*(int)strtol(token, (char **)NULL, 10);
		// 		token = COM_Parse(&s);
		// 		y = viddef.height / 2 - scale*120 + scale*(int)strtol(token, (char **)NULL, 10);
		// 		SCR_AddDirtyPoint(x, y);
		// 		SCR_AddDirtyPoint(x + scale*159, y + scale*31);

		// 		token = COM_Parse(&s);
		// 		value = (int)strtol(token, (char **)NULL, 10);

		// 		if ((value >= MAX_CLIENTS) || (value < 0))
		// 		{
		// 			Com_Error(ERR_DROP, "client >= MAX_CLIENTS");
		// 		}

		// 		ci = &cl.clientinfo[value];

		// 		token = COM_Parse(&s);
		// 		score = (int)strtol(token, (char **)NULL, 10);

		// 		token = COM_Parse(&s);
		// 		ping = (int)strtol(token, (char **)NULL, 10);

		// 		token = COM_Parse(&s);
		// 		time = (int)strtol(token, (char **)NULL, 10);

		// 		DrawAltStringScaled(x + scale*32, y, ci->name, scale);
		// 		DrawAltStringScaled(x + scale*32, y + scale*8, "Score: ", scale);
		// 		DrawAltStringScaled(x + scale*(32 + 7 * 8), y + scale*8, va("%i", score), scale);
		// 		DrawStringScaled(x + scale*32, y + scale*16, va("Ping:  %i", ping), scale);
		// 		DrawStringScaled(x + scale*32, y + scale*24, va("Time:  %i", time), scale);

		// 		if (!ci->icon)
		// 		{
		// 			ci = &cl.baseclientinfo;
		// 		}

		// 		Draw_PicScaled(x, y, ci->iconname, scale);
		// 		continue;
		// 	}

		// 	if (!strcmp(token, "ctf"))
		// 	{
		// 		/* draw a ctf client block */
		// 		int score, ping;
		// 		char block[80];

		// 		token = COM_Parse(&s);
		// 		x = viddef.width / 2 - scale*160 + scale*(int)strtol(token, (char **)NULL, 10);
		// 		token = COM_Parse(&s);
		// 		y = viddef.height / 2 - scale*120 + scale*(int)strtol(token, (char **)NULL, 10);
		// 		SCR_AddDirtyPoint(x, y);
		// 		SCR_AddDirtyPoint(x + scale*159, y + scale*31);

		// 		token = COM_Parse(&s);
		// 		value = (int)strtol(token, (char **)NULL, 10);

		// 		if ((value >= MAX_CLIENTS) || (value < 0))
		// 		{
		// 			Com_Error(ERR_DROP, "client >= MAX_CLIENTS");
		// 		}

		// 		ci = &cl.clientinfo[value];

		// 		token = COM_Parse(&s);
		// 		score = (int)strtol(token, (char **)NULL, 10);

		// 		token = COM_Parse(&s);
		// 		ping = (int)strtol(token, (char **)NULL, 10);

		// 		if (ping > 999)
		// 		{
		// 			ping = 999;
		// 		}

		// 		sprintf(block, "%3d %3d %-12.12s", score, ping, ci->name);

		// 		if (value == cl.playernum)
		// 		{
		// 			DrawAltStringScaled(x, y, block, scale);
		// 		}

		// 		else
		// 		{
		// 			DrawStringScaled(x, y, block, scale);
		// 		}

		// 		continue;
		// 	}

		// 	if (!strcmp(token, "picn"))
		// 	{
		// 		/* draw a pic from a name */
		// 		token = COM_Parse(&s);
		// 		SCR_AddDirtyPoint(x, y);
		// 		SCR_AddDirtyPoint(x + scale*23, y + scale*23);
		// 		Draw_PicScaled(x, y, (char *)token, scale);
		// 		continue;
		// 	}

		// 	if (!strcmp(token, "num"))
		// 	{
		// 		/* draw a number */
		// 		token = COM_Parse(&s);
		// 		width = (int)strtol(token, (char **)NULL, 10);
		// 		token = COM_Parse(&s);
		// 		value = cl.frame.playerstate.stats[(int)strtol(token, (char **)NULL, 10)];
		// 		SCR_DrawFieldScaled(x, y, 0, width, value, scale);
		// 		continue;
		// 	}

		if token == "hnum" {
			/* health number */
			value := T.cl.frame.playerstate.Stats[shared.STAT_HEALTH]

			var color int
			if value > 25 {
				color = 0 /* green */
			} else if value > 0 {
				color = (T.cl.frame.serverframe >> 2) & 1 /* flash */
			} else {
				color = 1
			}

			if (T.cl.frame.playerstate.Stats[shared.STAT_FLASHES] & 1) != 0 {
				T.Draw_PicScaled(x, y, "field_3", scale)
			}

			T.scrDrawFieldScaled(x, y, color, 3, int(value), scale)
			continue
		}

		if token == "anum" {
			/* ammo number */
			value := T.cl.frame.playerstate.Stats[shared.STAT_AMMO]

			var color int
			if value > 5 {
				color = 0 /* green */
			} else if value >= 0 {
				color = (T.cl.frame.serverframe >> 2) & 1 /* flash */
			} else {
				continue /* negative number = don't show */
			}

			if (T.cl.frame.playerstate.Stats[shared.STAT_FLASHES] & 4) != 0 {
				T.Draw_PicScaled(x, y, "field_3", scale)
			}

			T.scrDrawFieldScaled(x, y, color, 3, int(value), scale)
			continue
		}

		// 	if (!strcmp(token, "rnum"))
		// 	{
		// 		/* armor number */
		// 		int color;

		// 		width = 3;
		// 		value = cl.frame.playerstate.stats[STAT_ARMOR];

		// 		if (value < 1)
		// 		{
		// 			continue;
		// 		}

		// 		color = 0; /* green */

		// 		if (cl.frame.playerstate.stats[STAT_FLASHES] & 2)
		// 		{
		// 			Draw_PicScaled(x, y, "field_3", scale);
		// 		}

		// 		SCR_DrawFieldScaled(x, y, color, width, value, scale);
		// 		continue;
		// 	}

		// 	if (!strcmp(token, "stat_string"))
		// 	{
		// 		token = COM_Parse(&s);
		// 		index = (int)strtol(token, (char **)NULL, 10);

		// 		if ((index < 0) || (index >= MAX_CONFIGSTRINGS))
		// 		{
		// 			Com_Error(ERR_DROP, "Bad stat_string index");
		// 		}

		// 		index = cl.frame.playerstate.stats[index];

		// 		if ((index < 0) || (index >= MAX_CONFIGSTRINGS))
		// 		{
		// 			Com_Error(ERR_DROP, "Bad stat_string index");
		// 		}

		// 		DrawStringScaled(x, y, cl.configstrings[index], scale);
		// 		continue;
		// 	}

		// 	if (!strcmp(token, "cstring"))
		// 	{
		// 		token = COM_Parse(&s);
		// 		DrawHUDStringScaled(token, x, y, 320, 0, scale); // FIXME: or scale 320 here?
		// 		continue;
		// 	}

		// 	if (!strcmp(token, "string"))
		// 	{
		// 		token = COM_Parse(&s);
		// 		DrawStringScaled(x, y, token, scale);
		// 		continue;
		// 	}

		// 	if (!strcmp(token, "cstring2"))
		// 	{
		// 		token = COM_Parse(&s);
		// 		DrawHUDStringScaled(token, x, y, 320, 0x80, scale); // FIXME: or scale 320 here?
		// 		continue;
		// 	}

		// 	if (!strcmp(token, "string2"))
		// 	{
		// 		token = COM_Parse(&s);
		// 		DrawAltStringScaled(x, y, token, scale);
		// 		continue;
		// 	}

		if token == "if" {
			token, index = shared.COM_Parse(s, index)
			idx, _ := strconv.ParseInt(token, 10, 32)
			value := T.cl.frame.playerstate.Stats[idx]

			if value == 0 {
				/* skip to endif */
				for index > 0 && index < len(s) && token != "endif" {
					token, index = shared.COM_Parse(s, index)
				}
			}

			continue
		}

		if token == "endif" {
			continue
		}

		println("Token", token)
	}
}

/*
 * The status bar is a small layout program that
 * is based on the stats array
 */
func (T *qClient) scrDrawStats() {
	T.scrExecuteLayoutString(T.cl.configstrings[shared.CS_STATUSBAR])
}

const STAT_LAYOUTS = 13

func (T *qClient) scrDrawLayout() {
	if T.cl.frame.playerstate.Stats[STAT_LAYOUTS] == 0 {
		return
	}

	T.scrExecuteLayoutString(T.cl.layout)
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

		/* clear any dirty part of the background */
		T.scrTileClear()

		if err := T.renderView(separation[i]); err != nil {
			return err
		}

		T.scrDrawStats()

		if (T.cl.frame.playerstate.Stats[shared.STAT_LAYOUTS] & 1) != 0 {
			T.scrDrawLayout()
		}

		if (T.cl.frame.playerstate.Stats[shared.STAT_LAYOUTS] & 2) != 0 {
			// 			 CL_DrawInventory();
		}

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

		T.mDraw()

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
