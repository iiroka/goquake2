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
 * Interface between client <-> game and client calculations.
 *
 * =======================================================================
 */
package game

import (
	"goquake2/shared"
	"strconv"
)

/* ======================================================================= */

/*
 * This is only called when the game first
 * initializes in single player, but is called
 * after each death and level change in deathmatch
 */
func (G *qGame) initClientPersistant(client *gclient_t) {
	if client == nil {
		return
	}

	client.pers.copy(client_persistant_t{})

	//  item = FindItem("Blaster");
	//  client->pers.selected_item = ITEM_INDEX(item);
	//  client->pers.inventory[client->pers.selected_item] = 1;

	//  client->pers.weapon = item;

	client.pers.health = 100
	client.pers.max_health = 100

	client.pers.max_bullets = 200
	client.pers.max_shells = 100
	client.pers.max_rockets = 50
	client.pers.max_grenades = 50
	client.pers.max_cells = 200
	client.pers.max_slugs = 50

	client.pers.connected = true
}

/*
 * Chooses a player start, deathmatch start, coop start, etc
 */
func (G *qGame) selectSpawnPoint(ent *edict_t, origin, angles []float32) error {
	//  edict_t *spot = NULL;
	//  edict_t *coopspot = NULL;
	//  int index;
	//  int counter = 0;
	//  vec3_t d;

	if ent == nil {
		return nil
	}

	//  if (deathmatch->value) {
	// 	 spot = SelectDeathmatchSpawnPoint();
	//  }
	//  else if (coop->value)
	//  {
	// 	 spot = SelectCoopSpawnPoint(ent);
	//  }

	/* find a single player start spot */
	var spot *edict_t = nil
	if spot == nil {
		for {
			spot = G.gFind(spot, "Classname", "info_player_start")
			if spot == nil {
				break
			}
			if len(G.game.spawnpoint) == 0 && len(spot.Targetname) == 0 {
				break
			}

			if len(G.game.spawnpoint) == 0 || len(spot.Targetname) == 0 {
				continue
			}

			if G.game.spawnpoint == spot.Targetname {
				break
			}
		}

		if spot == nil {
			if len(G.game.spawnpoint) == 0 {
				/* there wasn't a spawnpoint without a target, so use any */
				spot = G.gFind(spot, "Classname", "info_player_start")
			}

			if spot == nil {
				return G.gi.Error("Couldn't find spawn point %s\n", G.game.spawnpoint)
			}
		}
	}

	/* If we are in coop and we didn't find a coop
	spawnpoint due to map bugs (not correctly
	connected or the map was loaded via console
	and thus no previously map is known to the
	client) use one in 550 units radius. */
	//  if (coop->value) {
	// 	 index = ent->client - game.clients;

	// 	 if (Q_stricmp(spot->classname, "info_player_start") == 0 && index != 0) {
	// 		 while(counter < 3)
	// 		 {
	// 			 coopspot = G_Find(coopspot, FOFS(classname), "info_player_coop");

	// 			 if (!coopspot)
	// 			 {
	// 				 break;
	// 			 }

	// 			 VectorSubtract(coopspot->s.origin, spot->s.origin, d);

	// 			 if ((VectorLength(d) < 550))
	// 			 {
	// 				 if (index == counter)
	// 				 {
	// 					 spot = coopspot;
	// 					 break;
	// 				 }
	// 				 else
	// 				 {
	// 					 counter++;
	// 				 }
	// 			 }
	// 		 }
	// 	 }
	//  }

	copy(origin, spot.s.Origin[:])
	origin[2] += 9
	copy(angles, spot.s.Angles[:])
	return nil
}

/* ============================================================== */

/*
 * Called when a player connects to
 * a server or respawns in a deathmatch.
 */
func (G *qGame) putClientInServer(ent *edict_t) error {
	//  char userinfo[MAX_INFO_STRING];

	if ent == nil {
		return nil
	}

	mins := []float32{-16, -16, -24}
	maxs := []float32{16, 16, 32}
	//  int index;
	//  gclient_t *client;
	//  int i;
	//  client_persistant_t saved;
	//  client_respawn_t resp;

	/* find a spawn point do it before setting
	health back up, so farthest ranging
	doesn't count this client */
	spawn_origin := make([]float32, 3)
	spawn_angles := make([]float32, 3)
	if err := G.selectSpawnPoint(ent, spawn_origin, spawn_angles); err != nil {
		return err
	}

	index := ent.index - 1
	client := ent.client

	resp := client_respawn_t{}
	/* deathmatch wipes most client data every spawn */
	//  if (deathmatch->value)
	//  {
	// 	 resp = client->resp;
	// 	 memcpy(userinfo, client->pers.userinfo, sizeof(userinfo));
	// 	 InitClientPersistant(client);
	// 	 ClientUserinfoChanged(ent, userinfo);
	//  }
	//  else if (coop->value)
	//  {
	// 	 resp = client->resp;
	// 	 memcpy(userinfo, client->pers.userinfo, sizeof(userinfo));
	// 	 resp.coop_respawn.game_helpchanged = client->pers.game_helpchanged;
	// 	 resp.coop_respawn.helpchanged = client->pers.helpchanged;
	// 	 client->pers = resp.coop_respawn;
	// 	 ClientUserinfoChanged(ent, userinfo);

	// 	 if (resp.score > client->pers.score)
	// 	 {
	// 		 client->pers.score = resp.score;
	// 	 }
	//  }

	// userinfo := string(client.pers.userinfo)
	//  ClientUserinfoChanged(ent, userinfo);

	/* clear everything but the persistant data */
	var saved client_persistant_t
	saved.copy(client.pers)
	client.copy(gclient_t{})
	client.pers.copy(saved)

	if client.pers.health <= 0 {
		G.initClientPersistant(client)
	}

	client.resp = resp

	//  /* copy some data from the client to the entity */
	//  FetchClientEntData(ent);

	/* clear entity values */
	//  ent->groundentity = NULL;
	ent.client = &G.game.clients[index]
	//  ent->takedamage = DAMAGE_AIM;
	ent.movetype = MOVETYPE_WALK
	//  ent->viewheight = 22;
	ent.inuse = true
	ent.Classname = "player"
	//  ent->mass = 200;
	ent.solid = shared.SOLID_BBOX
	//  ent->deadflag = DEAD_NO;
	//  ent->air_finished = level.time + 12;
	//  ent->clipmask = MASK_PLAYERSOLID;
	ent.Model = "players/male/tris.md2"
	//  ent->pain = player_pain;
	//  ent->die = player_die;
	//  ent->waterlevel = 0;
	//  ent->watertype = 0;
	//  ent->flags &= ~FL_NO_KNOCKBACK;
	ent.svflags = 0

	copy(ent.mins[:], mins)
	copy(ent.maxs[:], maxs)
	//  VectorClear(ent->velocity);

	/* clear playerstate values */
	client.ps.Copy(shared.Player_state_t{})

	client.ps.Pmove.Origin[0] = int16(spawn_origin[0] * 8)
	client.ps.Pmove.Origin[1] = int16(spawn_origin[1] * 8)
	client.ps.Pmove.Origin[2] = int16(spawn_origin[2] * 8)

	//  if (deathmatch->value && ((int)dmflags->value & DF_FIXED_FOV))
	//  {
	// 	 client->ps.fov = 90;
	//  }
	//  else
	//  {
	fv, _ := strconv.ParseInt(shared.Info_ValueForKey(client.pers.userinfo, "fov"), 10, 32)
	client.ps.Fov = float32(fv)
	if client.ps.Fov < 1 {
		client.ps.Fov = 90
	} else if client.ps.Fov > 160 {
		client.ps.Fov = 160
	}
	//  }

	//  client->ps.gunindex = gi.modelindex(client->pers.weapon->view_model);

	/* clear entity state values */
	ent.s.Effects = 0
	ent.s.Modelindex = 255  /* will use the skin specified model */
	ent.s.Modelindex2 = 255 /* custom gun model */

	/* sknum is player num and weapon number
	weapon number will be added in changeweapon */
	ent.s.Skinnum = ent.index - 1

	ent.s.Frame = 0
	copy(ent.s.Origin[:], spawn_origin)
	ent.s.Origin[2] += 1 /* make sure off ground */
	copy(ent.s.Old_origin[:], ent.s.Origin[:])

	//  /* set the delta angle */
	for i := 0; i < 3; i++ {
		client.ps.Pmove.Delta_angles[i] = shared.ANGLE2SHORT(
			spawn_angles[i] - client.resp.cmd_angles[i])
	}

	ent.s.Angles[shared.PITCH] = 0
	ent.s.Angles[shared.YAW] = spawn_angles[shared.YAW]
	ent.s.Angles[shared.ROLL] = 0
	copy(client.ps.Viewangles[:], ent.s.Angles[:])
	copy(client.v_angle[:], ent.s.Angles[:])

	//  /* spawn a spectator */
	//  if (client->pers.spectator)
	//  {
	// 	 client->chase_target = NULL;

	// 	 client->resp.spectator = true;

	// 	 ent->movetype = MOVETYPE_NOCLIP;
	// 	 ent->solid = SOLID_NOT;
	// 	 ent->svflags |= SVF_NOCLIENT;
	// 	 ent->client->ps.gunindex = 0;
	// 	 gi.linkentity(ent);
	// 	 return;
	//  }
	//  else
	//  {
	client.resp.spectator = false
	//  }

	//  if (!KillBox(ent))
	//  {
	// 	 /* could't spawn in? */
	//  }

	G.gi.Linkentity(ent)

	//  /* force the current weapon up */
	//  client->newweapon = client->pers.weapon;
	//  ChangeWeapon(ent);
	return nil
}

/*
 * QUAKED info_player_start (1 0 0) (-16 -16 -24) (16 16 32)
 * The normal starting point for a level.
 */
func spInfoPlayerStart(self *edict_t, G *qGame) error {
	if self == nil {
		return nil
	}

	/* Call function to hack unnamed spawn points */
	// self->think = SP_CreateUnnamedSpawn;
	self.nextthink = G.level.time + FRAMETIME

	// if (!coop->value) {
	// 	return;
	// }

	// if (Q_stricmp(level.mapname, "security") == 0) {
	// 	/* invoke one of our gross, ugly, disgusting hacks */
	// 	self->think = SP_CreateCoopSpots;
	// 	self->nextthink = level.time + FRAMETIME;
	// }
	return nil
}

func (G *qGame) initClientResp(client *gclient_t) {
	if client == nil {
		return
	}

	// memset(&client->resp, 0, sizeof(client->resp));
	// client->resp.enterframe = level.framenum;
	// client->resp.coop_respawn = client->pers;
}

/*
 * called when a client has finished connecting, and is ready
 * to be placed into the game.  This will happen every level load.
 */
func (G *qGame) ClientBegin(sent shared.Edict_s) error {
	//  int i;

	ent := sent.(*edict_t)
	if ent == nil {
		return nil
	}

	ent.client = &G.game.clients[ent.index-1]

	//  if (deathmatch->value) {
	// 	 ClientBeginDeathmatch(ent);
	// 	 return;
	//  }

	/* if there is already a body waiting for us (a loadgame),
	just take it, otherwise spawn one from scratch */
	if ent.inuse == true {
		/* the client has cleared the client side viewangles upon
		connecting to the server, which is different than the
		state when the game is saved, so we need to compensate
		with deltaangles */
		//  for i := 0; i < 3; i++ {
		// 	 ent->client->ps.pmove.delta_angles[i] = ANGLE2SHORT(
		// 			 ent->client->ps.viewangles[i]);
		//  }
	} else {
		/* a spawn point will completely reinitialize the entity
		except for the persistant data that was initialized at
		ClientConnect() time */
		G_InitEdict(ent, ent.index)
		ent.Classname = "player"
		// InitClientResp(ent.client)
		if err := G.putClientInServer(ent); err != nil {
			return err
		}
	}

	//  if (level.intermissiontime) {
	// 	 MoveClientToIntermission(ent);
	//  } else {
	// 	 /* send effect if in a multiplayer game */
	// 	 if (game.maxclients > 1) {
	// 		 gi.WriteByte(svc_muzzleflash);
	// 		 gi.WriteShort(ent - g_edicts);
	// 		 gi.WriteByte(MZ_LOGIN);
	// 		 gi.multicast(ent->s.origin, MULTICAST_PVS);

	// 		 gi.bprintf(PRINT_HIGH, "%s entered the game\n",
	// 				 ent->client->pers.netname);
	// 	 }
	//  }

	//  /* make sure all view stuff is valid */
	//  ClientEndServerFrame(ent);
	return nil
}

/*
 * Called when a player begins connecting to the server.
 * The game can refuse entrance to a client by returning false.
 * If the client is allowed, the connection process will continue
 * and eventually get to ClientBegin(). Changing levels will NOT
 * cause this to be called again, but loadgames will.
 */
func (G *qGame) ClientConnect(sent shared.Edict_s, userinfo string) bool {

	ent := sent.(*edict_t)
	if ent == nil {
		return false
	}

	/* check to see if they are on the banned IP list */
	// value := shared.Info_ValueForKey(userinfo, "ip")

	//  if (SV_FilterPacket(value)) {
	// 	 Info_SetValueForKey(userinfo, "rejmsg", "Banned.");
	// 	 return false;
	//  }

	//  /* check for a spectator */
	//  value = Info_ValueForKey(userinfo, "spectator");

	//  if (deathmatch->value && *value && strcmp(value, "0"))
	//  {
	// 	 int i, numspec;

	// 	 if (*spectator_password->string &&
	// 		 strcmp(spectator_password->string, "none") &&
	// 		 strcmp(spectator_password->string, value))
	// 	 {
	// 		 Info_SetValueForKey(userinfo, "rejmsg",
	// 				 "Spectator password required or incorrect.");
	// 		 return false;
	// 	 }

	// 	 /* count spectators */
	// 	 for (i = numspec = 0; i < maxclients->value; i++)
	// 	 {
	// 		 if (g_edicts[i + 1].inuse && g_edicts[i + 1].client->pers.spectator)
	// 		 {
	// 			 numspec++;
	// 		 }
	// 	 }

	// 	 if (numspec >= maxspectators->value)
	// 	 {
	// 		 Info_SetValueForKey(userinfo, "rejmsg",
	// 				 "Server spectator limit is full.");
	// 		 return false;
	// 	 }
	//  }
	//  else
	//  {
	// 	 /* check for a password */
	// 	 value = Info_ValueForKey(userinfo, "password");

	// 	 if (*password->string && strcmp(password->string, "none") &&
	// 		 strcmp(password->string, value))
	// 	 {
	// 		 Info_SetValueForKey(userinfo, "rejmsg",
	// 				 "Password required or incorrect.");
	// 		 return false;
	// 	 }
	//  }

	/* they can connect */
	ent.client = &G.game.clients[ent.index-1]

	/* if there is already a body waiting for us (a loadgame),
	just take it, otherwise spawn one from scratch */
	if ent.inuse == false {
		/* clear the respawning variables */
		G.initClientResp(ent.client)

		// 	 if (!game.autosaved || !ent->client->pers.weapon) {
		// 		 InitClientPersistant(ent->client);
		// 	 }
	}

	//  ClientUserinfoChanged(ent, userinfo);

	//  if (game.maxclients > 1) {
	// 	 gi.dprintf("%s connected\n", ent->client->pers.netname);
	//  }

	ent.svflags = 0 /* make sure we start with known default */
	//  ent->client->pers.connected = true;
	return true
}

/*
 * This will be called once for each server
 * frame, before running any other entities
 * in the world.
 */
func (G *qGame) clientBeginServerFrame(ent *edict_t) {
	//  gclient_t *client;
	//  int buttonMask;

	if ent == nil {
		return
	}

	//  if (level.intermissiontime) {
	// 	 return;
	//  }

	client := ent.client

	//  if (deathmatch->value &&
	// 	 (client->pers.spectator != client->resp.spectator) &&
	// 	 ((level.time - client->respawn_time) >= 5))
	//  {
	// 	 spectator_respawn(ent);
	// 	 return;
	//  }

	//  /* run weapon animations if it hasn't been done by a ucmd_t */
	//  if (!client->weapon_thunk && !client->resp.spectator) {
	// 	 Think_Weapon(ent);
	//  } else {
	// 	 client->weapon_thunk = false;
	//  }

	//  if (ent->deadflag)
	//  {
	// 	 /* wait for any button just going down */
	// 	 if (level.time > client->respawn_time)
	// 	 {
	// 		 /* in deathmatch, only wait for attack button */
	// 		 if (deathmatch->value)
	// 		 {
	// 			 buttonMask = BUTTON_ATTACK;
	// 		 }
	// 		 else
	// 		 {
	// 			 buttonMask = -1;
	// 		 }

	// 		 if ((client->latched_buttons & buttonMask) ||
	// 			 (deathmatch->value && ((int)dmflags->value & DF_FORCE_RESPAWN)))
	// 		 {
	// 			 respawn(ent);
	// 			 client->latched_buttons = 0;
	// 		 }
	// 	 }

	// 	 return;
	//  }

	//  /* add player trail so monsters can follow */
	//  if (!deathmatch->value)
	//  {
	// 	 if (!visible(ent, PlayerTrail_LastSpot()))
	// 	 {
	// 		 PlayerTrail_Add(ent->s.old_origin);
	// 	 }
	//  }

	client.latched_buttons = 0
}
