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
	map_cmodels [shared.MAX_MAP_MODELS]shared.Cmodel_t
	map_name    string
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
	//  numcmodels = 0;
	//  numvisibility = 0;
	//  numentitychars = 0;
	//  map_entitystring[0] = 0;
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
	//  CMod_LoadSubmodels(&header.lumps[LUMP_MODELS]);
	//  CMod_LoadNodes(&header.lumps[LUMP_NODES]);
	//  CMod_LoadAreas(&header.lumps[LUMP_AREAS]);
	//  CMod_LoadAreaPortals(&header.lumps[LUMP_AREAPORTALS]);
	//  CMod_LoadVisibility(&header.lumps[LUMP_VISIBILITY]);
	/* From kmquake2: adding an extra parameter for .ent support. */
	//  CMod_LoadEntityString(&header.lumps[LUMP_ENTITIES], name);

	//  FS_FreeFile(buf);

	//  CM_InitBoxHull();

	//  memset(portalopen, 0, sizeof(portalopen));
	//  FloodAreaConnections();

	T.collision.map_name = name
	return &T.collision.map_cmodels[0], nil
}
