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
 * This is the clients main loop as well as some miscelangelous utility
 * and support functions
 *
 * =======================================================================
 */
package client

import (
	"goquake2/shared"
	"strconv"
)

func (T *qClient) clearState() {
	// S_StopAllSounds();
	T.clearEffects()
	// CL_ClearTEnts();

	/* wipe the entire cl structure */
	T.cl = client_state_t{}
	T.cl.configstrings = make([]string, shared.MAX_CONFIGSTRINGS)
	T.cl_entities = make([]centity_t, shared.MAX_EDICTS)
	T.cl.model_draw = make([]interface{}, shared.MAX_MODELS)

	T.cls.netchan.Message.Clear()
}

/*
 * The server will send this command right
 * before allowing the client into the server
 */
func cl_Precache_f(args []string, a interface{}) error {
	T := a.(*qClient)
	/* Yet another hack to let old demos work */
	if len(args) < 2 {
		var map_checksum uint32 /* for detecting cheater maps */
		if _, err := T.common.CMLoadMap(T.cl.configstrings[shared.CS_MODELS+1], true, &map_checksum); err != nil {
			return err
		}
		//  CL_RegisterSounds();
		return T.prepRefresh()
	}

	T.precache_check = shared.CS_MODELS

	cnt, _ := strconv.ParseInt(args[1], 10, 32)
	T.precache_spawncount = int(cnt)
	T.precache_model = nil
	T.precache_model_skin = 0

	T.requestNextDownload()
	return nil
}

func (T *qClient) initLocal() {
	T.cls.state = ca_disconnected
	T.cls.realtime = T.common.Sys_Milliseconds()

	// 	CL_InitInput();

	/* register our variables */
	// T.cin_force43 = Cvar_Get("cin_force43", "1", 0)

	T.cl_add_blend = T.common.Cvar_Get("cl_blend", "1", 0)
	T.cl_add_lights = T.common.Cvar_Get("cl_lights", "1", 0)
	T.cl_add_particles = T.common.Cvar_Get("cl_particles", "1", 0)
	T.cl_add_entities = T.common.Cvar_Get("cl_entities", "1", 0)
	T.cl_kickangles = T.common.Cvar_Get("cl_kickangles", "1", 0)
	T.cl_gun = T.common.Cvar_Get("cl_gun", "2", shared.CVAR_ARCHIVE)
	T.cl_footsteps = T.common.Cvar_Get("cl_footsteps", "1", 0)
	T.cl_noskins = T.common.Cvar_Get("cl_noskins", "0", 0)
	T.cl_predict = T.common.Cvar_Get("cl_predict", "1", 0)
	T.cl_showfps = T.common.Cvar_Get("cl_showfps", "0", shared.CVAR_ARCHIVE)

	// T.cl_upspeed = Cvar_Get("cl_upspeed", "200", 0)
	// T.cl_forwardspeed = Cvar_Get("cl_forwardspeed", "200", 0)
	// T.cl_sidespeed = Cvar_Get("cl_sidespeed", "200", 0)
	// T.cl_yawspeed = Cvar_Get("cl_yawspeed", "140", 0)
	// T.cl_pitchspeed = Cvar_Get("cl_pitchspeed", "150", 0)
	// T.cl_anglespeedkey = Cvar_Get("cl_anglespeedkey", "1.5", 0)

	// T.cl_run = Cvar_Get("cl_run", "0", CVAR_ARCHIVE)

	T.cl_shownet = T.common.Cvar_Get("cl_shownet", "0", 0)
	T.cl_showmiss = T.common.Cvar_Get("cl_showmiss", "0", 0)
	T.cl_showclamp = T.common.Cvar_Get("showclamp", "0", 0)
	T.cl_timeout = T.common.Cvar_Get("cl_timeout", "120", 0)
	T.cl_paused = T.common.Cvar_Get("paused", "0", 0)
	T.cl_loadpaused = T.common.Cvar_Get("cl_loadpaused", "1", shared.CVAR_ARCHIVE)

	T.gl1_stereo = T.common.Cvar_Get("gl1_stereo", "0", shared.CVAR_ARCHIVE)
	T.gl1_stereo_separation = T.common.Cvar_Get("gl1_stereo_separation", "1", shared.CVAR_ARCHIVE)
	T.gl1_stereo_convergence = T.common.Cvar_Get("gl1_stereo_convergence", "1.4", shared.CVAR_ARCHIVE)

	T.rcon_client_password = T.common.Cvar_Get("rcon_password", "", 0)
	T.rcon_address = T.common.Cvar_Get("rcon_address", "", 0)

	T.cl_lightlevel = T.common.Cvar_Get("r_lightlevel", "0", 0)
	T.cl_r1q2_lightstyle = T.common.Cvar_Get("cl_r1q2_lightstyle", "1", shared.CVAR_ARCHIVE)
	T.cl_limitsparksounds = T.common.Cvar_Get("cl_limitsparksounds", "0", shared.CVAR_ARCHIVE)

	/* userinfo */
	T.name = T.common.Cvar_Get("name", "unnamed", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.skin = T.common.Cvar_Get("skin", "male/grunt", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.rate = T.common.Cvar_Get("rate", "8000", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.msg = T.common.Cvar_Get("msg", "1", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.hand = T.common.Cvar_Get("hand", "0", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.fov = T.common.Cvar_Get("fov", "90", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.horplus = T.common.Cvar_Get("horplus", "1", shared.CVAR_ARCHIVE)
	T.windowed_mouse = T.common.Cvar_Get("windowed_mouse", "1", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.gender = T.common.Cvar_Get("gender", "male", shared.CVAR_USERINFO|shared.CVAR_ARCHIVE)
	T.gender_auto = T.common.Cvar_Get("gender_auto", "1", shared.CVAR_ARCHIVE)
	T.gender.Modified = false

	T.allow_download = T.common.Cvar_Get("allow_download", "1", shared.CVAR_ARCHIVE)
	T.allow_download_players = T.common.Cvar_Get("allow_download_players", "0", shared.CVAR_ARCHIVE)
	T.allow_download_models = T.common.Cvar_Get("allow_download_models", "1", shared.CVAR_ARCHIVE)
	T.allow_download_sounds = T.common.Cvar_Get("allow_download_sounds", "1", shared.CVAR_ARCHIVE)
	T.allow_download_maps = T.common.Cvar_Get("allow_download_maps", "1", shared.CVAR_ARCHIVE)

	// USERINFO cvars are special, they just need to be registered
	T.common.Cvar_Get("password", "", shared.CVAR_USERINFO)
	T.common.Cvar_Get("spectator", "0", shared.CVAR_USERINFO)

	T.cl_vwep = T.common.Cvar_Get("cl_vwep", "1", shared.CVAR_ARCHIVE)

	// #ifdef USE_CURL
	// 	cl_http_proxy = Cvar_Get("cl_http_proxy", "", 0);
	// 	cl_http_filelists = Cvar_Get("cl_http_filelists", "1", 0);
	// 	cl_http_downloads = Cvar_Get("cl_http_downloads", "1", CVAR_ARCHIVE);
	// 	cl_http_max_connections = Cvar_Get("cl_http_max_connections", "4", 0);
	// #endif

	// 	/* register our commands */
	T.common.Cmd_AddCommand("cmd", cl_ForwardToServer_f, T)
	// 	Cmd_AddCommand("pause", CL_Pause_f);
	// 	Cmd_AddCommand("pingservers", CL_PingServers_f);
	// 	Cmd_AddCommand("skins", CL_Skins_f);

	// 	Cmd_AddCommand("userinfo", CL_Userinfo_f);
	// 	Cmd_AddCommand("snd_restart", CL_Snd_Restart_f);

	T.common.Cmd_AddCommand("changing", cl_Changing_f, T)
	// 	Cmd_AddCommand("disconnect", CL_Disconnect_f);
	// 	Cmd_AddCommand("record", CL_Record_f);
	// 	Cmd_AddCommand("stop", CL_Stop_f);

	// 	Cmd_AddCommand("quit", CL_Quit_f);

	// 	Cmd_AddCommand("connect", CL_Connect_f);
	T.common.Cmd_AddCommand("reconnect", cl_Reconnect_f, T)

	// 	Cmd_AddCommand("rcon", CL_Rcon_f);

	// 	Cmd_AddCommand("setenv", CL_Setenv_f);

	T.common.Cmd_AddCommand("precache", cl_Precache_f, T)

	// 	Cmd_AddCommand("download", CL_Download_f);

	// 	Cmd_AddCommand("currentmap", CL_CurrentMap_f);

	/* forward to server commands
	 * the only thing this does is allow command completion
	 * to work -- all unknown commands are automatically
	 * forwarded to the server */
	T.common.Cmd_AddCommand("wave", nil, nil)
	T.common.Cmd_AddCommand("inven", nil, nil)
	T.common.Cmd_AddCommand("kill", nil, nil)
	T.common.Cmd_AddCommand("use", nil, nil)
	T.common.Cmd_AddCommand("drop", nil, nil)
	T.common.Cmd_AddCommand("say", nil, nil)
	T.common.Cmd_AddCommand("say_team", nil, nil)
	T.common.Cmd_AddCommand("info", nil, nil)
	T.common.Cmd_AddCommand("prog", nil, nil)
	T.common.Cmd_AddCommand("give", nil, nil)
	T.common.Cmd_AddCommand("god", nil, nil)
	T.common.Cmd_AddCommand("notarget", nil, nil)
	T.common.Cmd_AddCommand("noclip", nil, nil)
	T.common.Cmd_AddCommand("invuse", nil, nil)
	T.common.Cmd_AddCommand("invprev", nil, nil)
	T.common.Cmd_AddCommand("invnext", nil, nil)
	T.common.Cmd_AddCommand("invdrop", nil, nil)
	T.common.Cmd_AddCommand("weapnext", nil, nil)
	T.common.Cmd_AddCommand("weapprev", nil, nil)
	T.common.Cmd_AddCommand("listentities", nil, nil)
	T.common.Cmd_AddCommand("teleport", nil, nil)
	T.common.Cmd_AddCommand("cycleweap", nil, nil)
}

func (T *qClient) Frame(packetdelta, renderdelta, timedelta int, packetframe, renderframe bool) error {
	// 	static int lasttimecalled;

	// Dedicated?
	if T.common.IsDedicated() {
		return nil
	}

	// Calculate simulation time.
	T.cls.nframetime = float32(packetdelta) / 1000000.0
	T.cls.rframetime = float32(renderdelta) / 1000000.0
	T.cls.realtime = T.common.Curtime()
	T.cl.time += timedelta / 1000

	// 	// Don't extrapolate too far ahead.
	if T.cls.nframetime > 0.5 {
		T.cls.nframetime = 0.5
	}

	if T.cls.rframetime > 0.5 {
		T.cls.rframetime = 0.5
	}

	// 	// if in the debugger last frame, don't timeout.
	// 	if (timedelta > 5000000) {
	// 		T.cls.netchan.last_received = Sys_Milliseconds();
	// 	}

	// 	// Reset power shield / power screen sound counter.
	// 	num_power_sounds = 0;

	// 	if (!cl_timedemo->value)
	// 	{
	// 		// Don't throttle too much when connecting / loading.
	// 		if ((cls.state == ca_connected) && (packetdelta > 100000)) {
	// 			packetframe = true;
	// 		}
	// 	}

	// Update input stuff.
	if packetframe || renderframe {
		if err := T.readPackets(); err != nil {
			return err
		}
		// 		CL_UpdateWindowedMouse();
		T.input.Update()
		if err := T.common.Cbuf_Execute(); err != nil {
			return err
		}
		// 		CL_FixCvarCheats();

		if T.cls.state > ca_connecting {
			T.refreshCmd()
		} else {
			T.refreshMove()
		}
	}

	// 	if (cls.forcePacket || userinfo_modified) {
	// 		packetframe = true;
	// 		cls.forcePacket = false;
	// 	}

	if packetframe {
		if err := T.sendCmd(); err != nil {
			return err
		}
		if err := T.checkForResend(); err != nil {
			return err
		}

		// 		// Run HTTP downloads during game.
		// #ifdef USE_CURL
		// 		CL_RunHTTPDownloads();
		// #endif
	}

	if renderframe {
		if err := T.vidCheckChanges(); err != nil {
			return err
		}
		T.predictMovement()

		// 		if (!cl.refresh_prepped && (cls.state == ca_active)) {
		// 			CL_PrepRefresh();
		// 		}

		// 		/* update the screen */
		// 		if (host_speeds->value) {
		// 			time_before_ref = Sys_Milliseconds();
		// 		}

		if err := T.scrUpdateScreen(); err != nil {
			return err
		}

		// 		if (host_speeds->value)
		// 		{
		// 			time_after_ref = Sys_Milliseconds();
		// 		}

		// 		/* update audio */
		// 		S_Update(cl.refdef.vieworg, cl.v_forward, cl.v_right, cl.v_up);

		/* advance local effects for next frame */
		T.runDLights()
		T.runLightStyles()
		// 		SCR_RunCinematic();
		// 		SCR_RunConsole();

		/* Update framecounter */
		T.cls.framecount++

		// 		if (log_stats->value) {
		// 			if (cls.state == ca_active) {
		// 				if (!lasttimecalled) {
		// 					lasttimecalled = Sys_Milliseconds();

		// 					if (log_stats_file)
		// 					{
		// 						fprintf(log_stats_file, "0\n");
		// 					}
		// 				} else {
		// 					int now = Sys_Milliseconds();

		// 					if (log_stats_file)
		// 					{
		// 						fprintf(log_stats_file, "%d\n", now - lasttimecalled);
		// 					}

		// 					lasttimecalled = now;
		// 				}
		// 			}
		// 		}
	}
	return nil
}

func (T *qClient) Init() error {

	if T.common.IsDedicated() {
		return nil /* nothing running on the client */
	}

	/* all archived variables will now be loaded */
	T.conInit()

	// 	S_Init();

	T.scrInit()

	if err := T.vidInit(); err != nil {
		return err
	}

	T.input.Init(T)

	// 	V_Init();

	// 	net_message.data = net_message_buffer;

	// 	net_message.maxsize = sizeof(net_message_buffer);

	T.mInit()

	// #ifdef USE_CURL
	// 	CL_InitHTTPDownloads();
	// #endif

	// 	cls.disable_screen = true; /* don't draw yet */

	T.initLocal()

	if err := T.common.Cbuf_Execute(); err != nil {
		return err
	}

	// 	Key_ReadConsoleHistory();
	return nil
}
