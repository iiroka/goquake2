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
 * Server commands received by clients. There are only two ways on which
 * those can be received. Typed via stdin into the server console or via
 * a network / internal communication datagram.
 *
 * =======================================================================
 */
package server

/*
 * Puts the server in demo mode on a specific map/cinematic
 */
func sv_DemoMap_f(args []string, arg interface{}) error {

	T := arg.(*qServer)
	if len(args) != 2 {
		T.common.Com_Printf("USAGE: demomap <demoname.dm2>\n")
		return nil
	}

	return T.svMap(true, args[1], false)
}

func (T *qServer) initOperatorCommands() {
	// Cmd_AddCommand("heartbeat", SV_Heartbeat_f);
	// Cmd_AddCommand("kick", SV_Kick_f);
	// Cmd_AddCommand("status", SV_Status_f);
	// Cmd_AddCommand("serverinfo", SV_Serverinfo_f);
	// Cmd_AddCommand("dumpuser", SV_DumpUser_f);

	// Cmd_AddCommand("map", SV_Map_f);
	// Cmd_AddCommand("listmaps", SV_ListMaps_f);
	T.common.Cmd_AddCommand("demomap", sv_DemoMap_f, T)
	// Cmd_AddCommand("gamemap", SV_GameMap_f);
	// Cmd_AddCommand("setmaster", SV_SetMaster_f);

	// if (dedicated->value)
	// {
	// 	Cmd_AddCommand("say", SV_ConSay_f);
	// }

	// Cmd_AddCommand("serverrecord", SV_ServerRecord_f);
	// Cmd_AddCommand("serverstop", SV_ServerStop_f);

	// Cmd_AddCommand("save", SV_Savegame_f);
	// Cmd_AddCommand("load", SV_Loadgame_f);

	// Cmd_AddCommand("killserver", SV_KillServer_f);

	// Cmd_AddCommand("sv", SV_ServerCommand_f);
}
