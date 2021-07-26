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
 * Connectionless server commands.
 *
 * =======================================================================
 */
package server

import (
	"goquake2/shared"
	"strconv"
)

/*
 * A connection request that did not come from the master
 */
func (T *qServer) directConnect(args []string, adr shared.Netadr_t) error {
	// 	 char userinfo[MAX_INFO_STRING];
	// 	 netadr_t adr;
	// 	 int i;
	// 	 client_t *cl, *newcl;
	// 	 client_t temp;
	// 	 edict_t *ent;
	// 	 int edictnum;
	// 	 int version;
	// 	 int qport;
	// 	 int challenge;

	// 	 adr = net_from;

	T.common.Com_DPrintf("SVC_DirectConnect ()\n")

	version, _ := strconv.ParseInt(args[1], 10, 32)

	if version != shared.PROTOCOL_VERSION {
		T.common.Netchan_OutOfBandPrint(shared.NS_SERVER, adr, "print\nServer is protocol version 34.\n")
		T.common.Com_DPrintf("    rejected connect from version %v\n", version)
		return nil
	}

	qport, _ := strconv.ParseInt(args[2], 10, 32)

	challenge, _ := strconv.ParseInt(args[3], 10, 32)

	userinfo := args[4]

	// 	 /* force the IP key/value pair so the game can filter based on ip */
	// 	 Info_SetValueForKey(userinfo, "ip", NET_AdrToString(net_from));

	// 	 /* attractloop servers are ONLY for local clients */
	// 	 if (sv.attractloop)
	// 	 {
	// 		 if (!NET_IsLocalAddress(adr))
	// 		 {
	// 			 Com_Printf("Remote connect in attract loop.  Ignored.\n");
	// 			 Netchan_OutOfBandPrint(NS_SERVER, adr,
	// 					 "print\nConnection refused.\n");
	// 			 return;
	// 		 }
	// 	 }

	// 	 /* see if the challenge is valid */
	// 	 if (!NET_IsLocalAddress(adr))
	// 	 {
	// 		 for (i = 0; i < MAX_CHALLENGES; i++)
	// 		 {
	// 			 if (NET_CompareBaseAdr(net_from, svs.challenges[i].adr))
	// 			 {
	// 				 if (challenge == svs.challenges[i].challenge)
	// 				 {
	// 					 break; /* good */
	// 				 }

	// 				 Netchan_OutOfBandPrint(NS_SERVER, adr,
	// 						 "print\nBad challenge.\n");
	// 				 return;
	// 			 }
	// 		 }

	// 		 if (i == MAX_CHALLENGES)
	// 		 {
	// 			 Netchan_OutOfBandPrint(NS_SERVER, adr,
	// 					 "print\nNo challenge for address.\n");
	// 			 return;
	// 		 }
	// 	 }

	index := -1
	// 	 newcl = &temp;
	// 	 memset(newcl, 0, sizeof(client_t));

	// 	 /* if there is already a slot for this ip, reuse it */
	// 	 for (i = 0, cl = svs.clients; i < maxclients->value; i++, cl++)
	// 	 {
	// 		 if (cl->state < cs_connected)
	// 		 {
	// 			 continue;
	// 		 }

	// 		 if (NET_CompareBaseAdr(adr, cl->netchan.remote_address) &&
	// 			 ((cl->netchan.qport == qport) ||
	// 			  (adr.port == cl->netchan.remote_address.port)))
	// 		 {
	// 			 if (!NET_IsLocalAddress(adr))
	// 			 {
	// 				 Com_DPrintf("%s:reconnect rejected : too soon\n",
	// 						 NET_AdrToString(adr));
	// 				 return;
	// 			 }

	// 			 Com_Printf("%s:reconnect\n", NET_AdrToString(adr));
	// 			 newcl = cl;
	// 			 goto gotnewcl;
	// 		 }
	// 	 }

	// 	 /* find a client slot */
	// 	 newcl = NULL;

	for i, cl := range T.svs.clients {
		if cl.state == cs_free {
			index = i
			break
		}
	}

	if index < 0 {
		T.common.Netchan_OutOfBandPrint(shared.NS_SERVER, adr, "print\nServer is full.\n")
		T.common.Com_DPrintf("Rejected a connection.\n")
		return nil
	}

	//  gotnewcl:

	/* build a new connection  accept the new client this
	is the only place a client_t is ever initialized */
	T.svs.clients[index] = client_t{}
	// 	 *newcl = temp;
	// 	 sv_client = newcl;
	// 	 edictnum = (newcl - svs.clients) + 1;
	// 	 ent = EDICT_NUM(edictnum);
	// 	 newcl->edict = ent;
	T.svs.clients[index].challenge = int(challenge) /* save challenge for checksumming */

	// 	 /* get the game a chance to reject this connection or modify the userinfo */
	// 	 if (!(ge->ClientConnect(ent, userinfo)))
	// 	 {
	// 		 if (*Info_ValueForKey(userinfo, "rejmsg"))
	// 		 {
	// 			 Netchan_OutOfBandPrint(NS_SERVER, adr,
	// 					 "print\n%s\nConnection refused.\n",
	// 					 Info_ValueForKey(userinfo, "rejmsg"));
	// 		 }
	// 		 else
	// 		 {
	// 			 Netchan_OutOfBandPrint(NS_SERVER, adr,
	// 					 "print\nConnection refused.\n");
	// 		 }

	// 		 Com_DPrintf("Game rejected a connection.\n");
	// 		 return;
	// 	 }

	/* parse some info from the info strings */
	T.svs.clients[index].userinfo = userinfo
	T.userinfoChanged(&T.svs.clients[index])

	// 	 /* send the connect packet to the client */
	// 	 if (sv_downloadserver->string[0])
	// 	 {
	// 		 Netchan_OutOfBandPrint(NS_SERVER, adr, "client_connect dlserver=%s", sv_downloadserver->string);
	// 	 }
	// 	 else
	// 	 {
	T.common.Netchan_OutOfBandPrint(shared.NS_SERVER, adr, "client_connect")
	// 	 }

	T.svs.clients[index].netchan.Setup(T.common, shared.NS_SERVER, adr, int(qport))

	T.svs.clients[index].state = cs_connected

	// 	 SZ_Init(&newcl->datagram, newcl->datagram_buf, sizeof(newcl->datagram_buf));
	// 	 newcl->datagram.allowoverflow = true;
	T.svs.clients[index].lastmessage = T.svs.realtime /* don't timeout */
	T.svs.clients[index].lastconnect = T.svs.realtime
	return nil
}

/*
 * A connectionless packet has four leading 0xff
 * characters to distinguish it from a game channel.
 * Clients that are in the game can still send
 * connectionless packets.
 */
func (T *qServer) connectionlessPacket(msg *shared.QReadbuf, from *shared.Netadr_t) error {
	//  char *s;
	//  char *c;

	msg.BeginReading()
	msg.ReadLong() /* skip the -1 marker */

	s := msg.ReadStringLine()

	args := T.common.Cmd_TokenizeString(s, false)

	T.common.Com_DPrintf("Packet %v : %v\n", from, args[0])

	switch args[0] {
	//  if (!strcmp(c, "ping"))
	//  {
	// 	 SVC_Ping();
	//  }
	//  else if (!strcmp(c, "ack"))
	//  {
	// 	 SVC_Ack();
	//  }
	//  else if (!strcmp(c, "status"))
	//  {
	// 	 SVC_Status();
	//  }
	//  else if (!strcmp(c, "info"))
	//  {
	// 	 SVC_Info();
	//  }
	//  else if (!strcmp(c, "getchallenge"))
	//  {
	// 	 SVC_GetChallenge();
	//  }
	case "connect":
		return T.directConnect(args, *from)
	//  {
	// 	 SVC_DirectConnect();
	//  }
	//  else if (!strcmp(c, "rcon"))
	//  {
	// 	 SVC_RemoteCommand();
	//  }
	//  else
	//  {
	default:
		T.common.Com_Printf("bad connectionless packet from %v:\n%v\n", from, s)
	}
	return nil
}
