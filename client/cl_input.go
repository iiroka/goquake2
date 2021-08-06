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

import (
	"goquake2/shared"
	"strconv"
)

func (T *qClient) keyDown(args []string, b *kbutton_t) {
	// int k;
	// char *c;

	var k int = -1
	if len(args[1]) > 0 {
		if kk, err := strconv.ParseInt(args[1], 10, 32); err == nil {
			k = int(kk)
		}
	}

	if (k == b.down[0]) || (k == b.down[1]) {
		return /* repeating key */
	}

	if b.down[0] == 0 {
		b.down[0] = k
	} else if b.down[1] == 0 {
		b.down[1] = k
	} else {
		T.common.Com_Printf("Three keys down for a button!\n")
		return
	}

	if (b.state & 1) != 0 {
		return /* still down */
	}

	/* save timestamp */
	dt, _ := strconv.ParseInt(args[2], 10, 32)
	b.downtime = uint(dt)

	if b.downtime == 0 {
		b.downtime = uint(T.input.Sys_frame_time - 100)
	}

	b.state |= 1 + 2 /* down + impulse down */
}

func in_ForwardDown(args []string, a interface{}) error {
	T := a.(*qClient)
	T.keyDown(args, &T.in_forward)
	return nil
}

func (T *qClient) initInput() {
	// Cmd_AddCommand("centerview", IN_CenterView);
	// Cmd_AddCommand("force_centerview", IN_ForceCenterView);

	// Cmd_AddCommand("+moveup", IN_UpDown);
	// Cmd_AddCommand("-moveup", IN_UpUp);
	// Cmd_AddCommand("+movedown", IN_DownDown);
	// Cmd_AddCommand("-movedown", IN_DownUp);
	// Cmd_AddCommand("+left", IN_LeftDown);
	// Cmd_AddCommand("-left", IN_LeftUp);
	// Cmd_AddCommand("+right", IN_RightDown);
	// Cmd_AddCommand("-right", IN_RightUp);
	T.common.Cmd_AddCommand("+forward", in_ForwardDown, T)
	// Cmd_AddCommand("-forward", IN_ForwardUp);
	// Cmd_AddCommand("+back", IN_BackDown);
	// Cmd_AddCommand("-back", IN_BackUp);
	// Cmd_AddCommand("+lookup", IN_LookupDown);
	// Cmd_AddCommand("-lookup", IN_LookupUp);
	// Cmd_AddCommand("+lookdown", IN_LookdownDown);
	// Cmd_AddCommand("-lookdown", IN_LookdownUp);
	// Cmd_AddCommand("+strafe", IN_StrafeDown);
	// Cmd_AddCommand("-strafe", IN_StrafeUp);
	// Cmd_AddCommand("+moveleft", IN_MoveleftDown);
	// Cmd_AddCommand("-moveleft", IN_MoveleftUp);
	// Cmd_AddCommand("+moveright", IN_MoverightDown);
	// Cmd_AddCommand("-moveright", IN_MoverightUp);
	// Cmd_AddCommand("+speed", IN_SpeedDown);
	// Cmd_AddCommand("-speed", IN_SpeedUp);
	// Cmd_AddCommand("+attack", IN_AttackDown);
	// Cmd_AddCommand("-attack", IN_AttackUp);
	// Cmd_AddCommand("+use", IN_UseDown);
	// Cmd_AddCommand("-use", IN_UseUp);
	// Cmd_AddCommand("impulse", IN_Impulse);
	// Cmd_AddCommand("+klook", IN_KLookDown);
	// Cmd_AddCommand("-klook", IN_KLookUp);

	// cl_nodelta = Cvar_Get("cl_nodelta", "0", 0);
}

/*
 * Returns the fraction of the
 * frame that the key was down
 */
func (T *qClient) keyState(key *kbutton_t) float32 {

	key.state &= 1 /* clear impulses */

	msec := int(key.msec)
	key.msec = 0

	if key.state != 0 {
		/* still down */
		msec += T.input.Sys_frame_time - int(key.downtime)
		key.downtime = uint(T.input.Sys_frame_time)
	}

	val := float32(msec) / float32(T.frame_msec)

	if val < 0 {
		val = 0
	}

	if val > 1 {
		val = 1
	}

	return val
}

/*
 * Send the intended movement message to the server
 */
func (T *qClient) baseMove(cmd *shared.Usercmd_t) {
	//  CL_AdjustAngles();

	cmd.Copy(shared.Usercmd_t{})

	for i := range cmd.Angles {
		cmd.Angles[i] = int16(T.cl.viewangles[i])
	}

	//  if (in_strafe.state & 1)
	//  {
	// 	 cmd->sidemove += cl_sidespeed->value * CL_KeyState(&in_right);
	// 	 cmd->sidemove -= cl_sidespeed->value * CL_KeyState(&in_left);
	//  }

	//  cmd->sidemove += cl_sidespeed->value * CL_KeyState(&in_moveright);
	//  cmd->sidemove -= cl_sidespeed->value * CL_KeyState(&in_moveleft);

	//  cmd->upmove += cl_upspeed->value * CL_KeyState(&in_up);
	//  cmd->upmove -= cl_upspeed->value * CL_KeyState(&in_down);

	//  if (!(in_klook.state & 1))
	//  {
	cmd.Forwardmove += int16(T.cl_forwardspeed.Float() * T.keyState(&T.in_forward))
	// 	 cmd->forwardmove -= cl_forwardspeed->value * CL_KeyState(&in_back);
	//  }

	//  /* adjust for speed key / running */
	//  if ((in_speed.state & 1) ^ (int)(cl_run->value))
	//  {
	// 	 cmd->forwardmove *= 2;
	// 	 cmd->sidemove *= 2;
	// 	 cmd->upmove *= 2;
	//  }
}

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
	T.baseMove(cmd)
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
	T.baseMove(cmd)
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

	cmd.Impulse = byte(T.in_impulse)
	T.in_impulse = 0

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
