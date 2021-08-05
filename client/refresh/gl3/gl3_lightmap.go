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
 * Lightmap handling
 *
 * =======================================================================
 */
package gl3

import (
	"goquake2/shared"

	"github.com/go-gl/gl/v3.2-core/gl"
)

func (T *qGl3) lmInitBlock() {
	for i := range T.gl3_lms.allocated {
		T.gl3_lms.allocated[i] = 0
	}
}

func (T *qGl3) lmUploadBlock() error {
	// int map;

	// NOTE: we don't use the dynamic lightmap anymore - all lightmaps are loaded at level load
	//       and not changed after that. they're blended dynamically depending on light styles
	//       though, and dynamic lights are (will be) applied in shader, hopefully per fragment.

	T.bindLightmap(T.gl3_lms.current_lightmap_texture)

	// upload all 4 lightmaps
	for mmap := 0; mmap < MAX_LIGHTMAPS_PER_SURFACE; mmap++ {
		T.selectTMU(gl.TEXTURE1 + uint32(mmap)) // this relies on GL_TEXTURE2 being GL_TEXTURE1+1 etc
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		T.gl3_lms.internal_format = GL_LIGHTMAP_FORMAT
		gl.TexImage2D(gl.TEXTURE_2D, 0, int32(T.gl3_lms.internal_format),
			BLOCK_WIDTH, BLOCK_HEIGHT, 0, GL_LIGHTMAP_FORMAT,
			gl.UNSIGNED_BYTE, gl.Ptr(T.gl3_lms.lightmap_buffers[mmap]))
	}

	T.gl3_lms.current_lightmap_texture++
	if T.gl3_lms.current_lightmap_texture == MAX_LIGHTMAPS {
		return T.ri.Sys_Error(shared.ERR_DROP, "LM_UploadBlock() - MAX_LIGHTMAPS exceeded\n")
	}
	return nil
}

/*
 * returns a texture number and the position inside it
 */
func (T *qGl3) lmAllocBlock(w, h int) (bool, int, int) {
	// int i, j;
	// int best, best2;

	best := BLOCK_HEIGHT
	var x, y int

	for i := 0; i < BLOCK_WIDTH-w; i++ {
		best2 := 0
		index := -1

		for j := 0; j < w; j++ {
			if T.gl3_lms.allocated[i+j] >= best {
				index = j
				break
			}

			if T.gl3_lms.allocated[i+j] > best2 {
				best2 = T.gl3_lms.allocated[i+j]
			}
		}

		if index < 0 {
			/* this is a valid spot */
			best = best2
			x = i
			y = best
		}
	}

	if best+h > BLOCK_HEIGHT {
		return false, x, y
	}

	for i := 0; i < w; i++ {
		T.gl3_lms.allocated[x+i] = best + h
	}

	return true, x, y
}

func (T *qGl3) lmBuildPolygonFromSurface(fa *msurface_t, mod *gl3model_t) {

	/* reconstruct the polygon */
	pedges := mod.edges
	lnumverts := fa.numedges

	// VectorClear(total);
	total := make([]float32, 3)

	/* draw texture */
	poly := &glpoly_t{}
	poly.verticesData = make([]uint32, lnumverts*gl3_3D_vtx_size)
	poly.next = fa.polys
	poly.flags = fa.flags
	fa.polys = poly
	poly.numverts = lnumverts

	normal := make([]float32, 3)
	copy(normal, fa.plane.Normal[:])

	if (fa.flags & SURF_PLANEBACK) != 0 {
		// if for some reason the normal sticks to the back of the plane, invert it
		// so it's usable for the shader
		for i := 0; i < 3; i++ {
			normal[i] = -normal[i]
		}
	}

	for i := 0; i < lnumverts; i++ {
		vert := poly.vertices(i)

		lindex := mod.surfedges[fa.firstedge+i]

		var vec []float32
		if lindex > 0 {
			r_pedge := &pedges[lindex]
			vec = mod.vertexes[r_pedge.v[0]].position[:]
		} else {
			r_pedge := &pedges[-lindex]
			vec = mod.vertexes[r_pedge.v[1]].position[:]
		}

		s := shared.DotProduct(vec, fa.texinfo.vecs[0][:]) + fa.texinfo.vecs[0][3]
		s /= float32(fa.texinfo.image.width)

		t := shared.DotProduct(vec, fa.texinfo.vecs[1][:]) + fa.texinfo.vecs[1][3]
		t /= float32(fa.texinfo.image.height)

		shared.VectorAdd(total, vec, total)
		vert.setPos(vec)
		vert.setTexCoord(s, t)

		/* lightmap texture coordinates */
		s = shared.DotProduct(vec, fa.texinfo.vecs[0][:]) + fa.texinfo.vecs[0][3]
		s -= float32(fa.texturemins[0])
		s += float32(fa.light_s * 16)
		s += 8
		s /= BLOCK_WIDTH * 16 /* fa->texinfo->texture->width; */

		t = shared.DotProduct(vec, fa.texinfo.vecs[1][:]) + fa.texinfo.vecs[1][3]
		t -= float32(fa.texturemins[1])
		t += float32(fa.light_t * 16)
		t += 8
		t /= BLOCK_HEIGHT * 16 /* fa->texinfo->texture->height; */

		vert.setLmTexCoord(s, t)

		vert.setNormal(normal)
		vert.setLightFlags(0)
	}
}

func (T *qGl3) lmCreateSurfaceLightmap(surf *msurface_t) error {

	if (surf.flags & (SURF_DRAWSKY | SURF_DRAWTURB)) != 0 {
		return nil
	}

	smax := (int(surf.extents[0]) >> 4) + 1
	tmax := (int(surf.extents[1]) >> 4) + 1

	if ok, s, t := T.lmAllocBlock(smax, tmax); ok {
		surf.light_s = s
		surf.light_t = t
	} else {
		T.lmUploadBlock()
		T.lmInitBlock()

		if ok, s, t := T.lmAllocBlock(smax, tmax); ok {
			surf.light_s = s
			surf.light_t = t
		} else {
			return T.ri.Sys_Error(shared.ERR_FATAL, "Consecutive calls to LM_AllocBlock(%d,%d) failed\n",
				smax, tmax)
		}
	}

	surf.lightmaptexturenum = T.gl3_lms.current_lightmap_texture

	return T.buildLightMap(surf, (surf.light_t*BLOCK_WIDTH+surf.light_s)*LIGHTMAP_BYTES, BLOCK_WIDTH*LIGHTMAP_BYTES)
}

func (T *qGl3) lmBeginBuildingLightmaps(m *gl3model_t) {

	var lightstyles [shared.MAX_LIGHTSTYLES]shared.Lightstyle_t

	for i := range T.gl3_lms.allocated {
		T.gl3_lms.allocated[i] = 0
	}

	T.gl3_framecount = 1 /* no dlightcache */

	/* setup the base lightstyles so the lightmaps
	   won't have to be regenerated the first time
	   they're seen */
	for i := 0; i < shared.MAX_LIGHTSTYLES; i++ {
		lightstyles[i].Rgb[0] = 1
		lightstyles[i].Rgb[1] = 1
		lightstyles[i].Rgb[2] = 1
		lightstyles[i].White = 3
	}

	T.gl3_newrefdef.Lightstyles = lightstyles[:]

	T.gl3_lms.current_lightmap_texture = 0
	T.gl3_lms.internal_format = GL_LIGHTMAP_FORMAT

	// Note: the dynamic lightmap used to be initialized here, we don't use that anymore.
}

func (T *qGl3) lmEndBuildingLightmaps() error {
	return T.lmUploadBlock()
}
