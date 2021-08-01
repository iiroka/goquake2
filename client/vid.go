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
 * API between the client and renderers.
 *
 * =======================================================================
 */
package client

import (
	"goquake2/client/refresh/gl3"
	"goquake2/shared"
)

// Hold the video state.
type viddef_t struct {
	height int
	width  int
}

type vrect_t struct {
	x, y, width, height int
}

type vidExport struct {
	T *qClient
}

func (T *vidExport) Sys_Error(code int, format string, a ...interface{}) error {
	return T.T.common.Com_Error(code, format, a...)
}

func (T *vidExport) Com_VPrintf(print_level int, format string, a ...interface{}) {
	T.T.common.Com_VPrintf(print_level, format, a...)
}

func (T *vidExport) Cvar_Get(var_name, var_value string, flags int) *shared.CvarT {
	return T.T.common.Cvar_Get(var_name, var_value, flags)
}

func (T *vidExport) Cvar_Set(var_name, value string) *shared.CvarT {
	return T.T.common.Cvar_Set(var_name, value)
}

func (T *vidExport) GLimp_InitGraphics(fullscreen, width, height int) (int, int) {
	return T.T.glimpInitGraphics(fullscreen, width, height)
}

func (T *vidExport) LoadFile(path string) ([]byte, error) {
	return T.T.common.LoadFile(path)
}

// --------

// Video mode array
// ----------------

type vidmode_t struct {
	description string
	width       int
	height      int
	mode        int
}

// This must be the same as VID_MenuInit()->resolutions[] in videomenu.c!
var vid_modes = []vidmode_t{
	{"Mode  0:  320x240", 320, 240, 0},
	{"Mode  1:  400x300", 400, 300, 1},
	{"Mode  2:  512x384", 512, 384, 2},
	{"Mode  3:  640x400", 640, 400, 3},
	{"Mode  4:  640x480", 640, 480, 4},
	{"Mode  5:  800x500", 800, 500, 5},
	{"Mode  6:  800x600", 800, 600, 6},
	{"Mode  7:  960x720", 960, 720, 7},
	{"Mode  8: 1024x480", 1024, 480, 8},
	{"Mode  9: 1024x640", 1024, 640, 9},
	{"Mode 10: 1024x768", 1024, 768, 10},
	{"Mode 11: 1152x768", 1152, 768, 11},
	{"Mode 12: 1152x864", 1152, 864, 12},
	{"Mode 13: 1280x800", 1280, 800, 13},
	{"Mode 14: 1280x720", 1280, 720, 14},
	{"Mode 15: 1280x960", 1280, 960, 15},
	{"Mode 16: 1280x1024", 1280, 1024, 16},
	{"Mode 17: 1366x768", 1366, 768, 17},
	{"Mode 18: 1440x900", 1440, 900, 18},
	{"Mode 19: 1600x1200", 1600, 1200, 19},
	{"Mode 20: 1680x1050", 1680, 1050, 20},
	{"Mode 21: 1920x1080", 1920, 1080, 21},
	{"Mode 22: 1920x1200", 1920, 1200, 22},
	{"Mode 23: 2048x1536", 2048, 1536, 23},
	{"Mode 24: 2560x1080", 2560, 1080, 24},
	{"Mode 25: 2560x1440", 2560, 1440, 25},
	{"Mode 26: 2560x1600", 2560, 1600, 26},
	{"Mode 27: 3440x1440", 3440, 1440, 27},
	{"Mode 28: 3840x1600", 3840, 1600, 28},
	{"Mode 29: 3840x2160", 3840, 2160, 29},
	{"Mode 30: 4096x2160", 4096, 2160, 30},
	{"Mode 31: 5120x2880", 5120, 2880, 31},
}

// #define VID_NUM_MODES (sizeof(vid_modes) / sizeof(vid_modes[0]))

/*
 * Callback function for the 'vid_listmodes' cmd.
 */
// void
// VID_ListModes_f(void)
// {
// 	int i;

// 	Com_Printf("Supported video modes (r_mode):\n");

// 	for (i = 0; i < VID_NUM_MODES; ++i)
// 	{
// 		Com_Printf("  %s\n", vid_modes[i].description);
// 	}
// 	Com_Printf("  Mode -1: r_customwidth x r_customheight\n");
// }

/*
 * Returns informations about the given mode.
 */
func (T *vidExport) Vid_GetModeInfo(mode int) (int, int) {
	if (mode < 0) || (mode >= len(vid_modes)) {
		return -1, -1
	}

	return vid_modes[mode].width, vid_modes[mode].height
}

/*
 * Loads and initializes a renderer.
 */
func (T *qClient) vidLoadRenderer() bool {
	// 	 refimport_t	ri;
	// 	 GetRefAPI_t	GetRefAPI;

	//  #ifdef __APPLE__
	// 	 const char* lib_ext = "dylib";
	//  #elif defined(_WIN32)
	// 	 const char* lib_ext = "dll";
	//  #else
	// 	 const char* lib_ext = "so";
	//  #endif

	// 	 char reflib_name[64] = {0};
	// 	 char reflib_path[MAX_OSPATH] = {0};

	// 	 // If the refresher is already active we need
	// 	 // to shut it down before loading a new one
	// 	 VID_ShutdownRenderer();

	// Log what we're doing.
	T.common.Com_Printf("----- refresher initialization -----\n")

	reflib_name := T.vid_renderer.String
	T.common.Com_Printf("Loading library: %v\n", reflib_name)

	// 	 // Fill in the struct exported to the renderer.
	// 	 // FIXME: Do we really need all these?
	// 	 ri.Cmd_AddCommand = Cmd_AddCommand;
	// 	 ri.Cmd_Argc = Cmd_Argc;
	// 	 ri.Cmd_Argv = Cmd_Argv;
	// 	 ri.Cmd_ExecuteText = Cbuf_ExecuteText;
	// 	 ri.Cmd_RemoveCommand = Cmd_RemoveCommand;
	// 	 ri.Com_VPrintf = Com_VPrintf;
	// 	 ri.Cvar_Get = Cvar_Get;
	// 	 ri.Cvar_Set = Cvar_Set;
	// 	 ri.Cvar_SetValue = Cvar_SetValue;
	// 	 ri.FS_FreeFile = FS_FreeFile;
	// 	 ri.FS_Gamedir = FS_Gamedir;
	// 	 ri.FS_LoadFile = FS_LoadFile;
	// 	 ri.GLimp_InitGraphics = GLimp_InitGraphics;
	// 	 ri.GLimp_GetDesktopMode = GLimp_GetDesktopMode;
	// 	 ri.Sys_Error = Com_Error;
	// 	 ri.Vid_GetModeInfo = VID_GetModeInfo;
	// 	 ri.Vid_MenuInit = VID_MenuInit;
	// 	 ri.Vid_WriteScreenshot = VID_WriteScreenshot;

	// Mkay, let's load the requested renderer.
	if reflib_name == "gl3" {
		T.re = gl3.QGl3Create(T.ri)
	}

	// Okay, we couldn't load it. It's up to the
	// caller to recover from this.
	if T.re == nil {
		T.common.Com_Printf("Loading %v as renderer lib failed!", reflib_name)
		return false
	}

	// Everything seems okay, initialize it.
	if !T.re.Init() {
		//  VID_ShutdownRenderer();

		T.common.Com_Printf("ERROR: Loading %v as rendering backend failed.\n", reflib_name)
		T.common.Com_Printf("------------------------------------\n\n")

		return false
	}

	// 	 /* Ensure that all key states are cleared */
	// 	 Key_MarkAllUp();

	T.common.Com_Printf("Successfully loaded %v as rendering backend.\n", reflib_name)
	T.common.Com_Printf("------------------------------------\n\n")

	return true
}

/*
 * Checks if a renderer changes was requested and executes it.
 * Inclusive fallback through all renderers. :)
 */
func (T *qClient) vidCheckChanges() error {
	// FIXME: Not with vid_fullscreen, should be a dedicated variable.
	// Sounds easy but this vid_fullscreen hack is really messy and
	// interacts with several critical places in both the client and
	// the renderers...
	if T.vid_fullscreen.Modified {
		// Stop sound, because the clients blocks while
		// we're reloading the renderer. The sound system
		// would screw up it's internal timings.
		//  S_StopAllSounds();

		// Reset the client side of the renderer state.
		//  cl.refresh_prepped = false;
		//  cl.cinematicpalette_active = false;

		// More or less blocks the client.
		//  cls.disable_screen = true;

		// Mkay, let's try our luck.
		for !T.vidLoadRenderer() {
			// We try: vk -> gl3 -> gl1 -> soft.
			if T.vid_renderer.String == "vk" {
				T.common.Com_Printf("Retrying with gl3...\n")
				T.common.Cvar_Set("vid_renderer", "gl3")
			} else if T.vid_renderer.String == "gl3" {
				T.common.Com_Printf("Retrying with gl1...\n")
				T.common.Cvar_Set("vid_renderer", "gl1")
			} else if T.vid_renderer.String == "gl1" {
				T.common.Com_Printf("Retrying with soft...\n")
				T.common.Cvar_Set("vid_renderer", "soft")
			} else if T.vid_renderer.String == "soft" {
				// Sorry, no usable renderer found.
				return T.common.Com_Error(shared.ERR_FATAL, "No usable renderer found!\n")
			} else {
				// User forced something stupid.
				T.common.Com_Printf("Retrying with gl3...\n")
				T.common.Cvar_Set("vid_renderer", "gl3")
			}
		}

		// Ignore possible changes in vid_renderer above.
		T.vid_renderer.Modified = false

		// Unblock the client.
		//  cls.disable_screen = false;
	}

	//  if (vid_renderer->modified) {
	// 	 vid_renderer->modified = false;
	// 	 cl.refresh_prepped = false;
	//  }
	return nil
}

/*
 * Initializes the video stuff.
 */
func (T *qClient) vidInit() error {
	T.ri = &vidExport{T}

	// Console variables
	T.vid_gamma = T.common.Cvar_Get("vid_gamma", "1.0", shared.CVAR_ARCHIVE)
	T.vid_fullscreen = T.common.Cvar_Get("vid_fullscreen", "0", shared.CVAR_ARCHIVE)
	T.vid_renderer = T.common.Cvar_Get("vid_renderer", "gl3", shared.CVAR_ARCHIVE)

	// Commands
	// Cmd_AddCommand("vid_restart", VID_Restart_f)
	// Cmd_AddCommand("vid_listmodes", VID_ListModes_f)

	// Initializes the video backend. This is NOT the renderer
	// itself, just the client side support stuff!
	if !T.glimpInit() {
		return T.common.Com_Error(shared.ERR_FATAL, "Couldn't initialize the graphics subsystem!\n")
	}

	// Load the renderer and get things going.
	return T.vidCheckChanges()
}

func (T *qClient) R_BeginRegistration(name string) error {
	if T.re != nil {
		return T.re.BeginRegistration(name)
	}
	return nil
}

func (T *qClient) R_RegisterModel(name string) (interface{}, error) {
	if T.re != nil {
		return T.re.RegisterModel(name)
	}

	return nil, nil
}

func (T *qClient) R_RegisterSkin(name string) interface{} {
	if T.re != nil {
		return T.re.RegisterSkin(name)
	}

	return nil
}

func (T *qClient) R_RenderFrame(fd shared.Refdef_t) error {
	if T.re != nil {
		return T.re.RenderFrame(fd)
	}
	return nil
}

func (T *qClient) R_BeginFrame(camera_separation float32) error {
	if T.re != nil {
		return T.re.BeginFrame(camera_separation)
	}
	return nil
}

func (T *qClient) R_EndFrame() {
	if T.re != nil {
		T.re.EndFrame()
	}
}

func (T *qClient) Draw_TileClear(x, y, w, h int, name string) {
	if T.re != nil {
		T.re.DrawTileClear(x, y, w, h, name)
	}
}

func (T *qClient) Draw_Fill(x, y, w, h, c int) {
	if T.re != nil {
		T.re.DrawFill(x, y, w, h, c)
	}
}

func (T *qClient) Draw_StretchPic(x, y, w, h int, name string) {
	if T.re != nil {
		T.re.DrawStretchPic(x, y, w, h, name)
	}
}

func (T *qClient) Draw_PicScaled(x, y int, pic string, factor float32) {
	if T.re != nil {
		T.re.DrawPicScaled(x, y, pic, factor)
	}
}

func (T *qClient) R_SetSky(name string, rotate float32, axis []float32) {
	if T.re != nil {
		T.re.SetSky(name, rotate, axis)
	}
}

func (T *qClient) IsVSyncActive() bool {
	if T.re != nil {
		return T.re.IsVSyncActive()
	}

	return false
}
