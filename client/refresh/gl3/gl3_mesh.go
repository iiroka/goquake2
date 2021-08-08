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
 * Mesh handling
 *
 * =======================================================================
 */
package gl3

import (
	"goquake2/shared"
	"math"

	"github.com/go-gl/gl/v3.2-core/gl"
)

func lerpVerts(powerUpEffect bool, nverts int, v, ov,
	verts []shared.Dtrivertx_t, lerp [][4]float32, move,
	frontv, backv [3]float32) {

	if powerUpEffect {
		for i := 0; i < nverts; i++ {
			normal := shared.R_avertexnormals[verts[i].Lightnormalindex]
			lerp[i][0] = move[0] + float32(ov[i].V[0])*backv[0] + float32(v[i].V[0])*frontv[0] + normal[0]*shared.POWERSUIT_SCALE
			lerp[i][1] = move[1] + float32(ov[i].V[1])*backv[1] + float32(v[i].V[1])*frontv[1] + normal[1]*shared.POWERSUIT_SCALE
			lerp[i][2] = move[2] + float32(ov[i].V[2])*backv[2] + float32(v[i].V[2])*frontv[2] + normal[2]*shared.POWERSUIT_SCALE
		}
	} else {
		for i := 0; i < nverts; i++ {
			lerp[i][0] = move[0] + float32(ov[i].V[0])*backv[0] + float32(v[i].V[0])*frontv[0]
			lerp[i][1] = move[1] + float32(ov[i].V[1])*backv[1] + float32(v[i].V[1])*frontv[1]
			lerp[i][2] = move[2] + float32(ov[i].V[2])*backv[2] + float32(v[i].V[2])*frontv[2]
		}
	}
}

/*
 * Interpolates between two frames and origins
 */
func (T *qGl3) drawAliasFrameLerp(aliasextra aliasExtra, entity *shared.Entity_t, shadelight [3]float32) {
	backlerp := entity.Backlerp
	frontlerp := 1.0 - backlerp
	// draw without texture? used for quad damage effect etc, I think
	colorOnly := (entity.Flags &
		(shared.RF_SHELL_RED | shared.RF_SHELL_GREEN | shared.RF_SHELL_BLUE | shared.RF_SHELL_DOUBLE |
			shared.RF_SHELL_HALF_DAM)) != 0

	// TODO: maybe we could somehow store the non-rotated normal and do the dot in shader?
	shadedots := shared.R_avertexnormal_dots[((int)(entity.Angles[1]*
		(shared.SHADEDOT_QUANT/360.0)))&(shared.SHADEDOT_QUANT-1)]

	frame := aliasextra.frames[entity.Frame]
	verts := frame.Verts

	oldframe := aliasextra.frames[entity.Oldframe]
	ov := oldframe.Verts

	var alpha float32
	if (entity.Flags & shared.RF_TRANSLUCENT) != 0 {
		alpha = entity.Alpha * 0.666
	} else {
		alpha = 1.0
	}

	if colorOnly {
		T.useProgram(T.gl3state.si3DaliasColor.shaderProgram)
	} else {
		T.useProgram(T.gl3state.si3Dalias.shaderProgram)
	}

	/* move should be the delta back to the previous frame * backlerp */
	var delta [3]float32
	shared.VectorSubtract(entity.Oldorigin[:], entity.Origin[:], delta[:])
	var vectors [3][3]float32
	shared.AngleVectors(entity.Angles[:], vectors[0][:], vectors[1][:], vectors[2][:])

	var move [3]float32
	move[0] = shared.DotProduct(delta[:], vectors[0][:])  /* forward */
	move[1] = -shared.DotProduct(delta[:], vectors[1][:]) /* left */
	move[2] = shared.DotProduct(delta[:], vectors[2][:])  /* up */

	shared.VectorAdd(move[:], oldframe.Translate[:], move[:])

	var frontv [3]float32
	var backv [3]float32
	for i := 0; i < 3; i++ {
		move[i] = backlerp*move[i] + frontlerp*frame.Translate[i]

		frontv[i] = frontlerp * frame.Scale[i]
		backv[i] = backlerp * oldframe.Scale[i]
	}

	var lerp [shared.MAX_VERTS][4]float32

	lerpVerts(colorOnly, int(aliasextra.header.Num_xyz), verts, ov, verts, lerp[:], move, frontv, backv)

	// all the triangle fans and triangle strips of this model will be converted to
	// just triangles: the vertices stay the same and are batched in vtxBuf,
	// but idxBuf will contain indices to draw them all as GL_TRIANGLE
	// this way there's only one draw call (and two glBufferData() calls)
	// instead of (at least) dozens. *greatly* improves performance.

	// so first clear out the data from last call to this function
	// (the buffers are static global so we don't have malloc()/free() for each rendered model)
	vtxBuf := make([]float32, gl3_alias_vtx_size*aliasextra.vertexCount)
	indxBuf := make([]uint16, aliasextra.indexCount)
	vtxIndx := 0
	indxIndx := 0

	index := 0
	for {
		nextVtxIdx := uint16(vtxIndx)

		/* get the vertex count and primitive type */
		count := aliasextra.glcmds[index]
		index++
		if count == 0 {
			break /* done */
		}

		ttype := gl.TRIANGLE_STRIP
		if count < 0 {
			count = -count
			ttype = gl.TRIANGLE_FAN
		}

		if colorOnly {
			for i := 0; i < int(count); i++ {
				cur := gl3_alias_vtx_t{vtxBuf[vtxIndx*gl3_alias_vtx_size:]}
				index_xyz := aliasextra.glcmds[index+2]
				vtxIndx++
				index += 3

				cur.setPos(lerp[index_xyz][:])
				cur.setColorAlpha(shadelight[:], alpha)
			}
		} else {
			for i := 0; i < int(count); i++ {
				cur := gl3_alias_vtx_t{vtxBuf[vtxIndx*gl3_alias_vtx_size:]}
				/* texture coordinates come from the draw list */
				cur.setTexCoord(
					math.Float32frombits(uint32(aliasextra.glcmds[index+0])),
					math.Float32frombits(uint32(aliasextra.glcmds[index+1])))

				index_xyz := aliasextra.glcmds[index+2]

				vtxIndx++
				index += 3

				/* normals and vertexes come from the frame list */
				// shadedots is set above according to rotation (around Z axis I think)
				// to one of 16 (SHADEDOT_QUANT) presets in r_avertexnormal_dots
				l := shadedots[verts[index_xyz].Lightnormalindex]

				cur.setPos(lerp[index_xyz][:])
				var light [3]float32
				for j := 0; j < 3; j++ {
					light[j] = l * shadelight[j]
				}
				cur.setColorAlpha(light[:], alpha)
			}
		}

		// translate triangle fan/strip to just triangle indices
		if ttype == gl.TRIANGLE_FAN {
			for i := 1; i < int(count-1); i++ {
				indxBuf[indxIndx] = nextVtxIdx
				indxBuf[indxIndx+1] = nextVtxIdx + uint16(i)
				indxBuf[indxIndx+2] = nextVtxIdx + uint16(i) + 1
				indxIndx += 3
			}
		} else { // triangle strip
			i := 1
			for ; i < int(count)-2; i += 2 {
				// add two triangles at once, because the vertex order is different
				// for odd vs even triangles
				add := indxBuf[indxIndx:]

				add[0] = nextVtxIdx + uint16(i) - 1
				add[1] = nextVtxIdx + uint16(i)
				add[2] = nextVtxIdx + uint16(i) + 1

				add[3] = nextVtxIdx + uint16(i)
				add[4] = nextVtxIdx + uint16(i) + 2
				add[5] = nextVtxIdx + uint16(i) + 1

				indxIndx += 6
			}
			// add remaining triangle, if any
			if i < int(count)-1 {
				add := indxBuf[indxIndx:]

				add[0] = nextVtxIdx + uint16(i) - 1
				add[1] = nextVtxIdx + uint16(i)
				add[2] = nextVtxIdx + uint16(i) + 1

				indxIndx += 3
			}
		}
	}

	T.bindVAO(T.gl3state.vaoAlias)
	T.bindVBO(T.gl3state.vboAlias)

	gl.BufferData(gl.ARRAY_BUFFER, int(vtxIndx)*gl3_alias_vtx_size*4, gl.Ptr(vtxBuf), gl.STREAM_DRAW)
	T.bindEBO(T.gl3state.eboAlias)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indxIndx*2, gl.Ptr(indxBuf), gl.STREAM_DRAW)
	gl.DrawElements(gl.TRIANGLES, int32(indxIndx), gl.UNSIGNED_SHORT, nil)
}

func (T *qGl3) cullAliasModel(bbox [8][3]float32, e *shared.Entity_t) bool {

	model := e.Model.(*gl3model_t)

	extra := model.extradata.(aliasExtra)
	paliashdr := &extra.header

	if (e.Frame >= int(paliashdr.Num_frames)) || (e.Frame < 0) {
		T.rPrintf(shared.PRINT_DEVELOPER, "R_CullAliasModel %s: no such frame %d\n",
			model.name, e.Frame)
		e.Frame = 0
	}

	if (e.Oldframe >= int(paliashdr.Num_frames)) || (e.Oldframe < 0) {
		T.rPrintf(shared.PRINT_DEVELOPER, "R_CullAliasModel %s: no such oldframe %d\n",
			model.name, e.Oldframe)
		e.Oldframe = 0
	}

	pframe := &extra.frames[e.Frame]

	poldframe := &extra.frames[e.Oldframe]

	/* compute axially aligned mins and maxs */
	var mins [3]float32
	var maxs [3]float32
	if pframe == poldframe {
		for i := 0; i < 3; i++ {
			mins[i] = pframe.Translate[i]
			maxs[i] = mins[i] + pframe.Scale[i]*255
		}
	} else {
		var thismins [3]float32
		var thismaxs [3]float32
		var oldmins [3]float32
		var oldmaxs [3]float32

		for i := 0; i < 3; i++ {
			thismins[i] = pframe.Translate[i]
			thismaxs[i] = thismins[i] + pframe.Scale[i]*255

			oldmins[i] = poldframe.Translate[i]
			oldmaxs[i] = oldmins[i] + poldframe.Scale[i]*255

			if thismins[i] < oldmins[i] {
				mins[i] = thismins[i]
			} else {
				mins[i] = oldmins[i]
			}

			if thismaxs[i] > oldmaxs[i] {
				maxs[i] = thismaxs[i]
			} else {
				maxs[i] = oldmaxs[i]
			}
		}
	}

	/* compute a full bounding box */
	for i := 0; i < 8; i++ {
		var tmp [3]float32

		if (i & 1) != 0 {
			tmp[0] = mins[0]
		} else {
			tmp[0] = maxs[0]
		}

		if (i & 2) != 0 {
			tmp[1] = mins[1]
		} else {
			tmp[1] = maxs[1]
		}

		if (i & 4) != 0 {
			tmp[2] = mins[2]
		} else {
			tmp[2] = maxs[2]
		}

		copy(bbox[i][:], tmp[:])
	}

	/* rotate the bounding box */
	var vectors [3][3]float32
	var angles [3]float32
	copy(angles[:], e.Angles[:])
	angles[shared.YAW] = -angles[shared.YAW]
	shared.AngleVectors(angles[:], vectors[0][:], vectors[1][:], vectors[2][:])

	for i := 0; i < 8; i++ {
		var tmp [3]float32

		copy(tmp[:], bbox[i][:])

		bbox[i][0] = shared.DotProduct(vectors[0][:], tmp[:])
		bbox[i][1] = -shared.DotProduct(vectors[1][:], tmp[:])
		bbox[i][2] = shared.DotProduct(vectors[2][:], tmp[:])

		shared.VectorAdd(e.Origin[:], bbox[i][:], bbox[i][:])
	}

	// int p, f, aggregatemask = ~0;
	aggregatemask := 0xFFFFFFFF

	for p := 0; p < 8; p++ {
		mask := 0

		for f := 0; f < 4; f++ {
			dp := shared.DotProduct(T.frustum[f].Normal[:], bbox[p][:])

			if (dp - T.frustum[f].Dist) < 0 {
				mask |= (1 << f)
			}
		}

		aggregatemask &= mask
	}

	if aggregatemask != 0 {
		return true
	}

	return false
}

func (T *qGl3) drawAliasModel(entity *shared.Entity_t) {

	var bbox [8][3]float32
	if (entity.Flags & shared.RF_WEAPONMODEL) == 0 {
		if T.cullAliasModel(bbox, entity) {
			return
		}
	}

	if (entity.Flags & shared.RF_WEAPONMODEL) != 0 {
		if T.gl_lefthand.Int() == 2 {
			return
		}
	}

	model := entity.Model.(*gl3model_t)
	aliasExtra := model.extradata.(aliasExtra)
	paliashdr := &aliasExtra.header

	/* get lighting information */
	var shadelight [3]float32
	if (entity.Flags &
		(shared.RF_SHELL_HALF_DAM | shared.RF_SHELL_GREEN | shared.RF_SHELL_RED |
			shared.RF_SHELL_BLUE | shared.RF_SHELL_DOUBLE)) != 0 {

		if (entity.Flags & shared.RF_SHELL_HALF_DAM) != 0 {
			shadelight[0] = 0.56
			shadelight[1] = 0.59
			shadelight[2] = 0.45
		}

		if (entity.Flags & shared.RF_SHELL_DOUBLE) != 0 {
			shadelight[0] = 0.9
			shadelight[1] = 0.7
		}

		if (entity.Flags & shared.RF_SHELL_RED) != 0 {
			shadelight[0] = 1.0
		}

		if (entity.Flags & shared.RF_SHELL_GREEN) != 0 {
			shadelight[1] = 1.0
		}

		if (entity.Flags & shared.RF_SHELL_BLUE) != 0 {
			shadelight[2] = 1.0
		}
	} else if (entity.Flags & shared.RF_FULLBRIGHT) != 0 {
		for i := 0; i < 3; i++ {
			shadelight[i] = 1.0
		}
	} else {
		T.lightPoint(entity.Origin[:], shadelight[:])

		// 	/* player lighting hack for communication back to server */
		if (entity.Flags & shared.RF_WEAPONMODEL) != 0 {
			/* pick the greatest component, which should be
			   the same as the mono value returned by software */
			// 		if (shadelight[0] > shadelight[1])
			// 		{
			// 			if (shadelight[0] > shadelight[2])
			// 			{
			// 				r_lightlevel->value = 150 * shadelight[0];
			// 			}
			// 			else
			// 			{
			// 				r_lightlevel->value = 150 * shadelight[2];
			// 			}
			// 		}
			// 		else
			// 		{
			// 			if (shadelight[1] > shadelight[2])
			// 			{
			// 				r_lightlevel->value = 150 * shadelight[1];
			// 			}
			// 			else
			// 			{
			// 				r_lightlevel->value = 150 * shadelight[2];
			// 			}
			// 		}
		}
	}

	if (entity.Flags & shared.RF_MINLIGHT) != 0 {
		found := false
		for i := 0; i < 3; i++ {
			if shadelight[i] > 0.1 {
				found = true
				break
			}
		}

		if !found {
			shadelight[0] = 0.1
			shadelight[1] = 0.1
			shadelight[2] = 0.1
		}
	}

	if (entity.Flags & shared.RF_GLOW) != 0 {
		/* bonus items will pulse with time */

		scale := float32(0.1 * math.Sin(float64(T.gl3_newrefdef.Time*7)))

		for i := 0; i < 3; i++ {
			min := shadelight[i] * 0.8
			shadelight[i] += scale

			if shadelight[i] < min {
				shadelight[i] = min
			}
		}
	}

	// Note: gl_overbrightbits are now applied in shader.

	/* ir goggles color override */
	// if ((gl3_newrefdef.rdflags & RDF_IRGOGGLES) && (entity->flags & RF_IR_VISIBLE))
	// {
	// 	shadelight[0] = 1.0;
	// 	shadelight[1] = 0.0;
	// 	shadelight[2] = 0.0;
	// }

	// an = entity->angles[1] / 180 * M_PI;
	// shadevector[0] = cos(-an);
	// shadevector[1] = sin(-an);
	// shadevector[2] = 1;
	// VectorNormalize(shadevector);

	/* locate the proper data */
	T.c_alias_polys += int(paliashdr.Num_tris)

	// /* draw all the triangles */
	// if (entity->flags & RF_DEPTHHACK) != 0 {
	// 	/* hack the depth range to prevent view model from poking into walls */
	// 	glDepthRange(gl3depthmin, gl3depthmin + 0.3 * (gl3depthmax - gl3depthmin));
	// }

	var origProjMat []float32
	if (entity.Flags & shared.RF_WEAPONMODEL) != 0 {
		// 	extern hmm_mat4 GL3_MYgluPerspective(GLdouble fovy, GLdouble aspect, GLdouble zNear, GLdouble zFar);

		origProjMat = T.gl3state.uni3DData.getTransProjMat4()

		// render weapon with a different FOV (r_gunfov) so it's not distorted at high view FOV
		screenaspect := float32(T.gl3_newrefdef.Width) / float32(T.gl3_newrefdef.Height)
		var dist float32 = 4096.0
		if T.r_farsee.Bool() {
			dist = 8192.0
		}

		if T.r_gunfov.Int() < 0 {
			T.gl3state.uni3DData.setTransProjMat4(GL3_MYgluPerspective(T.gl3_newrefdef.Fov_y, screenaspect, 4, dist))
		} else {
			T.gl3state.uni3DData.setTransProjMat4(GL3_MYgluPerspective(T.r_gunfov.Float(), screenaspect, 4, dist))
		}

		if T.gl_lefthand.Int() == 1 {
			// to mirror gun so it's rendered left-handed, just invert X-axis column
			// of projection matrix
			mat := T.gl3state.uni3DData.getTransProjMat4()
			for i := 0; i < 4; i++ {
				mat[i] = -mat[i]
			}
			T.gl3state.uni3DData.setTransProjMat4(mat)
			//GL3_UpdateUBO3D(); Note: GL3_RotateForEntity() will call this,no need to do it twice before drawing

			gl.CullFace(gl.BACK)
		}
	}

	//glPushMatrix();
	origModelMat := T.gl3state.uni3DData.getTransModelMat4()

	entity.Angles[shared.PITCH] = -entity.Angles[shared.PITCH]
	T.rotateForEntity(entity)
	entity.Angles[shared.PITCH] = -entity.Angles[shared.PITCH]

	/* select skin */
	var skin *gl3image_t
	if entity.Skin != nil {
		skin = entity.Skin.(*gl3image_t) /* custom player skin */
	} else {
		if entity.Skinnum >= len(model.skins) {
			skin = model.skins[0]
		} else {
			skin = model.skins[entity.Skinnum]
			if skin == nil {
				skin = model.skins[0]
			}
		}
	}

	if skin == nil {
		skin = T.gl3_notexture /* fallback... */
	}

	T.bind(skin.texnum)

	if (entity.Flags & shared.RF_TRANSLUCENT) != 0 {
		gl.Enable(gl.BLEND)
	}

	// if ((entity->frame >= paliashdr->num_frames) ||
	// 	(entity->frame < 0))
	// {
	// 	R_Printf(PRINT_DEVELOPER, "R_DrawAliasModel %s: no such frame %d\n",
	// 			model->name, entity->frame);
	// 	entity->frame = 0;
	// 	entity->oldframe = 0;
	// }

	// if ((entity->oldframe >= paliashdr->num_frames) ||
	// 	(entity->oldframe < 0))
	// {
	// 	R_Printf(PRINT_DEVELOPER, "R_DrawAliasModel %s: no such oldframe %d\n",
	// 			model->name, entity->oldframe);
	// 	entity->frame = 0;
	// 	entity->oldframe = 0;
	// }

	T.drawAliasFrameLerp(aliasExtra, entity, shadelight)

	//glPopMatrix();
	T.gl3state.uni3DData.setTransModelMat4(origModelMat)
	T.updateUBO3D()

	if (entity.Flags & shared.RF_WEAPONMODEL) != 0 {
		T.gl3state.uni3DData.setTransProjMat4(origProjMat)
		T.updateUBO3D()
		if T.gl_lefthand.Int() == 1 {
			gl.CullFace(gl.FRONT)
		}
	}

	if (entity.Flags & shared.RF_TRANSLUCENT) != 0 {
		gl.Disable(gl.BLEND)
	}

	// if (entity->flags & RF_DEPTHHACK) != 0 {
	// 	glDepthRange(gl3depthmin, gl3depthmax);
	// }

	// if (gl_shadows->value && gl3config.stencil && !(entity->flags & (RF_TRANSLUCENT | RF_WEAPONMODEL | RF_NOSHADOW)))
	// {
	// 	gl3_shadowinfo_t si = {0};
	// 	VectorCopy(lightspot, si.lightspot);
	// 	VectorCopy(shadevector, si.shadevector);
	// 	si.paliashdr = paliashdr;
	// 	si.entity = entity;

	// 	da_push(shadowModels, si);
	// }
}
