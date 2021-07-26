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
 *  The prototypes for most file formats used by Quake II
 *
 * =======================================================================
 */
package shared

type QFileHandle interface {
	Close()
	Read(len int) []byte
}

/* .MD2 triangle model file format */

const IDALIASHEADER = (('2' << 24) + ('P' << 16) + ('D' << 8) + 'I')
const ALIAS_VERSION = 8

const (
	MAX_TRIANGLES = 4096
	MAX_VERTS     = 2048
	MAX_FRAMES    = 512
	MAX_MD2SKINS  = 32
	MAX_SKINNAME  = 64
)

// typedef struct
// {
// 	short s;
// 	short t;
// } dstvert_t;

// typedef struct
// {
// 	short index_xyz[3];
// 	short index_st[3];
// } dtriangle_t;

// typedef struct
// {
// 	byte v[3]; /* scaled byte to fit in frame mins/maxs */
// 	byte lightnormalindex;
// } dtrivertx_t;

// #define DTRIVERTX_V0 0
// #define DTRIVERTX_V1 1
// #define DTRIVERTX_V2 2
// #define DTRIVERTX_LNI 3
// #define DTRIVERTX_SIZE 4

// typedef struct
// {
// 	float scale[3];       /* multiply byte verts by this */
// 	float translate[3];   /* then add this */
// 	char name[16];        /* frame name from grabbing */
// 	dtrivertx_t verts[1]; /* variable sized */
// } daliasframe_t;

// /* the glcmd format:
//  * - a positive integer starts a tristrip command, followed by that many
//  *   vertex structures.
//  * - a negative integer starts a trifan command, followed by -x vertexes
//  *   a zero indicates the end of the command list.
//  * - a vertex consists of a floating point s, a floating point t,
//  *   and an integer vertex index. */

// typedef struct
// {
// 	int ident;
// 	int version;

// 	int skinwidth;
// 	int skinheight;
// 	int framesize;  /* byte size of each frame */

// 	int num_skins;
// 	int num_xyz;
// 	int num_st;     /* greater than num_xyz for seams */
// 	int num_tris;
// 	int num_glcmds; /* dwords in strip/fan command list */
// 	int num_frames;

// 	int ofs_skins;  /* each skin is a MAX_SKINNAME string */
// 	int ofs_st;     /* byte offset from start for stverts */
// 	int ofs_tris;   /* offset for dtriangles */
// 	int ofs_frames; /* offset for first frame */
// 	int ofs_glcmds;
// 	int ofs_end;    /* end of file */
// } dmdl_t;

/* .SP2 sprite file format */

const IDSPRITEHEADER = (('2' << 24) + ('S' << 16) + ('D' << 8) + 'I') /* little-endian "IDS2" */
const SPRITE_VERSION = 2

// typedef struct
// {
// 	int width, height;
// 	int origin_x, origin_y;  /* raster coordinates inside pic */
// 	char name[MAX_SKINNAME]; /* name of pcx file */
// } dsprframe_t;

// typedef struct
// {
// 	int ident;
// 	int version;
// 	int numframes;
// 	dsprframe_t frames[1]; /* variable sized */
// } dsprite_t;

/* .WAL texture file format */

// #define MIPLEVELS 4
// typedef struct miptex_s
// {
// 	char name[32];
// 	unsigned width, height;
// 	unsigned offsets[MIPLEVELS]; /* four mip maps stored */
// 	char animname[32];           /* next frame in animation chain */
// 	int flags;
// 	int contents;
// 	int value;
// } miptex_t;

/* .BSP file format */

const IDBSPHEADER = (('P' << 24) + ('S' << 16) + ('B' << 8) + 'I') /* little-endian "IBSP" */
const BSPVERSION = 38

/* upper design bounds: leaffaces, leafbrushes, planes, and
 * verts are still bounded by 16 bit short limits */
const (
	MAX_MAP_MODELS    = 1024
	MAX_MAP_BRUSHES   = 8192
	MAX_MAP_ENTITIES  = 2048
	MAX_MAP_ENTSTRING = 0x40000
	MAX_MAP_TEXINFO   = 8192

	MAX_MAP_AREAS       = 256
	MAX_MAP_AREAPORTALS = 1024
	MAX_MAP_PLANES      = 65536
	MAX_MAP_NODES       = 65536
	MAX_MAP_BRUSHSIDES  = 65536
	MAX_MAP_LEAFS       = 65536
	MAX_MAP_VERTS       = 65536
	MAX_MAP_FACES       = 65536
	MAX_MAP_LEAFFACES   = 65536
	MAX_MAP_LEAFBRUSHES = 65536
	MAX_MAP_PORTALS     = 65536
	MAX_MAP_EDGES       = 128000
	MAX_MAP_SURFEDGES   = 256000
	MAX_MAP_LIGHTING    = 0x200000
	MAX_MAP_VISIBILITY  = 0x100000

	/* key / value pair sizes */

	MAX_KEY   = 32
	MAX_VALUE = 1024
)

/* ================================================================== */

type Lump_t struct {
	Fileofs int32
	Filelen int32
}

const Lump_size = 2 * 4

const (
	LUMP_ENTITIES    = 0
	LUMP_PLANES      = 1
	LUMP_VERTEXES    = 2
	LUMP_VISIBILITY  = 3
	LUMP_NODES       = 4
	LUMP_TEXINFO     = 5
	LUMP_FACES       = 6
	LUMP_LIGHTING    = 7
	LUMP_LEAFS       = 8
	LUMP_LEAFFACES   = 9
	LUMP_LEAFBRUSHES = 10
	LUMP_EDGES       = 11
	LUMP_SURFEDGES   = 12
	LUMP_MODELS      = 13
	LUMP_BRUSHES     = 14
	LUMP_BRUSHSIDES  = 15
	LUMP_POP         = 16
	LUMP_AREAS       = 17
	LUMP_AREAPORTALS = 18
	HEADER_LUMPS     = 19
)

type Dheader_t struct {
	Ident   int32
	Version int32
	Lumps   []Lump_t
}

const Dheader_size = 2*4 + HEADER_LUMPS*Lump_size

func DheaderCreate(data []byte) Dheader_t {
	d := Dheader_t{}
	d.Ident = ReadInt32(data[0:])
	d.Version = ReadInt32(data[4:])
	d.Lumps = make([]Lump_t, HEADER_LUMPS)
	for i := 0; i < HEADER_LUMPS; i++ {
		d.Lumps[i].Fileofs = ReadInt32(data[2*4+i*Lump_size:])
		d.Lumps[i].Filelen = ReadInt32(data[2*4+i*Lump_size+4:])
	}
	return d
}

type Dmodel_t struct {
	Mins                [3]float32
	Maxs                [3]float32
	Origin              [3]float32 /* for sounds or lights */
	Headnode            int32
	Firstface, Numfaces int32 /* submodels just draw faces without
	   walking the bsp tree */
}

const Dmodel_size = 12 * 4

func Dmodel(data []byte) Dmodel_t {
	d := Dmodel_t{}
	d.Mins[0] = ReadFloat32(data[0*4:])
	d.Mins[1] = ReadFloat32(data[1*4:])
	d.Mins[2] = ReadFloat32(data[2*4:])
	d.Maxs[0] = ReadFloat32(data[3*4:])
	d.Maxs[1] = ReadFloat32(data[4*4:])
	d.Maxs[2] = ReadFloat32(data[5*4:])
	d.Origin[0] = ReadFloat32(data[6*4:])
	d.Origin[1] = ReadFloat32(data[7*4:])
	d.Origin[2] = ReadFloat32(data[8*4:])
	d.Headnode = ReadInt32(data[9*4:])
	d.Firstface = ReadInt32(data[10*4:])
	d.Numfaces = ReadInt32(data[11*4:])
	return d
}

type Dvertex_t struct {
	point [3]float32
}

const Dvertex_size = 3 * 4

func Dvertex(data []byte) Dvertex_t {
	d := Dvertex_t{}
	d.point[0] = ReadFloat32(data[0:])
	d.point[1] = ReadFloat32(data[1*4:])
	d.point[2] = ReadFloat32(data[2*4:])
	return d
}

// /* 0-2 are axial planes */
// #define PLANE_X 0
// #define PLANE_Y 1
// #define PLANE_Z 2

// /* 3-5 are non-axial planes snapped to the nearest */
// #define PLANE_ANYX 3
// #define PLANE_ANYY 4
// #define PLANE_ANYZ 5

// /* planes (x&~1) and (x&~1)+1 are always opposites */

// typedef struct
// {
// 	float normal[3];
// 	float dist;
// 	int type; /* PLANE_X - PLANE_ANYZ */
// } dplane_t;

// /* contents flags are seperate bits
//  * - given brush can contribute multiple content bits
//  * - multiple brushes can be in a single leaf */

// /* lower bits are stronger, and will eat weaker brushes completely */
// #define CONTENTS_SOLID 1  /* an eye is never valid in a solid */
// #define CONTENTS_WINDOW 2 /* translucent, but not watery */
// #define CONTENTS_AUX 4
// #define CONTENTS_LAVA 8
// #define CONTENTS_SLIME 16
// #define CONTENTS_WATER 32
// #define CONTENTS_MIST 64
// #define LAST_VISIBLE_CONTENTS 64

// /* remaining contents are non-visible, and don't eat brushes */
// #define CONTENTS_AREAPORTAL 0x8000

// #define CONTENTS_PLAYERCLIP 0x10000
// #define CONTENTS_MONSTERCLIP 0x20000

// /* currents can be added to any other contents, and may be mixed */
// #define CONTENTS_CURRENT_0 0x40000
// #define CONTENTS_CURRENT_90 0x80000
// #define CONTENTS_CURRENT_180 0x100000
// #define CONTENTS_CURRENT_270 0x200000
// #define CONTENTS_CURRENT_UP 0x400000
// #define CONTENTS_CURRENT_DOWN 0x800000

// #define CONTENTS_ORIGIN 0x1000000       /* removed before bsping an entity */

// #define CONTENTS_MONSTER 0x2000000      /* should never be on a brush, only in game */
// #define CONTENTS_DEADMONSTER 0x4000000
// #define CONTENTS_DETAIL 0x8000000       /* brushes to be added after vis leafs */
// #define CONTENTS_TRANSLUCENT 0x10000000 /* auto set if any surface has trans */
// #define CONTENTS_LADDER 0x20000000

// #define SURF_LIGHT 0x1    /* value will hold the light strength */

// #define SURF_SLICK 0x2    /* effects game physics */

// #define SURF_SKY 0x4      /* don't draw, but add to skybox */
// #define SURF_WARP 0x8     /* turbulent water warp */
// #define SURF_TRANS33 0x10
// #define SURF_TRANS66 0x20
// #define SURF_FLOWING 0x40 /* scroll towards angle */
// #define SURF_NODRAW 0x80  /* don't bother referencing the texture */

// typedef struct
// {
// 	int planenum;
// 	int children[2];         /* negative numbers are -(leafs+1), not nodes */
// 	short mins[3];           /* for frustom culling */
// 	short maxs[3];
// 	unsigned short firstface;
// 	unsigned short numfaces; /* counting both sides */
// } dnode_t;

// typedef struct texinfo_s
// {
// 	float vecs[2][4]; /* [s/t][xyz offset] */
// 	int flags;        /* miptex flags + overrides light emission, etc */
// 	int value;
// 	char texture[32]; /* texture name (textures*.wal) */
// 	int nexttexinfo;  /* for animations, -1 = end of chain */
// } texinfo_t;

// /* note that edge 0 is never used, because negative edge
//    nums are used for counterclockwise use of the edge in
//    a face */
// typedef struct
// {
// 	unsigned short v[2]; /* vertex numbers */
// } dedge_t;

const MAXLIGHTMAPS = 4

// typedef struct
// {
// 	unsigned short planenum;
// 	short side;

// 	int firstedge; /* we must support > 64k edges */
// 	short numedges;
// 	short texinfo;

// 	/* lighting info */
// 	byte styles[MAXLIGHTMAPS];
// 	int lightofs; /* start of [numstyles*surfsize] samples */
// } dface_t;

// typedef struct
// {
// 	int contents; /* OR of all brushes (not needed?) */

// 	short cluster;
// 	short area;

// 	short mins[3]; /* for frustum culling */
// 	short maxs[3];

// 	unsigned short firstleafface;
// 	unsigned short numleaffaces;

// 	unsigned short firstleafbrush;
// 	unsigned short numleafbrushes;
// } dleaf_t;

// typedef struct
// {
// 	unsigned short planenum; /* facing out of the leaf */
// 	short texinfo;
// } dbrushside_t;

// typedef struct
// {
// 	int firstside;
// 	int numsides;
// 	int contents;
// } dbrush_t;

// #define ANGLE_UP -1
// #define ANGLE_DOWN -2

// /* the visibility lump consists of a header with a count, then
//  * byte offsets for the PVS and PHS of each cluster, then the raw
//  * compressed bit vectors */
// #define DVIS_PVS 0
// #define DVIS_PHS 1
// typedef struct
// {
// 	int numclusters;
// 	int bitofs[8][2]; /* bitofs[numclusters][2] */
// } dvis_t;

// /* each area has a list of portals that lead into other areas
//  * when portals are closed, other areas may not be visible or
//  * hearable even if the vis info says that it should be */
// typedef struct
// {
// 	int portalnum;
// 	int otherarea;
// } dareaportal_t;

// typedef struct
// {
// 	int numareaportals;
// 	int firstareaportal;
// } darea_t;
