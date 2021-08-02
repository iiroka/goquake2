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
	"log"
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
}

func (T *qClient) mBanner(name string) {
	scale := T.scrGetMenuScale()

	w, _ := T.Draw_GetPicSize(name)
	T.Draw_PicScaled(T.viddef.width/2-int(float32(w)*scale)/2, T.viddef.height/2-int(110*scale), name, scale)
}

func (T *qClient) mPopMenu() {
	// S_StartLocalSound(menu_out_sound);

	if len(T.menu.m_layers) < 1 {
		log.Fatal("M_PopMenu: depth < 1")
	}

	T.menu.m_layers = T.menu.m_layers[:len(T.menu.m_layers)-1]
	last := T.menu.m_layers[len(T.menu.m_layers)-1]

	T.menu.m_drawfunc = last.draw
	T.menu.m_keyfunc = last.key

	// if (!m_menudepth) {
	//     M_ForceMenuOff();
	// }
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
	println("m_Main_Key")
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
			// M_Menu_Multiplayer_f()

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
 * GAME MENU
 */

func (T *qClient) game_MenuInit() {
	// Mods_NamesInit();

	T.menu.s_game_menu.x = int(float32(T.viddef.width) * 0.50)
	T.menu.s_game_menu.owner = T

	T.menu.s_easy_game_action.flags = QMF_LEFT_JUSTIFY
	T.menu.s_easy_game_action.x = 0
	T.menu.s_easy_game_action.y = 0
	T.menu.s_easy_game_action.name = "easy"
	// s_easy_game_action.generic.callback = EasyGameFunc;

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
	// return Default_MenuKey(&s_game_menu, key)
	return ""
}

func m_Menu_Game_f(args []string, a interface{}) error {
	T := a.(*qClient)
	T.game_MenuInit()
	T.mPushMenu(game_MenuDraw, game_MenuKey, "game")
	T.menu.m_game_cursor = 1
	return nil
}

func (T *qClient) mInit() {
	T.common.Cmd_AddCommand("menu_main", m_Menu_Main_f, T)
	T.common.Cmd_AddCommand("menu_game", m_Menu_Game_f, T)
	// Cmd_AddCommand("menu_loadgame", M_Menu_LoadGame_f);
	// Cmd_AddCommand("menu_savegame", M_Menu_SaveGame_f);
	// Cmd_AddCommand("menu_joinserver", M_Menu_JoinServer_f);
	// Cmd_AddCommand("menu_addressbook", M_Menu_AddressBook_f);
	// Cmd_AddCommand("menu_startserver", M_Menu_StartServer_f);
	// Cmd_AddCommand("menu_dmoptions", M_Menu_DMOptions_f);
	// Cmd_AddCommand("menu_playerconfig", M_Menu_PlayerConfig_f);
	// Cmd_AddCommand("menu_downloadoptions", M_Menu_DownloadOptions_f);
	// Cmd_AddCommand("menu_credits", M_Menu_Credits_f);
	// Cmd_AddCommand("menu_mods", M_Menu_Mods_f);
	// Cmd_AddCommand("menu_multiplayer", M_Menu_Multiplayer_f);
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
