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

func (T *qGl3) drawSkyPolygon(nump int, vecs [][3]float32) {
	// int i, j;
	// vec3_t v, av;
	// float s, t, dv;
	// int axis;
	// float *vp;

	T.c_sky++

	/* decide which face it maps to */
	v := make([]float32, 3)

	for i := 0; i < nump; i++ {
		shared.VectorAdd(vecs[i][:], v, v)
	}

	av := []float32{
		float32(math.Abs(float64(v[0]))),
		float32(math.Abs(float64(v[1]))),
		float32(math.Abs(float64(v[2])))}

	var axis int
	if (av[0] > av[1]) && (av[0] > av[2]) {
		if v[0] < 0 {
			axis = 1
		} else {
			axis = 0
		}
	} else if (av[1] > av[2]) && (av[1] > av[0]) {
		if v[1] < 0 {
			axis = 3
		} else {
			axis = 2
		}
	} else {
		if v[2] < 0 {
			axis = 5
		} else {
			axis = 4
		}
	}

	/* project new texture coords */
	for i := 0; i < nump; i++ {
		j := vec_to_st[axis][2]

		var dv float32
		if j > 0 {
			dv = vecs[i][j-1]
		} else {
			dv = -vecs[i][-j-1]
		}

		if dv < 0.001 {
			continue /* don't divide by zero */
		}

		j = vec_to_st[axis][0]

		var s float32
		if j < 0 {
			s = -vecs[i][-j-1] / dv
		} else {
			s = vecs[i][j-1] / dv
		}

		j = vec_to_st[axis][1]

		var t float32
		if j < 0 {
			t = -vecs[i][-j-1] / dv
		} else {
			t = vecs[i][j-1] / dv
		}

		if s < T.skymins[0][axis] {
			T.skymins[0][axis] = s
		}

		if t < T.skymins[1][axis] {
			T.skymins[1][axis] = t
		}

		if s > T.skymaxs[0][axis] {
			T.skymaxs[0][axis] = s
		}

		if t > T.skymaxs[1][axis] {
			T.skymaxs[1][axis] = t
		}
	}
}

func (T *qGl3) clipSkyPolygon(nump int, vecs [][3]float32, stage int) {
	// float *norm;
	// float *v;
	// qboolean front, back;
	// float d, e;
	// float dists[MAX_CLIP_VERTS];
	// int sides[MAX_CLIP_VERTS];
	// vec3_t newv[2][MAX_CLIP_VERTS];
	// int newc[2];
	// int i, j;

	if nump > MAX_CLIP_VERTS-2 {
		T.ri.Sys_Error(shared.ERR_DROP, "R_ClipSkyPolygon: MAX_CLIP_VERTS")
		return
	}

	if stage == 6 {
		/* fully clipped, so draw it */
		T.drawSkyPolygon(nump, vecs)
		return
	}

	// front = back = false;
	front := false
	back := false
	norm := skyclip[stage]

	var dists [MAX_CLIP_VERTS]float32
	var sides [MAX_CLIP_VERTS]int

	for i := 0; i < nump; i++ {
		d := shared.DotProduct(vecs[i][:], norm[:])

		if d > ON_EPSILON {
			front = true
			sides[i] = SIDE_FRONT
		} else if d < -ON_EPSILON {
			back = true
			sides[i] = SIDE_BACK
		} else {
			sides[i] = SIDE_ON
		}

		dists[i] = d
	}

	if !front || !back {
		/* not clipped */
		T.clipSkyPolygon(nump, vecs, stage+1)
		return
	}

	/* clip it */
	sides[nump] = sides[0]
	dists[nump] = dists[0]
	copy(vecs[nump][:], vecs[0][:])
	var newc [2]int
	var newv [2][MAX_CLIP_VERTS][3]float32

	for i := 0; i < nump; i++ {
		switch sides[i] {
		case SIDE_FRONT:
			copy(newv[0][newc[0]][:], vecs[i][:])
			newc[0]++
		case SIDE_BACK:
			copy(newv[1][newc[1]][:], vecs[i][:])
			newc[1]++
		case SIDE_ON:
			copy(newv[0][newc[0]][:], vecs[i][:])
			newc[0]++
			copy(newv[1][newc[1]][:], vecs[i][:])
			newc[1]++
		}

		if (sides[i] == SIDE_ON) ||
			(sides[i+1] == SIDE_ON) ||
			(sides[i+1] == sides[i]) {
			continue
		}

		d := dists[i] / (dists[i] - dists[i+1])

		for j := 0; j < 3; j++ {
			e := vecs[i][j] + d*(vecs[i+1][j]-vecs[i][j])
			newv[0][newc[0]][j] = e
			newv[1][newc[1]][j] = e
		}

		newc[0]++
		newc[1]++
	}

	/* continue */
	T.clipSkyPolygon(newc[0], newv[0][:], stage+1)
	T.clipSkyPolygon(newc[1], newv[1][:], stage+1)
}

func (T *qGl3) addSkySurface(fa *msurface_t) {
	var verts [MAX_CLIP_VERTS][3]float32

	/* calculate vertex values for sky box */
	for p := fa.polys; p != nil; p = p.next {
		for i := 0; i < p.numverts; i++ {
			shared.VectorSubtract(p.vertices(i).getPos(), T.gl3_origin[:], verts[i][:])
		}

		T.clipSkyPolygon(p.numverts, verts[:], 0)
	}
}

func (T *qGl3) clearSkyBox() {

	for i := 0; i < 6; i++ {
		T.skymins[0][i] = 9999
		T.skymins[1][i] = 9999
		T.skymaxs[0][i] = -9999
		T.skymaxs[1][i] = -9999
	}
}

func (T *qGl3) makeSkyVec(s, t float32, axis int) {
	// vec3_t v, b;
	// int j, k;

	var b [3]float32
	if !T.r_farsee.Bool() {
		b[0] = s * 2300
		b[1] = t * 2300
		b[2] = 2300
	} else {
		b[0] = s * 4096
		b[1] = t * 4096
		b[2] = 4096
	}

	var v [3]float32
	for j := 0; j < 3; j++ {
		k := st_to_vec[axis][j]

		if k < 0 {
			v[j] = -b[-k-1]
		} else {
			v[j] = b[k-1]
		}
	}

	/* avoid bilerp seam */
	s = (s + 1) * 0.5
	t = (t + 1) * 0.5

	if s < T.sky_min {
		s = T.sky_min
	} else if s > T.sky_max {
		s = T.sky_max
	}

	if t < T.sky_min {
		t = T.sky_min
	} else if t > T.sky_max {
		t = T.sky_max
	}

	t = 1.0 - t

	T.tex_sky[T.index_tex] = s
	T.tex_sky[T.index_tex+1] = t

	T.vtx_sky[T.index_vtx+2] = v[0]
	T.vtx_sky[T.index_vtx+3] = v[1]
	T.vtx_sky[T.index_vtx+4] = v[2]
	T.index_tex += 5
}

func (T *qGl3) drawSkyBox() {

	// if T.skyrotate != 0 {
	// 	/* check for no sky at all */
	// 	for i := 0; i < 6; i++ {
	// 		if (T.skymins[0][i] < T.skymaxs[0][i]) &&
	// 			(T.skymins[1][i] < T.skymaxs[1][i]) {
	// 			break
	// 		}
	// 	}

	// 	if i == 6 {
	// 		return /* nothing visible */
	// 	}
	// }

	// gl.PushMatrix()
	// gl.Translatef(r_origin[0], r_origin[1], r_origin[2])
	// gl.Rotatef(r_newrefdef.time*skyrotate, skyaxis[0], skyaxis[1], skyaxis[2])

	// for i := 0; i < 6; i++ {
	// 	if T.skyrotate != 0 {
	// 		T.skymins[0][i] = -1
	// 		T.skymins[1][i] = -1
	// 		T.skymaxs[0][i] = 1
	// 		T.skymaxs[1][i] = 1
	// 	}

	// 	if (T.skymins[0][i] >= T.skymaxs[0][i]) ||
	// 		(T.skymins[1][i] >= T.skymaxs[1][i]) {
	// 		continue
	// 	}

	// 	T.bind(T.sky_images[skytexorder[i]].texnum)

	// 	gl.EnableClientState(gl.VERTEX_ARRAY)
	// 	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)

	// 	T.index_vtx = 0
	// 	T.index_tex = 0

	// 	T.makeSkyVec(T.skymins[0][i], T.skymins[1][i], i)
	// 	T.makeSkyVec(T.skymins[0][i], T.skymaxs[1][i], i)
	// 	T.makeSkyVec(T.skymaxs[0][i], T.skymaxs[1][i], i)
	// 	T.makeSkyVec(T.skymaxs[0][i], T.skymins[1][i], i)

	// 	gl.VertexPointer(3, GL_FLOAT, 0, vtx_sky)
	// 	gl.TexCoordPointer(2, GL_FLOAT, 0, tex_sky)
	// 	gl.DrawArrays(GL_TRIANGLE_FAN, 0, 4)

	// 	gl.DisableClientState(GL_VERTEX_ARRAY)
	// 	gl.DisableClientState(GL_TEXTURE_COORD_ARRAY)
	// }

	// glPopMatrix()
}
