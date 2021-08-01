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
 * Warps. Used on water surfaces und for skybox rotation.
 *
 * =======================================================================
 */
package gl3

import (
	"fmt"
	"goquake2/shared"
	"math"
)

func rBoundPoly(numverts int, verts [][3]float32, mins, maxs [3]float32) {

	mins[0] = 9999
	mins[1] = 9999
	mins[2] = 9999
	maxs[0] = -9999
	maxs[1] = -9999
	maxs[2] = -9999

	for i := 0; i < numverts; i++ {
		for j := 0; j < 3; j++ {
			if verts[i][j] < mins[j] {
				mins[j] = verts[i][j]
			}

			if verts[i][j] > maxs[j] {
				maxs[j] = verts[i][j]
			}
		}
	}
}

const SUBDIVIDE_SIZE = 64.0

func rSubdividePolygon(numverts int, verts [][3]float32, warpface *msurface_t) {

	var normal [3]float32
	copy(normal[:], warpface.plane.Normal[:])

	// if (numverts > 60) {
	// 	ri.Sys_Error(ERR_DROP, "numverts = %i", numverts);
	// }

	var mins [3]float32
	var maxs [3]float32
	rBoundPoly(numverts, verts, mins, maxs)

	var front [64][3]float32
	var back [64][3]float32
	var dist [64]float32
	for i := 0; i < 3; i++ {
		m := (mins[i] + maxs[i]) * 0.5
		m = SUBDIVIDE_SIZE * float32(math.Floor(float64(m)/SUBDIVIDE_SIZE+0.5))

		if maxs[i]-m < 8 {
			continue
		}

		if m-mins[i] < 8 {
			continue
		}

		/* cut it */
		for j := 0; j < numverts; j++ {
			dist[j] = verts[j][i] - m
		}

		/* wrap cases */
		dist[numverts] = dist[0]
		copy(verts[numverts][:], verts[0][:])

		f := 0
		b := 0
		for j := 0; j < numverts; j++ {
			if dist[j] >= 0 {
				copy(front[f][:], verts[j][:])
				f++
			}

			if dist[j] <= 0 {
				copy(back[b][:], verts[j][:])
				b++
			}

			if (dist[j] == 0) || (dist[j+1] == 0) {
				continue
			}

			if (dist[j] > 0) != (dist[j+1] > 0) {
				/* clip point */
				frac := dist[j] / (dist[j] - dist[j+1])

				for k := 0; k < 3; k++ {
					back[b][k] = verts[j][k] + frac*(verts[j+1][k]-verts[j][k])
					front[f][k] = back[b][k]
				}

				f++
				b++
			}
		}

		rSubdividePolygon(f, front[:], warpface)
		rSubdividePolygon(b, back[:], warpface)
		return
	}

	/* add a point in the center to help keep warp valid */
	poly := glpoly_t{}
	poly.verticesData = make([]uint32, (numverts+2)*gl3_3D_vtx_size)
	// poly = Hunk_Alloc(sizeof(glpoly_t) + ((numverts - 4) + 2) * sizeof(gl3_3D_vtx_t));
	poly.next = warpface.polys
	warpface.polys = &poly
	poly.numverts = numverts + 2
	var total [3]float32
	var total_s float32 = 0
	var total_t float32 = 0

	for i := 0; i < numverts; i++ {
		v := poly.vertices(i + 1)
		v.setPos(verts[i][:])
		s := shared.DotProduct(verts[i][:], warpface.texinfo.vecs[0][:])
		t := shared.DotProduct(verts[i][:], warpface.texinfo.vecs[1][:])

		total_s += s
		total_t += t
		shared.VectorAdd(total[:], verts[i][:], total[:])

		v.setTexCoord(s, t)

		v.setNormal(normal[:])
		v.setLightFlags(0)
	}

	v := poly.vertices(0)
	v.setPos(shared.VectorScaled(total[:], 1.0/float32(numverts)))
	v.setTexCoord(total_s/float32(numverts), total_t/float32(numverts))
	v.setNormal(normal[:])

	/* copy first vertex to last */
	copy(poly.verticesData[(numverts+1)*gl3_3D_vtx_size:(numverts+2)*gl3_3D_vtx_size],
		poly.verticesData[gl3_3D_vtx_size:2*gl3_3D_vtx_size])
}

/*
 * Breaks a polygon up along axial 64 unit
 * boundaries so that turbulent and sky warps
 * can be done reasonably.
 */
func gl3SubdivideSurface(fa *msurface_t, loadmodel *gl3model_t) {

	/* convert edges back to a normal polygon */
	numverts := 0
	var verts [64][3]float32

	for i := 0; i < fa.numedges; i++ {
		lindex := loadmodel.surfedges[fa.firstedge+i]

		var vec []float32
		if lindex > 0 {
			vec = loadmodel.vertexes[loadmodel.edges[lindex].v[0]].position[:]
		} else {
			vec = loadmodel.vertexes[loadmodel.edges[-lindex].v[1]].position[:]
		}

		copy(verts[numverts][:], vec)
		numverts++
	}

	rSubdividePolygon(numverts, verts[:], fa)
}

// ########### below: Sky-specific stuff ##########

const ON_EPSILON = 0.1 /* point on plane side epsilon */
const MAX_CLIP_VERTS = 64

var skytexorder = [6]int{0, 2, 1, 3, 4, 5}

/* 3dstudio environment map names */
var suf = [6]string{"rt", "bk", "lf", "ft", "up", "dn"}

var skyclip = [6][3]float32{
	{1, 1, 0},
	{1, -1, 0},
	{0, -1, 1},
	{0, 1, 1},
	{1, 0, 1},
	{-1, 0, 1},
}

var st_to_vec = [6][3]int{
	{3, -1, 2},
	{-3, 1, 2},

	{1, 3, 2},
	{-1, -3, 2},

	{-2, -1, 3}, /* 0 degrees yaw, look straight up */
	{2, -1, -3}, /* look straight down */
}

var vec_to_st = [6][3]int{
	{-2, 3, 1},
	{2, 3, -1},

	{1, 3, 2},
	{-1, 3, -2},

	{-2, -1, 3},
	{-2, 1, -3},
}

func (T *qGl3) SetSky(name string, rotate float32, axis []float32) {

	skyname := name
	T.skyrotate = rotate
	copy(T.skyaxis[:], axis)

	for i := 0; i < 6; i++ {
		// NOTE: there might be a paletted .pcx version, which was only used
		//       if gl_config.palettedtexture so it *shouldn't* be relevant for he GL3 renderer
		pathname := fmt.Sprintf("env/%s%s.tga", skyname, suf[i])

		T.sky_images[i] = T.findImage(pathname, it_sky)

		if T.sky_images[i] == nil || T.sky_images[i] == T.gl3_notexture {
			pathname = fmt.Sprintf("pics/Skies/%s%s.m8", skyname, suf[i])

			T.sky_images[i] = T.findImage(pathname, it_sky)
		}

		if T.sky_images[i] == nil {
			T.sky_images[i] = T.gl3_notexture
		}

		T.sky_min = 1.0 / 512
		T.sky_max = 511.0 / 512
	}
}
