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
 * Refresher setup and main part of the frame generation, for OpenGL3
 *
 * =======================================================================
 */
package gl3

import (
	"goquake2/shared"

	"github.com/go-gl/gl/v3.2-core/gl"
)

var gl3_identityMat4 = []float32{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 1, 0,
	0, 0, 0, 1,
}

func QGl3Create(ri shared.Refimport_t) shared.Refexport_t {
	r := &qGl3{}
	r.ri = ri
	r.gl3textures = make([]gl3image_t, MAX_GL3TEXTURES)
	r.mod_inline = make([]gl3model_t, MAX_MOD_KNOWN)
	r.mod_known = make([]gl3model_t, MAX_MOD_KNOWN)
	return r
}

func (T *qGl3) gl3Strings() {

	T.rPrintf(shared.PRINT_ALL, "GL_VENDOR: %s\n", T.gl3config.vendor_string)
	T.rPrintf(shared.PRINT_ALL, "GL_RENDERER: %s\n", T.gl3config.renderer_string)
	T.rPrintf(shared.PRINT_ALL, "GL_VERSION: %s\n", T.gl3config.version_string)
	T.rPrintf(shared.PRINT_ALL, "GL_SHADING_LANGUAGE_VERSION: %s\n", T.gl3config.glsl_version_string)

	var numExtensions int32
	gl.GetIntegerv(gl.NUM_EXTENSIONS, &numExtensions)

	T.rPrintf(shared.PRINT_ALL, "GL_EXTENSIONS:")
	for i := 0; i < int(numExtensions); i++ {
		T.rPrintf(shared.PRINT_ALL, " %s", gl.GoStr(gl.GetStringi(gl.EXTENSIONS, uint32(i))))
	}
	T.rPrintf(shared.PRINT_ALL, "\n")
}

func (T *qGl3) register() {
	T.gl_lefthand = T.ri.Cvar_Get("hand", "0", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.r_gunfov = T.ri.Cvar_Get("r_gunfov", "80", shared.CVAR_ARCHIVE)
	T.r_farsee = T.ri.Cvar_Get("r_farsee", "0", shared.CVAR_LATCH|shared.CVAR_ARCHIVE)

	T.gl_drawbuffer = T.ri.Cvar_Get("gl_drawbuffer", "GL_BACK", 0)
	T.r_vsync = T.ri.Cvar_Get("r_vsync", "1", shared.CVAR_ARCHIVE)
	T.gl_msaa_samples = T.ri.Cvar_Get("r_msaa_samples", "0", shared.CVAR_ARCHIVE)
	T.gl_retexturing = T.ri.Cvar_Get("r_retexturing", "1", shared.CVAR_ARCHIVE)
	T.gl3_debugcontext = T.ri.Cvar_Get("gl3_debugcontext", "0", 0)
	T.r_mode = T.ri.Cvar_Get("r_mode", "4", shared.CVAR_ARCHIVE)
	T.r_customwidth = T.ri.Cvar_Get("r_customwidth", "1024", shared.CVAR_ARCHIVE)
	T.r_customheight = T.ri.Cvar_Get("r_customheight", "768", shared.CVAR_ARCHIVE)
	T.gl3_particle_size = T.ri.Cvar_Get("gl3_particle_size", "40", shared.CVAR_ARCHIVE)
	T.gl3_particle_fade_factor = T.ri.Cvar_Get("gl3_particle_fade_factor", "1.2", shared.CVAR_ARCHIVE)
	T.gl3_particle_square = T.ri.Cvar_Get("gl3_particle_square", "0", shared.CVAR_ARCHIVE)

	//  0: use lots of calls to glBufferData()
	//  1: reduce calls to glBufferData() with one big VBO (see GL3_BufferAndDraw3D())
	// -1: auto (let yq2 choose to enable/disable this based on detected driver)
	T.gl3_usebigvbo = T.ri.Cvar_Get("gl3_usebigvbo", "-1", shared.CVAR_ARCHIVE)

	T.r_norefresh = T.ri.Cvar_Get("r_norefresh", "0", 0)
	T.r_drawentities = T.ri.Cvar_Get("r_drawentities", "1", 0)
	T.r_drawworld = T.ri.Cvar_Get("r_drawworld", "1", 0)
	T.r_fullbright = T.ri.Cvar_Get("r_fullbright", "0", 0)
	T.r_fixsurfsky = T.ri.Cvar_Get("r_fixsurfsky", "0", shared.CVAR_ARCHIVE)

	/* don't bilerp characters and crosshairs */
	T.gl_nolerp_list = T.ri.Cvar_Get("r_nolerp_list", "pics/conchars.pcx pics/ch1.pcx pics/ch2.pcx pics/ch3.pcx", 0)
	T.gl_nobind = T.ri.Cvar_Get("gl_nobind", "0", 0)

	T.gl_texturemode = T.ri.Cvar_Get("gl_texturemode", "GL_LINEAR_MIPMAP_NEAREST", shared.CVAR_ARCHIVE)
	T.gl_anisotropic = T.ri.Cvar_Get("r_anisotropic", "0", shared.CVAR_ARCHIVE)

	T.vid_fullscreen = T.ri.Cvar_Get("vid_fullscreen", "0", shared.CVAR_ARCHIVE)
	T.vid_gamma = T.ri.Cvar_Get("vid_gamma", "1.2", shared.CVAR_ARCHIVE)
	T.gl3_intensity = T.ri.Cvar_Get("gl3_intensity", "1.5", shared.CVAR_ARCHIVE)
	T.gl3_intensity_2D = T.ri.Cvar_Get("gl3_intensity_2D", "1.5", shared.CVAR_ARCHIVE)

	T.r_lightlevel = T.ri.Cvar_Get("r_lightlevel", "0", 0)
	T.gl3_overbrightbits = T.ri.Cvar_Get("gl3_overbrightbits", "1.3", shared.CVAR_ARCHIVE)

	T.gl_lightmap = T.ri.Cvar_Get("gl_lightmap", "0", 0)
	T.gl_shadows = T.ri.Cvar_Get("r_shadows", "0", shared.CVAR_ARCHIVE)

	T.r_modulate = T.ri.Cvar_Get("r_modulate", "1", shared.CVAR_ARCHIVE)
	T.gl_zfix = T.ri.Cvar_Get("gl_zfix", "0", 0)
	T.r_clear = T.ri.Cvar_Get("r_clear", "0", 0)
	T.gl_cull = T.ri.Cvar_Get("gl_cull", "1", 0)
	T.r_lockpvs = T.ri.Cvar_Get("r_lockpvs", "0", 0)
	T.r_novis = T.ri.Cvar_Get("r_novis", "0", 0)
	T.r_speeds = T.ri.Cvar_Get("r_speeds", "0", 0)
	T.gl_finish = T.ri.Cvar_Get("gl_finish", "0", shared.CVAR_ARCHIVE)

	// ri.Cmd_AddCommand("imagelist", GL3_ImageList_f);
	// ri.Cmd_AddCommand("screenshot", GL3_ScreenShot);
	// ri.Cmd_AddCommand("modellist", GL3_Mod_Modellist_f);
	// ri.Cmd_AddCommand("gl_strings", GL3_Strings);
}

/*
 * Changes the video mode
 */

// the following is only used in the next to functions,
// no need to put it in a header
const (
	rserr_ok           = 0
	rserr_invalid_mode = 1
	rserr_unknown      = 2
)

func (T *qGl3) setModeImpl(pwidth, pheight, mode, fullscreen int) (int, int, int) {
	T.rPrintf(shared.PRINT_ALL, "Setting mode %d:", mode)

	/* mode -1 is not in the vid mode table - so we keep the values in pwidth
	   and pheight and don't even try to look up the mode info */
	var w int
	var h int
	if mode >= 0 {
		w, h = T.ri.Vid_GetModeInfo(mode)
		if w < 0 || h < 0 {
			T.rPrintf(shared.PRINT_ALL, " invalid mode\n")
			return rserr_invalid_mode, w, h
		}
	}

	// /* We trying to get resolution from desktop */
	// if (mode == -2)
	// {
	// 	if(!ri.GLimp_GetDesktopMode(pwidth, pheight))
	// 	{
	// 		R_Printf( PRINT_ALL, " can't detect mode\n" );
	// 		return rserr_invalid_mode;
	// 	}
	// }

	T.rPrintf(shared.PRINT_ALL, " %vx%v (vid_fullscreen %v)\n", w, h, fullscreen)

	w, h = T.ri.GLimp_InitGraphics(fullscreen, w, h)
	if w < 0 || h < 0 {
		return rserr_invalid_mode, w, h
	}

	return rserr_ok, w, h
}

func (T *qGl3) setMode() bool {

	fullscreen := T.vid_fullscreen.Int()

	T.vid_fullscreen.Modified = false
	T.r_mode.Modified = false

	// /* a bit hackish approach to enable custom resolutions:
	//    Glimp_SetMode needs these values set for mode -1 */
	// vid.width = r_customwidth->value;
	// vid.height = r_customheight->value;

	if err, w, h := T.setModeImpl(T.r_customwidth.Int(), T.r_customheight.Int(), T.r_mode.Int(), fullscreen); err == rserr_ok {
		T.vid.width = w
		T.vid.height = h
		if T.r_mode.Int() == -1 {
			T.gl3state.prev_mode = 4 /* safe default for custom mode */
		} else {
			T.gl3state.prev_mode = T.r_mode.Int()
		}
	} else {
		if err == rserr_invalid_mode {
			T.rPrintf(shared.PRINT_ALL, "ref_gl3::GL3_SetMode() - invalid mode\n")

			// 		if (gl_msaa_samples->value != 0.0f)
			// 		{
			// 			R_Printf(PRINT_ALL, "gl_msaa_samples was %d - will try again with gl_msaa_samples = 0\n", (int)gl_msaa_samples->value);
			// 			ri.Cvar_SetValue("r_msaa_samples", 0.0f);
			// 			gl_msaa_samples->modified = false;

			// 			if ((err = SetMode_impl(&vid.width, &vid.height, r_mode->value, 0)) == rserr_ok)
			// 			{
			// 				return true;
			// 			}
			// 		}
			// 		if(r_mode->value == gl3state.prev_mode)
			// 		{
			// 			// trying again would result in a crash anyway, give up already
			// 			// (this would happen if your initing fails at all and your resolution already was 640x480)
			// 			return false;
			// 		}

			// 		ri.Cvar_SetValue("r_mode", gl3state.prev_mode);
			// 		r_mode->modified = false;
		}

		// 	/* try setting it back to something safe */
		// 	if ((err = SetMode_impl(&vid.width, &vid.height, gl3state.prev_mode, 0)) != rserr_ok)
		// 	{
		// 		R_Printf(PRINT_ALL, "ref_gl3::GL3_SetMode() - could not revert to safe mode\n");
		// 		return false;
		// 	}
	}

	return true
}

func (T *qGl3) Init() bool {
	// 	Swap_Init(); // FIXME: for fucks sake, this doesn't have to be done at runtime!

	// 	R_Printf(PRINT_ALL, "Refresh: " REF_VERSION "\n");
	// 	R_Printf(PRINT_ALL, "Client: " YQ2VERSION "\n\n");

	// 	if(sizeof(float) != sizeof(GLfloat))
	// 	{
	// 		// if this ever happens, things would explode because we feed vertex arrays and UBO data
	// 		// using floats to OpenGL, which expects GLfloat (can't easily change, those floats are from HMM etc)
	// 		// (but to be honest I very much doubt this will ever happen.)
	// 		R_Printf(PRINT_ALL, "ref_gl3: sizeof(float) != sizeof(GLfloat) - we're in real trouble here.\n");
	// 		return false;
	// 	}

	if err := T.drawGetPalette(); err != nil {
		return false
	}

	T.register()

	/* set our "safe" mode */
	T.gl3state.prev_mode = 4
	//gl_state.stereo_mode = gl1_stereo->value;

	/* create the window and set up the context */
	if !T.setMode() {
		T.rPrintf(shared.PRINT_ALL, "ref_gl3::R_Init() - could not R_SetMode()\n")
		return false
	}

	// 	ri.Vid_MenuInit();

	/* get our various GL strings */
	T.gl3config.vendor_string = gl.GoStr(gl.GetString(gl.VENDOR))
	T.gl3config.renderer_string = gl.GoStr(gl.GetString(gl.RENDERER))
	T.gl3config.version_string = gl.GoStr(gl.GetString(gl.VERSION))
	T.gl3config.glsl_version_string = gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))

	T.rPrintf(shared.PRINT_ALL, "\nOpenGL setting:\n")
	T.gl3Strings()

	/*
		if (gl_config.major_version < 3)
		{
			// if (gl_config.major_version == 3 && gl_config.minor_version < 2)
			{
				QGL_Shutdown();
				R_Printf(PRINT_ALL, "Support for OpenGL 3.2 is not available\n");

				return false;
			}
		}
	*/

	T.rPrintf(shared.PRINT_ALL, "\n\nProbing for OpenGL extensions:\n")

	/* Anisotropic */
	T.rPrintf(shared.PRINT_ALL, " - Anisotropic Filtering: ")

	if T.gl3config.anisotropic {
		// gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY_EXT, &T.gl3config.max_anisotropy)

		T.rPrintf(shared.PRINT_ALL, "Max level: %vx\n", T.gl3config.max_anisotropy)
	} else {
		T.gl3config.max_anisotropy = 0.0

		T.rPrintf(shared.PRINT_ALL, "Not supported\n")
	}

	// 	if(gl3config.debug_output)
	// 	{
	// 		R_Printf(PRINT_ALL, " - OpenGL Debug Output: Supported ");
	// 		if(gl3_debugcontext->value == 0.0f)
	// 		{
	// 			R_Printf(PRINT_ALL, "(but disabled with gl3_debugcontext = 0)\n");
	// 		}
	// 		else
	// 		{
	// 			R_Printf(PRINT_ALL, "and enabled with gl3_debugcontext = %i\n", (int)gl3_debugcontext->value);
	// 		}
	// 	}
	// 	else
	// 	{
	T.rPrintf(shared.PRINT_ALL, " - OpenGL Debug Output: Not Supported\n")
	// 	}

	T.gl3config.useBigVBO = false
	// 	if(gl3_usebigvbo->value == 1.0f)
	// 	{
	// 		R_Printf(PRINT_ALL, "Enabling useBigVBO workaround because gl3_usebigvbo = 1\n");
	// 		gl3config.useBigVBO = true;
	// 	}
	// 	else if(gl3_usebigvbo->value == -1.0f)
	// 	{
	// 		// enable for AMDs proprietary Windows and Linux drivers
	// #ifdef _WIN32
	// 		if(gl3config.version_string != NULL && gl3config.vendor_string != NULL
	// 		   && strstr(gl3config.vendor_string, "ATI Technologies Inc") != NULL)
	// 		{
	// 			int a, b, ver;
	// 			if(sscanf(gl3config.version_string, " %d.%d.%d ", &a, &b, &ver) >= 3 && ver >= 13431)
	// 			{
	// 				// turns out the legacy driver is a lot faster *without* the workaround :-/
	// 				// GL_VERSION for legacy 16.2.1 Beta driver: 3.2.13399 Core Profile Forward-Compatible Context 15.200.1062.1004
	// 				//            (this is the last version that supports the Radeon HD 6950)
	// 				// GL_VERSION for (non-legacy) 16.3.1 driver on Radeon R9 200: 4.5.13431 Compatibility Profile Context 16.150.2111.0
	// 				// GL_VERSION for non-legacy 17.7.2 WHQL driver: 4.5.13491 Compatibility Profile/Debug Context 22.19.662.4
	// 				// GL_VERSION for 18.10.1 driver: 4.6.13541 Compatibility Profile/Debug Context 25.20.14003.1010
	// 				// GL_VERSION for (current) 19.3.2 driver: 4.6.13547 Compatibility Profile/Debug Context 25.20.15027.5007
	// 				// (the 3.2/4.5/4.6 can probably be ignored, might depend on the card and what kind of context was requested
	// 				//  but AFAIK the number behind that can be used to roughly match the driver version)
	// 				// => let's try matching for x.y.z with z >= 13431
	// 				// (no, I don't feel like testing which release since 16.2.1 has introduced the slowdown.)
	// 				R_Printf(PRINT_ALL, "Detected AMD Windows GPU driver, enabling useBigVBO workaround\n");
	// 				gl3config.useBigVBO = true;
	// 			}
	// 		}
	// #elif defined(__linux__)
	// 		if(gl3config.vendor_string != NULL && strstr(gl3config.vendor_string, "Advanced Micro Devices, Inc.") != NULL)
	// 		{
	// 			R_Printf(PRINT_ALL, "Detected proprietary AMD GPU driver, enabling useBigVBO workaround\n");
	// 			R_Printf(PRINT_ALL, "(consider using the open source RadeonSI drivers, they tend to work better overall)\n");
	// 			gl3config.useBigVBO = true;
	// 		}
	// #endif
	// 	}

	// generate texture handles for all possible lightmaps
	T.gl3state.lightmap_textureIDs = make([]uint32, MAX_LIGHTMAPS*MAX_LIGHTMAPS_PER_SURFACE)
	gl.GenTextures(MAX_LIGHTMAPS*MAX_LIGHTMAPS_PER_SURFACE, (*uint32)(gl.Ptr(T.gl3state.lightmap_textureIDs)))

	T.setDefaultState()

	if T.initShaders() {
		T.rPrintf(shared.PRINT_ALL, "Loading shaders succeeded.\n")
	} else {
		T.rPrintf(shared.PRINT_ALL, "Loading shaders failed.\n")
		return false
	}

	T.registration_sequence = 1 // from R_InitImages() (everything else from there shouldn't be needed anymore)

	// 	GL3_Mod_Init();

	// 	GL3_InitParticleTexture();

	if err := T.drawInitLocal(); err != nil {
		return false
	}

	T.surfInit()

	T.rPrintf(shared.PRINT_ALL, "\n")
	return true
}

func (T *qGl3) setGL2D() {
	var x int32 = 0
	var w int32 = int32(T.vid.width)
	var y int32 = 0
	var h int32 = int32(T.vid.height)

	gl.Viewport(x, y, w, h)

	transMatr := HMM_Orthographic(0, float32(T.vid.width), float32(T.vid.height), 0, -99999, 99999)

	copy(T.gl3state.uni2DData.data, transMatr)

	T.updateUBO2D()

	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.BLEND)
}

func (T *qGl3) clear() {
	// Check whether the stencil buffer needs clearing, and do so if need be.
	var stencilFlags uint32 = 0

	if T.r_clear.Bool() {
		gl.Clear(gl.COLOR_BUFFER_BIT | stencilFlags | gl.DEPTH_BUFFER_BIT)
	} else {
		gl.Clear(gl.DEPTH_BUFFER_BIT | stencilFlags)
	}

	T.gl3depthmin = 0
	T.gl3depthmax = 1
	gl.DepthFunc(gl.LEQUAL)

	gl.DepthRange(T.gl3depthmin, T.gl3depthmax)

	// 	if (gl_zfix->value)
	// 	{
	// 		if (gl3depthmax > gl3depthmin)
	// 		{
	// 			glPolygonOffset(0.05, 1);
	// 		}
	// 		else
	// 		{
	// 			glPolygonOffset(-0.05, -1);
	// 		}
	// 	}

	// 	/* stencilbuffer shadows */
	// 	if (gl_shadows->value && gl3config.stencil)
	// 	{
	// 		glClearStencil(1);
	// 		glClear(GL_STENCIL_BUFFER_BIT);
	// 	}
}

func (T *qGl3) BeginFrame(camera_separation float32) error {
	/* change modes if necessary */
	// if (r_mode->modified) {
	// 	vid_fullscreen->modified = true;
	// }

	if T.vid_gamma.Modified || T.gl3_intensity.Modified || T.gl3_intensity_2D.Modified {
		T.vid_gamma.Modified = false
		T.gl3_intensity.Modified = false
		T.gl3_intensity_2D.Modified = false

		T.gl3state.uniCommonData.setGamma(1.0 / T.vid_gamma.Float())
		T.gl3state.uniCommonData.setIntensity(T.gl3_intensity.Float())
		T.gl3state.uniCommonData.setIntensity2D(T.gl3_intensity_2D.Float())
		T.updateUBOCommon()
	}

	// // in GL3, overbrightbits can have any positive value
	// if (gl3_overbrightbits->modified)
	// {
	// 	gl3_overbrightbits->modified = false;

	// 	if(gl3_overbrightbits->value < 0.0f)
	// 	{
	// 		ri.Cvar_Set("gl3_overbrightbits", "0");
	// 	}

	// 	gl3state.uni3DData.overbrightbits = (gl3_overbrightbits->value <= 0.0f) ? 1.0f : gl3_overbrightbits->value;
	// 	GL3_UpdateUBO3D();
	// }

	// if(gl3_particle_fade_factor->modified)
	// {
	// 	gl3_particle_fade_factor->modified = false;
	// 	gl3state.uni3DData.particleFadeFactor = gl3_particle_fade_factor->value;
	// 	GL3_UpdateUBO3D();
	// }

	// if(gl3_particle_square->modified)
	// {
	// 	gl3_particle_square->modified = false;
	// 	GL3_RecreateShaders();
	// }

	/* go into 2D mode */

	T.setGL2D()

	/* draw buffer stuff */
	if T.gl_drawbuffer.Modified {
		T.gl_drawbuffer.Modified = false

		// TODO: stereo stuff
		//if ((gl3state.camera_separation == 0) || gl3state.stereo_mode != STEREO_MODE_OPENGL)
		// 	{
		if T.gl_drawbuffer.String == "GL_FRONT" {
			gl.DrawBuffer(gl.FRONT)
		} else {
			gl.DrawBuffer(gl.BACK)
		}
		// 	}
	}

	// /* texturemode stuff */
	// if (gl_texturemode->modified || (gl3config.anisotropic && gl_anisotropic->modified))
	// {
	// 	GL3_TextureMode(gl_texturemode->string);
	// 	gl_texturemode->modified = false;
	// 	gl_anisotropic->modified = false;
	// }

	// if (r_vsync->modified) {
	// 	r_vsync->modified = false;
	// 	GL3_SetVsync();
	// }

	/* clear screen if desired */
	T.clear()
	return nil
}

func (T *qGl3) rPrintf(level int, format string, a ...interface{}) {
	T.ri.Com_VPrintf(level, format, a...)
}
