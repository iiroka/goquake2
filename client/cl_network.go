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
 * This file implements generic network functions
 *
 * =======================================================================
 */
package client

import (
	"fmt"
	"goquake2/shared"
	"strconv"
	"strings"
)

func cl_ForwardToServer_f(args []string, a interface{}) error {
	T := a.(*qClient)
	if (T.cls.state != ca_connected) && (T.cls.state != ca_active) {
		T.common.Com_Printf("Can't \"%s\", not connected\n", args[0])
		return nil
	}

	/* don't forward the first argument */
	if len(args) > 1 {
		T.cls.netchan.Message.WriteByte(shared.ClcStringcmd)
		var arg strings.Builder
		for i := 1; i < len(args); i++ {
			arg.WriteString(args[i])
			if i < len(args)-1 {
				arg.WriteRune(' ')
			}
		}
		T.cls.netchan.Message.Print(arg.String())
	}
	return nil
}

/*
 * We have gotten a challenge from the server, so try and
 * connect.
 */
func (T *qClient) sendConnectPacket() error {

	adr := shared.NET_StringToAdr(T.cls.servername)
	if adr == nil {
		T.common.Com_Printf("Bad server address\n")
		T.cls.connect_time = 0
		return nil
	}

	if adr.Port == 0 {
		adr.Port = shared.PORT_SERVER
	}

	port := T.common.Cvar_VariableInt("qport")

	T.common.Cvar_ClearUserinfoModified()

	return T.common.Netchan_OutOfBandPrint(shared.NS_CLIENT, *adr, "connect %v %v %v \"%v\"\n",
		shared.PROTOCOL_VERSION, port, T.cls.challenge, T.common.Cvar_Userinfo())
}

/*
 * Resend a connect message if the last one has timed out
 */
func (T *qClient) checkForResend() error {

	/* if the local server is running and we aren't just connect */
	if (T.cls.state == ca_disconnected) && T.common.ServerState() != 0 {
		T.cls.state = ca_connecting
		T.cls.servername = "localhost"
		/* we don't need a challenge on the localhost */
		return T.sendConnectPacket()
	}

	/* resend if we haven't gotten a reply yet */
	if T.cls.state != ca_connecting {
		return nil
	}

	if T.cls.realtime-int(T.cls.connect_time) < 3000 {
		return nil
	}

	adr := shared.NET_StringToAdr(T.cls.servername)
	if adr == nil {
		T.common.Com_Printf("Bad server address\n")
		T.cls.state = ca_disconnected
		return nil
	}

	if adr.Port == 0 {
		adr.Port = shared.PORT_SERVER
	}

	T.cls.connect_time = float32(T.cls.realtime)

	T.common.Com_Printf("Connecting to %v...\n", T.cls.servername)

	return T.common.Netchan_OutOfBandPrint(shared.NS_CLIENT, *adr, "getchallenge\n")
}

/*
 * Goes from a connected state to full screen
 * console state Sends a disconnect message to
 * the server This is also called on Com_Error, so
 * it shouldn't cause any errors
 */
func (T *qClient) disconnect() {
	//  byte final[32];

	if T.cls.state == ca_disconnected {
		return
	}

	//  if (cl_timedemo && cl_timedemo->value) {
	// 	 int time;

	// 	 time = Sys_Milliseconds() - cl.timedemo_start;

	// 	 if (time > 0)
	// 	 {
	// 		 Com_Printf("%i frames, %3.1f seconds: %3.1f fps\n",
	// 				 cl.timedemo_frames, time / 1000.0,
	// 				 cl.timedemo_frames * 1000.0 / time);
	// 	 }
	//  }

	//  VectorClear(cl.refdef.blend);

	//  R_SetPalette(NULL);

	T.mForceMenuOff()

	T.cls.connect_time = 0

	//  SCR_StopCinematic();

	//  OGG_Stop();

	//  if (cls.demorecording)
	//  {
	// 	 CL_Stop_f();
	//  }

	/* send a disconnect message to the server */
	final := shared.QWritebufCreate(32)
	final.WriteByte(shared.ClcStringcmd)
	final.WriteString("disconnect")

	T.cls.netchan.Transmit(final.Data())
	T.cls.netchan.Transmit(final.Data())
	T.cls.netchan.Transmit(final.Data())

	T.clearState()

	/* stop file download */
	// 	 if (cls.download) {
	// 		 fclose(cls.download);
	// 		 cls.download = NULL;
	// 	 }

	//  #ifdef USE_CURL
	// 	 CL_CancelHTTPDownloads(true);
	// 	 cls.downloadReferer[0] = 0;
	// 	 cls.downloadname[0] = 0;
	// 	 cls.downloadposition = 0;
	//  #endif

	T.cls.state = ca_disconnected

	//  snd_is_underwater = false;

	//  // save config for old game/mod
	//  CL_WriteConfiguration();

	// we disconnected, so revert to default game/mod (might have been different mod on MP server)
	//  T.common.Cvar_Set("game", userGivenGame);
}

/*
 * Just sent as a hint to the client that they should
 * drop to full console
 */
func cl_Changing_f(args []string, a interface{}) error {
	T := a.(*qClient)
	/* if we are downloading, we don't change!
	This so we don't suddenly stop downloading a map */
	// 	 if (cls.download) {
	// 		 return;
	// 	 }

	// 	 SCR_BeginLoadingPlaque();
	T.cls.state = ca_connected /* not active anymore, but not disconnected */
	T.common.Com_Printf("\nChanging map...\n")

	//  #ifdef USE_CURL
	// 	 if (cls.downloadServerRetry[0] != 0)
	// 	 {
	// 		 CL_SetHTTPServer(cls.downloadServerRetry);
	// 	 }
	//  #endif
	return nil
}

/*
 * The server is changing levels
 */
func cl_Reconnect_f(args []string, a interface{}) error {
	T := a.(*qClient)
	/* if we are downloading, we don't change!
	This so we don't suddenly stop downloading a map */
	//  if (cls.download) {
	// 	 return;
	//  }

	//  S_StopAllSounds();

	if T.cls.state == ca_connected {
		T.common.Com_Printf("reconnecting...\n")
		T.cls.state = ca_connected
		T.cls.netchan.Message.WriteChar(shared.ClcStringcmd)
		T.cls.netchan.Message.WriteString("new")
		return nil
	}

	if len(T.cls.servername) > 0 {
		if T.cls.state >= ca_connected {
			T.disconnect()
			T.cls.connect_time = float32(T.cls.realtime - 1500)
		} else {
			T.cls.connect_time = -99999 /* Hack: fire immediately */
		}

		T.cls.state = ca_connecting

		T.common.Com_Printf("reconnecting...\n")
	}
	return nil
}

func (T *qClient) clPingServers() {
	// int i;
	// netadr_t adr;
	// char name[32];
	// const char *adrstring;
	// cvar_t *noudp;
	// cvar_t *noipx;

	T.common.NET_Config(true) /* allow remote but do we even need lokal pings? */

	/* send a broadcast packet */
	T.common.Com_Printf("pinging broadcast...\n")

	noudp := T.common.Cvar_Get("noudp", "0", shared.CVAR_NOSET)

	if !noudp.Bool() {
		adr := shared.Netadr_t{}
		adr.Type = shared.NA_BROADCAST
		adr.Port = shared.PORT_SERVER
		T.common.Netchan_OutOfBandPrint(shared.NS_CLIENT, adr, fmt.Sprintf("info %v", shared.PROTOCOL_VERSION))

		T.common.Com_Printf("pinging multicast...\n")
		adr.Type = shared.NA_MULTICAST6
		adr.Port = shared.PORT_SERVER
		T.common.Netchan_OutOfBandPrint(shared.NS_CLIENT, adr, fmt.Sprintf("info %v", shared.PROTOCOL_VERSION))
	}

	// noipx = Cvar_Get("noipx", "0", CVAR_NOSET);

	// if (!noipx->value) {
	// 	adr.type = NA_BROADCAST_IPX;
	// 	adr.port = BigShort(PORT_SERVER);
	// 	Netchan_OutOfBandPrint(NS_CLIENT, adr, va("info %i", PROTOCOL_VERSION));
	// }

	/* send a packet to each address book entry */
	for i := 0; i < 16; i++ {
		name := fmt.Sprintf("adr%v", i)
		adrstring := T.common.Cvar_VariableString(name)

		if len(adrstring) == 0 {
			continue
		}

		T.common.Com_Printf("pinging %s...\n", adrstring)

		// 	if (!NET_StringToAdr(adrstring, &adr)) {
		// 		Com_Printf("Bad address: %s\n", adrstring);
		// 		continue;
		// 	}

		// 	if (!adr.port) {
		// 		adr.port = BigShort(PORT_SERVER);
		// 	}

		// 	Netchan_OutOfBandPrint(NS_CLIENT, adr, va("info %i", PROTOCOL_VERSION));
	}
}

/*
 * Responses to broadcasts, etc
 */
func (T *qClient) connectionlessPacket(msg *shared.QReadbuf, from *shared.Netadr_t) error {

	msg.BeginReading()
	msg.ReadLong() /* skip the -1 */

	s := msg.ReadStringLine()

	args := T.common.Cmd_TokenizeString(s, false)

	T.common.Com_Printf("%s: %s\n", from, args[0])

	/* server connection */
	if args[0] == "client_connect" {
		if T.cls.state == ca_connected {
			T.common.Com_Printf("Dup connect received.  Ignored.\n")
			return nil
		}

		T.cls.netchan.Setup(T.common, shared.NS_CLIENT, *from, T.cls.quakePort)
		// 		 char *buff = NET_AdrToString(cls.netchan.remote_address);

		// 		 for(int i = 1; i < Cmd_Argc(); i++)
		// 		 {
		// 			 char *p = Cmd_Argv(i);

		// 			 if(!strncmp(p, "dlserver=", 9))
		// 			 {
		//  #ifdef USE_CURL
		// 				 p += 9;
		// 				 Com_sprintf(cls.downloadReferer, sizeof(cls.downloadReferer), "quake2://%s", buff);
		// 				 CL_SetHTTPServer (p);

		// 				 if (cls.downloadServer[0])
		// 				 {
		// 					 Com_Printf("HTTP downloading enabled, URL: %s\n", cls.downloadServer);
		// 				 }
		//  #else
		// 				 Com_Printf("HTTP downloading supported by server but not the client.\n");
		//  #endif
		// 			 }
		// 		 }

		// 		 /* Put client into pause mode when connecting to a local server.
		// 			This prevents the world from being forwarded while the client
		// 			is connecting, loading assets, etc. It's not 100%, there're
		// 			still 4 world frames (for baseq2) processed in the game and
		// 			100 frames by the server if the player enters a level that he
		// 			or she already visited. In practise both shouldn't be a big
		// 			problem. 4 frames are hardly enough for monsters staring to
		// 			attack and in most levels the starting area in unreachable by
		// 			monsters and free from environmental effects.

		// 			Com_Serverstate() returns 2 if the server is local and we're
		// 			running a real game and no timedemo, cinematic, etc. The 2 is
		// 			taken from the server_state_t enum value 'ss_game'. If it's a
		// 			local server, maxclients aus either 0 (for single player), or
		// 			2 to 8 (coop and deathmatch) if we're reaching this code.
		// 			For remote servers it's always 1. So this should trigger only
		// 			if it's a local single player server.

		// 			Since the player can load savegames from a paused state (e.g.
		// 			through the console) we'll need to communicate if we entered
		// 			paused mode (and it should left as soon as the player joined
		// 			the server) or if it was already there.

		// 			Last but not least this can be disabled by cl_loadpaused 0. */
		// 		 if (Com_ServerState() == 2 && (Cvar_VariableValue("maxclients") <= 1))
		// 		 {
		// 			 if (cl_loadpaused->value)
		// 			 {
		// 				 if (!cl_paused->value)
		// 				 {
		// 					 paused_at_load = true;
		// 					 Cvar_Set("paused", "1");
		// 				 }
		// 			 }
		// 		 }

		T.cls.netchan.Message.WriteChar(shared.ClcStringcmd)
		T.cls.netchan.Message.WriteString("new")
		T.cls.state = ca_connected
		return nil
	}

	// 	 /* server responding to a status broadcast */
	// 	 if (!strcmp(c, "info"))
	// 	 {
	// 		 CL_ParseStatusMessage();
	// 		 return;
	// 	 }

	// 	 /* remote command from gui front end */
	// 	 if (!strcmp(c, "cmd"))
	// 	 {
	// 		 if (!NET_IsLocalAddress(net_from))
	// 		 {
	// 			 Com_Printf("Command packet from remote host.  Ignored.\n");
	// 			 return;
	// 		 }

	// 		 s = MSG_ReadString(&net_message);
	// 		 Cbuf_AddText(s);
	// 		 Cbuf_AddText("\n");
	// 		 return;
	// 	 }

	// 	 /* print command from somewhere */
	// 	 if (!strcmp(c, "print"))
	// 	 {
	// 		 s = MSG_ReadString(&net_message);
	// 		 Com_Printf("%s", s);
	// 		 return;
	// 	 }

	// 	 /* ping from somewhere */
	// 	 if (!strcmp(c, "ping"))
	// 	 {
	// 		 Netchan_OutOfBandPrint(NS_CLIENT, net_from, "ack");
	// 		 return;
	// 	 }

	/* challenge from the server we are connecting to */
	if args[0] == "challenge" {
		i, _ := strconv.ParseInt(args[1], 10, 32)
		T.cls.challenge = int(i)
		T.sendConnectPacket()
		return nil
	}

	// 	 /* echo request from server */
	// 	 if (!strcmp(c, "echo"))
	// 	 {
	// 		 Netchan_OutOfBandPrint(NS_CLIENT, net_from, "%s", Cmd_Argv(1));
	// 		 return;
	// 	 }

	T.common.Com_Printf("Unknown command.\n")
	return nil
}

func (T *qClient) readPackets() error {
	for {
		from, data := T.common.NET_GetPacket(shared.NS_CLIENT)
		if from == nil {
			break
		}
		/* remote command packet */
		id := shared.ReadInt32(data)
		if id == -1 {
			T.connectionlessPacket(shared.QReadbufCreate(data), from)
			continue
		}

		if (T.cls.state == ca_disconnected) || (T.cls.state == ca_connecting) {
			continue /* dump it if not connected */
		}

		msg := shared.QReadbufCreate((data))
		if msg.Size() < 8 {
			T.common.Com_Printf("%v: Runt packet\n", from)
			continue
		}

		// 	/* packet from server */
		// 	if (!NET_CompareAdr(net_from, cls.netchan.remote_address))
		// 	{
		// 		Com_DPrintf("%s:sequenced packet without connection\n",
		// 				NET_AdrToString(net_from));
		// 		continue;
		// 	}

		if !T.cls.netchan.Process(msg) {
			continue /* wasn't accepted for some reason */
		}

		if err := T.parseServerMessage(msg); err != nil {
			return err
		}
	}

	// /* check timeout */
	// if ((cls.state >= ca_connected) &&
	// 	(cls.realtime - cls.netchan.last_received > cl_timeout->value * 1000))
	// {
	// 	if (++cl.timeoutcount > 5)
	// 	{
	// 		Com_Printf("\nServer connection timed out.\n");
	// 		CL_Disconnect();
	// 		return;
	// 	}
	// }

	// else
	// {
	// 	cl.timeoutcount = 0;
	// }
	return nil
}
