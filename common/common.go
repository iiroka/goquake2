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
 * Prototypes witch are shared between the client, the server and the
 * game. This is the main game API, changes here will most likely
 * requiere changes to the game ddl.
 *
 * =======================================================================
 */
package common

import (
	"goquake2/shared"
	"time"
)

type xcommand_t struct {
	function func([]string, interface{}) error
	param    interface{}
}

type qCommon struct {
	client shared.QClient
	server shared.QServer

	// from frame
	developer       *shared.CvarT
	modder          *shared.CvarT
	timescale       *shared.CvarT
	fixedtime       *shared.CvarT
	cl_maxfps       *shared.CvarT
	dedicated       *shared.CvarT
	server_state    int
	packetdelta     int
	renderdelta     int
	clienttimedelta int
	servertimedelta int
	startTime       time.Time
	curtime         int

	busywait    *shared.CvarT
	cl_async    *shared.CvarT
	cl_timedemo *shared.CvarT
	vid_maxfps  *shared.CvarT
	host_speeds *shared.CvarT
	log_stats   *shared.CvarT
	showtrace   *shared.CvarT

	// from netchan
	showpackets *shared.CvarT
	showdrop    *shared.CvarT
	qport       *shared.CvarT

	// from filesystem
	fs_searchPaths []fsSearchPath_t
	filehandles    []qFileHandle

	datadir    string
	fs_gamedir string

	fs_basedir    *shared.CvarT
	fs_cddir      *shared.CvarT
	fs_gamedirvar *shared.CvarT
	fs_debug      *shared.CvarT

	// from cmdparser
	cmd_functions map[string]xcommand_t
	cmd_alias     map[string]string
	cmd_text      string
	cmd_wait      int
	alias_count   int

	// from cvars
	cvarVars         map[string]*shared.CvarT
	UserinfoModified bool

	// from clientserver
	recursive bool
	msg       string

	// from network
	loopback [](chan []byte)
}

func CreateQCommon(client shared.QClient, server shared.QServer) shared.QCommon {
	T := &qCommon{client: client, server: server}
	T.cmd_functions = make(map[string]xcommand_t)
	T.cmd_alias = make(map[string]string)
	T.cvarVars = make(map[string]*shared.CvarT)
	T.packetdelta = 1000000
	T.renderdelta = 1000000
	T.loopback = make([](chan []byte), 2)
	T.loopback[0] = make(chan []byte, 100)
	T.loopback[1] = make(chan []byte, 100)
	T.filehandles = make([]qFileHandle, MAX_HANDLES)
	return T
}
