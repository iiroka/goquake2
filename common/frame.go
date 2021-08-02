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
 * Platform independent initialization, main loop and frame handling.
 *
 * =======================================================================
 */
package common

import (
	"goquake2/shared"
	"math"
	"time"
)

func (T *qCommon) qcommon_Mainloop() error {
	// 	long long newtime;
	oldtime := time.Now()

	/* The mainloop. The legend. */
	for T.running {
		// #ifndef DEDICATED_ONLY
		// 		// Throttle the game a little bit.
		// 		if (busywait->value)
		// 		{
		// 			long long spintime = Sys_Microseconds();

		// 			while (1)
		// 			{
		// 				/* Give the CPU a hint that this is a very tight
		// 				   spinloop. One PAUSE instruction each loop is
		// 				   enough to reduce power consumption and head
		// 				   dispersion a lot, it's 95°C against 67°C on
		// 				   a Kaby Lake laptop. */
		// #if defined (__GNUC__) && (__i386 || __x86_64__)
		// 				asm("pause");
		// #elif defined(__aarch64__) || (defined(__ARM_ARCH) && __ARM_ARCH >= 7) || defined(__ARM_ARCH_6K__)
		// 				asm("yield");
		// #endif

		// 				if (Sys_Microseconds() - spintime >= 5)
		// 				{
		// 					break;
		// 				}
		// 			}
		// 		}
		// 		else
		// 		{
		time.Sleep(5 * time.Microsecond)
		// 			Sys_Nanosleep(5000);
		// 		}
		// #else
		// 		Sys_Nanosleep(850000);
		// #endif

		newtime := time.Now()
		T.qcommon_Frame(int(newtime.Sub(oldtime).Microseconds()))
		oldtime = newtime
	}
	return nil
}

func (T *qCommon) qcommon_ExecConfigs(gameStartUp bool) error {
	T.Cbuf_AddText("exec default.cfg\n")
	T.Cbuf_AddText("exec yq2.cfg\n")
	T.Cbuf_AddText("exec config.cfg\n")
	T.Cbuf_AddText("exec autoexec.cfg\n")

	// if (gameStartUp)
	// {
	// 	/* Process cmd arguments only startup. */
	// 	Cbuf_AddEarlyCommands(true);
	// }

	return T.Cbuf_Execute()
}

func (T *qCommon) IsDedicated() bool {
	return T.dedicated.Bool()
}

func (T *qCommon) Init() error {
	T.startTime = time.Now()
	T.running = true
	// Jump point used in emergency situations.
	// 	if (setjmp(abortframe))
	// 	{
	// 		Sys_Error("Error during initialization");
	// 	}

	// 	if (checkForHelp(argc, argv))
	// 	{
	// 		// ok, --help or similar commandline option was given
	// 		// and info was printed, exit the game now
	// 		exit(1);
	// 	}

	// Print the build and version string
	// 	Qcommon_Buildstring();

	// Seed PRNG
	// 	randk_seed();

	// Initialize zone malloc().
	// 	z_chain.next = z_chain.prev = &z_chain;

	// Start early subsystems.
	// 	COM_InitArgv(argc, argv);
	// 	Swap_Init();
	// 	Cbuf_Init();
	T.cmdInit()
	T.cvarInit()

	T.client.KeyInit()

	/* we need to add the early commands twice, because
	   a basedir or cddir needs to be set before execing
	   config files, but we want other parms to override
	   the settings of the config files */
	// 	Cbuf_AddEarlyCommands(false);
	if err := T.Cbuf_Execute(); err != nil {
		return err
	}

	// 	// remember the initial game name that might have been set on commandline
	// 	{
	// 		cvar_t* gameCvar = Cvar_Get("game", "", CVAR_LATCH | CVAR_SERVERINFO);
	// 		const char* game = "";

	// 		if(gameCvar->string && gameCvar->string[0])
	// 		{
	// 			game = gameCvar->string;
	// 		}

	// 		Q_strlcpy(userGivenGame, game, sizeof(userGivenGame));
	// 	}

	// The filesystems needs to be initialized after the cvars.
	if err := T.initFilesystem(); err != nil {
		return err
	}

	// Add and execute configuration files.
	if err := T.qcommon_ExecConfigs(true); err != nil {
		return err
	}

	// 	// Zone malloc statistics.
	// 	Cmd_AddCommand("z_stats", Z_Stats_f);

	// cvars

	T.cl_maxfps = T.Cvar_Get("cl_maxfps", "60", shared.CVAR_ARCHIVE)

	T.developer = T.Cvar_Get("developer", "0", 0)
	T.fixedtime = T.Cvar_Get("fixedtime", "0", 0)

	// 	logfile_active = Cvar_Get("logfile", "1", CVAR_ARCHIVE);
	T.modder = T.Cvar_Get("modder", "0", 0)
	T.timescale = T.Cvar_Get("timescale", "1", 0)

	// 	char *s;
	// 	s = va("%s %s %s %s", YQ2VERSION, YQ2ARCH, BUILD_DATE, YQ2OSTYPE);
	// 	Cvar_Get("version", s, CVAR_SERVERINFO | CVAR_NOSET);

	T.busywait = T.Cvar_Get("busywait", "1", shared.CVAR_ARCHIVE)
	T.cl_async = T.Cvar_Get("cl_async", "1", shared.CVAR_ARCHIVE)
	T.cl_timedemo = T.Cvar_Get("timedemo", "0", 0)
	T.dedicated = T.Cvar_Get("dedicated", "0", shared.CVAR_NOSET)
	T.vid_maxfps = T.Cvar_Get("vid_maxfps", "300", shared.CVAR_ARCHIVE)
	T.host_speeds = T.Cvar_Get("host_speeds", "0", 0)
	T.log_stats = T.Cvar_Get("log_stats", "0", 0)
	T.showtrace = T.Cvar_Get("showtrace", "0", 0)

	// 	// We can't use the clients "quit" command when running dedicated.
	// 	if (dedicated->value)
	// 	{
	// 		Cmd_AddCommand("quit", Com_Quit);
	// 	}

	// Start late subsystem.
	// 	Sys_Init();
	// 	NET_Init();
	T.netchanInit()
	if err := T.server.Init(T); err != nil {
		return err
	}
	if err := T.client.Init(); err != nil {
		return err
	}

	// 	// Everythings up, let's add + cmds from command line.
	// 	if (!Cbuf_AddLateCommands())
	// 	{
	if !T.dedicated.Bool() {
		// Start demo loop...
		T.Cbuf_AddText("d1\n")
	} else {
		// ...or dedicated server.
		T.Cbuf_AddText("dedicated_start\n")
	}

	if err := T.Cbuf_Execute(); err != nil {
		return err
	}
	// 	}
	// 	else
	// 	{
	// 		/* the user asked for something explicit
	// 		   so drop the loading plaque */
	// 		SCR_EndLoadingPlaque();
	// 	}

	T.Com_Printf("==== Yamagi Quake II Initialized ====\n\n")
	T.Com_Printf("*************************************\n\n")

	// Call the main loop
	return T.qcommon_Mainloop()
}

func (T *qCommon) SetServerState(state int) {
	T.server_state = state
}

func (T *qCommon) ServerState() int {
	return T.server_state
}

func (T *qCommon) Sys_Milliseconds() int {
	return int(time.Now().Sub(T.startTime).Milliseconds())
}

func (T *qCommon) Curtime() int {
	return T.curtime
}

func (T *qCommon) qcommon_Frame(usec int) error {
	// // Used for the dedicated server console.
	// char *s;

	// // Statistics.
	// int time_before = 0;
	// int time_between = 0;
	// int time_after;

	// // Target packetframerate.
	// int pfps;

	// //Target renderframerate.
	// int rfps;

	/* A packetframe runs the server and the client,
	   but not the renderer. The minimal interval of
	   packetframes is about 10.000 microsec. If run
	   more often the movement prediction in pmove.c
	   breaks. That's the Q2 variant if the famous
	   125hz bug. */
	packetframe := true

	/* A rendererframe runs the renderer, but not the
	   client or the server. The minimal interval is
	   about 1000 microseconds. */
	renderframe := true

	// /* Tells the client to shutdown.
	//    Used by the signal handlers. */
	// if (quitnextframe)
	// {
	// 	Cbuf_AddText("quit");
	// }

	// /* In case of ERR_DROP we're jumping here. Don't know
	//    if that's really save but it seems to work. So leave
	//    it alone. */
	// if (setjmp(abortframe))
	// {
	// 	return;
	// }

	// if (log_stats->modified)
	// {
	// 	log_stats->modified = false;

	// 	if (log_stats->value)
	// 	{
	// 		if (log_stats_file)
	// 		{
	// 			fclose(log_stats_file);
	// 			log_stats_file = 0;
	// 		}

	// 		log_stats_file = Q_fopen("stats.log", "w");

	// 		if (log_stats_file)
	// 		{
	// 			fprintf(log_stats_file, "entities,dlights,parts,frame time\n");
	// 		}
	// 	}
	// 	else
	// 	{
	// 		if (log_stats_file)
	// 		{
	// 			fclose(log_stats_file);
	// 			log_stats_file = 0;
	// 		}
	// 	}
	// }

	// // Timing debug crap. Just for historical reasons.
	// if (fixedtime->value)
	// {
	// 	usec = (int)fixedtime->value;
	// }
	// else if (timescale->value)
	// {
	// 	usec *= timescale->value;
	// }

	// if (showtrace->value)
	// {
	// 	extern int c_traces, c_brush_traces;
	// 	extern int c_pointcontents;

	// 	Com_Printf("%4i traces  %4i points\n", c_traces, c_pointcontents);
	// 	c_traces = 0;
	// 	c_brush_traces = 0;
	// 	c_pointcontents = 0;
	// }

	// /* We can render 1000 frames at maximum, because the minimum
	//    frametime of the client is 1 millisecond. And of course we
	//    need to render something, the framerate can never be less
	//    then 1. Cap vid_maxfps between 1 and 999. */
	// if (vid_maxfps->value > 999 || vid_maxfps->value < 1)
	// {
	// 	Cvar_SetValue("vid_maxfps", 999);
	// }

	// if (cl_maxfps->value > 250)
	// {
	// 	Cvar_SetValue("cl_maxfps", 250);
	// }
	// else if (cl_maxfps->value < 1)
	// {
	// 	Cvar_SetValue("cl_maxfps", 60);
	// }

	// Save global time for network- und input code.
	T.curtime = T.Sys_Milliseconds()

	// Calculate target and renderframerate.
	var rfps int
	if T.client.IsVSyncActive() {
		rfps = T.client.GetRefreshRate()

		if rfps > T.vid_maxfps.Int() {
			rfps = T.vid_maxfps.Int()
		}
	} else {
		rfps = T.vid_maxfps.Int()
	}

	// /* The target render frame rate may be too high. The current
	//    scene may be more complex then the previous one and SDL
	//    may give us a 1 or 2 frames too low display refresh rate.
	//    Add a security magin of 5%, e.g. 60fps * 0.95 = 57fps. */
	pfps := T.cl_maxfps.Int()
	if pfps > int(float32(rfps)*0.95) {
		pfps = int(math.Floor(float64(rfps) * 0.95))
	}

	// Calculate timings.
	T.packetdelta += usec
	T.renderdelta += usec
	T.clienttimedelta += usec
	T.servertimedelta += usec

	if !T.cl_timedemo.Bool() {
		if T.cl_async.Bool() {
			if T.client.IsVSyncActive() {
				// Netwwork frames.
				if T.packetdelta < int(0.8*(1000000.0/float32(pfps))) {
					packetframe = false
				}

				// Render frames.
				if T.renderdelta < int(0.8*(1000000.0/float32(rfps))) {
					renderframe = false
				}
			} else {
				// Network frames.
				if T.packetdelta < int(1000000.0/float32(pfps)) {
					packetframe = false
				}

				// Render frames.
				if T.renderdelta < int(1000000.0/float32(rfps)) {
					renderframe = false
				}
			}
		} else {
			// Cap frames at target framerate.
			if T.renderdelta < int(1000000.0/float32(rfps)) {
				renderframe = false
				packetframe = false
			}
		}
	} else if T.clienttimedelta < 1000 || T.servertimedelta < 1000 {
		return nil
	}

	// // Dedicated server terminal console.
	// do {
	// 	s = Sys_ConsoleInput();

	// 	if (s) {
	// 		Cbuf_AddText(va("%s\n", s));
	// 	}
	// } while (s);

	if err := T.Cbuf_Execute(); err != nil {
		return err
	}

	// if (host_speeds->value)
	// {
	// 	time_before = Sys_Milliseconds();
	// }

	// Run the serverframe.
	if packetframe {
		if err := T.server.Frame(T.servertimedelta); err != nil {
			return err
		}
		T.servertimedelta = 0
	}

	// if (host_speeds->value)
	// {
	// 	time_between = Sys_Milliseconds();
	// }

	// // Run the client frame.
	if packetframe || renderframe {
		if err := T.client.Frame(T.packetdelta, T.renderdelta, T.clienttimedelta, packetframe, renderframe); err != nil {
			return err
		}
		T.clienttimedelta = 0
	}

	// if (host_speeds->value)
	// {
	// 	int all, sv, gm, cl, rf;

	// 	time_after = Sys_Milliseconds();
	// 	all = time_after - time_before;
	// 	sv = time_between - time_before;
	// 	cl = time_after - time_between;
	// 	gm = time_after_game - time_before_game;
	// 	rf = time_after_ref - time_before_ref;
	// 	sv -= gm;
	// 	cl -= rf;
	// 	Com_Printf("all:%3i sv:%3i gm:%3i cl:%3i rf:%3i\n", all, sv, gm, cl, rf);
	// }

	// // Reset deltas and mark frame.
	// if (packetframe) {
	// 	packetdelta = 0;
	// }

	// if (renderframe) {
	// 	renderdelta = 0;
	// }
	return nil
}
