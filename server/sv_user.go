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
 * Server side user (player entity) moving.
 *
 * =======================================================================
 */
package server

import (
	"fmt"
	"goquake2/shared"
)

const maxSTRINGCMDS = 8

func (T *qServer) beginDemoserver() error {

	name := fmt.Sprintf("demos/%s", T.sv.name)
	f, err := T.common.FS_FOpenFile(name, false)
	if err != nil {
		return err
	}

	if f == nil {
		return T.common.Com_Error(shared.ERR_DROP, "Couldn't open %s\n", name)
	}
	T.sv.demofile = f
	return nil
}

/*
 * Sends the first message from the server to a connected client.
 * This will be sent on the initial connection and upon each server load.
 */
func sv_New_f(args []string, T *qServer) error {
	//  static char *gamedir;
	//  int playernum;
	//  edict_t *ent;

	T.common.Com_DPrintf("New() from %s\n", T.sv_client.name)

	if T.sv_client.state != cs_connected {
		T.common.Com_Printf("New not valid -- already spawned\n")
		return nil
	}

	/* demo servers just dump the file message */
	if T.sv.state == ss_demo {
		return T.beginDemoserver()
	}

	/* serverdata needs to go over for all types of servers
	to make sure the protocol is right, and to set the gamedir */
	//  gamedir = (char *)Cvar_VariableString("gamedir");

	/* send the serverdata */
	//  MSG_WriteByte(&sv_client->netchan.message, svc_serverdata);
	//  MSG_WriteLong(&sv_client->netchan.message, PROTOCOL_VERSION);
	//  MSG_WriteLong(&sv_client->netchan.message, svs.spawncount);
	//  MSG_WriteByte(&sv_client->netchan.message, sv.attractloop);
	//  MSG_WriteString(&sv_client->netchan.message, gamedir);

	//  if ((sv.state == ss_cinematic) || (sv.state == ss_pic))
	//  {
	// 	 playernum = -1;
	//  }
	//  else
	//  {
	// 	 playernum = sv_client - svs.clients;
	//  }

	//  MSG_WriteShort(&sv_client->netchan.message, playernum);

	//  /* send full levelname */
	//  MSG_WriteString(&sv_client->netchan.message, sv.configstrings[CS_NAME]);

	//  /* game server */
	//  if (sv.state == ss_game)
	//  {
	// 	 /* set up the entity for the client */
	// 	 ent = EDICT_NUM(playernum + 1);
	// 	 ent->s.number = playernum + 1;
	// 	 sv_client->edict = ent;
	// 	 memset(&sv_client->lastcmd, 0, sizeof(sv_client->lastcmd));

	// 	 /* begin fetching configstrings */
	// 	 MSG_WriteByte(&sv_client->netchan.message, svc_stufftext);
	// 	 MSG_WriteString(&sv_client->netchan.message,
	// 			 va("cmd configstrings %i 0\n", svs.spawncount));
	//  }
	return nil
}

var ucmds = map[string](func([]string, *qServer) error){
	/* auto issued */
	"new": sv_New_f,
	// {"configstrings", SV_Configstrings_f},
	// {"baselines", SV_Baselines_f},
	// {"begin", SV_Begin_f},
	// {"nextserver", SV_Nextserver_f},
	// {"disconnect", SV_Disconnect_f},

	// /* issued by hand at client consoles */
	// {"info", SV_ShowServerinfo_f},

	// {"download", SV_BeginDownload_f},
	// {"nextdl", SV_NextDownload_f},

	// {NULL, NULL}
}

func (T *qServer) executeUserCommand(s string) error {
	// ucmd_t *u;

	/* Security Fix... This is being set to false so that client's can't
	   macro expand variables on the server.  It seems unlikely that a
	   client ever ought to need to be able to do this... */
	args := T.common.Cmd_TokenizeString(s, false)
	println(args[0])
	// sv_player = sv_client->edict;

	if u, ok := ucmds[args[0]]; ok {
		return u(args, T)
	}

	// if (!u->name && (sv.state == ss_game))
	// {
	// 	ge->ClientCommand(sv_player);
	// }
	return nil
}

/*
 * The current net_message is parsed for the given client
 */
func (T *qServer) executeClientMessage(cl *client_t, msg *shared.QReadbuf) error {
	//  int c;
	//  char *s;

	//  usercmd_t nullcmd;
	//  usercmd_t oldest, oldcmd, newcmd;
	//  int net_drop;
	//  int stringCmdCount;
	//  int checksum, calculatedChecksum;
	//  int checksumIndex;
	//  qboolean move_issued;
	//  int lastframe;

	T.sv_client = cl
	//  sv_player = sv_client->edict;

	/* only allow one move command */
	//  move_issued = false;
	stringCmdCount := 0

	for {
		if msg.IsOver() {
			T.common.Com_Printf("SV_ReadClientMessage: badread\n")
			// SV_DropClient(cl)
			return nil
		}

		c := msg.ReadByte()

		if c == -1 {
			break
		}

		switch c {

		case shared.ClcNop:
			break

			// 		 case clc_userinfo:
			// 			 Q_strlcpy(cl->userinfo, MSG_ReadString(&net_message), sizeof(cl->userinfo));
			// 			 SV_UserinfoChanged(cl);
			// 			 break;

			// 		 case clc_move:

			// 			 if (move_issued)
			// 			 {
			// 				 return; /* someone is trying to cheat... */
			// 			 }

			// 			 move_issued = true;
			// 			 checksumIndex = net_message.readcount;
			// 			 checksum = MSG_ReadByte(&net_message);
			// 			 lastframe = MSG_ReadLong(&net_message);

			// 			 if (lastframe != cl->lastframe)
			// 			 {
			// 				 cl->lastframe = lastframe;

			// 				 if (cl->lastframe > 0)
			// 				 {
			// 					 cl->frame_latency[cl->lastframe & (LATENCY_COUNTS - 1)] =
			// 						 svs.realtime - cl->frames[cl->lastframe & UPDATE_MASK].senttime;
			// 				 }
			// 			 }

			// 			 memset(&nullcmd, 0, sizeof(nullcmd));
			// 			 MSG_ReadDeltaUsercmd(&net_message, &nullcmd, &oldest);
			// 			 MSG_ReadDeltaUsercmd(&net_message, &oldest, &oldcmd);
			// 			 MSG_ReadDeltaUsercmd(&net_message, &oldcmd, &newcmd);

			// 			 if (cl->state != cs_spawned)
			// 			 {
			// 				 cl->lastframe = -1;
			// 				 break;
			// 			 }

			// 			 /* if the checksum fails, ignore the rest of the packet */
			// 			 calculatedChecksum = COM_BlockSequenceCRCByte(
			// 				 net_message.data + checksumIndex + 1,
			// 				 net_message.readcount - checksumIndex - 1,
			// 				 cl->netchan.incoming_sequence);

			// 			 if (calculatedChecksum != checksum)
			// 			 {
			// 				 Com_DPrintf("Failed command checksum for %s (%d != %d)/%d\n",
			// 						 cl->name, calculatedChecksum, checksum,
			// 						 cl->netchan.incoming_sequence);
			// 				 return;
			// 			 }

			// 			 if (!sv_paused->value)
			// 			 {
			// 				 net_drop = cl->netchan.dropped;

			// 				 if (net_drop < 20)
			// 				 {
			// 					 while (net_drop > 2)
			// 					 {
			// 						 SV_ClientThink(cl, &cl->lastcmd);

			// 						 net_drop--;
			// 					 }

			// 					 if (net_drop > 1)
			// 					 {
			// 						 SV_ClientThink(cl, &oldest);
			// 					 }

			// 					 if (net_drop > 0)
			// 					 {
			// 						 SV_ClientThink(cl, &oldcmd);
			// 					 }
			// 				 }

			// 				 SV_ClientThink(cl, &newcmd);
			// 			 }

			// 			 cl->lastcmd = newcmd;
			// 			 break;

		case shared.ClcStringcmd:
			s := msg.ReadString()

			/* malicious users may try using too many string commands */
			stringCmdCount++
			if stringCmdCount < maxSTRINGCMDS {
				if err := T.executeUserCommand(s); err != nil {
					return err
				}
			}

			if cl.state == cs_zombie {
				return nil /* disconnect command */
			}

		default:
			T.common.Com_Printf("SV_ReadClientMessage: unknown command char\n")
			// 			 SV_DropClient(cl);
			return nil
		}
	}
	return nil
}
