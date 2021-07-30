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
	"goquake2/shared"

	"github.com/go-gl/gl/v3.2-core/gl"
)

const BACKFACE_EPSILON = 0.01

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

/*
 * Returns true if the box is completely outside the frustom
 */
func (T *qGl3) cullBox(mins, maxs []float32) bool {

	if !T.gl_cull.Bool() {
		return false
	}

	for i := 0; i < 4; i++ {
		if shared.BoxOnPlaneSide(mins, maxs, &T.frustum[i]) == 2 {
			return true
		}
	}

	return false
}

/*
 * Returns the proper texture for a given time and base texture
 */
func (T *qGl3) textureAnimation(tex *mtexinfo_t) *gl3image_t {

	if tex.next == nil {
		return tex.image
	}

	c := T.currententity.Frame % tex.numframes

	for c > 0 {
		tex = tex.next
		c--
	}

	return tex.image
}

func setAllLightFlags(surf *msurface_t) {
	var lightFlags uint32 = 0xffffffff

	numVerts := surf.polys.numverts
	for i := 0; i < numVerts; i++ {
		surf.polys.vertices(i).setLightFlags(lightFlags)
	}
}

func (T *qGl3) drawGLPoly(fa *msurface_t) {
	p := fa.polys

	T.bindVAO(T.gl3state.vao3D)
	T.bindVBO(T.gl3state.vbo3D)

	T.bufferAndDraw3D(gl.Ptr(p.verticesData), p.numverts, gl.TRIANGLE_FAN)
}

func (T *qGl3) renderLightmappedPoly(surf *msurface_t) {
	// int map;
	image := T.textureAnimation(surf.texinfo)

	// hmm_vec4 lmScales[MAX_LIGHTMAPS_PER_SURFACE] = {0};
	// lmScales[0] = HMM_Vec4(1.0f, 1.0f, 1.0f, 1.0f);

	// assert((surf->texinfo->flags & (SURF_SKY | SURF_TRANS33 | SURF_TRANS66 | SURF_WARP)) == 0
	// 		&& "RenderLightMappedPoly mustn't be called with transparent, sky or warping surfaces!");

	// Any dynamic lights on this surface?
	for mmap := 0; mmap < MAX_LIGHTMAPS_PER_SURFACE; mmap++ {
		if surf.styles[mmap] == 255 {
			break
		}
		// lmScales[mmap].R = T.gl3_newrefdef.Lightstyles[surf.styles[mmap]].Rgb[0]
		// lmScales[mmap].G = T.gl3_newrefdef.Lightstyles[surf.styles[mmap]].Rgb[1]
		// lmScales[mmap].B = T.gl3_newrefdef.Lightstyles[surf.styles[mmap]].Rgb[2]
		// lmScales[mmap].A = 1.0
	}

	T.c_brush_polys++

	T.bind(image.texnum)
	T.bindLightmap(surf.lightmaptexturenum)

	// if (surf->texinfo->flags & SURF_FLOWING)
	// {
	// 	GL3_UseProgram(gl3state.si3DlmFlow.shaderProgram);
	// 	UpdateLMscales(lmScales, &gl3state.si3DlmFlow);
	// 	GL3_DrawGLFlowingPoly(surf);
	// }
	// else
	// {
	T.useProgram(T.gl3state.si3Dlm.shaderProgram)
	// 	UpdateLMscales(lmScales, &gl3state.si3Dlm);
	T.drawGLPoly(surf)
	// }
}

func (T *qGl3) drawInlineBModel() {
	// int i, k;
	// cplane_t *pplane;
	// float dot;
	// msurface_t *psurf;
	// dlight_t *lt;

	// /* calculate dynamic lighting for bmodel */
	// lt = gl3_newrefdef.dlights;

	// for (k = 0; k < gl3_newrefdef.num_dlights; k++, lt++)
	// {
	// 	GL3_MarkLights(lt, 1 << k, currentmodel->nodes + currentmodel->firstnode);
	// }

	// psurf = &currentmodel->surfaces[currentmodel->firstmodelsurface];

	// if (currententity->flags & RF_TRANSLUCENT)
	// {
	// 	glEnable(GL_BLEND);
	// 	/* TODO: should I care about the 0.25 part? we'll just set alpha to 0.33 or 0.66 depending on surface flag..
	// 	glColor4f(1, 1, 1, 0.25);
	// 	R_TexEnv(GL_MODULATE);
	// 	*/
	// }

	/* draw texture */
	for i := 0; i < T.currentmodel.nummodelsurfaces; i++ {
		/* find which side of the node we are on */
		psurf := &T.currentmodel.surfaces[T.currentmodel.firstmodelsurface+i]
		pplane := psurf.plane

		dot := shared.DotProduct(T._surf_modelorg[:], pplane.Normal[:]) - pplane.Dist

		// 	/* draw the polygon */
		if ((psurf.flags&SURF_PLANEBACK) != 0 && (dot < -BACKFACE_EPSILON)) ||
			((psurf.flags&SURF_PLANEBACK) == 0 && (dot > BACKFACE_EPSILON)) {
			if (psurf.texinfo.flags & (shared.SURF_TRANS33 | shared.SURF_TRANS66)) != 0 {
				// 			/* add to the translucent chain */
				println("Draw trans")
				// 			psurf->texturechain = gl3_alpha_surfaces;
				// 			gl3_alpha_surfaces = psurf;
			} else if (psurf.flags & SURF_DRAWTURB) == 0 {
				setAllLightFlags(psurf)
				T.renderLightmappedPoly(psurf)
			} else {
				println("Draw brush")
				// 			RenderBrushPoly(psurf);
			}
		}
	}

	// if (currententity->flags & RF_TRANSLUCENT)
	// {
	// 	glDisable(GL_BLEND);
	// }
}

func (T *qGl3) drawBrushModel(e *shared.Entity_t) {
	// vec3_t mins, maxs;
	// int i;
	// qboolean rotated;

	if T.currentmodel.nummodelsurfaces == 0 {
		return
	}

	T.currententity = e
	T.gl3state.currenttexture = 0xFFFFFFFF

	var mins [3]float32
	var maxs [3]float32
	var rotated bool

	if e.Angles[0] != 0 || e.Angles[1] != 0 || e.Angles[2] != 0 {
		rotated = true

		for i := 0; i < 3; i++ {
			mins[i] = e.Origin[i] - T.currentmodel.radius
			maxs[i] = e.Origin[i] + T.currentmodel.radius
		}
	} else {
		rotated = false
		shared.VectorAdd(e.Origin[:], T.currentmodel.mins[:], mins[:])
		shared.VectorAdd(e.Origin[:], T.currentmodel.maxs[:], maxs[:])
	}

	if T.cullBox(mins[:], maxs[:]) {
		return
	}

	// if (gl_zfix->value) {
	// 	glEnable(GL_POLYGON_OFFSET_FILL);
	// }

	shared.VectorSubtract(T.gl3_newrefdef.Vieworg[:], e.Origin[:], T._surf_modelorg[:])

	if rotated {
		// 	vec3_t temp;
		// 	vec3_t forward, right, up;

		// 	VectorCopy(T._surf_modelorg[:], temp);
		// 	AngleVectors(e->angles, forward, right, up);
		// 	T._surf_modelorg[0] = DotProduct(temp, forward);
		// 	T._surf_modelorg[1] = -DotProduct(temp, right);
		// 	T._surf_modelorg[2] = DotProduct(temp, up);
	}

	// //glPushMatrix();
	oldMat := T.gl3state.uni3DData.getTransModelMat4()

	e.Angles[0] = -e.Angles[0]
	e.Angles[2] = -e.Angles[2]
	// GL3_RotateForEntity(e);
	e.Angles[0] = -e.Angles[0]
	e.Angles[2] = -e.Angles[2]

	T.drawInlineBModel()

	// glPopMatrix();
	T.gl3state.uni3DData.setTransModelMat4(oldMat)
	T.updateUBO3D()

	// if (gl_zfix->value) {
	// 	glDisable(GL_POLYGON_OFFSET_FILL);
	// }
}

func (T *qGl3) recursiveWorldNode(anode mnode_or_leaf) {

	if anode.Contents() == shared.CONTENTS_SOLID {
		return /* solid */
	}

	if anode.Visframe() != T.gl3_visframecount {
		return
	}

	if T.cullBox(anode.Minmaxs(), anode.Minmaxs()[3:]) {
		return
	}

	/* if a leaf node, draw stuff */
	if anode.Contents() != -1 {
		// pleaf := anode.(*mleaf_t)

		// 		/* check for door connected areas */
		// 		if (gl3_newrefdef.areabits)
		// 		{
		// 			if (!(gl3_newrefdef.areabits[pleaf->area >> 3] & (1 << (pleaf->area & 7))))
		// 			{
		// 				return; /* not visible */
		// 			}
		// 		}

		// 		mark = pleaf->firstmarksurface;
		// 		c = pleaf->nummarksurfaces;

		// 		if (c)
		// 		{
		// 			do
		// 			{
		// 				(*mark)->visframe = gl3_framecount;
		// 				mark++;
		// 			}
		// 			while (--c);
		// 		}

		return
	}

	node := anode.(*mnode_t)
	/* node is just a decision point, so go down the apropriate
	   sides find which side of the node we are on */
	plane := node.plane

	var dot float32
	switch plane.Type {
	case shared.PLANE_X:
		dot = T._surf_modelorg[0] - plane.Dist
		break
	case shared.PLANE_Y:
		dot = T._surf_modelorg[1] - plane.Dist
		break
	case shared.PLANE_Z:
		dot = T._surf_modelorg[2] - plane.Dist
		break
	default:
		dot = shared.DotProduct(T._surf_modelorg[:], plane.Normal[:]) - plane.Dist
		break
	}

	var side int
	var sidebit int
	if dot >= 0 {
		side = 0
		sidebit = 0
	} else {
		side = 1
		sidebit = SURF_PLANEBACK
	}

	/* recurse down the children, front side first */
	T.recursiveWorldNode(node.children[side])

	/* draw stuff */
	for c := 0; c < int(node.numsurfaces); c++ {
		surf := &T.gl3_worldmodel.surfaces[int(node.firstsurface)+c]
		if surf.visframe != T.gl3_framecount {
			continue
		}

		if (surf.flags & SURF_PLANEBACK) != sidebit {
			continue /* wrong side */
		}

		if (surf.texinfo.flags & shared.SURF_SKY) != 0 {
			// 			/* just adds to visible sky bounds */
			println("Draw sky")
			// 			GL3_AddSkySurface(surf);
		} else if (surf.texinfo.flags & (shared.SURF_TRANS33 | shared.SURF_TRANS66)) != 0 {
			// 			/* add to the translucent chain */
			println("Draw trans surface")
			// 			surf->texturechain = gl3_alpha_surfaces;
			// 			gl3_alpha_surfaces = surf;
			// 			gl3_alpha_surfaces->texinfo->image = TextureAnimation(surf->texinfo);
		} else {
			// calling RenderLightmappedPoly() here probably isn't optimal, rendering everything
			// through texturechains should be faster, because far less glBindTexture() is needed
			// (and it might allow batching the drawcalls of surfaces with the same texture)
			/* the polygon is visible, so add it to the texture sorted chain */
			println("Draw surface")
			// 				image = TextureAnimation(surf->texinfo);
			// 				surf->texturechain = image->texturechain;
			// 				image->texturechain = surf;
		}
	}

	/* recurse down the back side */
	T.recursiveWorldNode(node.children[side^1])
}

func (T *qGl3) drawWorld() {
	// entity_t ent;

	if !T.r_drawworld.Bool() {
		return
	}

	if (T.gl3_newrefdef.Rdflags & shared.RDF_NOWORLDMODEL) != 0 {
		return
	}

	T.currentmodel = T.gl3_worldmodel

	copy(T._surf_modelorg[:], T.gl3_newrefdef.Vieworg[:])

	/* auto cycle the world frame for texture animation */
	ent := shared.Entity_t{}
	ent.Frame = int(T.gl3_newrefdef.Time * 2)
	T.currententity = &ent

	T.gl3state.currenttexture = 0xFFFFFFFF

	// GL3_ClearSkyBox();
	T.recursiveWorldNode(&T.gl3_worldmodel.nodes[0])
	// DrawTextureChains();
	// GL3_DrawSkyBox();
	// DrawTriangleOutlines();

	T.currententity = nil
}
