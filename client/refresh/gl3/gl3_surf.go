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
 * Surface generation and drawing
 *
 * =======================================================================
 */
package gl3

import (
	"github.com/go-gl/gl/v3.2-core/gl"
)

func (T *qGl3) surfInit() {
	// init the VAO and VBO for the standard vertexdata: 10 floats and 1 uint
	// (X, Y, Z), (S, T), (LMS, LMT), (normX, normY, normZ) ; lightFlags - last two groups for lightmap/dynlights

	gl.GenVertexArrays(1, &T.gl3state.vao3D)
	T.bindVAO(T.gl3state.vao3D)

	gl.GenBuffers(1, &T.gl3state.vbo3D)
	T.bindVBO(T.gl3state.vbo3D)

	// if(gl3config.useBigVBO) {
	// 	gl3state.vbo3Dsize = 5*1024*1024; // a 5MB buffer seems to work well?
	// 	gl3state.vbo3DcurOffset = 0;
	// 	glBufferData(GL_ARRAY_BUFFER, gl3state.vbo3Dsize, NULL, GL_STREAM_DRAW); // allocate/reserve that data
	// }

	gl.EnableVertexAttribArray(GL3_ATTRIB_POSITION)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_POSITION, 3, gl.FLOAT, false, 11*4, 0)

	gl.EnableVertexAttribArray(GL3_ATTRIB_TEXCOORD)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_TEXCOORD, 2, gl.FLOAT, false, 11*4, 3*4)

	gl.EnableVertexAttribArray(GL3_ATTRIB_LMTEXCOORD)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_LMTEXCOORD, 2, gl.FLOAT, false, 11*4, 5*4)

	gl.EnableVertexAttribArray(GL3_ATTRIB_NORMAL)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_NORMAL, 3, gl.FLOAT, false, 11*4, 7*4)

	gl.EnableVertexAttribArray(GL3_ATTRIB_LIGHTFLAGS)
	gl.VertexAttribIPointer(GL3_ATTRIB_LIGHTFLAGS, 1, gl.UNSIGNED_INT, 11*4, gl.PtrOffset(10*4))

	// init VAO and VBO for model vertexdata: 9 floats
	// (X,Y,Z), (S,T), (R,G,B,A)

	gl.GenVertexArrays(1, &T.gl3state.vaoAlias)
	T.bindVAO(T.gl3state.vaoAlias)

	gl.GenBuffers(1, &T.gl3state.vboAlias)
	T.bindVBO(T.gl3state.vboAlias)

	gl.EnableVertexAttribArray(GL3_ATTRIB_POSITION)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_POSITION, 3, gl.FLOAT, false, 9*4, 0)

	gl.EnableVertexAttribArray(GL3_ATTRIB_TEXCOORD)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_TEXCOORD, 2, gl.FLOAT, false, 9*4, 3*4)

	gl.EnableVertexAttribArray(GL3_ATTRIB_COLOR)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_COLOR, 4, gl.FLOAT, false, 9*4, 5*4)

	gl.GenBuffers(1, &T.gl3state.eboAlias)

	// init VAO and VBO for particle vertexdata: 9 floats
	// (X,Y,Z), (point_size,distace_to_camera), (R,G,B,A)

	gl.GenVertexArrays(1, &T.gl3state.vaoParticle)
	T.bindVAO(T.gl3state.vaoParticle)

	gl.GenBuffers(1, &T.gl3state.vboParticle)
	T.bindVBO(T.gl3state.vboParticle)

	gl.EnableVertexAttribArray(GL3_ATTRIB_POSITION)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_POSITION, 3, gl.FLOAT, false, 9*4, 0)

	// TODO: maybe move point size and camera origin to UBO and calculate distance in vertex shader
	gl.EnableVertexAttribArray(GL3_ATTRIB_TEXCOORD) // it's abused for (point_size, distance) here..
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_TEXCOORD, 2, gl.FLOAT, false, 9*4, 3*4)

	gl.EnableVertexAttribArray(GL3_ATTRIB_COLOR)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_COLOR, 4, gl.FLOAT, false, 9*4, 5*4)
}
