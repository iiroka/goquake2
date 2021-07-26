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
 * Misc OpenGL3 refresher functions
 *
 * =======================================================================
 */
package gl3

import (
	"github.com/go-gl/gl/v3.2-core/gl"
)

func (T *qGl3) setDefaultState() {
	gl.ClearColor(1, 0, 0.5, 0.5)
	gl.Disable(gl.MULTISAMPLE)
	gl.CullFace(gl.FRONT)

	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.BLEND)

	T.gl_filter_min = gl.LINEAR_MIPMAP_NEAREST
	T.gl_filter_max = gl.LINEAR

	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	// TODO: gl1_texturealphamode?
	T.textureMode(T.gl_texturemode.String)
	//R_TextureAlphaMode(gl1_texturealphamode->string);
	//R_TextureSolidMode(gl1_texturesolidmode->string);

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, T.gl_filter_min)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, T.gl_filter_max)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	if T.gl_msaa_samples.Bool() {
		gl.Enable(gl.MULTISAMPLE)
		// glHint(GL_MULTISAMPLE_FILTER_HINT_NV, GL_NICEST); TODO what is this for?
	}
}
