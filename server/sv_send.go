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
 * Message sending and multiplexing.
 *
 * =======================================================================
 */
package server

import "goquake2/shared"

func (T *qServer) svDemoCompleted() {
	if T.sv.demofile != nil {
		T.sv.demofile.Close()
		T.sv.demofile = nil
	}

	// SV_Nextserver()
}

func (T *qServer) svSendClientMessages() {
	// int i;
	// client_t *c;
	// int msglen;
	// byte msgbuf[MAX_MSGLEN];
	// size_t r;

	// msglen = 0;
	var msgbuf []byte

	/* read the next demo message if needed */
	if T.sv.demofile != nil && (T.sv.state == ss_demo) {
		if !T.sv_paused.Bool() {
			/* get the next message */
			bfr := T.sv.demofile.Read(4)
			if len(bfr) != 4 {
				T.svDemoCompleted()
				return
			}

			msglen := int(shared.ReadInt32(bfr))
			if msglen == -1 {
				T.svDemoCompleted()
				return
			}

			if msglen > shared.MAX_MSGLEN {
				T.common.Com_Error(shared.ERR_DROP,
					"SV_SendClientMessages: msglen > MAX_MSGLEN")
			}

			msgbuf = T.sv.demofile.Read(msglen)
			if len(msgbuf) != msglen {
				T.svDemoCompleted()
				return
			}
		}
	}

	/* send a message to each connected client */
	for i, c := range T.svs.clients {
		if c.state == cs_free {
			continue
		}

		/* if the reliable message
		   overflowed, drop the
		   client */
		// 	if (c->netchan.message.overflowed) {
		// 		SZ_Clear(&c->netchan.message);
		// 		SZ_Clear(&c->datagram);
		// 		SV_BroadcastPrintf(PRINT_HIGH, "%s overflowed\n", c->name);
		// 		SV_DropClient(c);
		// 	}

		if (T.sv.state == ss_cinematic) ||
			(T.sv.state == ss_demo) ||
			(T.sv.state == ss_pic) {
			T.svs.clients[i].netchan.Transmit(msgbuf)
		} else if c.state == cs_spawned {
			// 		/* don't overrun bandwidth */
			// 		if (SV_RateDrop(c))
			// 		{
			// 			continue;
			// 		}

			// 		SV_SendClientDatagram(c);
		} else {
			/* just update reliable	if needed */
			// 		if (c->netchan.message.cursize ||
			// 			(curtime - c->netchan.last_sent > 1000)) {
			// 			Netchan_Transmit(&c->netchan, 0, NULL);
			// 		}
		}
	}
}
