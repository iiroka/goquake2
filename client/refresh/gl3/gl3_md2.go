/*
 * Copyright (C) 1997-2001 Id Software, Inc.
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
 * MD2 file format
 *
 * =======================================================================
 */
package gl3

import (
	"goquake2/shared"
)

type aliasExtra struct {
	header      shared.Dmdl_t
	sts         []shared.Dstvert_t
	tris        []shared.Dtriangle_t
	frames      []shared.Daliasframe_t
	glcmds      []int32
	skinNames   []string
	vertexCount int
	indexCount  int
}

func (T *qGl3) loadMD2(mod *gl3model_t, buffer []byte) error {

	pheader := shared.Dmdl(buffer)

	if pheader.Version != shared.ALIAS_VERSION {
		return T.ri.Sys_Error(shared.ERR_DROP, "%s has wrong version number (%i should be %i)",
			mod.name, pheader.Version, shared.ALIAS_VERSION)
	}

	if pheader.Ofs_end < 0 || int(pheader.Ofs_end) > len(buffer) {
		return T.ri.Sys_Error(shared.ERR_DROP, "model %s file size(%d) too small, should be %d", mod.name,
			len(buffer), pheader.Ofs_end)
	}

	if pheader.Skinheight > MAX_LBM_HEIGHT {
		return T.ri.Sys_Error(shared.ERR_DROP, "model %s has a skin taller than %d", mod.name,
			MAX_LBM_HEIGHT)
	}

	if pheader.Num_xyz <= 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "model %s has no vertices", mod.name)
	}

	if pheader.Num_xyz > shared.MAX_VERTS {
		return T.ri.Sys_Error(shared.ERR_DROP, "model %s has too many vertices", mod.name)
	}

	if pheader.Num_st <= 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "model %s has no st vertices", mod.name)
	}

	if pheader.Num_tris <= 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "model %s has no triangles", mod.name)
	}

	if pheader.Num_frames <= 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "model %s has no frames", mod.name)
	}

	/* load base s and t vertices (not used in gl version) */
	psts := make([]shared.Dstvert_t, pheader.Num_st)
	for i := range psts {
		psts[i] = shared.Dstvert(buffer[int(pheader.Ofs_st)+i*shared.Dstvert_size:])
	}

	/* load triangle lists */
	ptris := make([]shared.Dtriangle_t, pheader.Num_tris)
	for i := range ptris {
		ptris[i] = shared.Dtriangle(buffer[int(pheader.Ofs_tris)+i*shared.Dtriangle_size:])
	}

	/* load the frames */
	pframes := make([]shared.Daliasframe_t, pheader.Num_frames)
	for i := range pframes {
		pframes[i] = shared.Daliasframe(buffer[int(pheader.Ofs_frames)+i*int(pheader.Framesize):], int(pheader.Framesize))
	}

	mod.mtype = mod_alias

	/* load the glcmds */
	glcmds := make([]int32, pheader.Num_glcmds)
	for i := range glcmds {
		glcmds[i] = shared.ReadInt32(buffer[int(pheader.Ofs_glcmds)+i*4:])
	}

	// if (poutcmd[pheader->num_glcmds-1] != 0) {
	// 	R_Printf(PRINT_ALL, "%s: Entity %s has possible last element issues with %d verts.\n",
	// 		__func__,
	// 		mod->name,
	// 		poutcmd[pheader->num_glcmds-1]);
	// }

	/* register all skins */
	skinNames := make([]string, pheader.Num_skins)
	for i := range skinNames {
		skinNames[i] = shared.ReadString(buffer[int(pheader.Ofs_skins)+i*shared.MAX_SKINNAME:], shared.MAX_SKINNAME)
	}

	mod.skins = make([]*gl3image_t, pheader.Num_skins)
	for i := range skinNames {
		mod.skins[i] = T.findImage(skinNames[i], it_skin)
	}

	extra := aliasExtra{}
	extra.header = pheader
	extra.sts = psts
	extra.tris = ptris
	extra.frames = pframes
	extra.glcmds = glcmds
	extra.skinNames = skinNames

	extra.vertexCount, extra.indexCount = countVerticesAndIndices(glcmds)

	mod.extradata = extra

	mod.mins[0] = -32
	mod.mins[1] = -32
	mod.mins[2] = -32
	mod.maxs[0] = 32
	mod.maxs[1] = 32
	mod.maxs[2] = 32
	return nil
}

func countVerticesAndIndices(cmds []int32) (int, int) {
	indices := 0
	vertices := 0
	index := 0
	for {
		/* get the vertex count and primitive type */
		count := cmds[index]
		index++
		if count == 0 {
			break /* done */
		}

		if count < 0 {
			count = -count
		}

		index += 3 * int(count)
		vertices += int(count)
		indices += 3 * int(count-2)
	}
	return vertices, indices
}
