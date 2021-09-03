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
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307,
 * USA.
 *
 * =======================================================================
 *
 * This file implements the non generic part of the menu system, e.g.
 * the menu shown to the player. Beware! This code is very fragile and
 * should only be touched with great care and exessive testing.
 * Otherwise strange things and hard to track down bugs can occure. In a
 * better world someone would rewrite this file to something more like
 * Quake III Team Arena.
 *
 * =======================================================================
 */
package client

import (
	"fmt"
	"goquake2/shared"
	"log"
	"strconv"
	"strings"
)

/* Number of the frames of the spinning quake logo */
const NUM_CURSOR_FRAMES = 15

const menu_in_sound = "misc/menu1.wav"
const menu_move_sound = "misc/menu2.wav"
const menu_out_sound = "misc/menu3.wav"

type menulayer_t struct {
	draw func(T *qClient)
	key  func(T *qClient, key int) string
	name string
}

type MenuStr struct {
	m_drawfunc func(T *qClient)
	m_keyfunc  func(T *qClient, key int) string

	m_layers      []menulayer_t
	m_main_cursor int
	m_game_cursor int
	cached        bool

	m_popup_string  string
	m_popup_endtime int

	s_multiplayer_menu            menuframework_t
	s_join_network_server_action  menuaction_t
	s_start_network_server_action menuaction_t
	s_player_setup_action         menuaction_t

	s_game_menu          menuframework_t
	s_easy_game_action   menuaction_t
	s_medium_game_action menuaction_t
	s_hard_game_action   menuaction_t
	s_hardp_game_action  menuaction_t
	s_load_game_action   menuaction_t
	s_save_game_action   menuaction_t
	s_credits_action     menuaction_t
	s_mods_action        menuaction_t
	// static menuseparator_s s_blankline;

	s_joinserver_menu                menuframework_t
	s_joinserver_search_action       menuaction_t
	s_joinserver_address_book_action menuaction_t
	s_joinserver_server_actions      [MAX_LOCAL_SERVERS]menuaction_t
	local_server_names               [MAX_LOCAL_SERVERS]string
	local_server_netadr_strings      [MAX_LOCAL_SERVERS]string
	m_num_servers                    int

	s_startserver_menu             menuframework_t
	mapnames                       []string
	nummaps                        int
	s_startserver_start_action     menuaction_t
	s_startserver_dmoptions_action menuaction_t
	s_timelimit_field              menufield_t
	s_fraglimit_field              menufield_t
	// s_capturelimit_field           menufield_t
	s_maxclients_field menufield_t
	s_hostname_field   menufield_t
	//  static menulist_s s_startmap_list;
	s_rules_box menulist_t
	//  static menulist_s s_rules_box;

}

func (T *qClient) mBanner(name string) {
	scale := T.scrGetMenuScale()

	w, _ := T.Draw_GetPicSize(name)
	T.Draw_PicScaled(T.viddef.width/2-int(float32(w)*scale)/2, T.viddef.height/2-int(110*scale), name, scale)
}

func (T *qClient) mForceMenuOff() {
	T.menu.m_drawfunc = nil
	T.menu.m_keyfunc = nil
	T.cls.key_dest = key_game
	T.menu.m_layers = make([]menulayer_t, 0)
	// Key_MarkAllUp()
	T.common.Cvar_Set("paused", "0")
}

func (T *qClient) mPopMenu() {
	// S_StartLocalSound(menu_out_sound);

	if len(T.menu.m_layers) < 1 {
		log.Fatal("M_PopMenu: depth < 1")
	} else if len(T.menu.m_layers) == 1 {
		T.mForceMenuOff()
	} else {
		T.menu.m_layers = T.menu.m_layers[:len(T.menu.m_layers)-1]
		last := T.menu.m_layers[len(T.menu.m_layers)-1]

		T.menu.m_drawfunc = last.draw
		T.menu.m_keyfunc = last.key
	}
}

/*
 * This crappy function maintaines a stack of opened menus.
 * The steps in this horrible mess are:
 *
 * 1. But the game into pause if a menu is opened
 *
 * 2. If the requested menu is already open, close it.
 *
 * 3. If the requested menu is already open but not
 *    on top, close all menus above it and the menu
 *    itself. This is necessary since an instance of
 *    the reqeuested menu is in flight and will be
 *    displayed.
 *
 * 4. Save the previous menu on top (which was in flight)
 *    to the stack and make the requested menu the menu in
 *    flight.
 */
func (T *qClient) mPushMenu(draw func(T *qClient), key func(T *qClient, key int) string, name string) {
	// 	 int i;
	// 	 int alreadyPresent = 0;

	if (T.common.Cvar_VariableInt("maxclients") == 1) &&
		T.common.ServerState() != 0 {
		T.common.Cvar_Set("paused", "1")
	}

	//  #ifdef USE_OPENAL
	// 	 if (cl.cinematic_file && sound_started == SS_OAL) {
	// 		 AL_UnqueueRawSamples();
	// 	 }
	//  #endif

	/* if this menu is already open (and on top),
	close it => toggling behaviour */
	// if (m_drawfunc == draw) && (m_keyfunc == key) {
	// 	M_PopMenu()
	// 	return
	// }

	/* if this menu is already present, drop back to
	that level to avoid stacking menus by hotkeys */
	index := -1
	for i := range T.menu.m_layers {
		if T.menu.m_layers[i].name == name {
			index = i
			break
		}
	}

	/* menu was already opened further down the stack */
	for index >= 0 && len(T.menu.m_layers) > index {
		T.mPopMenu() /* decrements m_menudepth */
	}

	// 	 if (m_menudepth >= MAX_MENU_DEPTH) {
	// 		 Com_Printf("Too many open menus!\n");
	// 		 return;
	// 	 }

	l := menulayer_t{}
	l.draw = draw
	l.key = key
	l.name = name
	T.menu.m_layers = append(T.menu.m_layers, l)

	T.menu.m_drawfunc = draw
	T.menu.m_keyfunc = key

	// 	 m_entersound = true;

	T.cls.key_dest = key_menu
}

func keyGetMenuKey(key int) int {
	switch key {
	case K_KP_UPARROW:
	case K_UPARROW:
	case K_HAT_UP:
		return K_UPARROW

	case K_TAB:
	case K_KP_DOWNARROW:
	case K_DOWNARROW:
	case K_HAT_DOWN:
		return K_DOWNARROW

	case K_KP_LEFTARROW:
	case K_LEFTARROW:
	case K_HAT_LEFT:
	case K_TRIG_LEFT:
		return K_LEFTARROW

	case K_KP_RIGHTARROW:
	case K_RIGHTARROW:
	case K_HAT_RIGHT:
	case K_TRIG_RIGHT:
		return K_RIGHTARROW

	case K_MOUSE1:
	case K_MOUSE2:
	case K_MOUSE3:
	case K_MOUSE4:
	case K_MOUSE5:

	case K_JOY1:
	case K_JOY2:
	case K_JOY3:
	case K_JOY4:
	case K_JOY5:
	case K_JOY6:
	case K_JOY7:
	case K_JOY8:
	case K_JOY9:
	case K_JOY10:
	case K_JOY11:
	case K_JOY12:
	case K_JOY13:
	case K_JOY14:
	case K_JOY15:
	case K_JOY16:
	case K_JOY17:
	case K_JOY18:
	case K_JOY19:
	case K_JOY20:
	case K_JOY21:
	case K_JOY22:
	case K_JOY23:
	case K_JOY24:
	case K_JOY25:
	case K_JOY26:
	case K_JOY27:
	case K_JOY28:
	case K_JOY29:
	case K_JOY30:
	case K_JOY31:

	case K_KP_ENTER:
	case K_ENTER:
		return K_ENTER

	case K_ESCAPE:
	case K_JOY_BACK:
		return K_ESCAPE
	}

	return key
}

func (T *qClient) defaultMenuKey(m *menuframework_t, key int) string {
	// const char *sound = NULL;
	menu_key := keyGetMenuKey(key)

	// if (m) {
	//     menucommon_s *item;

	//     if ((item = Menu_ItemAtCursor(m)) != 0) {
	//         if (item->type == MTYPE_FIELD) {
	//             if (Field_Key((menufield_s *)item, key)) {
	//                 return NULL;
	//             }
	//         }
	//     }
	// }

	switch menu_key {
	case K_ESCAPE:
		T.mPopMenu()
		return menu_out_sound

	case K_UPARROW:
		if m != nil {
			m.cursor--
			m.adjustCursor(-1)
			return menu_move_sound
		}

	case K_DOWNARROW:
		if m != nil {
			m.cursor++
			m.adjustCursor(1)
			return menu_move_sound
		}

	case K_LEFTARROW:
		if m != nil {
			m.SlideItem(-1)
			return menu_move_sound
		}

	case K_RIGHTARROW:
		if m != nil {
			m.SlideItem(1)
			return menu_move_sound
		}

	case K_ENTER:
		if m != nil {
			m.selectItem()
		}
		return menu_move_sound
	}

	return ""
}

/*
 * Draws one solid graphics character cx and cy are in 320*240
 * coordinates, and will be centered on higher res screens.
 */
func (T *qClient) mDrawCharacter(cx, cy, num int) {
	scale := T.scrGetMenuScale()
	T.Draw_CharScaled(cx+(int(T.viddef.width-int(320*scale))>>1), cy+(int(T.viddef.height-int(240*scale))>>1), num, scale)
}

func (T *qClient) mPrint(x, y int, str string) {
	scale := T.scrGetMenuScale()

	cx := x
	cy := y
	for index := 0; index < len(str); index++ {
		if str[index] == '\n' {
			cx = x
			cy += 8
		} else {
			T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale), int(str[index])+128)
			cx += 8
		}
	}
}

/*
 * Draws an animating cursor with the point at
 * x,y. The pic will extend to the left of x,
 * and both above and below y.
 */
func (T *qClient) mDrawCursor(x, y, f int) {
	//  char cursorname[80];
	//  static qboolean cached;
	scale := T.scrGetMenuScale()

	if !T.menu.cached {
		for i := 0; i < NUM_CURSOR_FRAMES; i++ {
			cursorname := fmt.Sprintf("m_cursor%d", i)
			T.Draw_FindPic(cursorname)
		}

		T.menu.cached = true
	}

	cursorname := fmt.Sprintf("m_cursor%d", f)
	T.Draw_PicScaled(int(float32(x)*scale), int(float32(y)*scale), cursorname, scale)
}

func (T *qClient) mDrawTextBox(x, y, width, lines int) {
	// int cx, cy;
	// int n;
	scale := T.scrGetMenuScale()

	/* draw left side */
	cx := x
	cy := y
	T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale), 1)

	for n := 0; n < lines; n++ {
		cy += 8
		T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale), 4)
	}

	T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale+8*scale), 7)

	/* draw middle */
	cx += 8

	for width > 0 {
		cy = y
		T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale), 2)

		for n := 0; n < lines; n++ {
			cy += 8
			T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale), 5)
		}

		T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale+8*scale), 8)
		width -= 1
		cx += 8
	}

	/* draw right side */
	cy = y
	T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale), 3)

	for n := 0; n < lines; n++ {
		cy += 8
		T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale), 6)
	}

	T.mDrawCharacter(int(float32(cx)*scale), int(float32(cy)*scale+8*scale), 9)
}

func (T *qClient) mPopup() {

	if len(T.menu.m_popup_string) == 0 {
		return
	}

	if T.menu.m_popup_endtime != 0 && T.menu.m_popup_endtime < T.cls.realtime {
		T.menu.m_popup_string = ""
		return
	}

	// if (!R_EndWorldRenderpass()) {
	//     return;
	// }

	width := 0
	lines := 0
	n := 0
	for str := range T.menu.m_popup_string {
		if str == '\n' {
			lines++
			n = 0
		} else {
			n++
			if n > width {
				width = n
			}
		}
	}
	if n != 0 {
		lines++
	}

	if width != 0 {
		width += 2

		x := (320 - (width+2)*8) / 2
		y := (240 - (lines+2)*8) / 3

		T.mDrawTextBox(x, y, width, lines)
		T.mPrint(x+16, y+8, T.menu.m_popup_string)
	}
}

/*
 * MAIN MENU
 */

const mainITEMS = 5

func m_Main_Draw(T *qClient) {

	scale := T.scrGetMenuScale()
	names := []string{
		"m_main_game",
		"m_main_multiplayer",
		"m_main_options",
		"m_main_video",
		"m_main_quit",
	}

	widest := 0
	totalheight := 0
	for i := range names {
		w, h := T.Draw_GetPicSize(names[i])
		if w > widest {
			widest = w
		}
		totalheight += (h + 12)
	}

	ystart := (int(float32(T.viddef.height)/(2*scale)) - 110)
	xoffset := (int(float32(T.viddef.width)/scale) - widest + 70) / 2

	for i := range names {
		if i != T.menu.m_main_cursor {
			T.Draw_PicScaled(int(float32(xoffset)*scale), int(float32(ystart+i*40+13)*scale), names[i], scale)
		}
	}

	litname := names[T.menu.m_main_cursor] + "_sel"
	T.Draw_PicScaled(int(float32(xoffset)*scale), int(float32(ystart+T.menu.m_main_cursor*40+13)*scale), litname, scale)

	T.mDrawCursor(xoffset-25, ystart+T.menu.m_main_cursor*40+11,
		int(T.cls.realtime/100)%NUM_CURSOR_FRAMES)

	w, h := T.Draw_GetPicSize("m_main_plaque")
	T.Draw_PicScaled(int(float32(xoffset-30-w)*scale), int(float32(ystart)*scale), "m_main_plaque", scale)

	T.Draw_PicScaled(int(float32(xoffset-30-w)*scale), int(float32(ystart+h+5)*scale), "m_main_logo", scale)
}

func m_Main_Key(T *qClient, key int) string {
	menu_key := keyGetMenuKey(key)

	switch menu_key {
	case K_ESCAPE:
		T.mPopMenu()
		break

	case K_DOWNARROW:
		T.menu.m_main_cursor++
		if T.menu.m_main_cursor >= mainITEMS {
			T.menu.m_main_cursor = 0
		}
		return menu_move_sound

	case K_UPARROW:
		T.menu.m_main_cursor--
		if T.menu.m_main_cursor < 0 {
			T.menu.m_main_cursor = mainITEMS - 1
		}
		return menu_move_sound

	case K_ENTER:
		// m_entersound = true

		switch T.menu.m_main_cursor {
		case 0:
			m_Menu_Game_f([]string{}, T)

		case 1:
			m_Menu_Multiplayer_f([]string{}, T)

		case 2:
			// M_Menu_Options_f()

		case 3:
			// M_Menu_Video_f()

		case 4:
			// M_Menu_Quit_f()
		}
	}

	return ""
}

func m_Menu_Main_f(args []string, a interface{}) error {
	a.(*qClient).mPushMenu(m_Main_Draw, m_Main_Key, "main")
	return nil
}

/*
 * MULTIPLAYER MENU
 */

func multiplayer_MenuDraw(T *qClient) {
	T.mBanner("m_banner_multiplayer")

	T.menu.s_multiplayer_menu.adjustCursor(1)
	T.menu.s_multiplayer_menu.draw()
}

//  static void
//  PlayerSetupFunc(void *unused)
//  {
// 	 M_Menu_PlayerConfig_f();
//  }

func joinNetworkServerFunc(data *menucommon_t) {
	m_Menu_JoinServer_f([]string{}, data.parent.owner)
}

func startNetworkServerFunc(data *menucommon_t) {
	m_Menu_StartServer_f([]string{}, data.parent.owner)
}

func (T *qClient) multiplayer_MenuInit() {
	scale := T.scrGetMenuScale()

	T.menu.s_multiplayer_menu.x = int(float32(T.viddef.width)*0.50) - int(64*scale)
	T.menu.s_multiplayer_menu.owner = T

	T.menu.s_join_network_server_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_join_network_server_action.x = 0
	T.menu.s_join_network_server_action.y = 0
	T.menu.s_join_network_server_action.name = " join network server"
	T.menu.s_join_network_server_action.callback = joinNetworkServerFunc

	T.menu.s_start_network_server_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_start_network_server_action.x = 0
	T.menu.s_start_network_server_action.y = 10
	T.menu.s_start_network_server_action.name = " start network server"
	T.menu.s_start_network_server_action.callback = startNetworkServerFunc

	T.menu.s_player_setup_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_player_setup_action.x = 0
	T.menu.s_player_setup_action.y = 20
	T.menu.s_player_setup_action.name = " player setup"
	// T.menu.s_player_setup_action.callback = PlayerSetupFunc

	T.menu.s_multiplayer_menu.addItem(&T.menu.s_join_network_server_action)
	T.menu.s_multiplayer_menu.addItem(&T.menu.s_start_network_server_action)
	T.menu.s_multiplayer_menu.addItem(&T.menu.s_player_setup_action)

	T.menu.s_multiplayer_menu.setStatusBar("")

	T.menu.s_multiplayer_menu.center()
}

func multiplayer_MenuKey(T *qClient, key int) string {
	return T.defaultMenuKey(&T.menu.s_multiplayer_menu, key)
}

func m_Menu_Multiplayer_f(args []string, a interface{}) error {
	q := a.(*qClient)
	q.multiplayer_MenuInit()
	q.mPushMenu(multiplayer_MenuDraw, multiplayer_MenuKey, "multiplayer")
	return nil
}

/*
 * GAME MENU
 */

func (T *qClient) startGame() {
	if T.cls.state != ca_disconnected && T.cls.state != ca_uninitialized {
		T.disconnect()
	}

	/* disable updates and start the cinematic going */
	T.cl.servercount = -1
	T.mForceMenuOff()
	T.common.Cvar_Set("deathmatch", "0")
	T.common.Cvar_Set("coop", "0")

	T.common.Cbuf_AddText("loading ; killserver ; wait ; newgame\n")
	T.cls.key_dest = key_game
}

func easyGameFunc(data *menucommon_t) {
	T := data.parent.owner
	T.common.Cvar_ForceSet("skill", "0")
	T.startGame()
}

func (T *qClient) game_MenuInit() {
	// Mods_NamesInit();

	T.menu.s_game_menu.x = int(float32(T.viddef.width) * 0.50)
	T.menu.s_game_menu.owner = T

	T.menu.s_easy_game_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_easy_game_action.x = 0
	T.menu.s_easy_game_action.y = 0
	T.menu.s_easy_game_action.name = "easy"
	T.menu.s_easy_game_action.callback = easyGameFunc

	T.menu.s_medium_game_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_medium_game_action.x = 0
	T.menu.s_medium_game_action.y = 10
	T.menu.s_medium_game_action.name = "medium"
	// s_medium_game_action.generic.callback = MediumGameFunc;

	T.menu.s_hard_game_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_hard_game_action.x = 0
	T.menu.s_hard_game_action.y = 20
	T.menu.s_hard_game_action.name = "hard"
	// s_hard_game_action.generic.callback = HardGameFunc;

	T.menu.s_hardp_game_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_hardp_game_action.x = 0
	T.menu.s_hardp_game_action.y = 30
	T.menu.s_hardp_game_action.name = "nightmare"
	// s_hardp_game_action.generic.callback = HardpGameFunc;

	// s_blankline.generic.type = MTYPE_SEPARATOR;

	T.menu.s_load_game_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_load_game_action.x = 0
	T.menu.s_load_game_action.y = 50
	T.menu.s_load_game_action.name = "load game"
	// s_load_game_action.generic.callback = LoadGameFunc;

	T.menu.s_save_game_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_save_game_action.x = 0
	T.menu.s_save_game_action.y = 60
	T.menu.s_save_game_action.name = "save game"
	// s_save_game_action.generic.callback = SaveGameFunc;

	// s_credits_action.generic.type = MTYPE_ACTION;
	// s_credits_action.generic.flags = QMF_LEFT_JUSTIFY;
	// s_credits_action.generic.x = 0;
	// s_credits_action.generic.y = 70;
	// s_credits_action.generic.name = "credits";
	// s_credits_action.generic.callback = CreditsFunc;

	T.menu.s_game_menu.addItem(&T.menu.s_easy_game_action)
	T.menu.s_game_menu.addItem(&T.menu.s_medium_game_action)
	T.menu.s_game_menu.addItem(&T.menu.s_hard_game_action)
	T.menu.s_game_menu.addItem(&T.menu.s_hardp_game_action)
	// Menu_AddItem(&s_game_menu, (void *)&s_blankline);
	T.menu.s_game_menu.addItem(&T.menu.s_load_game_action)
	T.menu.s_game_menu.addItem(&T.menu.s_save_game_action)
	// Menu_AddItem(&s_game_menu, (void *)&s_credits_action);

	// if(nummods > 1)
	// {
	//     s_mods_action.generic.type = MTYPE_ACTION;
	//     s_mods_action.generic.flags = QMF_LEFT_JUSTIFY;
	//     s_mods_action.generic.x = 0;
	//     s_mods_action.generic.y = 90;
	//     s_mods_action.generic.name = "mods";
	//     s_mods_action.generic.callback = ModsFunc;

	//     Menu_AddItem(&s_game_menu, (void *)&s_blankline);
	//     Menu_AddItem(&s_game_menu, (void *)&s_mods_action);
	// }

	T.menu.s_game_menu.center()
}

func game_MenuDraw(T *qClient) {
	T.mBanner("m_banner_game")
	T.menu.s_game_menu.adjustCursor(1)
	T.menu.s_game_menu.draw()
}

func game_MenuKey(T *qClient, key int) string {
	return T.defaultMenuKey(&T.menu.s_game_menu, key)
}

func m_Menu_Game_f(args []string, a interface{}) error {
	T := a.(*qClient)
	T.game_MenuInit()
	T.mPushMenu(game_MenuDraw, game_MenuKey, "game")
	T.menu.m_game_cursor = 1
	return nil
}

/*
 * JOIN SERVER MENU
 */

const MAX_LOCAL_SERVERS = 8

//  int m_num_servers;
const NO_SERVER_STRING = "<no server>"

//  /* network address */
//  static netadr_t local_server_netadr[MAX_LOCAL_SERVERS];

//  /* user readable information */
//  static char local_server_names[MAX_LOCAL_SERVERS][80];
//  static char local_server_netadr_strings[MAX_LOCAL_SERVERS][80];

//  void
//  M_AddToServerList(netadr_t adr, char *info)
//  {
// 	 char *s;
// 	 int i;

// 	 if (m_num_servers == MAX_LOCAL_SERVERS)
// 	 {
// 		 return;
// 	 }

// 	 while (*info == ' ')
// 	 {
// 		 info++;
// 	 }

// 	 s = NET_AdrToString(adr);

// 	 /* ignore if duplicated */
// 	 for (i = 0; i < m_num_servers; i++)
// 	 {
// 		 if (!strcmp(local_server_names[i], info) &&
// 			 !strcmp(local_server_netadr_strings[i], s))
// 		 {
// 			 return;
// 		 }
// 	 }

// 	 local_server_netadr[m_num_servers] = adr;
// 	 Q_strlcpy(local_server_names[m_num_servers], info,
// 			 sizeof(local_server_names[m_num_servers]));
// 	 Q_strlcpy(local_server_netadr_strings[m_num_servers], s,
// 			 sizeof(local_server_netadr_strings[m_num_servers]));
// 	 m_num_servers++;
//  }

//  static void
//  JoinServerFunc(void *self)
//  {
// 	 char buffer[128];
// 	 int index;

// 	 index = (int)((menuaction_s *)self - s_joinserver_server_actions);

// 	 if (Q_stricmp(local_server_names[index], NO_SERVER_STRING) == 0)
// 	 {
// 		 return;
// 	 }

// 	 if (index >= m_num_servers)
// 	 {
// 		 return;
// 	 }

// 	 Com_sprintf(buffer, sizeof(buffer), "connect %s\n",
// 				 NET_AdrToString(local_server_netadr[index]));
// 	 Cbuf_AddText(buffer);
// 	 M_ForceMenuOff();
//  }

//  static void
//  AddressBookFunc(void *self)
//  {
// 	 M_Menu_AddressBook_f();
//  }

func (T *qClient) searchLocalGames() {

	T.menu.m_num_servers = 0

	for i := 0; i < MAX_LOCAL_SERVERS; i++ {
		T.menu.local_server_names[i] = NO_SERVER_STRING
		T.menu.local_server_netadr_strings[i] = ""
	}

	T.menu.m_popup_string = "Searching for local servers. This\n" +
		"could take up to a minute, so\n" +
		"please be patient."
	T.menu.m_popup_endtime = T.cls.realtime + 2000
	T.mPopup()

	/* the text box won't show up unless we do a buffer swap */
	T.R_EndFrame()

	/* send out info packets */
	T.clPingServers()
}

func searchLocalGamesFunc(data *menucommon_t) {
	data.parent.owner.searchLocalGames()
}

func (T *qClient) joinServer_MenuInit() {
	// 	 int i;
	scale := T.scrGetMenuScale()

	T.menu.s_joinserver_menu.x = int(float32(T.viddef.width)*0.50) - int(120*scale)
	T.menu.s_joinserver_menu.owner = T

	// 	 s_joinserver_address_book_action.generic.type = MTYPE_ACTION;
	T.menu.s_joinserver_address_book_action.name = "address book"
	T.menu.s_joinserver_address_book_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_joinserver_address_book_action.x = 0
	T.menu.s_joinserver_address_book_action.y = 0
	// 	 s_joinserver_address_book_action.generic.callback = AddressBookFunc;

	// 	 s_joinserver_search_action.generic.type = MTYPE_ACTION;
	T.menu.s_joinserver_search_action.name = "refresh server list"
	T.menu.s_joinserver_search_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_joinserver_search_action.x = 0
	T.menu.s_joinserver_search_action.y = 10
	T.menu.s_joinserver_search_action.callback = searchLocalGamesFunc
	// T.menu.s_joinserver_search_action.statusbar = "search for servers"

	// 	 s_joinserver_server_title.generic.type = MTYPE_SEPARATOR;
	// 	 s_joinserver_server_title.generic.name = "connect to...";
	// 	 s_joinserver_server_title.generic.y = 30;
	// 	 s_joinserver_server_title.generic.x = 80 * scale;

	for i := 0; i < MAX_LOCAL_SERVERS; i++ {
		// 		 s_joinserver_server_actions[i].generic.type = MTYPE_ACTION;
		T.menu.s_joinserver_server_actions[i].name = T.menu.local_server_names[i]
		T.menu.s_joinserver_server_actions[i].flags = QMF_LEFT_JUSTIFY
		T.menu.s_joinserver_server_actions[i].x = 0
		T.menu.s_joinserver_server_actions[i].y = 40 + i*10
		// 		 s_joinserver_server_actions[i].generic.callback = JoinServerFunc;
		// 		 s_joinserver_server_actions[i].generic.statusbar =
		// 			 local_server_netadr_strings[i];
	}

	T.menu.s_joinserver_menu.addItem(&T.menu.s_joinserver_address_book_action)
	// T.menu.s_joinserver_menu.addItem(&T.menu.s_joinserver_server_title)
	T.menu.s_joinserver_menu.addItem(&T.menu.s_joinserver_search_action)

	for i := 0; i < MAX_LOCAL_SERVERS; i++ {
		T.menu.s_joinserver_menu.addItem(&T.menu.s_joinserver_server_actions[i])
	}

	T.menu.s_joinserver_menu.center()

	// 	 SearchLocalGames();
}

func joinServer_MenuDraw(T *qClient) {
	T.mBanner("m_banner_join_server")
	T.menu.s_joinserver_menu.draw()
	T.mPopup()
}

func joinServer_MenuKey(T *qClient, key int) string {
	if len(T.menu.m_popup_string) > 0 {
		T.menu.m_popup_string = ""
		return ""
	}
	return T.defaultMenuKey(&T.menu.s_joinserver_menu, key)
}

func m_Menu_JoinServer_f(args []string, a interface{}) error {
	T := a.(*qClient)
	T.joinServer_MenuInit()
	T.mPushMenu(joinServer_MenuDraw, joinServer_MenuKey, "joinserver")
	return nil
}

/*
 * START SERVER MENU
 */

//  static void
//  DMOptionsFunc(void *self)
//  {
// 	 M_Menu_DMOptions_f();
//  }

func rulesChangeFunc(data *menucommon_t) {
	Q := data.parent.owner
	/* Deathmatch */
	if Q.menu.s_rules_box.curvalue == 0 {
		Q.menu.s_maxclients_field.statusbar = ""
		Q.menu.s_startserver_dmoptions_action.statusbar = ""
	}

	/* Ground Zero game modes */
	//  else if (M_IsGame("rogue"))
	//  {
	// 	 if (s_rules_box.curvalue == 2)
	// 	 {
	// 		 s_maxclients_field.generic.statusbar = NULL;
	// 		 s_startserver_dmoptions_action.generic.statusbar = NULL;
	// 	 }
	//  }
}

func startServerActionFunc(data *menucommon_t) {
	Q := data.parent.owner

	// 	 char startmap[1024];
	// 	 float timelimit;
	// 	 float fraglimit;
	// 	 float capturelimit;
	// 	 float maxclients;
	// 	 char *spot;

	startmap := strings.Split(Q.menu.mapnames[0], "\n")[1]
	// 	 strcpy(startmap, strchr(mapnames[s_startmap_list.curvalue], '\n') + 1);

	maxclients, _ := strconv.ParseInt(Q.menu.s_maxclients_field.buffer, 10, 32)
	timelimit, _ := strconv.ParseInt(Q.menu.s_timelimit_field.buffer, 10, 32)
	fraglimit, _ := strconv.ParseInt(Q.menu.s_fraglimit_field.buffer, 10, 32)

	// 	 if (M_IsGame("ctf"))
	// 	 {
	// 		 capturelimit = (float)strtod(s_capturelimit_field.buffer, (char **)NULL);
	// 		 Cvar_SetValue("capturelimit", ClampCvar(0, capturelimit, capturelimit));
	// 	 }

	Q.common.Cvar_Set("maxclients", strconv.Itoa(clampCvarInt(0, int(maxclients), int(maxclients))))
	Q.common.Cvar_Set("timelimit", strconv.Itoa(clampCvarInt(0, int(timelimit), int(timelimit))))
	Q.common.Cvar_Set("fraglimit", strconv.Itoa(clampCvarInt(0, int(fraglimit), int(fraglimit))))
	// 	 Cvar_Set("hostname", s_hostname_field.buffer);

	// 	 if ((s_rules_box.curvalue < 2) || M_IsGame("rogue"))
	// 	 {
	// 		 Cvar_SetValue("deathmatch", (float)!s_rules_box.curvalue);
	// 		 Cvar_SetValue("coop", (float)s_rules_box.curvalue);
	// 	 }
	// 	 else
	// 	 {
	Q.common.Cvar_Set("singleplayer", "0")
	Q.common.Cvar_Set("deathmatch", "1") /* deathmatch is always true for rogue games */
	Q.common.Cvar_Set("coop", "0")       /* This works for at least the main game and both addons */
	// 	 }

	// 	 spot = NULL;

	if Q.menu.s_rules_box.curvalue == 1 {
		// 		 if (Q_stricmp(startmap, "bunk1") == 0)
		// 		 {
		// 			 spot = "start";
		// 		 }

		// 		 else if (Q_stricmp(startmap, "mintro") == 0)
		// 		 {
		// 			 spot = "start";
		// 		 }

		// 		 else if (Q_stricmp(startmap, "fact1") == 0)
		// 		 {
		// 			 spot = "start";
		// 		 }

		// 		 else if (Q_stricmp(startmap, "power1") == 0)
		// 		 {
		// 			 spot = "pstart";
		// 		 }

		// 		 else if (Q_stricmp(startmap, "biggun") == 0)
		// 		 {
		// 			 spot = "bstart";
		// 		 }

		// 		 else if (Q_stricmp(startmap, "hangar1") == 0)
		// 		 {
		// 			 spot = "unitstart";
		// 		 }

		// 		 else if (Q_stricmp(startmap, "city1") == 0)
		// 		 {
		// 			 spot = "unitstart";
		// 		 }

		// 		 else if (Q_stricmp(startmap, "boss1") == 0)
		// 		 {
		// 			 spot = "bosstart";
		// 		 }
	}

	// 	 if (spot)
	// 	 {
	// 		 if (Com_ServerState())
	// 		 {
	// 			 Cbuf_AddText("disconnect\n");
	// 		 }

	// 		 Cbuf_AddText(va("gamemap \"*%s$%s\"\n", startmap, spot));
	// 	 }
	// 	 else
	// 	 {
	Q.common.Cbuf_AddText(fmt.Sprintf("map %s\n", startmap))
	// 	 }

	Q.mForceMenuOff()
}

func (T *qClient) startServer_MenuInit() error {
	dm_coop_names := []string{
		"deathmatch",
		"cooperative",
	}
	// 	 static const char *dm_coop_names_rogue[] =
	// 	 {
	// 		 "deathmatch",
	// 		 "cooperative",
	// 		 "tag",
	// 		 0
	// 	 };

	// 	 char *buffer;
	// 	 char *s;
	scale := T.scrGetMenuScale()

	/* initialize list of maps once, reuse it afterwards (=> it isn't freed unless the game dir is changed) */
	if T.menu.mapnames == nil || len(T.menu.mapnames) == 0 {
		// 		 int i, length;
		// 		 size_t nummapslen;

		T.menu.mapnames = []string{}
		T.menu.nummaps = 0
		// T.menu.s_startmap_list.curvalue = 0

		/* load the list of map names */
		buffer, err := T.common.LoadFile("maps.lst")
		if err != nil || buffer == nil {
			return T.common.Com_Error(shared.ERR_DROP, "couldn't find maps.lst\n")
		}

		s := string(buffer)
		i := 0

		for i < len(s) {
			if s[i] == '\n' {
				T.menu.nummaps++
			}
			i++
		}

		if T.menu.nummaps == 0 {
			return T.common.Com_Error(shared.ERR_DROP, "no maps in maps.lst\n")
		}

		T.menu.mapnames = make([]string, T.menu.nummaps)
		index := 0

		for i := 0; i < T.menu.nummaps; i++ {
			// 			 char shortname[MAX_TOKEN_CHARS];
			// 			 char longname[MAX_TOKEN_CHARS];
			// 			 char scratch[200];
			// 			 int j, l;

			shortname, indx2 := shared.COM_Parse(s, index)
			// 			 strcpy(shortname, COM_Parse(&s));
			// 			 l = strlen(shortname);

			// 			 for (j = 0; j < l; j++)
			// 			 {
			// 				 shortname[j] = toupper((unsigned char)shortname[j]);
			// 			 }

			longname, indx3 := shared.COM_Parse(s, indx2)

			T.menu.mapnames[i] = fmt.Sprintf("%s\n%s", longname, shortname)
			println(T.menu.mapnames[i])
			index = indx3
		}
	}

	/* initialize the menu stuff */
	T.menu.s_startserver_menu.x = int(float32(T.viddef.width) * 0.50)
	T.menu.s_startserver_menu.owner = T

	// 	 s_startmap_list.generic.type = MTYPE_SPINCONTROL;
	// 	 s_startmap_list.generic.x = 0;

	// 	 if (M_IsGame("ctf"))
	// 		 s_startmap_list.generic.y = -8;
	// 	 else
	// 		 s_startmap_list.generic.y = 0;

	// 	 s_startmap_list.generic.name = "initial map";
	// 	 s_startmap_list.itemnames = (const char **)mapnames;

	// 	 if (M_IsGame("ctf"))
	// 	 {
	// 		 s_capturelimit_field.generic.type = MTYPE_FIELD;
	// 		 s_capturelimit_field.generic.name = "capture limit";
	// 		 s_capturelimit_field.generic.flags = QMF_NUMBERSONLY;
	// 		 s_capturelimit_field.generic.x = 0;
	// 		 s_capturelimit_field.generic.y = 18;
	// 		 s_capturelimit_field.generic.statusbar = "0 = no limit";
	// 		 s_capturelimit_field.length = 3;
	// 		 s_capturelimit_field.visible_length = 3;
	// 		 strcpy(s_capturelimit_field.buffer, Cvar_VariableString("capturelimit"));
	// 	 }
	// 	 else
	// 	 {
	//  s_rules_box.generic.type = MTYPE_SPINCONTROL;
	T.menu.s_rules_box.x = 0
	T.menu.s_rules_box.y = 20
	T.menu.s_rules_box.name = "rules"

	// 		 /* Ground Zero games only available with rogue game */
	// 		 if (M_IsGame("rogue"))
	// 		 {
	// 			 s_rules_box.itemnames = dm_coop_names_rogue;
	// 		 }
	// 		 else
	// 		 {
	T.menu.s_rules_box.itemnames = dm_coop_names
	// 		 }

	if T.common.Cvar_VariableBool("coop") {
		T.menu.s_rules_box.curvalue = 1
	} else {
		T.menu.s_rules_box.curvalue = 0
	}

	T.menu.s_rules_box.callback = rulesChangeFunc
	// 	 }

	// 	 s_timelimit_field.generic.type = MTYPE_FIELD;
	T.menu.s_timelimit_field.name = "time limit"
	T.menu.s_timelimit_field.flags = QMF_NUMBERSONLY
	T.menu.s_timelimit_field.x = 0
	T.menu.s_timelimit_field.y = 36
	T.menu.s_timelimit_field.statusbar = "0 = no limit"
	T.menu.s_timelimit_field.length = 3
	T.menu.s_timelimit_field.visible_length = 3
	T.menu.s_timelimit_field.buffer = T.common.Cvar_VariableString("timelimit")

	// 	 s_fraglimit_field.generic.type = MTYPE_FIELD;
	T.menu.s_fraglimit_field.name = "frag limit"
	T.menu.s_fraglimit_field.flags = QMF_NUMBERSONLY
	T.menu.s_fraglimit_field.x = 0
	T.menu.s_fraglimit_field.y = 54
	T.menu.s_fraglimit_field.statusbar = "0 = no limit"
	T.menu.s_fraglimit_field.length = 3
	T.menu.s_fraglimit_field.visible_length = 3
	T.menu.s_fraglimit_field.buffer = T.common.Cvar_VariableString("fraglimit")

	/* maxclients determines the maximum number of players that can join
	the game. If maxclients is only "1" then we should default the menu
	option to 8 players, otherwise use whatever its current value is.
	Clamping will be done when the server is actually started. */
	T.menu.s_maxclients_field.name = "max players"
	T.menu.s_maxclients_field.flags = QMF_NUMBERSONLY
	T.menu.s_maxclients_field.x = 0
	T.menu.s_maxclients_field.y = 72
	T.menu.s_maxclients_field.statusbar = ""
	T.menu.s_maxclients_field.length = 3
	T.menu.s_maxclients_field.visible_length = 3

	if T.common.Cvar_VariableInt("maxclients") == 1 {
		T.menu.s_maxclients_field.buffer = "8"
	} else {
		T.menu.s_maxclients_field.buffer = T.common.Cvar_VariableString("maxclients")
	}

	// 	 s_hostname_field.generic.type = MTYPE_FIELD;
	// 	 s_hostname_field.generic.name = "hostname";
	// 	 s_hostname_field.generic.flags = 0;
	// 	 s_hostname_field.generic.x = 0;
	// 	 s_hostname_field.generic.y = 90;
	// 	 s_hostname_field.generic.statusbar = NULL;
	// 	 s_hostname_field.length = 12;
	// 	 s_hostname_field.visible_length = 12;
	// 	 strcpy(s_hostname_field.buffer, Cvar_VariableString("hostname"));
	// 	 s_hostname_field.cursor = strlen(s_hostname_field.buffer);

	// 	 s_startserver_dmoptions_action.generic.type = MTYPE_ACTION;
	T.menu.s_startserver_dmoptions_action.name = " deathmatch flags"
	T.menu.s_startserver_dmoptions_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_startserver_dmoptions_action.x = int(24 * scale)
	T.menu.s_startserver_dmoptions_action.y = 108
	//  s_startserver_dmoptions_action.generic.statusbar = NULL;
	// 	 s_startserver_dmoptions_action.generic.callback = DMOptionsFunc;

	// 	 s_startserver_start_action.generic.type = MTYPE_ACTION;
	T.menu.s_startserver_start_action.name = " begin"
	T.menu.s_startserver_start_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_startserver_start_action.x = int(24 * scale)
	T.menu.s_startserver_start_action.y = 128
	T.menu.s_startserver_start_action.callback = startServerActionFunc

	// 	 Menu_AddItem(&s_startserver_menu, &s_startmap_list);

	// 	 if (M_IsGame("ctf"))
	// 		 Menu_AddItem(&s_startserver_menu, &s_capturelimit_field);
	// 	 else
	// 		 Menu_AddItem(&s_startserver_menu, &s_rules_box);
	T.menu.s_startserver_menu.addItem(&T.menu.s_rules_box)

	T.menu.s_startserver_menu.addItem(&T.menu.s_timelimit_field)
	T.menu.s_startserver_menu.addItem(&T.menu.s_fraglimit_field)
	T.menu.s_startserver_menu.addItem(&T.menu.s_maxclients_field)
	// 	 Menu_AddItem(&s_startserver_menu, &s_hostname_field);
	T.menu.s_startserver_menu.addItem(&T.menu.s_startserver_dmoptions_action)
	T.menu.s_startserver_menu.addItem(&T.menu.s_startserver_start_action)

	T.menu.s_startserver_menu.center()

	/* call this now to set proper inital state */
	rulesChangeFunc(&T.menu.s_rules_box.menucommon_t)
	return nil
}

func startServer_MenuDraw(T *qClient) {
	T.menu.s_startserver_menu.draw()
}

func startServer_MenuKey(T *qClient, key int) string {
	return T.defaultMenuKey(&T.menu.s_startserver_menu, key)
}

func m_Menu_StartServer_f(args []string, a interface{}) error {
	T := a.(*qClient)
	if err := T.startServer_MenuInit(); err != nil {
		return err
	}
	T.mPushMenu(startServer_MenuDraw, startServer_MenuKey, "startserver")
	return nil
}

func (T *qClient) mInit() {
	T.common.Cmd_AddCommand("menu_main", m_Menu_Main_f, T)
	T.common.Cmd_AddCommand("menu_game", m_Menu_Game_f, T)
	// Cmd_AddCommand("menu_loadgame", M_Menu_LoadGame_f);
	// Cmd_AddCommand("menu_savegame", M_Menu_SaveGame_f);
	T.common.Cmd_AddCommand("menu_joinserver", m_Menu_JoinServer_f, T)
	// Cmd_AddCommand("menu_addressbook", M_Menu_AddressBook_f);
	T.common.Cmd_AddCommand("menu_startserver", m_Menu_StartServer_f, T)
	// Cmd_AddCommand("menu_dmoptions", M_Menu_DMOptions_f);
	// Cmd_AddCommand("menu_playerconfig", M_Menu_PlayerConfig_f);
	// Cmd_AddCommand("menu_downloadoptions", M_Menu_DownloadOptions_f);
	// Cmd_AddCommand("menu_credits", M_Menu_Credits_f);
	// Cmd_AddCommand("menu_mods", M_Menu_Mods_f);
	T.common.Cmd_AddCommand("menu_multiplayer", m_Menu_Multiplayer_f, T)
	// Cmd_AddCommand("menu_video", M_Menu_Video_f);
	// Cmd_AddCommand("menu_options", M_Menu_Options_f);
	// Cmd_AddCommand("menu_keys", M_Menu_Keys_f);
	// Cmd_AddCommand("menu_joy", M_Menu_Joy_f);
	// Cmd_AddCommand("menu_quit", M_Menu_Quit_f);

	// /* initialize the server address book cvars (adr0, adr1, ...)
	//  * so the entries are not lost if you don't open the address book */
	// for (int index = 0; index < NUM_ADDRESSBOOK_ENTRIES; index++) {
	//     char buffer[20];
	//     Com_sprintf(buffer, sizeof(buffer), "adr%d", index);
	//     Cvar_Get(buffer, "", CVAR_ARCHIVE);
	// }
}

func (T *qClient) mDraw() {
	if T.cls.key_dest != key_menu {
		return
	}

	/* repaint everything next frame */
	T.scrDirtyScreen()

	/* dim everything behind it down */
	// if T.cl.cinematictime > 0 {
	// 	T.Draw_Fill(0, 0, viddef.width, viddef.height, 0)
	// } else {
	// 	Draw_FadeScreen()
	// }

	T.menu.m_drawfunc(T)

	/* delay playing the enter sound until after the
	   menu has been drawn, to avoid delay while
	   caching images */
	// if m_entersound {
	// 	S_StartLocalSound(menu_in_sound)
	// 	m_entersound = false
	// }
}

func (T *qClient) mKeydown(key int) {
	if T.menu.m_keyfunc != nil {
		T.menu.m_keyfunc(T, key)
		//     const char *s;
		//     if ((s = m_keyfunc(key)) != 0) {
		//         S_StartLocalSound((char *)s);
		//     }
	}
}

func clampCvarInt(min, max, value int) int {
	if value < min {
		return min
	}

	if value > max {
		return max
	}

	return value
}
