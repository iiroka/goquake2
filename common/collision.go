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
 * The collision model. Slaps "boxes" through the world and checks if
 * they collide with the world model, entities or other boxes.
 *
 * =======================================================================
 */
package common

import "goquake2/shared"

type qCollision struct {
	map_cmodels      [shared.MAX_MAP_MODELS]shared.Cmodel_t
	map_name         string
	map_entitystring string
	numcmodels       int
}

func (T *qCommon) cmodLoadSubmodels(l shared.Lump_t, name string, buf []byte) error {
	// dmodel_t *in;
	// cmodel_t *out;
	// int i, j, count;

	// in = (void *)(cmod_base + l->fileofs);

	if (l.Filelen % shared.Dmodel_size) != 0 {
		return T.Com_Error(shared.ERR_DROP, "Mod_LoadSubmodels: funny lump size")
	}

	count := l.Filelen / shared.Dmodel_size

	if count < 1 {
		return T.Com_Error(shared.ERR_DROP, "Map with no models")
	}

	if count > shared.MAX_MAP_MODELS {
		return T.Com_Error(shared.ERR_DROP, "Map has too many models")
	}

	T.collision.numcmodels = int(count)

	for i := 0; i < int(count); i++ {
		src := shared.Dmodel(buf[int(l.Fileofs)+i*shared.Dmodel_size:])
		out := &T.collision.map_cmodels[i]

		for j := 0; j < 3; j++ {
			/* spread the mins / maxs by a pixel */
			out.Mins[j] = src.Mins[j] - 1
			out.Maxs[j] = src.Maxs[j] + 1
			out.Origin[j] = src.Origin[j]
		}

		out.Headnode = int(src.Headnode)
	}
	return nil
}

func (T *qCommon) cmodLoadEntityString(l shared.Lump_t, name string, buf []byte) error {
	// if (sv_entfile->value) {
	// 	char s[MAX_QPATH];
	// 	char *buffer = NULL;
	// 	int nameLen, bufLen;

	// 	nameLen = strlen(name);
	// 	strcpy(s, name);
	// 	s[nameLen-3] = 'e';	s[nameLen-2] = 'n';	s[nameLen-1] = 't';
	// 	bufLen = FS_LoadFile(s, (void **)&buffer);

	// 	if (buffer != NULL && bufLen > 1)
	// 	{
	// 		if (bufLen + 1 > sizeof(map_entitystring))
	// 		{
	// 			Com_Printf("CMod_LoadEntityString: .ent file %s too large: %i > %lu.\n", s, bufLen, (unsigned long)sizeof(map_entitystring));
	// 			FS_FreeFile(buffer);
	// 		}
	// 		else
	// 		{
	// 			Com_Printf ("CMod_LoadEntityString: .ent file %s loaded.\n", s);
	// 			numentitychars = bufLen;
	// 			memcpy(map_entitystring, buffer, bufLen);
	// 			map_entitystring[bufLen] = 0; /* jit entity bug - null terminate the entity string! */
	// 			FS_FreeFile(buffer);
	// 			return;
	// 		}
	// 	}
	// 	else if (bufLen != -1)
	// 	{
	// 		/* If the .ent file is too small, don't load. */
	// 		Com_Printf("CMod_LoadEntityString: .ent file %s too small.\n", s);
	// 		FS_FreeFile(buffer);
	// 	}
	// }

	// numentitychars = l->filelen;
	// if (l.filelen + 1 > sizeof(map_entitystring)) {
	// 	Com_Error(ERR_DROP, "Map has too large entity lump");
	// }

	T.collision.map_entitystring = string(buf[l.Fileofs : l.Fileofs+l.Filelen])
	// memcpy(map_entitystring, cmod_base + l->fileofs, l->filelen);
	// map_entitystring[l->filelen] = 0;
	return nil
}

/*
 * Loads in the map and all submodels
 */
func (T *qCommon) CMLoadMap(name string, clientload bool, checksum *uint32) (*shared.Cmodel_t, error) {
	//  unsigned *buf;
	//  int i;
	//  dheader_t header;
	//  int length;
	//  static unsigned last_checksum;

	//  map_noareas = Cvar_Get("map_noareas", "0", 0);

	if T.collision.map_name == name && (clientload || !T.Cvar_VariableBool("flushmap")) {
		// 	 *checksum = last_checksum;

		// if !clientload {
		// 		 memset(portalopen, 0, sizeof(portalopen));
		// 		 FloodAreaConnections();
		// 	 }

		return &T.collision.map_cmodels[0], nil /* still have the right version */
	}

	//  /* free old stuff */
	//  numplanes = 0;
	//  numnodes = 0;
	//  numleafs = 0;
	T.collision.numcmodels = 0
	//  numvisibility = 0;
	//  numentitychars = 0;
	T.collision.map_entitystring = ""
	T.collision.map_name = ""

	if len(name) == 0 {
		// 	 numleafs = 1;
		// 	 numclusters = 1;
		// 	 numareas = 1;
		*checksum = 0
		return &T.collision.map_cmodels[0], nil /* cinematic servers won't have anything at all */
	}

	buf, err := T.LoadFile(name)
	if err != nil {
		return nil, err
	}
	if buf == nil {
		return nil, T.Com_Error(shared.ERR_DROP, "Couldn't load %s", name)
	}

	//  last_checksum = LittleLong(Com_BlockChecksum(buf, length));
	//  *checksum = last_checksum;

	header := shared.DheaderCreate(buf)
	if header.Version != shared.BSPVERSION {
		return nil, T.Com_Error(shared.ERR_DROP,
			"CMod_LoadBrushModel: %s has wrong version number (%v should be %v)",
			name, header.Version, shared.BSPVERSION)
	}

	//  cmod_base = (byte *)buf;

	/* load into heap */
	//  CMod_LoadSurfaces(&header.lumps[LUMP_TEXINFO]);
	//  CMod_LoadLeafs(&header.lumps[LUMP_LEAFS]);
	//  CMod_LoadLeafBrushes(&header.lumps[LUMP_LEAFBRUSHES]);
	//  CMod_LoadPlanes(&header.lumps[LUMP_PLANES]);
	//  CMod_LoadBrushes(&header.lumps[LUMP_BRUSHES]);
	//  CMod_LoadBrushSides(&header.lumps[LUMP_BRUSHSIDES]);
	if err := T.cmodLoadSubmodels(header.Lumps[shared.LUMP_MODELS], name, buf); err != nil {
		return nil, err
	}
	//  CMod_LoadNodes(&header.lumps[LUMP_NODES]);
	//  CMod_LoadAreas(&header.lumps[LUMP_AREAS]);
	//  CMod_LoadAreaPortals(&header.lumps[LUMP_AREAPORTALS]);
	//  CMod_LoadVisibility(&header.lumps[LUMP_VISIBILITY]);
	/* From kmquake2: adding an extra parameter for .ent support. */
	if err := T.cmodLoadEntityString(header.Lumps[shared.LUMP_ENTITIES], name, buf); err != nil {
		return nil, err
	}

	//  FS_FreeFile(buf);

	//  CM_InitBoxHull();

	//  memset(portalopen, 0, sizeof(portalopen));
	//  FloodAreaConnections();

	T.collision.map_name = name
	return &T.collision.map_cmodels[0], nil
}

func (T *qCommon) CMEntityString() string {
	return T.collision.map_entitystring
}
