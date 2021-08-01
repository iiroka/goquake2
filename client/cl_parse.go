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
 * This file implements the entity and network protocol parsing
 *
 * =======================================================================
 */
package client

import (
	"fmt"
	"goquake2/shared"
	"math"
	"strings"
)

var svc_strings = []string{
	"svc_bad",

	"svc_muzzleflash",
	"svc_muzzlflash2",
	"svc_temp_entity",
	"svc_layout",
	"svc_inventory",

	"svc_nop",
	"svc_disconnect",
	"svc_reconnect",
	"svc_sound",
	"svc_print",
	"svc_stufftext",
	"svc_serverdata",
	"svc_configstring",
	"svc_spawnbaseline",
	"svc_centerprint",
	"svc_download",
	"svc_playerinfo",
	"svc_packetentities",
	"svc_deltapacketentities",
	"svc_frame",
}

/*
 * Returns the entity number and the header bits
 */
func (T *qClient) parseEntityBits(msg *shared.QReadbuf) (int, int) {
	//  unsigned b, total;
	//  int i;
	var number int

	total := msg.ReadByte()

	if (total & shared.U_MOREBITS1) != 0 {
		b := msg.ReadByte() & 0xFF
		total |= b << 8
	}

	if (total & shared.U_MOREBITS2) != 0 {
		b := msg.ReadByte() & 0xFF
		total |= b << 16
	}

	if (total & shared.U_MOREBITS3) != 0 {
		b := msg.ReadByte() & 0xFF
		total |= b << 24
	}

	//  /* count the bits for net profiling */
	//  for (i = 0; i < 32; i++) {
	// 	 if (total & (1 << i)) {
	// 		 bitcounts[i]++;
	// 	 }
	//  }

	if (total & shared.U_NUMBER16) != 0 {
		number = msg.ReadShort()
	} else {
		number = msg.ReadByte()
	}

	return number, total
}

/*
 * Can go from either a baseline or a previous packet_entity
 */
func (T *qClient) parseDelta(msg *shared.QReadbuf, from, to *shared.Entity_state_t, number, bits int) {
	/* set everything to the state we are delta'ing from */
	to.Copy(*from)

	copy(to.Old_origin[:], from.Origin[:])
	to.Number = number

	if (bits & shared.U_MODEL) != 0 {
		to.Modelindex = msg.ReadByte()
	}

	if (bits & shared.U_MODEL2) != 0 {
		to.Modelindex2 = msg.ReadByte()
	}

	if (bits & shared.U_MODEL3) != 0 {
		to.Modelindex3 = msg.ReadByte()
	}

	if (bits & shared.U_MODEL4) != 0 {
		to.Modelindex4 = msg.ReadByte()
	}

	if (bits & shared.U_FRAME8) != 0 {
		to.Frame = msg.ReadByte()
	}

	if (bits & shared.U_FRAME16) != 0 {
		to.Frame = msg.ReadShort()
	}

	/* used for laser colors */
	if (bits&shared.U_SKIN8) != 0 && (bits&shared.U_SKIN16) != 0 {
		to.Skinnum = msg.ReadLong()
	} else if (bits & shared.U_SKIN8) != 0 {
		to.Skinnum = msg.ReadByte()
	} else if (bits & shared.U_SKIN16) != 0 {
		to.Skinnum = msg.ReadShort()
	}

	if (bits & (shared.U_EFFECTS8 | shared.U_EFFECTS16)) == (shared.U_EFFECTS8 | shared.U_EFFECTS16) {
		to.Effects = uint(msg.ReadLong())
	} else if (bits & shared.U_EFFECTS8) != 0 {
		to.Effects = uint(msg.ReadByte())
	} else if (bits & shared.U_EFFECTS16) != 0 {
		to.Effects = uint(msg.ReadShort())
	}

	if (bits & (shared.U_RENDERFX8 | shared.U_RENDERFX16)) == (shared.U_RENDERFX8 | shared.U_RENDERFX16) {
		to.Renderfx = msg.ReadLong()
	} else if (bits & shared.U_RENDERFX8) != 0 {
		to.Renderfx = msg.ReadByte()
	} else if (bits & shared.U_RENDERFX16) != 0 {
		to.Renderfx = msg.ReadShort()
	}

	if (bits & shared.U_ORIGIN1) != 0 {
		to.Origin[0] = msg.ReadCoord()
	}

	if (bits & shared.U_ORIGIN2) != 0 {
		to.Origin[1] = msg.ReadCoord()
	}

	if (bits & shared.U_ORIGIN3) != 0 {
		to.Origin[2] = msg.ReadCoord()
	}

	if (bits & shared.U_ANGLE1) != 0 {
		to.Angles[0] = msg.ReadAngle()
	}

	if (bits & shared.U_ANGLE2) != 0 {
		to.Angles[1] = msg.ReadAngle()
	}

	if (bits & shared.U_ANGLE3) != 0 {
		to.Angles[2] = msg.ReadAngle()
	}

	if (bits & shared.U_OLDORIGIN) != 0 {
		copy(to.Old_origin[0:], msg.ReadPos())
	}

	if (bits & shared.U_SOUND) != 0 {
		to.Sound = msg.ReadByte()
	}

	if (bits & shared.U_EVENT) != 0 {
		to.Event = msg.ReadByte()
	} else {
		to.Event = 0
	}

	if (bits & shared.U_SOLID) != 0 {
		to.Solid = msg.ReadShort()
	}
}

/*
 * Parses deltas from the given base and adds the resulting entity to
 * the current frame
 */
func (T *qClient) deltaEntity(msg *shared.QReadbuf, frame *frame_t, newnum int, old *shared.Entity_state_t, bits int) {

	ent := &T.cl_entities[newnum]

	state := &T.cl_parse_entities[T.cl.parse_entities&(MAX_PARSE_ENTITIES-1)]
	T.cl.parse_entities++
	frame.num_entities++

	T.parseDelta(msg, old, state, newnum, bits)

	/* some data changes will force no lerping */
	if (state.Modelindex != ent.current.Modelindex) ||
		(state.Modelindex2 != ent.current.Modelindex2) ||
		(state.Modelindex3 != ent.current.Modelindex3) ||
		(state.Modelindex4 != ent.current.Modelindex4) ||
		(state.Event == shared.EV_PLAYER_TELEPORT) ||
		(state.Event == shared.EV_OTHER_TELEPORT) ||
		(math.Abs(float64(state.Origin[0]-ent.current.Origin[0])) > 512) ||
		(math.Abs(float64(state.Origin[1]-ent.current.Origin[1])) > 512) ||
		(math.Abs(float64(state.Origin[2]-ent.current.Origin[2])) > 512) {
		ent.serverframe = -99
	}

	/* wasn't in last update, so initialize some things */
	if ent.serverframe != T.cl.frame.serverframe-1 {
		ent.trailcount = 1024 /* for diminishing rocket / grenade trails */

		/* duplicate the current state so
		lerping doesn't hurt anything */
		ent.prev.Copy(*state)

		if state.Event == shared.EV_OTHER_TELEPORT {
			copy(ent.prev.Origin[:], state.Origin[:])
			copy(ent.lerp_origin[:], state.Origin[:])
		} else {
			copy(ent.prev.Origin[:], state.Old_origin[:])
			copy(ent.lerp_origin[:], state.Old_origin[:])
		}
	} else {
		/* shuffle the last state to previous */
		ent.prev.Copy(ent.current)
	}

	ent.serverframe = T.cl.frame.serverframe
	ent.current.Copy(*state)
}

/*
 * An svc_packetentities has just been
 * parsed, deal with the rest of the
 * data stream.
 */
func (T *qClient) parsePacketEntities(msg *shared.QReadbuf, oldframe, newframe *frame_t) error {

	newframe.parse_entities = T.cl.parse_entities
	newframe.num_entities = 0

	/* delta from the entities present in oldframe */
	oldindex := 0
	var oldstate *shared.Entity_state_t
	var oldnum int

	if oldframe == nil {
		oldnum = 99999
	} else {
		if oldindex >= oldframe.num_entities {
			oldnum = 99999
		} else {
			oldstate = &T.cl_parse_entities[(oldframe.parse_entities+oldindex)&(MAX_PARSE_ENTITIES-1)]
			oldnum = oldstate.Number
		}
	}

	for {
		newnum, bits := T.parseEntityBits(msg)
		if newnum >= shared.MAX_EDICTS {
			return T.common.Com_Error(shared.ERR_DROP, "CL_ParsePacketEntities: bad number:%v", newnum)
		}

		if msg.IsOver() {
			return T.common.Com_Error(shared.ERR_DROP, "CL_ParsePacketEntities: end of message")
		}

		if newnum == 0 {
			break
		}

		for oldnum < newnum {
			/* one or more entities from the old packet are unchanged */
			if T.cl_shownet.Int() == 3 {
				T.common.Com_Printf("   unchanged: %v\n", oldnum)
			}

			T.deltaEntity(msg, newframe, oldnum, oldstate, 0)

			oldindex++

			if oldindex >= oldframe.num_entities {
				oldnum = 99999
			} else {
				oldstate = &T.cl_parse_entities[(oldframe.parse_entities+oldindex)&(MAX_PARSE_ENTITIES-1)]
				oldnum = oldstate.Number
			}
		}

		if (bits & shared.U_REMOVE) != 0 {
			/* the entity present in oldframe is not in the current frame */
			if T.cl_shownet.Int() == 3 {
				T.common.Com_Printf("   remove: %v\n", newnum)
			}

			if oldnum != newnum {
				T.common.Com_Printf("U_REMOVE: oldnum != newnum\n")
			}

			oldindex++

			if oldindex >= oldframe.num_entities {
				oldnum = 99999
			} else {
				oldstate = &T.cl_parse_entities[(oldframe.parse_entities+oldindex)&(MAX_PARSE_ENTITIES-1)]
				oldnum = oldstate.Number
			}

			continue
		}

		if oldnum == newnum {
			/* delta from previous state */
			if T.cl_shownet.Int() == 3 {
				T.common.Com_Printf("   delta: %v\n", newnum)
			}

			T.deltaEntity(msg, newframe, newnum, oldstate, bits)

			oldindex++

			if oldindex >= oldframe.num_entities {
				oldnum = 99999
			} else {
				oldstate = &T.cl_parse_entities[(oldframe.parse_entities+oldindex)&(MAX_PARSE_ENTITIES-1)]
				oldnum = oldstate.Number
			}

			continue
		}

		if oldnum > newnum {
			/* delta from baseline */
			if T.cl_shownet.Int() == 3 {
				T.common.Com_Printf("   baseline: %v\n", newnum)
			}

			T.deltaEntity(msg, newframe, newnum,
				&T.cl_entities[newnum].baseline,
				bits)
			continue
		}
	}

	/* any remaining entities in the old frame are copied over */
	for oldnum != 99999 {
		/* one or more entities from the old packet are unchanged */
		if T.cl_shownet.Int() == 3 {
			T.common.Com_Printf("   unchanged: %v\n", oldnum)
		}

		T.deltaEntity(msg, newframe, oldnum, oldstate, 0)

		oldindex++

		if oldindex >= oldframe.num_entities {
			oldnum = 99999
		} else {
			oldstate = &T.cl_parse_entities[(oldframe.parse_entities+oldindex)&(MAX_PARSE_ENTITIES-1)]
			oldnum = oldstate.Number
		}
	}
	return nil
}

func (T *qClient) parsePlayerstate(msg *shared.QReadbuf, oldframe, newframe *frame_t) {

	state := &newframe.playerstate

	/* clear to old value before delta parsing */
	if oldframe != nil {
		state.Copy(oldframe.playerstate)
	} else {
		state.Copy(shared.Player_state_t{})
	}

	flags := msg.ReadShort()

	/* parse the pmove_state_t */
	if (flags & shared.PS_M_TYPE) != 0 {
		state.Pmove.Pm_type = shared.Pmtype_t(msg.ReadByte())
	}

	if (flags & shared.PS_M_ORIGIN) != 0 {
		state.Pmove.Origin[0] = int16(msg.ReadShort())
		state.Pmove.Origin[1] = int16(msg.ReadShort())
		state.Pmove.Origin[2] = int16(msg.ReadShort())
	}

	if (flags & shared.PS_M_VELOCITY) != 0 {
		state.Pmove.Velocity[0] = int16(msg.ReadShort())
		state.Pmove.Velocity[1] = int16(msg.ReadShort())
		state.Pmove.Velocity[2] = int16(msg.ReadShort())
	}

	if (flags & shared.PS_M_TIME) != 0 {
		state.Pmove.Pm_time = uint8(msg.ReadByte())
	}

	if (flags & shared.PS_M_FLAGS) != 0 {
		state.Pmove.Pm_flags = uint8(msg.ReadByte())
	}

	if (flags & shared.PS_M_GRAVITY) != 0 {
		state.Pmove.Gravity = int16(msg.ReadShort())
	}

	if (flags & shared.PS_M_DELTA_ANGLES) != 0 {
		state.Pmove.Delta_angles[0] = int16(msg.ReadShort())
		state.Pmove.Delta_angles[1] = int16(msg.ReadShort())
		state.Pmove.Delta_angles[2] = int16(msg.ReadShort())
	}

	if T.cl.attractloop {
		state.Pmove.Pm_type = shared.PM_FREEZE /* demo playback */
	}

	/* parse the rest of the player_state_t */
	if (flags & shared.PS_VIEWOFFSET) != 0 {
		state.Viewoffset[0] = float32(msg.ReadChar()) * 0.25
		state.Viewoffset[1] = float32(msg.ReadChar()) * 0.25
		state.Viewoffset[2] = float32(msg.ReadChar()) * 0.25
	}

	if (flags & shared.PS_VIEWANGLES) != 0 {
		state.Viewangles[0] = msg.ReadAngle16()
		state.Viewangles[1] = msg.ReadAngle16()
		state.Viewangles[2] = msg.ReadAngle16()
	}

	if (flags & shared.PS_KICKANGLES) != 0 {
		state.Kick_angles[0] = float32(msg.ReadChar()) * 0.25
		state.Kick_angles[1] = float32(msg.ReadChar()) * 0.25
		state.Kick_angles[2] = float32(msg.ReadChar()) * 0.25
	}

	if (flags & shared.PS_WEAPONINDEX) != 0 {
		state.Gunindex = msg.ReadByte()
	}

	if (flags & shared.PS_WEAPONFRAME) != 0 {
		state.Gunframe = msg.ReadByte()
		state.Gunoffset[0] = float32(msg.ReadChar()) * 0.25
		state.Gunoffset[1] = float32(msg.ReadChar()) * 0.25
		state.Gunoffset[2] = float32(msg.ReadChar()) * 0.25
		state.Gunangles[0] = float32(msg.ReadChar()) * 0.25
		state.Gunangles[1] = float32(msg.ReadChar()) * 0.25
		state.Gunangles[2] = float32(msg.ReadChar()) * 0.25
	}

	if (flags & shared.PS_BLEND) != 0 {
		state.Blend[0] = float32(msg.ReadByte()) / 255.0
		state.Blend[1] = float32(msg.ReadByte()) / 255.0
		state.Blend[2] = float32(msg.ReadByte()) / 255.0
		state.Blend[3] = float32(msg.ReadByte()) / 255.0
	}

	if (flags & shared.PS_FOV) != 0 {
		state.Fov = float32(msg.ReadByte())
	}

	if (flags & shared.PS_RDFLAGS) != 0 {
		state.Rdflags = msg.ReadByte()
	}

	/* parse stats */
	statbits := msg.ReadLong()

	for i := 0; i < shared.MAX_STATS; i++ {
		if (statbits & (1 << i)) != 0 {
			state.Stats[i] = int16(msg.ReadShort())
		}
	}
}

func (T *qClient) parseFrame(msg *shared.QReadbuf) error {

	T.cl.frame.copy(frame_t{})

	T.cl.frame.serverframe = msg.ReadLong()
	T.cl.frame.deltaframe = msg.ReadLong()
	T.cl.frame.servertime = T.cl.frame.serverframe * 100

	/* BIG HACK to let old demos continue to work */
	if T.cls.serverProtocol != 26 {
		T.cl.surpressCount = msg.ReadByte()
	}

	if T.cl_shownet.Int() == 3 {
		T.common.Com_Printf("   frame:%v  delta:%v\n", T.cl.frame.serverframe,
			T.cl.frame.deltaframe)
	}

	/* If the frame is delta compressed from data that we
	   no longer have available, we must suck up the rest of
	   the frame, but not use it, then ask for a non-compressed
	   message */
	var old *frame_t
	if T.cl.frame.deltaframe <= 0 {
		T.cl.frame.valid = true /* uncompressed frame */
		old = nil
		// 	cls.demowaiting = false; /* we can start recording now */
	} else {
		old = &T.cl.frames[T.cl.frame.deltaframe&shared.UPDATE_MASK]

		if !old.valid {
			/* should never happen */
			T.common.Com_Printf("Delta from invalid frame (not supposed to happen!).\n")
		}

		if old.serverframe != T.cl.frame.deltaframe {
			/* The frame that the server did the delta from
			   is too old, so we can't reconstruct it properly. */
			T.common.Com_Printf("Delta frame too old.\n")
		} else if T.cl.parse_entities-old.parse_entities > MAX_PARSE_ENTITIES-128 {
			T.common.Com_Printf("Delta parse_entities too old.\n")
		} else {
			T.cl.frame.valid = true /* valid delta parse */
		}
	}

	/* clamp time */
	if T.cl.time > T.cl.frame.servertime {
		T.cl.time = T.cl.frame.servertime
	} else if T.cl.time < T.cl.frame.servertime-100 {
		T.cl.time = T.cl.frame.servertime - 100
	}

	/* read areabits */
	blen := msg.ReadByte()
	msg.ReadData(T.cl.frame.areabits[:], blen)

	/* read playerinfo */
	cmd := msg.ReadByte()
	if cmd >= 0 && cmd < len(svc_strings) {
		T.SHOWNET(svc_strings[cmd], msg)
	}

	if cmd != shared.SvcPlayerinfo {
		return T.common.Com_Error(shared.ERR_DROP, "CL_ParseFrame: 0x%X not playerinfo", cmd)
	}

	T.parsePlayerstate(msg, old, &T.cl.frame)

	// /* read packet entities */
	cmd = msg.ReadByte()
	if cmd >= 0 && cmd < len(svc_strings) {
		T.SHOWNET(svc_strings[cmd], msg)
	}

	if cmd != shared.SvcPacketentities {
		return T.common.Com_Error(shared.ERR_DROP, "CL_ParseFrame: 0x%X not packetentities", cmd)
	}

	if err := T.parsePacketEntities(msg, old, &T.cl.frame); err != nil {
		return err
	}

	/* save the frame off in the backup array for later delta comparisons */
	index := T.cl.frame.serverframe & shared.UPDATE_MASK
	T.cl.frames[index].copy(T.cl.frame)

	if T.cl.frame.valid {
		/* getting a valid frame message ends the connection process */
		if T.cls.state != ca_active {
			T.cls.state = ca_active
			T.cl.force_refdef = true
			for i := 0; i < 3; i++ {
				T.cl.predicted_origin[i] = float32(T.cl.frame.playerstate.Pmove.Origin[i]) * 0.125
				T.cl.predicted_angles[i] = T.cl.frame.playerstate.Viewangles[i]
			}

			// 		if ((cls.disable_servercount != cl.servercount) && cl.refresh_prepped) {
			// 			SCR_EndLoadingPlaque();  /* get rid of loading plaque */
			// 		}

			T.cl.sound_prepped = true

			// 		if (paused_at_load) {
			// 			if (cl_loadpaused->value == 1) {
			// 				Cvar_Set("paused", "0");
			// 			}

			// 			paused_at_load = false;
			// 		}
		}

		// 	/* fire entity events */
		// 	CL_FireEntityEvents(&cl.frame);

		if !(!T.cl_predict.Bool() ||
			(T.cl.frame.playerstate.Pmove.Pm_flags&
				shared.PMF_NO_PREDICTION) != 0) {
			T.checkPredictionError()
		}
	}
	return nil
}

func (T *qClient) parseServerData(msg *shared.QReadbuf) error {

	// /* Clear all key states */
	// In_FlushQueue();

	T.common.Com_DPrintf("Serverdata packet received.\n")

	/* wipe the client_state_t struct */
	T.clearState()
	T.cls.state = ca_connected

	/* parse protocol version number */
	i := msg.ReadLong()
	T.cls.serverProtocol = i

	/* another demo hack */
	if T.common.ServerState() != 0 && (shared.PROTOCOL_VERSION == 34) {
	} else if i != shared.PROTOCOL_VERSION {
		return T.common.Com_Error(shared.ERR_DROP, "Server returned version %v, not %v",
			i, shared.PROTOCOL_VERSION)
	}

	T.cl.servercount = msg.ReadLong()
	T.cl.attractloop = msg.ReadByte() != 0

	/* game directory */
	str := msg.ReadString()
	T.cl.gamedir = str

	// /* set gamedir */
	// if ((*str && (!fs_gamedirvar->string || !*fs_gamedirvar->string ||
	// 	  strcmp(fs_gamedirvar->string, str))) ||
	// 	(!*str && (fs_gamedirvar->string && !*fs_gamedirvar->string)))
	// {
	// 	Cvar_Set("game", str);
	// }

	/* parse player entity number */
	T.cl.playernum = msg.ReadShort()

	/* get the full level name */
	str = msg.ReadString()

	if T.cl.playernum == -1 {
		// 	/* playing a cinematic or showing a pic, not a level */
		// 	SCR_PlayCinematic(str);
	} else {
		// 	/* seperate the printfs so the server
		// 	 * message can have a color */
		// 	Com_Printf("\n\n\35\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\36\37\n\n");
		T.common.Com_Printf("%c%v\n", 2, str)

		/* need to prep refresh at next oportunity */
		T.cl.refresh_prepped = false
	}
	return nil
}

func (T *qClient) parseBaseline(msg *shared.QReadbuf) error {
	newnum, bits := T.parseEntityBits(msg)
	es := &T.cl_entities[newnum].baseline
	T.parseDelta(msg, &shared.Entity_state_t{}, es, newnum, bits)
	return nil
}

func (T *qClient) loadClientinfo(ci *clientinfo_t, s string) {

	ci.cinfo = s

	/* isolate the player's name */
	t := strings.IndexRune(s, '\\')
	if t > 0 {
		ci.name = s[:t]
		s = s[t+1:]
	} else {
		ci.name = s
	}

	if T.cl_noskins.Bool() || t < 0 {
		ci.iconname = "/players/male/grunt_i.pcx"
		ci.model, _ = T.R_RegisterModel("players/male/tris.md2")
		for i := range ci.weaponmodel {
			ci.weaponmodel[i] = nil
		}
		ci.weaponmodel[0], _ = T.R_RegisterModel("players/male/weapon.md2")
		ci.skin = T.R_RegisterSkin("players/male/grunt.pcx")
		ci.icon = T.Draw_FindPic(ci.iconname)
	} else {
		/* isolate the model name */
		var model_name string
		var skin_name string
		t = strings.IndexRune(s, '/')
		if t < 0 {
			t = strings.IndexRune(s, '\\')
		}

		if t < 0 {
			model_name = ""
			skin_name = s
		} else {
			model_name = s[:t]
			skin_name = s[t+1:]
		}

		/* model file */
		model_filename := fmt.Sprintf("players/%s/tris.md2", model_name)
		ci.model, _ = T.R_RegisterModel(model_filename)

		if ci.model == nil {
			model_name = "male"
			model_filename = "players/male/tris.md2"
			ci.model, _ = T.R_RegisterModel(model_filename)
		}

		/* skin file */
		skin_filename := fmt.Sprintf("players/%s/%s.pcx", model_name, skin_name)
		ci.skin = T.R_RegisterSkin(skin_filename)

		// 	/* if we don't have the skin and the model wasn't male,
		// 	 * see if the male has it (this is for CTF's skins) */
		// 	if (!ci->skin && Q_stricmp(model_name, "male"))
		// 	{
		// 		/* change model to male */
		// 		strcpy(model_name, "male");
		// 		Com_sprintf(model_filename, sizeof(model_filename),
		// 				"players/male/tris.md2");
		// 		ci->model = R_RegisterModel(model_filename);

		// 		/* see if the skin exists for the male model */
		// 		Com_sprintf(skin_filename, sizeof(skin_filename),
		// 				"players/%s/%s.pcx", model_name, skin_name);
		// 		ci->skin = R_RegisterSkin(skin_filename);
		// 	}

		/* if we still don't have a skin, it means that the male model didn't have
		 * it, so default to grunt */
		if ci.skin == nil {
			/* see if the skin exists for the male model */
			skin_filename = fmt.Sprintf("players/%s/grunt.pcx", model_name)
			ci.skin = T.R_RegisterSkin(skin_filename)
		}

		/* weapon file */
		for i := 0; i < T.num_cl_weaponmodels; i++ {
			weapon_filename := fmt.Sprintf("players/%s/%s", model_name, T.cl_weaponmodels[i])
			ci.weaponmodel[i], _ = T.R_RegisterModel(weapon_filename)

			// 		if (!ci->weaponmodel[i] && (strcmp(model_name, "cyborg") == 0))
			// 		{
			// 			/* try male */
			// 			Com_sprintf(weapon_filename, sizeof(weapon_filename),
			// 					"players/male/%s", cl_weaponmodels[i]);
			// 			ci->weaponmodel[i] = R_RegisterModel(weapon_filename);
			// 		}

			if !T.cl_vwep.Bool() {
				break /* only one when vwep is off */
			}
		}

		/* icon file */
		ci.iconname = fmt.Sprintf("/players/%s/%s_i.pcx", model_name, skin_name)
		ci.icon = T.Draw_FindPic(ci.iconname)
	}

	// /* must have loaded all data types to be valid */
	// if (!ci->skin || !ci->icon || !ci->model || !ci->weaponmodel[0])
	// {
	// 	ci->skin = NULL;
	// 	ci->icon = NULL;
	// 	ci->model = NULL;
	// 	ci->weaponmodel[0] = NULL;
	// 	return;
	// }
}

/*
 * Load the skin, icon, and model for a client
 */
func (T *qClient) parseClientinfo(player int) {

	s := T.cl.configstrings[player+shared.CS_PLAYERSKINS]

	ci := &T.cl.clientinfo[player]

	T.loadClientinfo(ci, s)
}

func (T *qClient) parseConfigString(msg *shared.QReadbuf) error {
	// 	int i, length;
	// 	char *s;
	// 	char olds[MAX_QPATH];

	i := msg.ReadShort()

	if (i < 0) || (i >= shared.MAX_CONFIGSTRINGS) {
		return T.common.Com_Error(shared.ERR_DROP, "configstring > MAX_CONFIGSTRINGS")
	}

	s := msg.ReadString()

	// 	Q_strlcpy(olds, cl.configstrings[i], sizeof(olds));

	// 	length = strlen(s);
	// 	if (length > sizeof(cl.configstrings) - sizeof(cl.configstrings[0])*i - 1)
	// 	{
	// 		Com_Error(ERR_DROP, "CL_ParseConfigString: oversize configstring");
	// 	}

	T.cl.configstrings[i] = s

	/* do something apropriate */
	if (i >= shared.CS_LIGHTS) && (i < shared.CS_LIGHTS+shared.MAX_LIGHTSTYLES) {
		T.setLightstyle(i - shared.CS_LIGHTS)
	} else if i == shared.CS_CDTRACK {
		// 		if (cl.refresh_prepped)
		// 		{
		// 			OGG_PlayTrack((int)strtol(cl.configstrings[CS_CDTRACK], (char **)NULL, 10));
		// 		}
	} else if (i >= shared.CS_MODELS) && (i < shared.CS_MODELS+shared.MAX_MODELS) {
		// 		if (cl.refresh_prepped)
		// 		{
		// 			cl.model_draw[i - CS_MODELS] = R_RegisterModel(cl.configstrings[i]);

		// 			if (cl.configstrings[i][0] == '*')
		// 			{
		// 				cl.model_clip[i - CS_MODELS] = CM_InlineModel(cl.configstrings[i]);
		// 			}

		// 			else
		// 			{
		// 				cl.model_clip[i - CS_MODELS] = NULL;
		// 			}
		// 		}
	} else if (i >= shared.CS_SOUNDS) && (i < shared.CS_SOUNDS+shared.MAX_MODELS) {
		// 		if (cl.refresh_prepped)
		// 		{
		// 			cl.sound_precache[i - CS_SOUNDS] =
		// 				S_RegisterSound(cl.configstrings[i]);
		// 		}
	} else if (i >= shared.CS_IMAGES) && (i < shared.CS_IMAGES+shared.MAX_MODELS) {
		// 		if (cl.refresh_prepped)
		// 		{
		// 			cl.image_precache[i - CS_IMAGES] = Draw_FindPic(cl.configstrings[i]);
	} else if (i >= shared.CS_PLAYERSKINS) && (i < shared.CS_PLAYERSKINS+shared.MAX_CLIENTS) {
		// 		if (cl.refresh_prepped && strcmp(olds, s))
		// 		{
		// 			CL_ParseClientinfo(i - CS_PLAYERSKINS);
		// 		}
	}
	return nil
}

func (T *qClient) parseStartSoundPacket(msg *shared.QReadbuf) error {

	flags := msg.ReadByte()
	_ = msg.ReadByte()
	// sound_num := msg.ReadByte()

	if (flags & shared.SND_VOLUME) != 0 {
		_ = float32(msg.ReadByte()) / 255.0
		// 	volume = MSG_ReadByte(&net_message) / 255.0f;
	}

	// else
	// {
	// 	volume = DEFAULT_SOUND_PACKET_VOLUME;
	// }

	if (flags & shared.SND_ATTENUATION) != 0 {
		_ = float32(msg.ReadByte()) / 64.0
		// attenuation = MSG_ReadByte(&net_message) / 64.0f;
	}

	// else
	// {
	// 	attenuation = DEFAULT_SOUND_PACKET_ATTENUATION;
	// }

	if (flags & shared.SND_OFFSET) != 0 {
		_ = float32(msg.ReadByte()) / 1000.0
		// 	ofs = MSG_ReadByte(&net_message) / 1000.0f;
	}

	// else
	// {
	// 	ofs = 0;
	// }

	if (flags & shared.SND_ENT) != 0 {
		/* entity reletive */
		_ = msg.ReadShort()
		// 	channel = MSG_ReadShort(&net_message);
		// 	ent = channel >> 3;

		// 	if (ent > MAX_EDICTS)
		// 	{
		// 		Com_Error(ERR_DROP, "CL_ParseStartSoundPacket: ent = %i", ent);
		// 	}

		// 	channel &= 7;
	}
	// else
	// {
	// 	ent = 0;
	// 	channel = 0;
	// }

	if (flags & shared.SND_POS) != 0 {
		/* positioned in space */
		_ = msg.ReadPos()
		// 	MSG_ReadPos(&net_message, pos_v);

		// 	pos = pos_v;
	}
	// else
	// {
	// 	/* use entity number */
	// 	pos = NULL;
	// }

	// if (!cl.sound_precache[sound_num])
	// {
	// 	return;
	// }

	// S_StartSound(pos, ent, channel, cl.sound_precache[sound_num],
	// 		volume, attenuation, ofs);
	return nil
}

func (T *qClient) SHOWNET(s string, msg *shared.QReadbuf) {
	if T.cl_shownet.Int() >= 2 {
		T.common.Com_Printf("%3v:%s\n", msg.Count()-1, s)
	}
}

func (T *qClient) parseServerMessage(msg *shared.QReadbuf) error {
	// int cmd;
	// char *s;
	// int i;

	/* if recording demos, copy the message out */
	if T.cl_shownet.Int() == 1 {
		T.common.Com_Printf("%v ", msg.Size())
	} else if T.cl_shownet.Int() >= 2 {
		T.common.Com_Printf("------------------\n")
	}

	/* parse the message */
	for {
		if msg.IsOver() {
			return T.common.Com_Error(shared.ERR_DROP, "CL_ParseServerMessage: Bad server message")
		}

		cmd := msg.ReadByte()
		if cmd == -1 {
			T.SHOWNET("END OF MESSAGE", msg)
			break
		}

		if T.cl_shownet.Int() >= 2 {
			if cmd < 0 || cmd >= len(svc_strings) {
				T.common.Com_Printf("%3v:BAD CMD %v\n", msg.Count()-1, cmd)
			} else {
				T.SHOWNET(svc_strings[cmd], msg)
			}
		}

		/* other commands */
		switch cmd {

		case shared.SvcNop:

		case shared.SvcDisconnect:
			return T.common.Com_Error(shared.ERR_DISCONNECT, "Server disconnected\n")

			// 		case svc_reconnect:
			// 			Com_Printf("Server disconnected, reconnecting\n");

			// 			if (cls.download)
			// 			{
			// 				/* close download */
			// 				fclose(cls.download);
			// 				cls.download = NULL;
			// 			}

			// 			cls.state = ca_connecting;
			// 			cls.connect_time = -99999; /* CL_CheckForResend() will fire immediately */
			// 			break;

			// 		case svc_print:
			// 			i = MSG_ReadByte(&net_message);

			// 			if (i == PRINT_CHAT)
			// 			{
			// 				S_StartLocalSound("misc/talk.wav");
			// 				con.ormask = 128;
			// 			}

			// 			Com_Printf("%s", MSG_ReadString(&net_message));
			// 			con.ormask = 0;
			// 			break;

			// 		case svc_centerprint:
			// 			SCR_CenterPrint(MSG_ReadString(&net_message));
			// 			break;

		case shared.SvcStufftext:
			s := msg.ReadString()
			T.common.Com_DPrintf("stufftext: %s\n", s)
			T.common.Cbuf_AddText(s)

		case shared.SvcServerdata:
			if err := T.common.Cbuf_Execute(); err != nil { /* make sure any stuffed commands are done */
				return err
			}
			if err := T.parseServerData(msg); err != nil {
				return err
			}

		case shared.SvcConfigstring:
			if err := T.parseConfigString(msg); err != nil {
				return nil
			}

		case shared.SvcSound:
			if err := T.parseStartSoundPacket(msg); err != nil {
				return err
			}

		case shared.SvcSpawnbaseline:
			if err := T.parseBaseline(msg); err != nil {
				return err
			}

		case shared.SvcTempEntity:
			if err := T.parseTEnt(msg); err != nil {
				return err
			}

		case shared.SvcMuzzleflash:
			if err := T.addMuzzleFlash(msg); err != nil {
				return err
			}

		case shared.SvcMuzzleflash2:
			if err := T.addMuzzleFlash2(msg); err != nil {
				return err
			}

			// 		case svc_download:
			// 			CL_ParseDownload();
			// 			break;

		case shared.SvcFrame:
			if err := T.parseFrame(msg); err != nil {
				return err
			}

			// 		case svc_inventory:
			// 			CL_ParseInventory();
			// 			break;

			// 		case svc_layout:
			// 			s = MSG_ReadString(&net_message);
			// 			Q_strlcpy(cl.layout, s, sizeof(cl.layout));
			// 			break;

		case shared.SvcPlayerinfo:
		case shared.SvcPacketentities:
		case shared.SvcDeltapacketentities:
			return T.common.Com_Error(shared.ERR_DROP, "Out of place frame data")
		default:
			return T.common.Com_Error(shared.ERR_DROP, "CL_ParseServerMessage: Illegible server message %v\n", cmd)
		}
	}

	// CL_AddNetgraph();

	// /* we don't know if it is ok to save a demo message
	//    until after we have parsed the frame */
	// if (cls.demorecording && !cls.demowaiting)
	// {
	// 	CL_WriteDemoMessage();
	// }
	return nil
}
