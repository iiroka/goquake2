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
 * This file implements the game media download from the server
 *
 * =======================================================================
 */
package client

import (
	"fmt"
	"goquake2/shared"
)

const PLAYER_MULT = 5

/* ENV_CNT is map load, ENV_CNT+1 is first env map */
const ENV_CNT = (shared.CS_PLAYERSKINS + shared.MAX_CLIENTS*PLAYER_MULT)
const TEXTURE_CNT = (ENV_CNT + 13)

func (T *qClient) requestNextDownload() {
	// 	unsigned int map_checksum; /* for detecting cheater maps */
	// 	char fn[MAX_OSPATH];
	// 	dmdl_t *pheader;

	// 	if (precacherIteration == 0) {
	// #if USE_CURL
	// 		// r1q2-style URLs.
	// 		Q_strlcpy(dlquirks.gamedir, cl.gamedir, sizeof(dlquirks.gamedir));
	// #endif
	// 	} else if (precacherIteration == 1) {
	// #if USE_CURL
	// 		// q2pro-style URLs.
	// 		if (cl.gamedir[0] == '\0')
	// 		{
	// 			Q_strlcpy(dlquirks.gamedir, BASEDIRNAME, sizeof(dlquirks.gamedir));
	// 		}
	// 		else
	// 		{
	// 			Q_strlcpy(dlquirks.gamedir, cl.gamedir, sizeof(dlquirks.gamedir));
	// 		}

	// 		// Force another try with the filelist.
	// 		dlquirks.filelist = true;
	// 		gamedirForFilelist = true;
	// #endif
	// 	} else if (precacherIteration == 2) {
	// 		// UDP Fallback.
	// 		forceudp = true;
	// 	} else {
	// 		// Cannot get here.
	// 		assert(1 && "Recursed from UDP fallback case");
	// 	}

	if T.cls.state != ca_connected {
		return
	}

	if !T.allow_download.Bool() && (T.precache_check < ENV_CNT) {
		T.precache_check = ENV_CNT
	}

	if T.precache_check == shared.CS_MODELS {
		T.precache_check = shared.CS_MODELS + 2

		if T.allow_download_maps.Bool() {
			// 			if (!CL_CheckOrDownloadFile(cl.configstrings[CS_MODELS + 1])) {
			// 				return; /* started a download */
			// 			}
		}
	}

	if (T.precache_check >= shared.CS_MODELS) &&
		(T.precache_check < shared.CS_MODELS+shared.MAX_MODELS) {
		// 		if (allow_download_models->value) {
		// 			while (precache_check < CS_MODELS + MAX_MODELS &&
		// 				   cl.configstrings[precache_check][0]) {
		// 				if ((cl.configstrings[precache_check][0] == '*') ||
		// 					(cl.configstrings[precache_check][0] == '#')) {
		// 					precache_check++;
		// 					continue;
		// 				}

		// 				if (precache_model_skin == 0) {
		// 					if (!CL_CheckOrDownloadFile(cl.configstrings[precache_check])) {
		// 						precache_model_skin = 1;
		// 						return; /* started a download */
		// 					}

		// 					precache_model_skin = 1;
		// 				}

		// #ifdef USE_CURL
		// 				/* Wait for the models to download before checking * skins. */
		// 				if (CL_PendingHTTPDownloads())
		// 				{
		// 					return;
		// 				}
		// #endif

		// 				/* checking for skins in the model */
		// 				if (!precache_model) {
		// 					FS_LoadFile(cl.configstrings[precache_check],
		// 							(void **)&precache_model);

		// 					if (!precache_model) {
		// 						precache_model_skin = 0;
		// 						precache_check++;
		// 						continue; /* couldn't load it */
		// 					}

		// 					if (LittleLong(*(unsigned *)precache_model) !=
		// 						IDALIASHEADER)
		// 					{
		// 						/* not an alias model */
		// 						FS_FreeFile(precache_model);
		// 						precache_model = 0;
		// 						precache_model_skin = 0;
		// 						precache_check++;
		// 						continue;
		// 					}

		// 					pheader = (dmdl_t *)precache_model;

		// 					if (LittleLong(pheader->version) != ALIAS_VERSION)
		// 					{
		// 						precache_check++;
		// 						precache_model_skin = 0;
		// 						continue; /* couldn't load it */
		// 					}
		// 				}

		// 				pheader = (dmdl_t *)precache_model;

		// 				while (precache_model_skin - 1 < LittleLong(pheader->num_skins))
		// 				{
		// 					if (!CL_CheckOrDownloadFile((char *)precache_model +
		// 								LittleLong(pheader->ofs_skins) +
		// 								(precache_model_skin - 1) * MAX_SKINNAME))
		// 					{
		// 						precache_model_skin++;
		// 						return; /* started a download */
		// 					}

		// 					precache_model_skin++;
		// 				}

		// 				if (precache_model)
		// 				{
		// 					FS_FreeFile(precache_model);
		// 					precache_model = 0;
		// 				}

		// 				precache_model_skin = 0;

		// 				precache_check++;
		// 			}
		// 		}

		T.precache_check = shared.CS_SOUNDS
	}

	if (T.precache_check >= shared.CS_SOUNDS) &&
		(T.precache_check < shared.CS_SOUNDS+shared.MAX_SOUNDS) {
		// 		if (allow_download_sounds->value)
		// 		{
		// 			if (precache_check == CS_SOUNDS)
		// 			{
		// 				precache_check++;
		// 			}

		// 			while (precache_check < CS_SOUNDS + MAX_SOUNDS &&
		// 				   cl.configstrings[precache_check][0])
		// 			{
		// 				if (cl.configstrings[precache_check][0] == '*')
		// 				{
		// 					precache_check++;
		// 					continue;
		// 				}

		// 				Com_sprintf(fn, sizeof(fn), "sound/%s",
		// 						cl.configstrings[precache_check++]);

		// 				if (!CL_CheckOrDownloadFile(fn))
		// 				{
		// 					return; /* started a download */
		// 				}
		// 			}
		// 		}

		// 		precache_check = CS_IMAGES;
		// 	}

		// 	if ((precache_check >= CS_IMAGES) &&
		// 		(precache_check < CS_IMAGES + MAX_IMAGES))
		// 	{
		// 		if (precache_check == CS_IMAGES)
		// 		{
		// 			precache_check++;
		// 		}

		// 		while (precache_check < CS_IMAGES + MAX_IMAGES &&
		// 			   cl.configstrings[precache_check][0])
		// 		{
		// 			Com_sprintf(fn, sizeof(fn), "pics/%s.pcx",
		// 					cl.configstrings[precache_check++]);

		// 			if (!CL_CheckOrDownloadFile(fn))
		// 			{
		// 				return; /* started a download */
		// 			}
		// 		}

		T.precache_check = shared.CS_PLAYERSKINS
	}

	/* skins are special, since a player has three
	   things to download:  model, weapon model and
	   skin so precache_check is now *3 */
	if (T.precache_check >= shared.CS_PLAYERSKINS) &&
		(T.precache_check < shared.CS_PLAYERSKINS+shared.MAX_CLIENTS*PLAYER_MULT) {
		// 		if (allow_download_players->value)
		// 		{
		// 			while (precache_check < CS_PLAYERSKINS + MAX_CLIENTS * PLAYER_MULT)
		// 			{
		// 				int i, n;
		// 				char model[MAX_QPATH], skin[MAX_QPATH], *p;

		// 				i = (precache_check - CS_PLAYERSKINS) / PLAYER_MULT;
		// 				n = (precache_check - CS_PLAYERSKINS) % PLAYER_MULT;

		// 				if (!cl.configstrings[CS_PLAYERSKINS + i][0])
		// 				{
		// 					precache_check = CS_PLAYERSKINS + (i + 1) * PLAYER_MULT;
		// 					continue;
		// 				}

		// 				if ((p = strchr(cl.configstrings[CS_PLAYERSKINS + i], '\\')) != NULL)
		// 				{
		// 					p++;
		// 				}
		// 				else
		// 				{
		// 					p = cl.configstrings[CS_PLAYERSKINS + i];
		// 				}

		// 				strcpy(model, p);

		// 				p = strchr(model, '/');

		// 				if (!p)
		// 				{
		// 					p = strchr(model, '\\');
		// 				}

		// 				if (p)
		// 				{
		// 					*p++ = 0;
		// 					strcpy(skin, p);
		// 				}

		// 				else
		// 				{
		// 					*skin = 0;
		// 				}

		// 				switch (n)
		// 				{
		// 					case 0: /* model */
		// 						Com_sprintf(fn, sizeof(fn), "players/%s/tris.md2", model);

		// 						if (!CL_CheckOrDownloadFile(fn))
		// 						{
		// 							precache_check = CS_PLAYERSKINS + i * PLAYER_MULT + 1;
		// 							return;
		// 						}

		// 						n++;

		// 					case 1: /* weapon model */
		// 						Com_sprintf(fn, sizeof(fn), "players/%s/weapon.md2", model);

		// 						if (!CL_CheckOrDownloadFile(fn))
		// 						{
		// 							precache_check = CS_PLAYERSKINS + i * PLAYER_MULT + 2;
		// 							return;
		// 						}

		// 						n++;

		// 					case 2: /* weapon skin */
		// 						Com_sprintf(fn, sizeof(fn), "players/%s/weapon.pcx", model);

		// 						if (!CL_CheckOrDownloadFile(fn))
		// 						{
		// 							precache_check = CS_PLAYERSKINS + i * PLAYER_MULT + 3;
		// 							return;
		// 						}

		// 						n++;

		// 					case 3: /* skin */
		// 						Com_sprintf(fn, sizeof(fn), "players/%s/%s.pcx", model, skin);

		// 						if (!CL_CheckOrDownloadFile(fn))
		// 						{
		// 							precache_check = CS_PLAYERSKINS + i * PLAYER_MULT + 4;
		// 							return;
		// 						}

		// 						n++;

		// 					case 4: /* skin_i */
		// 						Com_sprintf(fn, sizeof(fn), "players/%s/%s_i.pcx", model, skin);

		// 						if (!CL_CheckOrDownloadFile(fn))
		// 						{
		// 							precache_check = CS_PLAYERSKINS + i * PLAYER_MULT + 5;
		// 							return; /* started a download */
		// 						}

		// 						/* move on to next model */
		// 						precache_check = CS_PLAYERSKINS + (i + 1) * PLAYER_MULT;
		// 				}
		// 			}
		// 		}
	}

	// #ifdef USE_CURL
	// 	/* Wait for pending downloads. */
	// 	if (CL_PendingHTTPDownloads())
	// 	{
	// 		return;
	// 	}

	// 	if (dlquirks.error)
	// 	{
	// 		dlquirks.error = false;

	// 		/* Mkay, there were download errors. Let's start over. */
	// 		precacherIteration++;
	// 		CL_ResetPrecacheCheck();
	// 		CL_RequestNextDownload();
	// 		return;
	// 	}
	// #endif

	// 	/* precache phase completed */
	// 	if (!dont_restart_texture_stage) {
	// 		precache_check = ENV_CNT + 1;
	// 	}

	var map_checksum uint32
	T.common.CMLoadMap(T.cl.configstrings[shared.CS_MODELS+1], true, &map_checksum)

	// 	if (map_checksum != (int)strtol(cl.configstrings[CS_MAPCHECKSUM], (char **)NULL, 10)) {
	// 		Com_Error(ERR_DROP, "Local map version differs from server: %i != '%s'\n",
	// 				map_checksum, cl.configstrings[CS_MAPCHECKSUM]);
	// 		return;
	// 	}

	// 	if ((precache_check > ENV_CNT) && (precache_check < TEXTURE_CNT)) {
	// 		if (allow_download->value && allow_download_maps->value) {
	// 			while (precache_check < TEXTURE_CNT) {
	// 				int n = precache_check++ - ENV_CNT - 1;

	// 				if (n & 1)
	// 				{
	// 					Com_sprintf(fn, sizeof(fn), "env/%s%s.pcx",
	// 							cl.configstrings[CS_SKY], env_suf[n / 2]);
	// 				}
	// 				else
	// 				{
	// 					Com_sprintf(fn, sizeof(fn), "env/%s%s.tga",
	// 							cl.configstrings[CS_SKY], env_suf[n / 2]);
	// 				}

	// 				if (!CL_CheckOrDownloadFile(fn))
	// 				{
	// 					return;
	// 				}
	// 			}
	// 		}

	// 		precache_check = TEXTURE_CNT;
	// 	}

	// 	if (precache_check == TEXTURE_CNT) {
	// 		precache_check = TEXTURE_CNT + 1;
	// 		precache_tex = 0;
	// 	}

	// 	/* confirm existance of textures, download any that don't exist */
	// 	if (precache_check == TEXTURE_CNT + 1) {
	// 		extern int numtexinfo;
	// 		extern mapsurface_t map_surfaces[];

	// 		if (allow_download->value && allow_download_maps->value)
	// 		{
	// 			while (precache_tex < numtexinfo)
	// 			{
	// 				char fn[MAX_OSPATH];

	// 				sprintf(fn, "textures/%s.wal",
	// 						map_surfaces[precache_tex++].rname);

	// 				if (!CL_CheckOrDownloadFile(fn))
	// 				{
	// 					return; /* started a download */
	// 				}
	// 			}
	// 		}

	// 		precache_check = TEXTURE_CNT + 999;
	// 	}

	// #ifdef USE_CURL
	// 	/* Wait for pending downloads. */
	// 	if (CL_PendingHTTPDownloads())
	// 	{
	// 		return;
	// 	}
	// #endif

	// 	/* This map is done, start over for next map. */
	// 	forceudp = false;
	// 	precacherIteration = 0;
	// 	gamedirForFilelist = false;
	// 	httpSecondChance = true;
	// 	dont_restart_texture_stage = false;

	// #ifdef USE_CURL
	// 	dlquirks.filelist = true;
	// #endif

	// 	CL_RegisterSounds();
	T.prepRefresh()

	T.cls.netchan.Message.WriteByte(shared.ClcStringcmd)
	T.cls.netchan.Message.WriteString(fmt.Sprintf("begin %v\n", T.precache_spawncount))
	T.cls.forcePacket = true
}
