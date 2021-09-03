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
	"math"
	"strconv"
)

const SIDE_FRONT = 0
const SIDE_BACK = 1
const SIDE_ON = 2

const MAX_MOD_KNOWN = 512

const (
	SURF_PLANEBACK      = 2
	SURF_DRAWSKY        = 4
	SURF_DRAWTURB       = 0x10
	SURF_DRAWBACKGROUND = 0x40
	SURF_UNDERWATER     = 0x80
)

// used for vertex array elements when drawing brushes, sprites, sky and more
// (ok, it has the layout used for rendering brushes, but is not used there)
type gl3_3D_vtx_t struct {
	// vec3_t pos;
	// float texCoord[2];
	// float lmTexCoord[2]; // lightmap texture coordinate (sometimes unused)
	// vec3_t normal;
	// GLuint lightFlags; // bit i set means: dynlight i affects surface
	data []uint32
}

const gl3_3D_vtx_size = 11

func (T *gl3_3D_vtx_t) getPos() []float32 {
	r := make([]float32, 3)
	r[0] = math.Float32frombits(T.data[0])
	r[1] = math.Float32frombits(T.data[1])
	r[2] = math.Float32frombits(T.data[2])
	return r
}

func (T *gl3_3D_vtx_t) setPos(v []float32) {
	T.data[0] = math.Float32bits(v[0])
	T.data[1] = math.Float32bits(v[1])
	T.data[2] = math.Float32bits(v[2])
}

func (T *gl3_3D_vtx_t) setTexCoord(s, t float32) {
	T.data[3] = math.Float32bits(s)
	T.data[4] = math.Float32bits(t)
}

func (T *gl3_3D_vtx_t) setLmTexCoord(s, t float32) {
	T.data[5] = math.Float32bits(s)
	T.data[6] = math.Float32bits(t)
}

func (T *gl3_3D_vtx_t) setNormal(v []float32) {
	T.data[7] = math.Float32bits(v[0])
	T.data[8] = math.Float32bits(v[1])
	T.data[9] = math.Float32bits(v[2])
}

func (T *gl3_3D_vtx_t) setLightFlags(v uint32) {
	T.data[10] = v
}

// used for vertex array elements when drawing models
type gl3_alias_vtx_t struct {
	// GLfloat pos[3];
	// GLfloat texCoord[2];
	// GLfloat color[4];
	data []float32
}

const gl3_alias_vtx_size = 9

func (T *gl3_alias_vtx_t) setPos(v []float32) {
	T.data[0] = v[0]
	T.data[1] = v[1]
	T.data[2] = v[2]
}

func (T *gl3_alias_vtx_t) setTexCoord(s, t float32) {
	T.data[3] = s
	T.data[4] = t
}

func (T *gl3_alias_vtx_t) setColor(v []float32) {
	T.data[5] = v[0]
	T.data[6] = v[1]
	T.data[7] = v[2]
	T.data[8] = v[3]
}

func (T *gl3_alias_vtx_t) setColorAlpha(v []float32, a float32) {
	T.data[5] = v[0]
	T.data[6] = v[1]
	T.data[7] = v[2]
	T.data[8] = a
}

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

type medge_t struct {
	v                [2]uint16
	cachededgeoffset uint32
}

type mtexinfo_t struct {
	vecs      [2][4]float32
	flags     int
	numframes int
	next      *mtexinfo_t /* animation chain */
	image     *gl3image_t
}

type glpoly_t struct {
	next     *glpoly_t
	chain    *glpoly_t
	numverts int
	flags    int /* for SURF_UNDERWATER (not needed anymore?) */
	// gl3_3D_vtx_t vertices[4]; /* variable sized */
	verticesData []uint32
}

func (T *glpoly_t) vertices(index int) *gl3_3D_vtx_t {
	d := &gl3_3D_vtx_t{}
	d.data = T.verticesData[index*gl3_3D_vtx_size:]
	return d
}

type msurface_t struct {
	visframe int /* should be drawn when node is crossed */

	plane *shared.Cplane_t
	flags int

	firstedge int /* look up in model->surfedges[], negative numbers */
	numedges  int /* are backwards edges */

	texturemins [2]int16
	extents     [2]int16

	light_s, light_t   int /* gl lightmap coordinates */
	dlight_s, dlight_t int /* gl lightmap coordinates for dynamic lightmaps */

	polys        *glpoly_t /* multiple if warped */
	texturechain *msurface_t
	// struct  msurface_s *lightmapchain; not used/needed anymore

	texinfo *mtexinfo_t

	/* lighting info */
	dlightframe int
	dlightbits  int

	lightmaptexturenum int
	styles             [shared.MAXLIGHTMAPS]byte // MAXLIGHTMAPS = MAX_LIGHTMAPS_PER_SURFACE (defined in local.h)
	// I think cached_light is not used/needed anymore
	// //float cached_light[MAXLIGHTMAPS];       /* values currently used in lightmap */
	samples []byte /* [numstyles*surfsize] */
}

type mnode_or_leaf interface {
	/* common with leaf */
	Contents() int
	Visframe() int /* node needs to be traversed if current */

	Minmaxs() []float32 /* for bounding box culling */

	Parent() *mnode_t
	SetParent(p *mnode_t)
}

type mnode_t struct {
	/* common with leaf */
	contents int /* -1, to differentiate from leafs */
	visframe int /* node needs to be traversed if current */

	minmaxs [6]float32 /* for bounding box culling */

	parent *mnode_t

	/* node specific */
	plane    *shared.Cplane_t
	children [2]mnode_or_leaf

	firstsurface uint16
	numsurfaces  uint16
}

func (T *mnode_t) Contents() int {
	return T.contents
}

func (T *mnode_t) Visframe() int {
	return T.visframe
}

func (T *mnode_t) Minmaxs() []float32 {
	return T.minmaxs[:]
}

func (T *mnode_t) Parent() *mnode_t {
	return T.parent
}

func (T *mnode_t) SetParent(p *mnode_t) {
	T.parent = p
}

type mleaf_t struct {
	/* common with node */
	contents int /* wil be a negative contents number */
	visframe int /* node needs to be traversed if current */

	minmaxs [6]float32 /* for bounding box culling */

	parent *mnode_t

	/* leaf specific */
	cluster int
	area    int

	firstmarksurface []msurface_t
	nummarksurfaces  int
}

func (T *mleaf_t) Contents() int {
	return T.contents
}

func (T *mleaf_t) Visframe() int {
	return T.visframe
}

func (T *mleaf_t) Minmaxs() []float32 {
	return T.minmaxs[:]
}

func (T *mleaf_t) Parent() *mnode_t {
	return T.parent
}

func (T *mleaf_t) SetParent(p *mnode_t) {
	T.parent = p
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
	mins   [3]float32
	maxs   [3]float32
	radius float32

	/* solid volume for clipping */
	clipbox  bool
	clipmins [3]float32
	clipmaxs [3]float32

	/* brush model */
	firstmodelsurface, nummodelsurfaces int
	lightmap                            int /* only for submodels */

	submodels []mmodel_t

	planes []shared.Cplane_t

	numleafs int /* number of visible leafs, not counting 0 */
	leafs    []mleaf_t

	vertexes []mvertex_t
	edges    []medge_t

	firstnode int
	nodes     []mnode_t

	texinfo []mtexinfo_t

	surfaces []msurface_t

	surfedges []int

	marksurfaces [][]msurface_t

	vis    *shared.Dvis_t
	visBfr []byte

	lightdata []byte

	/* for alias models and skins */
	skins []*gl3image_t

	extradata interface{}
}

func (M *gl3model_t) Copy(other gl3model_t) {
	M.name = other.name
	M.registration_sequence = other.registration_sequence
	M.mtype = other.mtype
	M.numframes = other.numframes
	M.flags = other.flags
	copy(M.mins[:], other.mins[:])
	copy(M.maxs[:], other.maxs[:])
	M.radius = other.radius
	M.clipbox = other.clipbox
	copy(M.clipmins[:], other.clipmins[:])
	copy(M.clipmaxs[:], other.clipmaxs[:])
	M.firstmodelsurface = other.firstmodelsurface
	M.nummodelsurfaces = other.nummodelsurfaces
	M.lightmap = other.lightmap
	M.submodels = other.submodels
	M.planes = other.planes
	M.numleafs = other.numleafs
	M.leafs = other.leafs
	M.vertexes = other.vertexes
	M.edges = other.edges
	M.firstnode = other.firstnode
	M.nodes = other.nodes
	M.texinfo = other.texinfo
	M.surfaces = other.surfaces
	M.surfedges = other.surfedges
	M.marksurfaces = other.marksurfaces
	M.visBfr = other.visBfr
	M.vis = other.vis
	M.lightdata = other.lightdata
	M.skins = other.skins
	M.extradata = other.extradata
}

func (T *qGl3) modPointInLeaf(p []float32, model *gl3model_t) (*mleaf_t, error) {

	if model == nil || len(model.nodes) == 0 {
		return nil, T.ri.Sys_Error(shared.ERR_DROP, "modPointInLeaf: bad model")
	}

	var anode mnode_or_leaf = &model.nodes[0]

	for {
		if anode.Contents() != -1 {
			return anode.(*mleaf_t), nil
		}

		node := anode.(*mnode_t)
		plane := node.plane
		d := shared.DotProduct(p, plane.Normal[:]) - plane.Dist

		if d > 0 {
			anode = node.children[0]
		} else {
			anode = node.children[1]
		}
	}
}

func (T *qGl3) modClusterPVS(cluster int, model *gl3model_t) []byte {
	if (cluster == -1) || model.vis == nil {
		return T.mod_novis[:]
	}

	return shared.Mod_DecompressVis(model.visBfr,
		int(model.vis.Bitofs[cluster][shared.DVIS_PVS]),
		int((model.vis.Numclusters+7)>>3))
}

func (T *qGl3) modInit() {
	for i, _ := range T.mod_novis {
		T.mod_novis[i] = 0xFF
	}
}

func (T *qGl3) modLoadLighting(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {
	if l.Filelen == 0 {
		mod.lightdata = nil
		return nil
	}

	mod.lightdata = make([]byte, l.Filelen)
	copy(mod.lightdata, buffer[l.Fileofs:])
	return nil
}

func (T *qGl3) modLoadMarksurfaces(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {
	if (l.Filelen % 2) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadMarksurfaces: funny lump size in %s", mod.name)
	}

	count := l.Filelen / 2

	mod.marksurfaces = make([][]msurface_t, count)

	for i := 0; i < int(count); i++ {
		j := shared.ReadInt16(buffer[int(l.Fileofs)+i*2:])
		if (j < 0) || (int(j) >= len(mod.surfaces)) {
			return T.ri.Sys_Error(shared.ERR_DROP, "modLoadMarksurfaces: bad surface number")
		}

		mod.marksurfaces[i] = mod.surfaces[j:]
	}
	return nil
}

func (T *qGl3) modLoadVertexes(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {

	if (l.Filelen % shared.Dvertex_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadVertexes: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Dvertex_size

	mod.vertexes = make([]mvertex_t, count)

	for i := 0; i < int(count); i++ {
		src := shared.Dvertex(buffer[int(l.Fileofs)+i*shared.Dvertex_size:])
		mod.vertexes[i].position[0] = src.Point[0]
		mod.vertexes[i].position[1] = src.Point[1]
		mod.vertexes[i].position[2] = src.Point[2]
	}
	return nil
}

func (T *qGl3) modLoadTexinfo(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {

	if (l.Filelen % shared.Texinfo_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadTexinfo: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Texinfo_size

	mod.texinfo = make([]mtexinfo_t, count)

	for i := 0; i < int(count); i++ {
		src := shared.Texinfo(buffer[int(l.Fileofs)+i*shared.Texinfo_size:])

		for j := 0; j < 4; j++ {
			mod.texinfo[i].vecs[0][j] = src.Vecs[0][j]
			mod.texinfo[i].vecs[1][j] = src.Vecs[1][j]
		}

		mod.texinfo[i].flags = int(src.Flags)
		next := src.Nexttexinfo

		if next > 0 {
			mod.texinfo[i].next = &mod.texinfo[next]
		} else {
			mod.texinfo[i].next = nil
		}

		name := fmt.Sprintf("textures/%v.wal", src.Texture)

		mod.texinfo[i].image = T.findImage(name, it_wall)

		// if (!mod.texinfo[i].image || mod.texinfo[i].image == gl3_notexture) {
		// 	Com_sprintf(name, sizeof(name), "textures/%s.m8", in->texture);
		// 	mod.texinfo[i].image = GL3_FindImage(name, it_wall);
		// }

		if mod.texinfo[i].image == nil {
			T.rPrintf(shared.PRINT_ALL, "Couldn't load %s\n", name)
			mod.texinfo[i].image = T.gl3_notexture
		}
	}

	/* count animation frames */
	for i := 0; i < int(count); i++ {
		out := &mod.texinfo[i]
		out.numframes = 1

		for step := out.next; step != nil && step != out; step = step.next {
			out.numframes++
		}
	}
	return nil
}

func (T *qGl3) modLoadEdges(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {

	if (l.Filelen % shared.Dedge_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadEdges: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Dedge_size

	mod.edges = make([]medge_t, count)

	for i := 0; i < int(count); i++ {
		src := shared.Dedge(buffer[int(l.Fileofs)+i*shared.Dedge_size:])
		mod.edges[i].v[0] = src.V[0]
		mod.edges[i].v[1] = src.V[1]
	}
	return nil
}

func modSetParent(node mnode_or_leaf, parent *mnode_t) {
	node.SetParent(parent)
	if node.Contents() != -1 {
		return
	}

	n := node.(*mnode_t)
	modSetParent(n.children[0], n)
	modSetParent(n.children[1], n)
}

func (T *qGl3) modLoadNodes(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {
	if (l.Filelen % shared.Dnode_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadNodes: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Dnode_size

	mod.nodes = make([]mnode_t, count)

	for i := 0; i < int(count); i++ {
		src := shared.Dnode(buffer[int(l.Fileofs)+i*shared.Dnode_size:])

		for j := 0; j < 3; j++ {
			mod.nodes[i].minmaxs[j] = float32(src.Mins[j])
			mod.nodes[i].minmaxs[3+j] = float32(src.Maxs[j])
		}

		p := src.Planenum
		mod.nodes[i].plane = &mod.planes[p]

		mod.nodes[i].firstsurface = src.Firstface
		mod.nodes[i].numsurfaces = src.Numfaces
		mod.nodes[i].contents = -1 /* differentiate from leafs */

		for j := 0; j < 2; j++ {
			p := src.Children[j]

			if p >= 0 {
				mod.nodes[i].children[j] = &mod.nodes[p]
			} else {
				mod.nodes[i].children[j] = &mod.leafs[-1-p]
			}
		}
	}

	modSetParent(&mod.nodes[0], nil) /* sets nodes and leafs */
	return nil
}

func (T *qGl3) modLoadLeafs(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {
	if (l.Filelen % shared.Dleaf_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadLeafs: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Dleaf_size

	mod.leafs = make([]mleaf_t, count)
	mod.numleafs = int(count)

	for i := 0; i < int(count); i++ {
		src := shared.Dleaf(buffer[int(l.Fileofs)+i*shared.Dleaf_size:])

		for j := 0; j < 3; j++ {
			mod.leafs[i].minmaxs[j] = float32(src.Mins[j])
			mod.leafs[i].minmaxs[3+j] = float32(src.Maxs[j])
		}

		mod.leafs[i].contents = int(src.Contents)

		mod.leafs[i].cluster = int(src.Cluster)
		mod.leafs[i].area = int(src.Area)

		// make unsigned long from signed short
		firstleafface := src.Firstleafface & 0xFFFF
		mod.leafs[i].nummarksurfaces = int(src.Numleaffaces & 0xFFFF)

		mod.leafs[i].firstmarksurface = mod.marksurfaces[firstleafface]
		// 	if ((firstleafface + out->nummarksurfaces) > loadmodel->nummarksurfaces)
		// 	{
		// 		ri.Sys_Error(ERR_DROP, "%s: wrong marksurfaces position in %s",
		// 			__func__, loadmodel->name);
		// 	}
	}
	return nil
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

		mod.submodels[i].radius = shared.Mod_RadiusFromBounds(mod.submodels[i].mins[:], mod.submodels[i].maxs[:])
		mod.submodels[i].headnode = int(src.Headnode)
		mod.submodels[i].firstface = int(src.Firstface)
		mod.submodels[i].numfaces = int(src.Numfaces)
	}
	return nil
}

/*
 * Fills in s->texturemins[] and s->extents[]
 */
func mod_CalcSurfaceExtents(s *msurface_t, mod *gl3model_t) {

	mins := []float32{999999, 999999}
	maxs := []float32{-99999, -99999}

	tex := s.texinfo

	for i := 0; i < s.numedges; i++ {
		e := mod.surfedges[s.firstedge+i]

		var v *mvertex_t
		if e >= 0 {
			v = &mod.vertexes[mod.edges[e].v[0]]
		} else {
			v = &mod.vertexes[mod.edges[-e].v[1]]
		}

		for j := 0; j < 2; j++ {
			value := v.position[0]*tex.vecs[j][0] +
				v.position[1]*tex.vecs[j][1] +
				v.position[2]*tex.vecs[j][2] +
				tex.vecs[j][3]

			if value < mins[j] {
				mins[j] = value
			}

			if value > maxs[j] {
				maxs[j] = value
			}
		}
	}

	for i := 0; i < 2; i++ {
		bmins := int(math.Floor(float64(mins[i]) / 16))
		bmaxs := int(math.Ceil(float64(maxs[i]) / 16))

		s.texturemins[i] = int16(bmins * 16)
		s.extents[i] = int16((bmaxs - bmins) * 16)
	}
}

func (T *qGl3) modLoadFaces(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {

	if (l.Filelen % shared.Dface_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadFaces: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Dface_size

	mod.surfaces = make([]msurface_t, count)

	T.lmBeginBuildingLightmaps(mod)

	for surfnum := 0; surfnum < int(count); surfnum++ {
		src := shared.Dface(buffer[int(l.Fileofs)+surfnum*shared.Dface_size:])
		out := &mod.surfaces[surfnum]

		out.firstedge = int(src.Firstedge)
		out.numedges = int(src.Numedges)
		out.flags = 0
		out.polys = nil

		planenum := src.Planenum
		side := src.Side

		if side != 0 {
			out.flags |= SURF_PLANEBACK
		}

		out.plane = &mod.planes[planenum]

		ti := int(src.Texinfo)
		if (ti < 0) || (ti >= len(mod.texinfo)) {
			return T.ri.Sys_Error(shared.ERR_DROP, "modLoadFaces: bad texinfo number")
		}

		out.texinfo = &mod.texinfo[ti]

		mod_CalcSurfaceExtents(out, mod)

		/* lighting info */
		for i := 0; i < MAX_LIGHTMAPS_PER_SURFACE; i++ {
			out.styles[i] = src.Styles[i]
		}

		i := src.Lightofs

		if i == -1 {
			out.samples = nil
		} else {
			out.samples = mod.lightdata[i:]
		}

		/* set the drawing flags */
		if (out.texinfo.flags & shared.SURF_WARP) != 0 {
			out.flags |= SURF_DRAWTURB

			for i := 0; i < 2; i++ {
				out.extents[i] = 16384
				out.texturemins[i] = -8192
			}

			gl3SubdivideSurface(out, mod) /* cut up polygon for warps */
		}

		if T.r_fixsurfsky.Bool() {
			if (out.texinfo.flags & shared.SURF_SKY) != 0 {
				out.flags |= SURF_DRAWSKY
			}
		}

		/* create lightmaps and polygons */
		if (out.texinfo.flags & (shared.SURF_SKY | shared.SURF_TRANS33 | shared.SURF_TRANS66 | shared.SURF_WARP)) == 0 {
			if err := T.lmCreateSurfaceLightmap(out); err != nil {
				return err
			}
		}

		if (out.texinfo.flags & shared.SURF_WARP) == 0 {
			T.lmBuildPolygonFromSurface(out, mod)
		}
	}

	return T.lmEndBuildingLightmaps()
}

func (T *qGl3) modLoadSurfedges(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {

	if (l.Filelen % 4) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadSurfedges: funny lump size in %s", mod.name)
	}

	count := l.Filelen / 4

	mod.surfedges = make([]int, count)

	for i := 0; i < int(count); i++ {
		mod.surfedges[i] = int(shared.ReadInt32(buffer[int(l.Fileofs)+i*4:]))
	}
	return nil
}

func (T *qGl3) modLoadPlanes(l shared.Lump_t, mod *gl3model_t, buffer []byte) error {

	if (l.Filelen % shared.Dplane_size) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "modLoadPlanes: funny lump size in %s", mod.name)
	}

	count := l.Filelen / shared.Dplane_size

	mod.planes = make([]shared.Cplane_t, count)

	for i := 0; i < int(count); i++ {
		src := shared.Dplane(buffer[int(l.Fileofs)+i*shared.Dplane_size:])

		bits := 0

		for j := 0; j < 3; j++ {
			mod.planes[i].Normal[j] = src.Normal[j]
			if mod.planes[i].Normal[j] < 0 {
				bits |= 1 << j
			}
		}

		mod.planes[i].Dist = src.Dist
		mod.planes[i].Type = byte(src.Type)
		mod.planes[i].Signbits = byte(bits)
	}
	return nil
}

func (T *qGl3) modLoadVisibility(l shared.Lump_t, mod *gl3model_t, buffer []byte) {

	if l.Filelen == 0 {
		mod.vis = nil
		mod.visBfr = nil
		return
	}

	mod.visBfr = buffer[int(l.Fileofs) : int(l.Fileofs)+int(l.Filelen)]
	mod.vis = shared.Dvis(mod.visBfr)
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
	if err := T.modLoadVertexes(header.Lumps[shared.LUMP_VERTEXES], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadEdges(header.Lumps[shared.LUMP_EDGES], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadSurfedges(header.Lumps[shared.LUMP_SURFEDGES], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadLighting(header.Lumps[shared.LUMP_LIGHTING], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadPlanes(header.Lumps[shared.LUMP_PLANES], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadTexinfo(header.Lumps[shared.LUMP_TEXINFO], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadFaces(header.Lumps[shared.LUMP_FACES], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadMarksurfaces(header.Lumps[shared.LUMP_LEAFFACES], mod, buffer); err != nil {
		return err
	}
	T.modLoadVisibility(header.Lumps[shared.LUMP_VISIBILITY], mod, buffer)
	if err := T.modLoadLeafs(header.Lumps[shared.LUMP_LEAFS], mod, buffer); err != nil {
		return err
	}
	if err := T.modLoadNodes(header.Lumps[shared.LUMP_NODES], mod, buffer); err != nil {
		return err
	}
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
		starmod.firstnode = bm.headnode

		if starmod.firstnode >= len(mod.nodes) {
			return T.ri.Sys_Error(shared.ERR_DROP, "modLoadBrushModel: Inline model %v has bad firstnode", i)
		}

		copy(starmod.maxs[:], bm.maxs[:])
		copy(starmod.mins[:], bm.mins[:])
		starmod.radius = bm.radius

		if i == 0 {
			mod.Copy(*starmod)
		}

		starmod.numleafs = bm.visleafs
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
		if err := T.loadSP2(&T.mod_known[index], buf); err != nil {
			return nil, err
		}

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
	T.gl3_oldviewcluster = -1 /* force markleafs */

	T.gl3state.currentlightmap = -1

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

	T.gl3_viewcluster = -1
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
		if mod.mtype == mod_sprite {
			extra := mod.extradata.(shared.Dsprite_t)

			for i := range mod.skins {
				mod.skins[i] = T.findImage(extra.Frames[i].Name, it_sprite)
			}
		} else if mod.mtype == mod_alias {
			extra := mod.extradata.(aliasExtra)

			for i := range mod.skins {
				mod.skins[i] = T.findImage(extra.skinNames[i], it_skin)
			}

			mod.numframes = int(extra.header.Num_frames)
		} else if mod.mtype == mod_brush {
			for i := range mod.texinfo {
				mod.texinfo[i].image.registration_sequence = T.registration_sequence
			}
		}
	}

	return mod, nil
}
