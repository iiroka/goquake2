/*
 * Copyright (C) 1997-2001 Id Software, Inc.
 * Copyright (C) 2016-2017 Daniel Gibson
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
 * SDL backend for the GL3 renderer. Everything that needs to be on the
 * renderer side of thing. Also all glad (or whatever OpenGL loader I
 * end up using) specific things.
 *
 * =======================================================================
 */
package gl3

import (
	"fmt"
	"goquake2/shared"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

/*
 * This function returns the flags used at the SDL window
 * creation by GLimp_InitGraphics(). In case of error -1
 * is returned.
 */
func (T *qGl3) PrepareForWindow() int {
	//  // Mkay, let's try to load the libGL,
	//  const char *libgl;
	//  cvar_t *gl3_libgl = ri.Cvar_Get("gl3_libgl", "", CVAR_ARCHIVE);

	//  if (strlen(gl3_libgl->string) == 0)
	//  {
	// 	 libgl = NULL;
	//  }
	//  else
	//  {
	// 	 libgl = gl3_libgl->string;
	//  }

	//  while (1)
	//  {
	// 	 if (SDL_GL_LoadLibrary(libgl) < 0)
	// 	 {
	// 		 if (libgl == NULL)
	// 		 {
	// 			 ri.Sys_Error(ERR_FATAL, "Couldn't load libGL: %s!", SDL_GetError());

	// 			 return -1;
	// 		 }
	// 		 else
	// 		 {
	// 			 R_Printf(PRINT_ALL, "Couldn't load libGL: %s!\n", SDL_GetError());
	// 			 R_Printf(PRINT_ALL, "Retrying with default...\n");

	// 			 ri.Cvar_Set("gl3_libgl", "");
	// 			 libgl = NULL;
	// 		 }
	// 	 }
	// 	 else
	// 	 {
	// 		 break;
	// 	 }
	//  }

	// Set GL context attributs bound to the window.
	sdl.GLSetAttribute(sdl.GL_RED_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_GREEN_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_BLUE_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)

	if sdl.GLSetAttribute(sdl.GL_STENCIL_SIZE, 8) == nil {
		T.gl3config.stencil = true
	} else {
		T.gl3config.stencil = false
	}

	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	//  // Set GL context flags.
	//  int contextFlags = SDL_GL_CONTEXT_FORWARD_COMPATIBLE_FLAG;

	//  if (gl3_debugcontext && gl3_debugcontext->value)
	//  {
	// 	 contextFlags |= SDL_GL_CONTEXT_DEBUG_FLAG;
	//  }

	//  if (contextFlags != 0)
	//  {
	// 	 SDL_GL_SetAttribute(SDL_GL_CONTEXT_FLAGS, contextFlags);
	//  }

	// Let's see if the driver supports MSAA.
	msaa_samples := 0

	if T.gl_msaa_samples.Bool() {
		msaa_samples = T.gl_msaa_samples.Int()

		if err := sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1); err != nil {
			T.rPrintf(shared.PRINT_ALL, "MSAA is unsupported: %v\n", err.Error())

			T.ri.Cvar_Set("r_msaa_samples", "0")

			sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 0)
			sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 0)
		} else if err := sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, msaa_samples); err != nil {
			T.rPrintf(shared.PRINT_ALL, "MSAA %vx is unsupported: %v\n", msaa_samples, err.Error())

			T.ri.Cvar_Set("r_msaa_samples", "0")

			sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 0)
			sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 0)
		}
	} else {
		sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 0)
		sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 0)
	}

	return sdl.WINDOW_OPENGL
}

func (T *qGl3) InitContext(win interface{}) bool {
	// Coders are stupid.
	// if win == nil {
	// 	T.ri.Sys_Error(shared.ERR_FATAL, "R_InitContext() must not be called with NULL argument!")

	// 	return false
	// }

	T.window = win.(*sdl.Window)

	// Initialize GL context.
	ctx, err := T.window.GLCreateContext()
	if err != nil {
		T.rPrintf(shared.PRINT_ALL, "GL3_InitContext(): Creating OpenGL Context failed: %v\n", err.Error())
		T.window = nil
		return false
	}
	T.context = ctx

	if err := gl.Init(); err != nil {
		T.rPrintf(shared.PRINT_ALL, "GL3_InitContext(): GL init failed: %v\n", err.Error())
		T.window = nil
		return false
	}

	// Check if we've got the requested MSAA.
	if T.gl_msaa_samples.Bool() {
		if msaa_samples, err := sdl.GLGetAttribute(sdl.GL_MULTISAMPLESAMPLES); err == nil {
			T.ri.Cvar_Set("r_msaa_samples", fmt.Sprintf("%v", msaa_samples))
		}
	}

	// Check if we've got at least 8 stencil bits
	if T.gl3config.stencil {
		if stencil_bits, err := sdl.GLGetAttribute(sdl.GL_STENCIL_SIZE); err != nil || stencil_bits < 8 {
			T.gl3config.stencil = false
		}
	}

	// Enable vsync if requested.
	T.setVsync()

	// // Load GL pointrs through GLAD and check context.
	// if( !gladLoadGLLoader(SDL_GL_GetProcAddress))
	// {
	// 	R_Printf(PRINT_ALL, "GL3_InitContext(): ERROR: loading OpenGL function pointers failed!\n");

	// 	return false;
	// }
	// else if (GLVersion.major < 3 || (GLVersion.major == 3 && GLVersion.minor < 2))
	// {
	// 	R_Printf(PRINT_ALL, "GL3_InitContext(): ERROR: glad only got GL version %d.%d!\n", GLVersion.major, GLVersion.minor);

	// 	return false;
	// }
	// else
	// {
	// 	R_Printf(PRINT_ALL, "Successfully loaded OpenGL function pointers using glad, got version %d.%d!\n", GLVersion.major, GLVersion.minor);
	// }

	var numExtensions int32
	gl.GetIntegerv(gl.NUM_EXTENSIONS, &numExtensions)
	for i := 0; i < int(numExtensions); i++ {
		ext := gl.GoStr(gl.GetStringi(gl.EXTENSIONS, uint32(i)))
		if ext == "GL_EXT_texture_filter_anisotropic" {
			T.gl3config.anisotropic = true
		} else if ext == "GL_ARB_debug_output" {
			T.gl3config.debug_output = true
		}
	}

	// gl3config.debug_output = GLAD_GL_ARB_debug_output != 0;
	// gl3config.anisotropic = GLAD_GL_EXT_texture_filter_anisotropic != 0;

	// gl3config.major_version = GLVersion.major;
	// gl3config.minor_version = GLVersion.minor;

	// // Debug context setup.
	// if (gl3_debugcontext && gl3_debugcontext->value && gl3config.debug_output)
	// {
	// 	glDebugMessageCallbackARB(DebugCallback, NULL);

	// 	// Call GL3_DebugCallback() synchronously, i.e. directly when and
	// 	// where the error happens (so we can get the cause in a backtrace)
	// 	glEnable(GL_DEBUG_OUTPUT_SYNCHRONOUS_ARB);
	// }

	// Window title - set here so we can display renderer name in it.
	title := fmt.Sprintf("Yamagi Quake II %s - OpenGL 3.2", shared.YQ2VERSION)
	T.window.SetTitle(title)

	return true
}

// ---------

/*
 * Swaps the buffers and shows the next frame.
 */
func (T *qGl3) EndFrame() {
	//  if(gl3config.useBigVBO) {
	// 	 // I think this is a good point to orphan the VBO and get a fresh one
	// 	 GL3_BindVAO(gl3state.vao3D);
	// 	 GL3_BindVBO(gl3state.vbo3D);
	// 	 glBufferData(GL_ARRAY_BUFFER, gl3state.vbo3Dsize, NULL, GL_STREAM_DRAW);
	// 	 gl3state.vbo3DcurOffset = 0;
	//  }

	T.window.GLSwap()
}

/*
 * Returns whether the vsync is enabled.
 */
func (T *qGl3) IsVSyncActive() bool {
	return T.vsyncActive
}

/*
 * Enables or disabes the vsync.
 */
func (T *qGl3) setVsync() {
	// Make sure that the user given
	// value is SDL compatible...
	vsync := 0

	if T.r_vsync.Int() == 1 {
		vsync = 1
	} else if T.r_vsync.Int() == 2 {
		vsync = -1
	}

	if sdl.GLSetSwapInterval(vsync) != nil {
		if vsync == -1 {
			// Not every system supports adaptive
			// vsync, fallback to normal vsync.
			T.rPrintf(shared.PRINT_ALL, "Failed to set adaptive vsync, reverting to normal vsync.\n")
			sdl.GLSetSwapInterval(1)
		}
	}

	interval, _ := sdl.GLGetSwapInterval()
	T.vsyncActive = interval != 0
}
