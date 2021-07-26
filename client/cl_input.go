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
 * This file implements the input handling like mouse events and
 * keyboard strokes.
 *
 * =======================================================================
 */
package client

func (T *qClient) sendCmd() error {
	// sizebuf_t buf;
	// byte data[128];
	// int i;
	// usercmd_t *cmd, *oldcmd;
	// usercmd_t nullcmd;
	// int checksumIndex;

	// memset(&buf, 0, sizeof(buf));

	// /* save this command off for prediction */
	// i = cls.netchan.outgoing_sequence & (CMD_BACKUP - 1);
	// cmd = &cl.cmds[i];
	// cl.cmd_time[i] = cls.realtime; /* for netgraph ping calculation */

	// CL_FinalizeCmd();

	// cl.cmd = *cmd;

	if (T.cls.state == ca_disconnected) || (T.cls.state == ca_connecting) {
		return nil
	}

	if T.cls.state == ca_connected {
		if T.cls.netchan.Message.Cursize > 0 || (T.common.Curtime()-T.cls.netchan.LastSent > 1000) {
			T.cls.netchan.Transmit(nil)
		}

		return nil
	}

	// /* send a userinfo update if needed */
	// if (userinfo_modified)
	// {
	// 	CL_FixUpGender();
	// 	userinfo_modified = false;
	// 	MSG_WriteByte(&cls.netchan.message, clc_userinfo);
	// 	MSG_WriteString(&cls.netchan.message, Cvar_Userinfo());
	// }

	// SZ_Init(&buf, data, sizeof(data));

	// if ((cls.realtime > abort_cinematic) && (cl.cinematictime > 0) &&
	// 		!cl.attractloop && (cls.realtime - cl.cinematictime > 1000) &&
	// 		(cls.key_dest == key_game))
	// {
	// 	/* skip the rest of the cinematic */
	// 	SCR_FinishCinematic();
	// }

	// /* begin a client move command */
	// MSG_WriteByte(&buf, clc_move);

	// /* save the position for a checksum byte */
	// checksumIndex = buf.cursize;
	// MSG_WriteByte(&buf, 0);

	// /* let the server know what the last frame we
	//    got was, so the next message can be delta
	//    compressed */
	// if (cl_nodelta->value || !cl.frame.valid || cls.demowaiting)
	// {
	// 	MSG_WriteLong(&buf, -1); /* no compression */
	// }
	// else
	// {
	// 	MSG_WriteLong(&buf, cl.frame.serverframe);
	// }

	// /* send this and the previous cmds in the message, so
	//    if the last packet was dropped, it can be recovered */
	// i = (cls.netchan.outgoing_sequence - 2) & (CMD_BACKUP - 1);
	// cmd = &cl.cmds[i];
	// memset(&nullcmd, 0, sizeof(nullcmd));
	// MSG_WriteDeltaUsercmd(&buf, &nullcmd, cmd);
	// oldcmd = cmd;

	// i = (cls.netchan.outgoing_sequence - 1) & (CMD_BACKUP - 1);
	// cmd = &cl.cmds[i];
	// MSG_WriteDeltaUsercmd(&buf, oldcmd, cmd);
	// oldcmd = cmd;

	// i = (cls.netchan.outgoing_sequence) & (CMD_BACKUP - 1);
	// cmd = &cl.cmds[i];
	// MSG_WriteDeltaUsercmd(&buf, oldcmd, cmd);

	// /* calculate a checksum over the move commands */
	// buf.data[checksumIndex] = COM_BlockSequenceCRCByte(
	// 		buf.data + checksumIndex + 1, buf.cursize - checksumIndex - 1,
	// 		cls.netchan.outgoing_sequence);

	// /* deliver the message */
	// Netchan_Transmit(&cls.netchan, buf.cursize, buf.data);

	// /* Reinit the current cmd buffer */
	// cmd = &cl.cmds[cls.netchan.outgoing_sequence & (CMD_BACKUP - 1)];
	// memset(cmd, 0, sizeof(*cmd));
	return nil
}
