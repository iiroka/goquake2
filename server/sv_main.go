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
 * Server main function and correspondig stuff
 *
 * =======================================================================
 */
package server

import (
	"fmt"
	"goquake2/shared"
	"strconv"
	"time"
)

/*
 * Pull specific info from a newly changed userinfo string
 * into a more C freindly form.
 */
func (T *qServer) userinfoChanged(cl *client_t) {
	//  char *val;
	//  int i;

	//  /* call prog code to allow overrides */
	//  ge->ClientUserinfoChanged(cl->edict, cl->userinfo);

	/* name for C code */
	cl.name = shared.Info_ValueForKey(cl.userinfo, "name")

	/* rate command */
	v := shared.Info_ValueForKey(cl.userinfo, "rate")

	if len(v) > 0 {
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			cl.rate = int(i)
			if cl.rate < 100 {
				cl.rate = 100
			}

			if cl.rate > 15000 {
				cl.rate = 15000
			}
		} else {
			cl.rate = 5000
		}
	} else {
		cl.rate = 5000
	}

	/* msg command */
	//  val = Info_ValueForKey(cl->userinfo, "msg");

	//  if (strlen(val))
	//  {
	// 	 cl->messagelevel = (int)strtol(val, (char **)NULL, 10);
	//  }
}

/*
 * Only called at quake2.exe startup, not for each game
 */
func (T *qServer) Init(common shared.QCommon) error {
	T.common = common

	T.initOperatorCommands()

	T.rcon_password = T.common.Cvar_Get("rcon_password", "", 0)
	T.common.Cvar_Get("skill", "1", 0)
	T.common.Cvar_Get("singleplayer", "0", 0)
	T.common.Cvar_Get("deathmatch", "0", shared.CVAR_LATCH)
	T.common.Cvar_Get("coop", "0", shared.CVAR_LATCH)
	// T.common.Cvar_Get("dmflags", va("%i", DF_INSTANT_ITEMS), CVAR_SERVERINFO);
	T.common.Cvar_Get("fraglimit", "0", shared.CVAR_SERVERINFO)
	T.common.Cvar_Get("timelimit", "0", shared.CVAR_SERVERINFO)
	T.common.Cvar_Get("cheats", "0", shared.CVAR_SERVERINFO|shared.CVAR_LATCH)
	T.common.Cvar_Get("protocol", fmt.Sprintf("%v", shared.PROTOCOL_VERSION), shared.CVAR_SERVERINFO|shared.CVAR_NOSET)
	T.maxclients = T.common.Cvar_Get("maxclients", "1", shared.CVAR_SERVERINFO|shared.CVAR_LATCH)
	T.hostname = T.common.Cvar_Get("hostname", "noname", shared.CVAR_SERVERINFO|shared.CVAR_ARCHIVE)
	T.timeout = T.common.Cvar_Get("timeout", "125", 0)
	T.zombietime = T.common.Cvar_Get("zombietime", "2", 0)
	T.sv_showclamp = T.common.Cvar_Get("showclamp", "1", 0)
	T.sv_paused = T.common.Cvar_Get("paused", "0", 0)
	T.sv_timedemo = T.common.Cvar_Get("timedemo", "0", 0)
	T.sv_enforcetime = T.common.Cvar_Get("sv_enforcetime", "0", 0)
	T.allow_download = T.common.Cvar_Get("allow_download", "1", shared.CVAR_ARCHIVE)
	T.allow_download_players = T.common.Cvar_Get("allow_download_players", "0", shared.CVAR_ARCHIVE)
	T.allow_download_models = T.common.Cvar_Get("allow_download_models", "1", shared.CVAR_ARCHIVE)
	T.allow_download_sounds = T.common.Cvar_Get("allow_download_sounds", "1", shared.CVAR_ARCHIVE)
	T.allow_download_maps = T.common.Cvar_Get("allow_download_maps", "1", shared.CVAR_ARCHIVE)
	T.sv_downloadserver = T.common.Cvar_Get("sv_downloadserver", "", 0)

	T.sv_noreload = T.common.Cvar_Get("sv_noreload", "0", 0)

	T.sv_airaccelerate = T.common.Cvar_Get("sv_airaccelerate", "0", shared.CVAR_LATCH)

	T.public_server = T.common.Cvar_Get("public", "0", 0)

	T.sv_entfile = T.common.Cvar_Get("sv_entfile", "1", shared.CVAR_ARCHIVE)

	// SZ_Init(&net_message, net_message_buffer, sizeof(net_message_buffer));
	return nil
}

func (T *qServer) readPackets() error {
	for {
		from, data := T.common.NET_GetPacket(shared.NS_SERVER)
		if from == nil {
			return nil
		}
		/* check for connectionless packet (0xffffffff) first */
		id := shared.ReadInt32(data)
		println("SV MSG", id)
		if id == -1 {
			T.connectionlessPacket(shared.QReadbufCreate(data), from)
			continue
		}

		msg := shared.QReadbufCreate(data)
		/* read the qport out of the message so we can fix up
		   stupid address translating routers */
		msg.BeginReading()
		msg.ReadLong() /* sequence number */
		msg.ReadLong() /* sequence number */
		qport := msg.ReadShort() & 0xffff

		/* check for packets from connected clients */
		for i, cl := range T.svs.clients {
			if cl.state == cs_free {
				continue
			}

			// 		if (!NET_CompareBaseAdr(net_from, cl->netchan.remote_address))
			// 		{
			// 			continue;
			// 		}

			if cl.netchan.Qport != qport {
				println("Port does not match")
				continue
			}

			// 		if (cl->netchan.remote_address.port != net_from.port)
			// 		{
			// 			Com_Printf("SV_ReadPackets: fixing up a translated port\n");
			// 			cl->netchan.remote_address.port = net_from.port;
			// 		}

			if cl.netchan.Process(msg) {
				/* this is a valid, sequenced packet, so process it */
				if cl.state != cs_zombie {
					cl.lastmessage = T.svs.realtime /* don't timeout */

					if !(T.sv.demofile != nil && (T.sv.state == ss_demo)) {
						if err := T.executeClientMessage(&T.svs.clients[i], msg); err != nil {
							return err
						}
					}
				}
			}

			break
		}
	}
}

func (T *qServer) runGameFrame() {
	// #ifndef DEDICATED_ONLY
	// 	if (host_speeds->value)
	// 	{
	// 		time_before_game = Sys_Milliseconds();
	// 	}
	// #endif

	/* we always need to bump framenum, even if we
	   don't run the world, otherwise the delta
	   compression can get confused when a client
	   has the "current" frame */
	T.sv.framenum++
	T.sv.time = uint(T.sv.framenum * 100)

	/* don't run if paused */
	if !T.sv_paused.Bool() || (T.maxclients.Int() > 1) {
		// 		ge->RunFrame();

		/* never get more than one tic behind */
		if int(T.sv.time) < T.svs.realtime {
			if T.sv_showclamp.Bool() {
				T.common.Com_Printf("sv highclamp\n")
			}

			T.svs.realtime = int(T.sv.time)
		}
	}

	// #ifndef DEDICATED_ONLY
	// 	if (host_speeds->value)
	// 	{
	// 		time_after_game = Sys_Milliseconds();
	// 	}
	// #endif
}

func (T *qServer) Frame(usec int) error {
	// time_before_game = time_after_game = 0;

	/* if server is not active, do nothing */
	if !T.svs.initialized {
		return nil
	}

	T.svs.realtime += usec / 1000

	/* keep the random time dependent */
	shared.Randk()

	// /* check timeouts */
	// SV_CheckTimeouts();

	/* get packets from clients */
	if err := T.readPackets(); err != nil {
		return err
	}

	// /* move autonomous things around if enough time has passed */
	if !T.sv_timedemo.Bool() && (T.svs.realtime < int(T.sv.time)) {
		/* never let the time get too far off */
		if int(T.sv.time)-T.svs.realtime > 100 {
			if T.sv_showclamp.Bool() {
				T.common.Com_Printf("sv lowclamp\n")
			}

			T.svs.realtime = int(T.sv.time - 100)
		}

		time.Sleep(time.Duration(int(T.sv.time)-T.svs.realtime) * time.Millisecond)
		return nil
	}

	// /* update ping based on the last known frame from all clients */
	// SV_CalcPings();

	// /* give the clients some timeslices */
	// SV_GiveMsec();

	/* let everything in the world think and move */
	T.runGameFrame()

	/* send messages back to the clients that had packets read this frame */
	T.svSendClientMessages()

	// /* save the entire world state if recording a serverdemo */
	// SV_RecordDemoMessage();

	// /* send a heartbeat to the master if needed */
	// Master_Heartbeat();

	// /* clear teleport flags, etc for next frame */
	// SV_PrepWorldFrame();
	return nil
}
