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
 * Server startup.
 *
 * =======================================================================
 */
package server

import (
	"fmt"
	"goquake2/shared"
	"strings"
)

/*
 * Change the server to a new map, taking all connected
 * clients along with it.
 */
func (T *qServer) spawnServer(server, spawnpoint string, serverstate server_state_t,
	attractloop, loadgame bool) error {
	//  int i;
	//  unsigned checksum;

	if attractloop {
		T.common.Cvar_Set("paused", "0")
	}

	T.common.Com_Printf("------- server initialization ------\n")
	T.common.Com_DPrintf("SpawnServer: %s\n", server)

	//  if (sv.demofile) {
	// 	 FS_FCloseFile(sv.demofile);
	//  }

	//  svs.spawncount++; /* any partially connected client will be restarted */
	T.sv.state = ss_dead
	T.common.SetServerState(int(T.sv.state))

	/* wipe the entire per-level structure */
	T.sv = server_t{}
	T.svs.realtime = 0
	T.sv.loadgame = loadgame
	T.sv.attractloop = attractloop

	//  /* save name for levels that don't set message */
	//  strcpy(sv.configstrings[CS_NAME], server);

	//  if (Cvar_VariableValue("deathmatch"))
	//  {
	// 	 sprintf(sv.configstrings[CS_AIRACCEL], "%g", sv_airaccelerate->value);
	// 	 pm_airaccelerate = sv_airaccelerate->value;
	//  }
	//  else
	//  {
	// 	 strcpy(sv.configstrings[CS_AIRACCEL], "0");
	// 	 pm_airaccelerate = 0;
	//  }

	//  SZ_Init(&sv.multicast, sv.multicast_buf, sizeof(sv.multicast_buf));

	T.sv.name = string(server)

	//  /* leave slots at start for clients only */
	//  for (i = 0; i < maxclients->value; i++)
	//  {
	// 	 /* needs to reconnect */
	// 	 if (svs.clients[i].state > cs_connected)
	// 	 {
	// 		 svs.clients[i].state = cs_connected;
	// 	 }

	// 	 svs.clients[i].lastframe = -1;
	//  }

	T.sv.time = 1000

	//  strcpy(sv.configstrings[CS_NAME], server);

	//  if (serverstate != ss_game)
	//  {
	// 	 sv.models[1] = CM_LoadMap("", false, &checksum); /* no real map */
	//  }
	//  else
	//  {
	// 	 Com_sprintf(sv.configstrings[CS_MODELS + 1],
	// 			 sizeof(sv.configstrings[CS_MODELS + 1]), "maps/%s.bsp", server);
	// 	 sv.models[1] = CM_LoadMap(sv.configstrings[CS_MODELS + 1],
	// 			 false, &checksum);
	//  }

	//  Com_sprintf(sv.configstrings[CS_MAPCHECKSUM],
	// 		 sizeof(sv.configstrings[CS_MAPCHECKSUM]),
	// 		 "%i", checksum);

	//  /* clear physics interaction links */
	//  SV_ClearWorld();

	//  for (i = 1; i < CM_NumInlineModels(); i++)
	//  {
	// 	 Com_sprintf(sv.configstrings[CS_MODELS + 1 + i],
	// 			 sizeof(sv.configstrings[CS_MODELS + 1 + i]),
	// 			 "*%i", i);
	// 	 sv.models[i + 1] = CM_InlineModel(sv.configstrings[CS_MODELS + 1 + i]);
	//  }

	/* spawn the rest of the entities on the map */
	T.sv.state = ss_loading
	T.common.SetServerState(int(T.sv.state))

	//  /* load and spawn all other entities */
	//  ge->SpawnEntities(sv.name, CM_EntityString(), spawnpoint);

	//  /* run two frames to allow everything to settle */
	//  ge->RunFrame();
	//  ge->RunFrame();

	//  /* verify game didn't clobber important stuff */
	//  if ((int)checksum !=
	// 	 (int)strtol(sv.configstrings[CS_MAPCHECKSUM], (char **)NULL, 10))
	//  {
	// 	 Com_Error(ERR_DROP, "Game DLL corrupted server configstrings");
	//  }

	/* all precaches are complete */
	T.sv.state = serverstate
	T.common.SetServerState(int(T.sv.state))

	//  /* create a baseline for more efficient communications */
	//  SV_CreateBaseline();

	//  /* check for a savegame */
	//  SV_CheckForSavegame();

	/* set serverinfo variable */
	T.common.Cvar_FullSet("mapname", T.sv.name, shared.CVAR_SERVERINFO|shared.CVAR_NOSET)

	T.common.Com_Printf("------------------------------------\n\n")
	return nil
}

/*
 * A brand new game has been started
 */
func (T *qServer) initGame() {
	// 	 int i;
	// 	 edict_t *ent;
	// 	 char idmaster[32];

	// 	 if (svs.initialized) {
	// 		 /* cause any connected clients to reconnect */
	// 		 SV_Shutdown("Server restarted\n", true);
	// 	 } else {
	// 		 /* make sure the client is down */
	// 		 CL_Drop();
	// 		 SCR_BeginLoadingPlaque();
	// 	 }

	// 	 /* get any latched variable changes (maxclients, etc) */
	// 	 Cvar_GetLatchedVars();

	T.svs.initialized = true

	if T.common.Cvar_VariableBool("singleplayer") {
		T.common.Cvar_FullSet("coop", "0", shared.CVAR_SERVERINFO|shared.CVAR_LATCH)
		T.common.Cvar_FullSet("deathmatch", "0", shared.CVAR_SERVERINFO|shared.CVAR_LATCH)
	}

	if T.common.Cvar_VariableBool("coop") && T.common.Cvar_VariableBool("deathmatch") {
		// 		 T.common.Com_Printf("Deathmatch and Coop both set, disabling Coop\n");
		// 		 T.common.Cvar_FullSet("coop", "0", shared.CVAR_SERVERINFO | shared.CVAR_LATCH);
	}

	// 	 /* dedicated servers can't be single player and are usually DM
	// 		so unless they explicity set coop, force it to deathmatch */
	// 	 if (dedicated->value) {
	// 		 if (!Cvar_VariableValue("singleplayer")) {
	// 			 if (!Cvar_VariableValue("coop")) {
	// 				 T.common.Cvar_FullSet("deathmatch", "1", shared.CVAR_SERVERINFO | shared.CVAR_LATCH);
	// 			 }
	// 		 }
	// 	 }

	// 	 /* init clients */
	if T.common.Cvar_VariableBool("deathmatch") {
		// 		 if (maxclients->value <= 1) {
		// 			 T.common.Cvar_FullSet("maxclients", "8", shared.CVAR_SERVERINFO | shared.CVAR_LATCH);
		// 		 } else if (maxclients->value > MAX_CLIENTS) {
		// 			 T.common.Cvar_FullSet("maxclients", va("%i", MAX_CLIENTS), shared.CVAR_SERVERINFO | shared.CVAR_LATCH);
		// 		 }

		T.common.Cvar_FullSet("singleplayer", "0", 0)
	} else if T.common.Cvar_VariableBool("coop") {
		if (T.maxclients.Int() <= 1) || (T.maxclients.Int() > 4) {
			T.common.Cvar_FullSet("maxclients", "4", shared.CVAR_SERVERINFO|shared.CVAR_LATCH)
		}

		T.common.Cvar_FullSet("singleplayer", "0", 0)
	} else { /* non-deathmatch, non-coop is one player */
		T.common.Cvar_FullSet("maxclients", "1", shared.CVAR_SERVERINFO|shared.CVAR_LATCH)
		T.common.Cvar_FullSet("singleplayer", "1", 0)
	}

	T.svs.spawncount = shared.Randk()
	T.svs.clients = make([]client_t, T.maxclients.Int())
	T.svs.num_client_entities = T.maxclients.Int() * shared.UPDATE_BACKUP * 64
	// 	 svs.client_entities = Z_Malloc( sizeof(entity_state_t) * svs.num_client_entities);

	// 	 /* init network stuff */
	// 	 if (dedicated->value)
	// 	 {
	// 		 if (Cvar_VariableValue("singleplayer"))
	// 		 {
	// 			 NET_Config(true);
	// 		 }
	// 		 else
	// 		 {
	// 			 NET_Config((maxclients->value > 1));
	// 		 }
	// 	 }
	// 	 else
	// 	 {
	// 		 NET_Config((maxclients->value > 1));
	// 	 }

	/* heartbeats will always be sent to the id master */
	T.svs.last_heartbeat = -99999 /* send immediately */
	// 	 Com_sprintf(idmaster, sizeof(idmaster), "192.246.40.37:%i", PORT_MASTER);
	// 	 NET_StringToAdr(idmaster, &master_adr[0]);

	// 	 /* init game */
	// 	 SV_InitGameProgs();

	// 	 for (i = 0; i < maxclients->value; i++)
	// 	 {
	// 		 ent = EDICT_NUM(i + 1);
	// 		 ent->s.number = i + 1;
	// 		 svs.clients[i].edict = ent;
	// 		 memset(&svs.clients[i].lastcmd, 0, sizeof(svs.clients[i].lastcmd));
	// 	 }
}

/*
 * the full syntax is:
 *
 * map [*]<map>$<startspot>+<nextserver>
 *
 * command from the console or progs.
 * Map can also be a.cin, .pcx, or .dm2 file
 * Nextserver is used to allow a cinematic to play, then proceed to
 * another level:
 *
 *  map tram.cin+jail_e3
 */
func (T *qServer) svMap(attractloop bool, levelstring string, loadgame bool) error {

	T.sv.loadgame = loadgame
	T.sv.attractloop = attractloop

	if (T.sv.state == ss_dead) && !T.sv.loadgame {
		T.initGame() /* the game is just starting */
	}

	level := string(levelstring)

	/* if there is a + in the map, set nextserver to the remainder */
	ch := strings.IndexRune(level, '+')
	if ch >= 0 {
		T.common.Cvar_Set("nextserver", fmt.Sprintf("gamemap \"%v\"", level[ch+1:]))
		level = level[:ch]
	} else {
		// use next demo command if list of map commands as empty
		T.common.Cvar_Set("nextserver", T.common.Cvar_VariableString("nextdemo"))
		// and cleanup nextdemo
		T.common.Cvar_Set("nextdemo", "")
	}

	// 	/* hack for end game screen in coop mode */
	// 	if (Cvar_VariableValue("coop") && !Q_stricmp(level, "victory.pcx")) {
	// 		Cvar_Set("nextserver", "gamemap \"*base1\"");
	// 	}

	/* if there is a $, use the remainder as a spawnpoint */
	ch = strings.IndexRune(level, '$')
	spawnpoint := ""
	if ch >= 0 {
		spawnpoint = level[ch+1:]
		level = level[:ch]
	}

	// 	/* skip the end-of-unit flag if necessary */
	// 	l = strlen(level);

	if level[0] == '*' {
		level = level[1:]
	}

	if strings.HasSuffix(level, ".cin") {
		// 		SCR_BeginLoadingPlaque(); /* for local system */
		// 		SV_BroadcastCommand("changing\n");
		if err := T.spawnServer(level, spawnpoint, ss_cinematic, attractloop, loadgame); err != nil {
			return err
		}
	} else if strings.HasSuffix(level, ".dm2") {
		// 		SCR_BeginLoadingPlaque(); /* for local system */
		// 		SV_BroadcastCommand("changing\n");
		if err := T.spawnServer(level, spawnpoint, ss_demo, attractloop, loadgame); err != nil {
			return err
		}
	} else if strings.HasSuffix(level, ".pcx") {
		// 		SCR_BeginLoadingPlaque(); /* for local system */
		// 		SV_BroadcastCommand("changing\n");
		if err := T.spawnServer(level, spawnpoint, ss_pic, attractloop, loadgame); err != nil {
			return err
		}
	} else {
		// 		SCR_BeginLoadingPlaque(); /* for local system */
		// 		SV_BroadcastCommand("changing\n");
		// 		SV_SendClientMessages();
		if err := T.spawnServer(level, spawnpoint, ss_game, attractloop, loadgame); err != nil {
			return err
		}
		// 		Cbuf_CopyToDefer();
	}

	// 	SV_BroadcastCommand("reconnect\n");
	return nil
}
