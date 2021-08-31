/*
 * Copyright (C) 2010 Yamagi Burmeister
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
 * This is the client side of the render backend, implemented trough SDL.
 * The SDL window and related functrion (mouse grap, fullscreen switch)
 * are implemented here, everything else is in the renderers.
 *
 * =======================================================================
 */
package client

import (
	"goquake2/shared"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

var last_position_x = int32(sdl.WINDOWPOS_UNDEFINED)
var last_position_y = int32(sdl.WINDOWPOS_UNDEFINED)

/*
 * Initializes the SDL video subsystem. Must
 * be called before anything else.
 */
func (T *qClient) glimpInit() bool {
	T.vid_displayrefreshrate = T.common.Cvar_Get("vid_displayrefreshrate", "-1", shared.CVAR_ARCHIVE)
	T.vid_displayindex = T.common.Cvar_Get("vid_displayindex", "0", shared.CVAR_ARCHIVE)
	T.vid_rate = T.common.Cvar_Get("vid_rate", "-1", shared.CVAR_ARCHIVE)

	if sdl.WasInit(sdl.INIT_VIDEO) == 0 {
		if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
			T.common.Com_Printf("Couldn't init SDL video: %v.\n", err.Error())

			return false
		}

		var version sdl.Version

		sdl.GetVersion(&version)
		T.common.Com_Printf("-------- vid initialization --------\n")
		T.common.Com_Printf("SDL version is: %v.%v.%v\n", int(version.Major), int(version.Minor), int(version.Patch))
		drvr, _ := sdl.GetCurrentVideoDriver()
		T.common.Com_Printf("SDL video driver is \"%v\".\n", drvr)

		T.num_displays, _ = sdl.GetNumVideoDisplays()
		// InitDisplayIndices()
		// ClampDisplayIndexCvar()
		T.common.Com_Printf("SDL display modes:\n")

		T.printDisplayModes()
		T.common.Com_Printf("------------------------------------\n\n")
	}

	return true
}

/*
 * Lists all available display modes.
 */
func (T *qClient) printDisplayModes() {
	//  curdisplay := window ? SDL_GetWindowDisplayIndex(window) : 0;
	curdisplay := 0

	// On X11 (at least for me)
	// curdisplay is always -1.
	// DG: probably because window was NULL?
	if curdisplay < 0 {
		curdisplay = 0
	}

	nummodes, err := sdl.GetNumDisplayModes(curdisplay)
	if err != nil {
		T.common.Com_Printf("Can't get display modes: %v\n", err.Error())
		return
	}

	for i := 0; i < nummodes; i++ {
		mode, err := sdl.GetDisplayMode(curdisplay, i)
		if err != nil {
			T.common.Com_Printf("Can't get display mode: %v\n", err.Error())
			return
		}

		T.common.Com_Printf(" - Mode %2v: %vx%v@%v\n", i, mode.W, mode.H, mode.RefreshRate)
	}
}

/*
 * (Re)initializes the actual window.
 */
func (T *qClient) glimpInitGraphics(fullscreen, width, height int) (int, int) {

	// int flags;
	// int curWidth, curHeight;
	// int width = *pwidth;
	// int height = *pheight;
	fs_flag := 0

	if fullscreen == 1 {
		fs_flag = sdl.WINDOW_FULLSCREEN
	} else if fullscreen == 2 {
		fs_flag = sdl.WINDOW_FULLSCREEN_DESKTOP
	}

	/* Only do this if we already have a working window and a fully
	initialized rendering backend GLimp_InitGraphics() is also
	called when recovering if creating GL context fails or the
	one we got is unusable. */
	// if (T.glimp_initSuccessful && GetWindowSize(&curWidth, &curHeight)
	// 		&& (curWidth == width) && (curHeight == height))
	// {
	// 	/* If we want fullscreen, but aren't */
	// 	if (GetFullscreenType())
	// 	{
	// 		SDL_SetWindowFullscreen(window, fs_flag);
	// 		Cvar_SetValue("vid_fullscreen", fullscreen);
	// 	}

	// 	/* Are we now? */
	// 	if (GetFullscreenType())
	// 	{
	// 		return true;
	// 	}
	// }

	/* Is the surface used? */
	if T.window != nil {
		// 	re.ShutdownContext();
		// 	ShutdownGraphics();

		T.window = nil
	}

	/* We need the window size for the menu, the HUD, etc. */
	T.viddef.width = width
	T.viddef.height = height

	if T.glimp_last_flags != -1 && (T.glimp_last_flags&sdl.WINDOW_OPENGL) != 0 {
		/* Reset SDL. */
		// 	SDL_GL_ResetAttributes();
	}

	/* Let renderer prepare things (set OpenGL attributes).
	   FIXME: This is no longer necessary, the renderer
	   could and should pass the flags when calling this
	   function. */
	flags := T.re.PrepareForWindow()

	if flags == -1 {
		/* It's PrepareForWindow() job to log an error */
		return -1, -1
	}

	if fs_flag != 0 {
		flags |= fs_flag
	}

	/* Mkay, now the hard work. Let's create the window. */
	// cvar_t *gl_msaa_samples = Cvar_Get("r_msaa_samples", "0", CVAR_ARCHIVE);

	for {
		if !T.createSDLWindow(flags, width, height) {
			// 		if((flags & SDL_WINDOW_OPENGL) && gl_msaa_samples->value)
			// 		{
			// 			Com_Printf("SDL SetVideoMode failed: %s\n", SDL_GetError());
			// 			Com_Printf("Reverting to %s r_mode %i (%ix%i) without MSAA.\n",
			// 				        (flags & fs_flag) ? "fullscreen" : "windowed",
			// 				        (int) Cvar_VariableValue("r_mode"), width, height);

			// 			/* Try to recover */
			// 			Cvar_SetValue("r_msaa_samples", 0);

			// 			SDL_GL_SetAttribute(SDL_GL_MULTISAMPLEBUFFERS, 0);
			// 			SDL_GL_SetAttribute(SDL_GL_MULTISAMPLESAMPLES, 0);
			// 		}
			// 		else if (width != 640 || height != 480 || (flags & fs_flag))
			// 		{
			// 			Com_Printf("SDL SetVideoMode failed: %s\n", SDL_GetError());
			// 			Com_Printf("Reverting to windowed r_mode 4 (640x480).\n");

			// 			/* Try to recover */
			// 			Cvar_SetValue("r_mode", 4);
			// 			Cvar_SetValue("vid_fullscreen", 0);
			// 			Cvar_SetValue("vid_rate", -1);

			// 			fullscreen = 0;
			// 			*pwidth = width = 640;
			// 			*pheight = height = 480;
			// 			flags &= ~fs_flag;
			// 		}
			// 		else
			// 		{
			// 			Com_Error(ERR_FATAL, "Failed to revert to r_mode 4. Exiting...\n");
			// 			return false;
			// 		}
		} else {
			break
		}
	}

	T.glimp_last_flags = flags

	/* Now that we've got a working window print it's mode. */
	curdisplay, err := T.window.GetDisplayIndex()
	if err != nil {
		curdisplay = 0
	}

	// SDL_DisplayMode mode;

	if mode, err := sdl.GetCurrentDisplayMode(curdisplay); err != nil {
		T.common.Com_Printf("Can't get current display mode: %s\n", err.Error())
	} else {
		T.common.Com_Printf("Real display mode: %vx%v@%v\n", mode.W, mode.H, mode.RefreshRate)
	}

	/* Initialize rendering context. */
	if !T.re.InitContext(T.window) {
		/* InitContext() should have logged an error. */
		return -1, -1
	}

	// /* Set the window icon - For SDL2, this must be done after creating the window */
	// SetSDLIcon();

	// /* No cursor */
	// SDL_ShowCursor(0);

	T.glimp_initSuccessful = true

	return width, height
}

func (T *qClient) createSDLWindow(flags, w, h int) bool {
	// if (SDL_WINDOWPOS_ISUNDEFINED(last_position_x) || SDL_WINDOWPOS_ISUNDEFINED(last_position_y) || last_position_x < 0 ||last_position_y < 24)
	// {
	// 	last_position_x = last_position_y = SDL_WINDOWPOS_UNDEFINED_DISPLAY((int)vid_displayindex->value);
	// }

	// /* Force the window to minimize when focus is lost. This was the
	//  * default behavior until SDL 2.0.12 and changed with 2.0.14.
	//  * The windows staying maximized has some odd implications for
	//  * window ordering under Windows and some X11 window managers
	//  * like kwin. See:
	//  *  * https://github.com/libsdl-org/SDL/issues/4039
	//  *  * https://github.com/libsdl-org/SDL/issues/3656 */
	// SDL_SetHint(SDL_HINT_VIDEO_MINIMIZE_ON_FOCUS_LOSS, "1");

	wnd, err := sdl.CreateWindow("Yamagi Quake II", last_position_x, last_position_y, int32(w), int32(h), uint32(flags))

	if wnd != nil && err == nil {

		T.window = wnd

		/* save current display as default */
		T.last_display, _ = T.window.GetDisplayIndex()
		last_position_x, last_position_y = T.window.GetPosition()

		// 	/* Check if we're really in the requested diplay mode. There is
		// 	   (or was) an SDL bug were SDL switched into the wrong mode
		// 	   without giving an error code. See the bug report for details:
		// 	   https://bugzilla.libsdl.org/show_bug.cgi?id=4700 */
		// 	SDL_DisplayMode real_mode;

		// 	if ((flags & (SDL_WINDOW_FULLSCREEN | SDL_WINDOW_FULLSCREEN_DESKTOP)) == SDL_WINDOW_FULLSCREEN)
		// 	{
		// 		if (SDL_GetWindowDisplayMode(window, &real_mode) != 0)
		// 		{
		// 			SDL_DestroyWindow(window);
		// 			window = NULL;

		// 			Com_Printf("Can't get display mode: %s\n", SDL_GetError());

		// 			return false;
		// 		}
		// 	}

		// 	/* SDL_WINDOW_FULLSCREEN_DESKTOP implies SDL_WINDOW_FULLSCREEN! */
		// 	if (((flags & (SDL_WINDOW_FULLSCREEN | SDL_WINDOW_FULLSCREEN_DESKTOP)) == SDL_WINDOW_FULLSCREEN)
		// 			&& ((real_mode.w != w) || (real_mode.h != h)))
		// 	{

		// 		Com_Printf("Current display mode isn't requested display mode\n");
		// 		Com_Printf("Likely SDL bug #4700, trying to work around it\n");

		// 		/* Mkay, try to hack around that. */
		// 		SDL_DisplayMode wanted_mode = {};

		// 		wanted_mode.w = w;
		// 		wanted_mode.h = h;

		// 		if (SDL_SetWindowDisplayMode(window, &wanted_mode) != 0)
		// 		{
		// 			SDL_DestroyWindow(window);
		// 			window = NULL;

		// 			Com_Printf("Can't force resolution to %ix%i: %s\n", w, h, SDL_GetError());

		// 			return false;
		// 		}

		// 		/* The SDL doku says, that SDL_SetWindowSize() shouldn't be
		// 		   used on fullscreen windows. But at least in my test with
		// 		   SDL 2.0.9 the subsequent SDL_GetWindowDisplayMode() fails
		// 		   if I don't call it. */
		// 		SDL_SetWindowSize(window, wanted_mode.w, wanted_mode.h);

		// 		if (SDL_GetWindowDisplayMode(window, &real_mode) != 0)
		// 		{
		// 			SDL_DestroyWindow(window);
		// 			window = NULL;

		// 			Com_Printf("Can't get display mode: %s\n", SDL_GetError());

		// 			return false;
		// 		}

		// 		if ((real_mode.w != w) || (real_mode.h != h))
		// 		{
		// 			SDL_DestroyWindow(window);
		// 			window = NULL;

		// 			Com_Printf("Can't get display mode: %s\n", SDL_GetError());

		// 			return false;
		// 		}
		// 	}

		// 	/* Normally SDL stays at desktop refresh rate or chooses something
		// 	   sane. Some player may want to override that.

		// 	   Reminder: SDL_WINDOW_FULLSCREEN_DESKTOP implies SDL_WINDOW_FULLSCREEN! */
		// 	if ((flags & (SDL_WINDOW_FULLSCREEN | SDL_WINDOW_FULLSCREEN_DESKTOP)) == SDL_WINDOW_FULLSCREEN)
		// 	{
		// 		if (vid_rate->value > 0)
		// 		{
		// 			SDL_DisplayMode closest_mode;
		// 			SDL_DisplayMode requested_mode = real_mode;

		// 			requested_mode.refresh_rate = (int)vid_rate->value;

		// 			if (SDL_GetClosestDisplayMode(last_display, &requested_mode, &closest_mode) == NULL)
		// 			{
		// 				Com_Printf("SDL was unable to find a mode close to %ix%i@%i\n", w, h, requested_mode.refresh_rate);
		// 				Cvar_SetValue("vid_rate", -1);
		// 			}
		// 			else
		// 			{
		// 				Com_Printf("User requested %ix%i@%i, setting closest mode %ix%i@%i\n",
		// 						w, h, requested_mode.refresh_rate, w, h, closest_mode.refresh_rate);

		// 				if (SDL_SetWindowDisplayMode(window, &closest_mode) != 0)
		// 				{
		// 					Com_Printf("Couldn't switch to mode %ix%i@%i, staying at current mode\n",
		// 							w, h, closest_mode.refresh_rate);
		// 					Cvar_SetValue("vid_rate", -1);
		// 				}
		// 				else
		// 				{
		// 					Cvar_SetValue("vid_rate", closest_mode.refresh_rate);
		// 				}
		// 			}

		// 		}
		// 	}
	} else {
		return false
	}

	return true
}

/*
 * Returns the current display refresh rate. There're 2 limitations:
 *
 * * The timing code in frame.c only understands full integers, so
 *   values given by vid_displayrefreshrate are always round up. For
 *   example 59.95 become 60. Rounding up is the better choice for
 *   most users because assuming a too high display refresh rate
 *   avoids micro stuttering caused by missed frames if the vsync
 *   is enabled. The price are small and hard to notice timing
 *   problems.
 *
 * * SDL returns only full integers. In most cases they're rounded
 *   up, but in some cases - likely depending on the GPU driver -
 *   they're rounded down. If the value is rounded up, we'll see
 *   some small and nard to notice timing problems. If the value
 *   is rounded down frames will be missed. Both is only relevant
 *   if the vsync is enabled.
 */
func (T *qClient) GetRefreshRate() int {

	if T.vid_displayrefreshrate.Int() > -1 ||
		T.vid_displayrefreshrate.Modified {
		T.glimp_refreshRate = int(math.Ceil(float64(T.vid_displayrefreshrate.Int())))
		T.vid_displayrefreshrate.Modified = false
	}

	if T.glimp_refreshRate == -1 {
		//  SDL_DisplayMode mode;

		i, err := T.window.GetDisplayIndex()

		if err == nil && i >= 0 {
			if mode, err := sdl.GetCurrentDisplayMode(i); err == nil {
				T.glimp_refreshRate = int(mode.RefreshRate)
			}
		}

		// Something went wrong, use default.
		if T.glimp_refreshRate <= 0 {
			T.glimp_refreshRate = 60
		}
	}

	return T.glimp_refreshRate
}
