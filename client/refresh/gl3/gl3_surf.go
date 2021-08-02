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

func (T *qGl3) setLightFlags(surf *msurface_t) {
	lightFlags := 0
	if surf.dlightframe == T.gl3_framecount {
		lightFlags = surf.dlightbits
	}

	numVerts := surf.polys.numverts
	for i := 0; i < numVerts; i++ {
		surf.polys.vertices(i).setLightFlags(uint32(lightFlags))
	}
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

func (T *qGl3) updateLMscales(lmScales [MAX_LIGHTMAPS_PER_SURFACE][4]float32, si *gl3ShaderInfo_t) {
	hasChanged := false

	for i := 0; i < MAX_LIGHTMAPS_PER_SURFACE; i++ {
		if hasChanged {
			copy(si.lmScales[i*4:(i+1)*4], lmScales[i][:])
		} else if si.lmScales[i*4+0] != lmScales[i][0] ||
			si.lmScales[i*4+1] != lmScales[i][1] ||
			si.lmScales[i*4+2] != lmScales[i][2] ||
			si.lmScales[i*4+3] != lmScales[i][3] {
			copy(si.lmScales[i*4:(i+1)*4], lmScales[i][:])
			hasChanged = true
		}
	}

	if hasChanged {
		gl.Uniform4fv(si.uniLmScales, MAX_LIGHTMAPS_PER_SURFACE, (*float32)(gl.Ptr(si.lmScales[:])))
	}
}

func (T *qGl3) renderBrushPoly(fa *msurface_t) {
	// int map;
	// gl3image_t *image;

	T.c_brush_polys++

	image := T.textureAnimation(fa.texinfo)

	if (fa.flags & SURF_DRAWTURB) != 0 {
		// 	GL3_Bind(image->texnum);

		// 	GL3_EmitWaterPolys(fa);

		println("SURF_DRAWTURB BrushPoly")
		return
	} else {
		T.bind(image.texnum)
	}

	var lmScales [MAX_LIGHTMAPS_PER_SURFACE][4]float32
	for j := range lmScales[0] {
		lmScales[0][j] = 1.0
	}

	T.bindLightmap(fa.lightmaptexturenum)

	// Any dynamic lights on this surface?
	for mmap := 0; mmap < MAX_LIGHTMAPS_PER_SURFACE; mmap++ {
		if fa.styles[mmap] == 255 {
			break
		}
		lmScales[mmap][0] = T.gl3_newrefdef.Lightstyles[fa.styles[mmap]].Rgb[0]
		lmScales[mmap][1] = T.gl3_newrefdef.Lightstyles[fa.styles[mmap]].Rgb[1]
		lmScales[mmap][2] = T.gl3_newrefdef.Lightstyles[fa.styles[mmap]].Rgb[2]
		lmScales[mmap][3] = 1.0
	}

	if (fa.texinfo.flags & shared.SURF_FLOWING) != 0 {
		// 	GL3_UseProgram(gl3state.si3DlmFlow.shaderProgram);
		// 	UpdateLMscales(lmScales, &gl3state.si3DlmFlow);
		// 	GL3_DrawGLFlowingPoly(fa);
		println("SURF_FLOWING")
	} else {
		T.useProgram(T.gl3state.si3Dlm.shaderProgram)
		T.updateLMscales(lmScales, &T.gl3state.si3Dlm)
		T.drawGLPoly(fa)
	}

	// Note: lightmap chains are gone, lightmaps are rendered together with normal texture in one pass
}

/*
 * Draw water surfaces and windows.
 * The BSP tree is waled front to back, so unwinding the chain
 * of alpha_surfaces will draw back to front, giving proper ordering.
 */
func (T *qGl3) drawAlphaSurfaces() {

	//  /* go back to the world matrix */
	T.gl3state.uni3DData.setTransModelMat4(gl3_identityMat4)
	T.updateUBO3D()

	gl.Enable(gl.BLEND)

	for s := T.gl3_alpha_surfaces; s != nil; s = s.texturechain {
		T.bind(s.texinfo.image.texnum)
		T.c_brush_polys++
		var alpha float32 = 1.0
		if (s.texinfo.flags & shared.SURF_TRANS33) != 0 {
			alpha = 0.333
		} else if (s.texinfo.flags & shared.SURF_TRANS66) != 0 {
			alpha = 0.666
		}
		if alpha != T.gl3state.uni3DData.getAlpha() {
			T.gl3state.uni3DData.setAlpha(alpha)
			T.updateUBO3D()
		}

		if (s.flags & SURF_DRAWTURB) != 0 {
			T.emitWaterPolys(s)
		} else if (s.texinfo.flags & shared.SURF_FLOWING) != 0 {
			println("SURF_FLOWING alpha")
			// 		 GL3_UseProgram(gl3state.si3DtransFlow.shaderProgram);
			// 		 GL3_DrawGLFlowingPoly(s);
		} else {
			T.useProgram(T.gl3state.si3Dtrans.shaderProgram)
			T.drawGLPoly(s)
		}
	}

	T.gl3state.uni3DData.setAlpha(1.0)
	T.updateUBO3D()

	gl.Disable(gl.BLEND)

	T.gl3_alpha_surfaces = nil
}

func (T *qGl3) drawTextureChains() {

	T.c_visible_textures = 0

	for i := 0; i < T.numgl3textures; i++ {
		image := &T.gl3textures[i]
		if image.registration_sequence == 0 {
			continue
		}

		s := image.texturechain
		if s == nil {
			continue
		}

		T.c_visible_textures++

		for s != nil {
			T.setLightFlags(s)
			T.renderBrushPoly(s)
			s = s.texturechain
		}

		image.texturechain = nil
	}

	// TODO: maybe one loop for normal faces and one for SURF_DRAWTURB ???
}

func (T *qGl3) renderLightmappedPoly(surf *msurface_t) {
	// int map;
	image := T.textureAnimation(surf.texinfo)

	var lmScales [MAX_LIGHTMAPS_PER_SURFACE][4]float32
	for j := range lmScales[0] {
		lmScales[0][j] = 1.0
	}

	// assert((surf->texinfo->flags & (SURF_SKY | SURF_TRANS33 | SURF_TRANS66 | SURF_WARP)) == 0
	// 		&& "RenderLightMappedPoly mustn't be called with transparent, sky or warping surfaces!");

	// Any dynamic lights on this surface?
	for mmap := 0; mmap < MAX_LIGHTMAPS_PER_SURFACE; mmap++ {
		if surf.styles[mmap] == 255 {
			break
		}
		lmScales[mmap][0] = T.gl3_newrefdef.Lightstyles[surf.styles[mmap]].Rgb[0]
		lmScales[mmap][1] = T.gl3_newrefdef.Lightstyles[surf.styles[mmap]].Rgb[1]
		lmScales[mmap][2] = T.gl3_newrefdef.Lightstyles[surf.styles[mmap]].Rgb[2]
		lmScales[mmap][3] = 1.0
	}

	T.c_brush_polys++

	T.bind(image.texnum)
	T.bindLightmap(surf.lightmaptexturenum)

	if (surf.texinfo.flags & shared.SURF_FLOWING) != 0 {
		println("SURF_FLOWING")
		// 	GL3_UseProgram(gl3state.si3DlmFlow.shaderProgram);
		// 	UpdateLMscales(lmScales, &gl3state.si3DlmFlow);
		// 	GL3_DrawGLFlowingPoly(surf);
	} else {
		T.useProgram(T.gl3state.si3Dlm.shaderProgram)
		T.updateLMscales(lmScales, &T.gl3state.si3Dlm)
		T.drawGLPoly(surf)
	}
}

func (T *qGl3) drawInlineBModel() {
	// int i, k;
	// cplane_t *pplane;
	// float dot;
	// msurface_t *psurf;
	// dlight_t *lt;

	// /* calculate dynamic lighting for bmodel */
	// lt = gl3_newrefdef.dlights;

	// for (k = 0; k < gl3_newrefdef.num_dlights; k++, lt++) {
	// 	GL3_MarkLights(lt, 1 << k, currentmodel->nodes + currentmodel->firstnode);
	// }

	if (T.currententity.Flags & shared.RF_TRANSLUCENT) != 0 {
		gl.Enable(gl.BLEND)
		/* TODO: should I care about the 0.25 part? we'll just set alpha to 0.33 or 0.66 depending on surface flag..
		glColor4f(1, 1, 1, 0.25);
		R_TexEnv(GL_MODULATE);
		*/
	}

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
				/* add to the translucent chain */
				psurf.texturechain = T.gl3_alpha_surfaces
				T.gl3_alpha_surfaces = psurf
			} else if (psurf.flags & SURF_DRAWTURB) == 0 {
				setAllLightFlags(psurf)
				T.renderLightmappedPoly(psurf)
			} else {
				T.renderBrushPoly(psurf)
			}
		}
	}

	if (T.currententity.Flags & shared.RF_TRANSLUCENT) != 0 {
		gl.Disable(gl.BLEND)
	}
}

func (T *qGl3) drawBrushModel(e *shared.Entity_t) {

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

	if T.gl_zfix.Bool() {
		gl.Enable(gl.POLYGON_OFFSET_FILL)
	}

	shared.VectorSubtract(T.gl3_newrefdef.Vieworg[:], e.Origin[:], T._surf_modelorg[:])

	if rotated {
		temp := make([]float32, 3)
		copy(temp, T._surf_modelorg[:])
		forward := make([]float32, 3)
		right := make([]float32, 3)
		up := make([]float32, 3)
		shared.AngleVectors(e.Angles[:], forward, right, up)
		T._surf_modelorg[0] = shared.DotProduct(temp, forward)
		T._surf_modelorg[1] = -shared.DotProduct(temp, right)
		T._surf_modelorg[2] = shared.DotProduct(temp, up)
	}

	//glPushMatrix();
	oldMat := T.gl3state.uni3DData.getTransModelMat4()

	e.Angles[0] = -e.Angles[0]
	e.Angles[2] = -e.Angles[2]
	T.rotateForEntity(e)
	e.Angles[0] = -e.Angles[0]
	e.Angles[2] = -e.Angles[2]

	T.drawInlineBModel()

	// glPopMatrix();
	T.gl3state.uni3DData.setTransModelMat4(oldMat)
	T.updateUBO3D()

	if T.gl_zfix.Bool() {
		gl.Disable(gl.POLYGON_OFFSET_FILL)
	}
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
		pleaf := anode.(*mleaf_t)

		/* check for door connected areas */
		if T.gl3_newrefdef.Areabits != nil {
			if (T.gl3_newrefdef.Areabits[pleaf.area>>3] & (1 << (pleaf.area & 7))) == 0 {
				return /* not visible */
			}
		}

		mark := pleaf.firstmarksurface

		for c := 0; c < pleaf.nummarksurfaces; c++ {
			mark[c].visframe = T.gl3_framecount
		}

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
			/* just adds to visible sky bounds */
			T.addSkySurface(surf)
		} else if (surf.texinfo.flags & (shared.SURF_TRANS33 | shared.SURF_TRANS66)) != 0 {
			/* add to the translucent chain */
			surf.texturechain = T.gl3_alpha_surfaces
			T.gl3_alpha_surfaces = surf
			T.gl3_alpha_surfaces.texinfo.image = T.textureAnimation(surf.texinfo)
		} else {
			// calling RenderLightmappedPoly() here probably isn't optimal, rendering everything
			// through texturechains should be faster, because far less glBindTexture() is needed
			// (and it might allow batching the drawcalls of surfaces with the same texture)
			/* the polygon is visible, so add it to the texture sorted chain */
			image := T.textureAnimation(surf.texinfo)
			surf.texturechain = image.texturechain
			image.texturechain = surf
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

	T.clearSkyBox()
	T.recursiveWorldNode(&T.gl3_worldmodel.nodes[0])
	T.drawTextureChains()
	T.drawSkyBox()
	// DrawTriangleOutlines();

	T.currententity = nil
}

/*
 * Mark the leaves and nodes that are
 * in the PVS for the current cluster
 */
func (T *qGl3) markLeaves() {
	//  byte *vis;
	//  YQ2_ALIGNAS_TYPE(int) byte fatvis[MAX_MAP_LEAFS / 8];
	//  mnode_t *node;
	//  int i, c;
	//  mleaf_t *leaf;
	//  int cluster;

	if (T.gl3_oldviewcluster == T.gl3_viewcluster) &&
		(T.gl3_oldviewcluster2 == T.gl3_viewcluster2) &&
		!T.r_novis.Bool() &&
		(T.gl3_viewcluster != -1) {
		return
	}

	/* development aid to let you run around
	and see exactly where the pvs ends */
	if T.r_lockpvs.Bool() {
		return
	}

	T.gl3_visframecount++
	T.gl3_oldviewcluster = T.gl3_viewcluster
	T.gl3_oldviewcluster2 = T.gl3_viewcluster2

	//  if (r_novis->value || (gl3_viewcluster == -1) || !gl3_worldmodel->vis) {
	// 	 /* mark everything */
	for i := 0; i < T.gl3_worldmodel.numleafs; i++ {
		T.gl3_worldmodel.leafs[i].visframe = T.gl3_visframecount
	}

	for i := 0; i < len(T.gl3_worldmodel.nodes); i++ {
		T.gl3_worldmodel.nodes[i].visframe = T.gl3_visframecount
	}

	return
	//  }

	//  vis = GL3_Mod_ClusterPVS(gl3_viewcluster, gl3_worldmodel);

	//  /* may have to combine two clusters because of solid water boundaries */
	//  if (gl3_viewcluster2 != gl3_viewcluster)
	//  {
	// 	 memcpy(fatvis, vis, (gl3_worldmodel->numleafs + 7) / 8);
	// 	 vis = GL3_Mod_ClusterPVS(gl3_viewcluster2, gl3_worldmodel);
	// 	 c = (gl3_worldmodel->numleafs + 31) / 32;

	// 	 for (i = 0; i < c; i++)
	// 	 {
	// 		 ((int *)fatvis)[i] |= ((int *)vis)[i];
	// 	 }

	// 	 vis = fatvis;
	//  }

	//  for (i = 0, leaf = gl3_worldmodel->leafs;
	// 	  i < gl3_worldmodel->numleafs;
	// 	  i++, leaf++)
	//  {
	// 	 cluster = leaf->cluster;

	// 	 if (cluster == -1)
	// 	 {
	// 		 continue;
	// 	 }

	// 	 if (vis[cluster >> 3] & (1 << (cluster & 7)))
	// 	 {
	// 		 node = (mnode_t *)leaf;

	// 		 do
	// 		 {
	// 			 if (node->visframe == gl3_visframecount)
	// 			 {
	// 				 break;
	// 			 }

	// 			 node->visframe = gl3_visframecount;
	// 			 node = node->parent;
	// 		 }
	// 		 while (node);
	// 	 }
	//  }
}
