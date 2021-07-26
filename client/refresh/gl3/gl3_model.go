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
 * Model loading and caching for OpenGL3. Includes the .bsp file format
 *
 * =======================================================================
 */
package gl3

import (
	"fmt"
	"goquake2/shared"
	"strconv"
)

const MAX_MOD_KNOWN = 512

/* in memory representation */
type mvertex_t struct {
	position [3]float32
}

type mmodel_t struct {
	mins                [3]float32
	maxs                [3]float32
	origin              [3]float32 /* for sounds or lights */
	radius              float32
	headnode            int
	visleafs            int /* not including the solid leaf 0 */
	firstface, numfaces int
}

/* Whole model */

// this, must be struct model_s, not gl3model_s,
// because struct model_s* is returned by re.RegisterModel()
type gl3model_t struct {
	name string

	registration_sequence int

	mtype     modtype_t
	numframes int

	flags int

	// /* volume occupied by the model graphics */
	// vec3_t mins, maxs;
	radius float32

	// /* solid volume for clipping */
	// qboolean clipbox;
	// vec3_t clipmins, clipmaxs;

	/* brush model */
	firstmodelsurface, nummodelsurfaces int
	// int lightmap; /* only for submodels */

	submodels []mmodel_t

	// int numplanes;
	// cplane_t *planes;

	// int numleafs; /* number of visible leafs, not counting 0 */
	// mleaf_t *leafs;

	// int numvertexes;
	// mvertex_t *vertexes;

	// int numedges;
	// medge_t *edges;

	// int numnodes;
	// int firstnode;
	// mnode_t *nodes;

	// int numtexinfo;
	// mtexinfo_t *texinfo;

	// int numsurfaces;
	// msurface_t *surfaces;

	// int numsurfedges;
	// int *surfedges;

	// int nummarksurfaces;
	// msurface_t **marksurfaces;

	// dvis_t *vis;

	// byte *lightdata;

	// /* for alias models and skins */
	// gl3image_t *skins[MAX_MD2SKINS];

	// int extradatasize;
	// void *extradata;
}

func (M *gl3model_t) Copy(other gl3model_t) {
	M.name = other.name
	M.registration_sequence = other.registration_sequence
	M.mtype = other.mtype
	M.numframes = other.numframes
	M.flags = other.flags
	// qboolean clipbox;
	// vec3_t clipmins, clipmaxs;
	M.firstmodelsurface = other.firstmodelsurface
	M.nummodelsurfaces = other.nummodelsurfaces
	// int lightmap;
	M.submodels = other.submodels
	// int numplanes;
	// cplane_t *planes;
	// int numleafs; /* number of visible leafs, not counting 0 */
	// mleaf_t *leafs;
	// int numvertexes;
	// mvertex_t *vertexes;
	// int numedges;
	// medge_t *edges;
	// int numnodes;
	// int firstnode;
	// mnode_t *nodes;
	// int numtexinfo;
	// mtexinfo_t *texinfo;
	// int numsurfaces;
	// msurface_t *surfaces;
	// int numsurfedges;
	// int *surfedges;
	// int nummarksurfaces;
	// msurface_t **marksurfaces;
	// dvis_t *vis;
	// byte *lightdata;
	// gl3image_t *skins[MAX_MD2SKINS];
	// int extradatasize;
	// void *extradata;
}

func (T *qGl3) modLoadSubmodels(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {

	if (l.Filelen % shared.Dmodel_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadSubmodels: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Dmodel_size

	mod.submodels = make([]mmodel_t, count)

	for i := 0; i < int(count); i++ {
		src := shared.Dmodel(buffer[int(l.Fileofs)+i*shared.Dmodel_size:])
		for j := 0; j < 3; j++ {
			/* spread the mins / maxs by a pixel */
			mod.submodels[i].mins[j] = src.Mins[j] - 1
			mod.submodels[i].maxs[j] = src.Maxs[j] + 1
			mod.submodels[i].origin[j] = src.Origin[j]
		}

		// 	out->radius = Mod_RadiusFromBounds(out->mins, out->maxs);
		mod.submodels[i].headnode = int(src.Headnode)
		mod.submodels[i].firstface = int(src.Firstface)
		mod.submodels[i].numfaces = int(src.Numfaces)
	}
	return nil
}

func (T *qGl3) modLoadBrushModel(mod *gl3model_t, buffer []byte) error {
	// int i;
	// dheader_t *header;
	// mmodel_t *bm;

	if mod.name != T.mod_known[0].name {
		return T.ri.Sys_Error(shared.ERR_DROP, "Loaded a brush model after the world")
	}

	header := shared.DheaderCreate(buffer)

	if header.Version != shared.BSPVERSION {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadBrushModel: %v has wrong version number (%v should be %v)",
			mod.name, header.Version, shared.BSPVERSION)
	}

	// // calculate the needed hunksize from the lumps
	// int hunkSize = 0;
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_VERTEXES], sizeof(dvertex_t), sizeof(mvertex_t));
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_EDGES], sizeof(dedge_t), sizeof(medge_t));
	// hunkSize += sizeof(medge_t) + 31; // for count+1 in Mod_LoadEdges()
	// int surfEdgeCount = (header->lumps[LUMP_SURFEDGES].filelen+sizeof(int)-1)/sizeof(int);
	// if(surfEdgeCount < MAX_MAP_SURFEDGES) // else it errors out later anyway
	// 	hunkSize += calcLumpHunkSize(&header->lumps[LUMP_SURFEDGES], sizeof(int), sizeof(int));
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_LIGHTING], 1, 1);
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_PLANES], sizeof(dplane_t), sizeof(cplane_t)*2);
	// hunkSize += calcTexinfoAndFacesSize(&header->lumps[LUMP_FACES], &header->lumps[LUMP_TEXINFO]);
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_LEAFFACES], sizeof(short), sizeof(msurface_t *)); // yes, out is indeeed a pointer!
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_VISIBILITY], 1, 1);
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_LEAFS], sizeof(dleaf_t), sizeof(mleaf_t));
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_NODES], sizeof(dnode_t), sizeof(mnode_t));
	// hunkSize += calcLumpHunkSize(&header->lumps[LUMP_MODELS], sizeof(dmodel_t), sizeof(mmodel_t));

	// loadmodel->extradata = Hunk_Begin(hunkSize);
	mod.mtype = mod_brush

	/* load into heap */
	// Mod_LoadVertexes(&header->lumps[LUMP_VERTEXES]);
	// Mod_LoadEdges(&header->lumps[LUMP_EDGES]);
	// Mod_LoadSurfedges(&header->lumps[LUMP_SURFEDGES]);
	// Mod_LoadLighting(&header->lumps[LUMP_LIGHTING]);
	// Mod_LoadPlanes(&header->lumps[LUMP_PLANES]);
	// Mod_LoadTexinfo(&header->lumps[LUMP_TEXINFO]);
	// Mod_LoadFaces(&header->lumps[LUMP_FACES]);
	// Mod_LoadMarksurfaces(&header->lumps[LUMP_LEAFFACES]);
	// Mod_LoadVisibility(&header->lumps[LUMP_VISIBILITY]);
	// Mod_LoadLeafs(&header->lumps[LUMP_LEAFS]);
	// Mod_LoadNodes(&header->lumps[LUMP_NODES]);
	if err := T.modLoadSubmodels(header.Lumps[shared.LUMP_MODELS], mod, buffer); err != nil {
		return err
	}
	mod.numframes = 2 /* regular and alternate animation */

	/* set up the submodels */
	for i, bm := range mod.submodels {

		starmod := &T.mod_inline[i]

		starmod.Copy(*mod)

		starmod.firstmodelsurface = bm.firstface
		starmod.nummodelsurfaces = bm.numfaces
		// 	starmod->firstnode = bm->headnode;

		// 	if (starmod->firstnode >= loadmodel->numnodes) {
		// 		ri.Sys_Error(ERR_DROP, "%s: Inline model %i has bad firstnode",
		// 				__func__, i);
		// 	}

		// 	VectorCopy(bm->maxs, starmod->maxs);
		// 	VectorCopy(bm->mins, starmod->mins);
		starmod.radius = bm.radius

		if i == 0 {
			mod.Copy(*starmod)
		}

		// 	starmod->numleafs = bm->visleafs;
	}
	return nil
}

/*
 * Loads in a model for the given name
 */
func (T *qGl3) modForName(name string, crash bool) (*gl3model_t, error) {

	if len(name) == 0 {
		return nil, T.ri.Sys_Error(shared.ERR_DROP, "modForName: NULL name")
	}

	/* inline models are grabbed only from worldmodel */
	if name[0] == '*' {
		i, _ := strconv.ParseInt(name[1:], 10, 32)
		if (i < 1) || T.gl3_worldmodel == nil || (int(i) >= len(T.gl3_worldmodel.submodels)) {
			return nil, T.ri.Sys_Error(shared.ERR_DROP, "modForName: bad inline model number %v", i)
		}

		return &T.mod_inline[i], nil
	}

	/* search the currently loaded models */
	for i, mod := range T.mod_known {
		if len(mod.name) == 0 {
			continue
		}

		if mod.name == name {
			return &T.mod_known[i], nil
		}
	}

	/* find a free model slot spot */
	index := -1
	for i, mod := range T.mod_known {
		if len(mod.name) == 0 {
			index = i
			break /* free spot */
		}
	}

	if index < 0 {
		return nil, T.ri.Sys_Error(shared.ERR_DROP, "mod_numknown == MAX_MOD_KNOWN")
	}

	T.mod_known[index].name = name

	/* load the file */
	buf, err := T.ri.LoadFile(name)
	if buf == nil || err != nil {
		if crash {
			return nil, T.ri.Sys_Error(shared.ERR_DROP, "modForName: %s not found", name)
		}

		T.mod_known[index].name = ""
		return nil, nil
	}

	/* call the apropriate loader */
	id := shared.ReadInt32(buf)
	switch id {
	case shared.IDALIASHEADER:
		if err := T.loadMD2(&T.mod_known[index], buf); err != nil {
			return nil, err
		}

	case shared.IDSPRITEHEADER:
	// 		 GL3_LoadSP2(mod, buf, modfilelen);
	// 		 break;

	case shared.IDBSPHEADER:
		if err := T.modLoadBrushModel(&T.mod_known[index], buf); err != nil {
			return nil, err
		}

	default:
		return nil, T.ri.Sys_Error(shared.ERR_DROP, "modForName: unknown fileid for %s %x", name, id)
	}

	return &T.mod_known[index], nil
}

func (T *qGl3) BeginRegistration(model string) error {

	T.registration_sequence++
	// gl3_oldviewcluster = -1; /* force markleafs */

	// gl3state.currentlightmap = -1;

	fullname := fmt.Sprintf("maps/%s.bsp", model)

	/* explicitly free the old map if different
	   this guarantees that mod_known[0] is the
	   world map */
	flushmap := T.ri.Cvar_Get("flushmap", "0", 0)

	if (T.mod_known[0].name != fullname) || flushmap.Bool() {
		T.mod_known[0].name = ""
	}

	mod, err := T.modForName(fullname, true)
	if err != nil {
		return err
	}
	T.gl3_worldmodel = mod

	// gl3_viewcluster = -1;
	return nil
}

func (T *qGl3) RegisterModel(name string) (interface{}, error) {
	mod, err := T.modForName(name, false)
	if err != nil {
		return nil, err
	}

	if mod != nil {
		mod.registration_sequence = T.registration_sequence

		/* register any images used by the models */
		// if (mod->type == mod_sprite)
		// {
		// 	sprout = (dsprite_t *)mod->extradata;

		// 	for (i = 0; i < sprout->numframes; i++)
		// 	{
		// 		mod->skins[i] = GL3_FindImage(sprout->frames[i].name, it_sprite);
		// 	}
		// }
		// else if (mod->type == mod_alias)
		// {
		// 	pheader = (dmdl_t *)mod->extradata;

		// 	for (i = 0; i < pheader->num_skins; i++)
		// 	{
		// 		mod->skins[i] = GL3_FindImage((char *)pheader + pheader->ofs_skins + i * MAX_SKINNAME, it_skin);
		// 	}

		// 	mod->numframes = pheader->num_frames;
		// }
		// else if (mod->type == mod_brush)
		// {
		// 	for (i = 0; i < mod->numtexinfo; i++)
		// 	{
		// 		mod->texinfo[i].image->registration_sequence = registration_sequence;
		// 	}
		// }
	}

	return mod, nil
}
