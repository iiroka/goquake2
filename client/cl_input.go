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

import "goquake2/shared"

func (T *qClient) refreshCmd() {
	// int ms;
	// usercmd_t *cmd;

	// CMD to fill
	cmd := &T.cl.cmds[T.cls.netchan.Outgoing_sequence&(CMD_BACKUP-1)]

	// Calculate delta
	T.frame_msec = T.input.Sys_frame_time - T.old_sys_frame_time

	// Check bounds
	if T.frame_msec < 1 {
		return
	} else if T.frame_msec > 200 {
		T.frame_msec = 200
	}

	// Add movement
	// CL_BaseMove(cmd);
	// IN_Move(cmd);

	// Clamp angels for prediction
	// CL_ClampPitch();

	cmd.Angles[0] = shared.ANGLE2SHORT(T.cl.viewangles[0])
	cmd.Angles[1] = shared.ANGLE2SHORT(T.cl.viewangles[1])
	cmd.Angles[2] = shared.ANGLE2SHORT(T.cl.viewangles[2])

	// Update time for prediction
	ms := int(T.cls.nframetime * 1000.0)

	if ms > 250 {
		ms = 100
	}

	cmd.Msec = byte(ms)

	// Update frame time for the next call
	T.old_sys_frame_time = T.input.Sys_frame_time

	// Important events are send immediately
	// if (((in_attack.state & 2)) || (in_use.state & 2))
	// {
	// 	cls.forcePacket = true;
	// }
}

func (T *qClient) refreshMove() {
	// usercmd_t *cmd;

	// CMD to fill
	// cmd := &T.cl.cmds[T.cls.netchan.Outgoing_sequence&(CMD_BACKUP-1)]

	// Calculate delta
	T.frame_msec = T.input.Sys_frame_time - T.old_sys_frame_time

	// Check bounds
	if T.frame_msec < 1 {
		return
	} else if T.frame_msec > 200 {
		T.frame_msec = 200
	}

	// Add movement
	// CL_BaseMove(cmd);
	// IN_Move(cmd);

	T.old_sys_frame_time = T.input.Sys_frame_time
}

func (T *qClient) finalizeCmd() {

	// CMD to fill
	cmd := &T.cl.cmds[T.cls.netchan.Outgoing_sequence&(CMD_BACKUP-1)]

	// Mouse button events
	// if (in_attack.state & 3) != 0 {
	// 	cmd->buttons |= BUTTON_ATTACK;
	// }

	// in_attack.state &= ~2;

	// if (in_use.state & 3) != 0 {
	// 	cmd->buttons |= BUTTON_USE;
	// }

	// in_use.state &= ~2;

	// // Keyboard events
	// if (anykeydown && cls.key_dest == key_game) {
	// 	cmd->buttons |= BUTTON_ANY;
	// }

	// cmd->impulse = in_impulse;
	// in_impulse = 0;

	// Set light level for muzzle flash
	cmd.Lightlevel = byte(T.cl_lightlevel.Int())
}

func (T *qClient) sendCmd() error {

	/* save this command off for prediction */
	i := T.cls.netchan.Outgoing_sequence & (CMD_BACKUP - 1)
	cmd := &T.cl.cmds[i]
	T.cl.cmd_time[i] = T.cls.realtime /* for netgraph ping calculation */

	T.finalizeCmd()

	T.cl.cmd.Copy(*cmd)

	if (T.cls.state == ca_disconnected) || (T.cls.state == ca_connecting) {
		return nil
	}

	if T.cls.state == ca_connected {
		if T.cls.netchan.Message.Cursize > 0 || (T.common.Curtime()-T.cls.netchan.LastSent > 1000) {
			T.cls.netchan.Transmit(nil)
		}

		return nil
	}

	/* send a userinfo update if needed */
	// if (userinfo_modified) {
	// 	CL_FixUpGender();
	// 	userinfo_modified = false;
	// 	MSG_WriteByte(&cls.netchan.message, clc_userinfo);
	// 	MSG_WriteString(&cls.netchan.message, Cvar_Userinfo());
	// }

	buf := shared.QWritebufCreate(shared.MAX_MSGLEN)

	// if ((cls.realtime > abort_cinematic) && (cl.cinematictime > 0) &&
	// 		!cl.attractloop && (cls.realtime - cl.cinematictime > 1000) &&
	// 		(cls.key_dest == key_game)) {
	// 	/* skip the rest of the cinematic */
	// 	SCR_FinishCinematic();
	// }

	/* begin a client move command */
	buf.WriteByte(shared.ClcMove)

	/* save the position for a checksum byte */
	// checksumIndex = buf.cursize;
	buf.WriteByte(0)

	/* let the server know what the last frame we
	   got was, so the next message can be delta
	   compressed */
	// if (cl_nodelta->value || !cl.frame.valid || cls.demowaiting)
	// {
	// 	MSG_WriteLong(&buf, -1); /* no compression */
	// }
	// else
	// {
	buf.WriteLong(T.cl.frame.serverframe)
	// }

	/* send this and the previous cmds in the message, so
	   if the last packet was dropped, it can be recovered */
	i = (T.cls.netchan.Outgoing_sequence - 2) & (CMD_BACKUP - 1)
	cmd = &T.cl.cmds[i]
	nullcmd := shared.Usercmd_t{}
	buf.WriteDeltaUsercmd(&nullcmd, cmd)
	oldcmd := cmd

	i = (T.cls.netchan.Outgoing_sequence - 1) & (CMD_BACKUP - 1)
	cmd = &T.cl.cmds[i]
	buf.WriteDeltaUsercmd(oldcmd, cmd)
	oldcmd = cmd

	i = (T.cls.netchan.Outgoing_sequence) & (CMD_BACKUP - 1)
	cmd = &T.cl.cmds[i]
	buf.WriteDeltaUsercmd(oldcmd, cmd)

	// /* calculate a checksum over the move commands */
	// buf.data[checksumIndex] = COM_BlockSequenceCRCByte(
	// 		buf.data + checksumIndex + 1, buf.cursize - checksumIndex - 1,
	// 		cls.netchan.outgoing_sequence);

	/* deliver the message */
	T.cls.netchan.Transmit(buf.Data())

	/* Reinit the current cmd buffer */
	cmd = &T.cl.cmds[T.cls.netchan.Outgoing_sequence&(CMD_BACKUP-1)]
	cmd.Copy(shared.Usercmd_t{})
	return nil
}
