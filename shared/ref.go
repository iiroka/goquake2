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

type Refdef_t struct {
	X, Y, Width, Height int /* in virtual screen coordinates */
	Fov_x, Fov_y        float32
	Vieworg             [3]float32
	Viewangles          [3]float32
	Blend               [4]float32 /* rgba 0-1 full screen blend */
	Time                float32    /* time is used to auto animate */
	Rdflags             int        /* RDF_UNDERWATER, etc */

	// byte		*areabits; /* if not NULL, only areas with set bits will be drawn */

	// lightstyle_t	*lightstyles; /* [MAX_LIGHTSTYLES] */

	// int			num_entities;
	// entity_t	*entities;

	// int			num_dlights; // <= 32 (MAX_DLIGHTS)
	// dlight_t	*dlights;

	// int			num_particles;
	// particle_t	*particles;
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

	DrawStretchPic(x, y, w, h int, name string)

	BeginFrame(camera_separation float32) error
	EndFrame()
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
