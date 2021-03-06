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
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307,
 * USA.
 *
 * =======================================================================
 *
 * ABI between client and refresher
 *
 * =======================================================================
 */
package shared

const MAX_DLIGHTS = 32
const MAX_ENTITIES = 128
const MAX_PARTICLES = 4096

const POWERSUIT_SCALE = 4.0

type Entity_t struct {
	Model  interface{} /* opaque type outside refresh */
	Angles [3]float32

	/* most recent data */
	Origin [3]float32 /* also used as RF_BEAM's "from" */
	Frame  int        /* also used as RF_BEAM's diameter */

	/* previous data for lerping */
	Oldorigin [3]float32 /* also used as RF_BEAM's "to" */
	Oldframe  int

	/* misc */
	Backlerp float32 /* 0.0 = current, 1.0 = old */
	Skinnum  int     /* also used as RF_BEAM's palette index */

	Lightstyle int     /* for flashing entities */
	Alpha      float32 /* ignore if RF_TRANSLUCENT isn't set */

	Skin  interface{} /* NULL for inline skin */
	Flags int
}

type Dlight_t struct {
	Origin    [3]float32
	Color     [3]float32
	Intensity float32
}

type Lightstyle_t struct {
	Rgb   [3]float32 /* 0.0 - 2.0 */
	White float32    /* r+g+b */
}

type Particle_t struct {
	Origin [3]float32
	Color  int
	Alpha  float32
}

type Refdef_t struct {
	X, Y, Width, Height int /* in virtual screen coordinates */
	Fov_x, Fov_y        float32
	Vieworg             [3]float32
	Viewangles          [3]float32
	Blend               [4]float32 /* rgba 0-1 full screen blend */
	Time                float32    /* time is used to auto animate */
	Rdflags             int        /* RDF_UNDERWATER, etc */

	Areabits []byte /* if not NULL, only areas with set bits will be drawn */

	Lightstyles []Lightstyle_t /* [MAX_LIGHTSTYLES] */

	Entities []Entity_t

	// int			num_dlights; // <= 32 (MAX_DLIGHTS)
	Dlights []Dlight_t

	// int			num_particles;
	Particles []Particle_t
}

//
// these are the functions exported by the refresh module
//
type Refexport_t interface {
	// called when the library is loaded
	Init() bool

	PrepareForWindow() int

	// called by GLimp_InitGraphics() *after* creating window,
	// passing the SDL_Window* (void* so we don't spill SDL.h here)
	// (or SDL_Surface* for SDL1.2, another reason to use void*)
	// returns true (1) on success
	InitContext(sdl_window interface{}) bool

	IsVSyncActive() bool

	// All data that will be used in a level should be
	// registered before rendering any frames to prevent disk hits,
	// but they can still be registered at a later time
	// if necessary.
	//
	// EndRegistration will free any remaining data that wasn't registered.
	// Any model_s or skin_s pointers from before the BeginRegistration
	// are no longer valid after EndRegistration.
	//
	// Skins and images need to be differentiated, because skins
	// are flood filled to eliminate mip map edge errors, and pics have
	// an implicit "pics/" prepended to the name. (a pic name that starts with a
	// slash will not use the "pics/" prefix or the ".pcx" postfix)
	BeginRegistration(name string) error
	RegisterModel(name string) (interface{}, error)
	RegisterSkin(name string) interface{}
	DrawFindPic(name string) interface{}

	RenderFrame(fd Refdef_t) error

	DrawStretchPic(x, y, w, h int, name string)
	DrawTileClear(x, y, w, h int, name string)
	DrawPicScaled(x, y int, pic string, factor float32)
	DrawFill(x, y, w, h, c int)
	DrawGetPicSize(name string) (int, int)
	DrawCharScaled(x, y, num int, scale float32)

	BeginFrame(camera_separation float32) error
	EndFrame()

	SetSky(name string, rotate float32, axis []float32)
}

type Refimport_t interface {
	Sys_Error(code int, format string, a ...interface{}) error

	Com_VPrintf(print_level int, format string, a ...interface{})

	Cvar_Get(var_name, var_value string, flags int) *CvarT
	Cvar_Set(var_name, value string) *CvarT

	LoadFile(path string) ([]byte, error)

	Vid_GetModeInfo(mode int) (int, int)

	GLimp_InitGraphics(fullscreen, width, height int) (int, int)
}
