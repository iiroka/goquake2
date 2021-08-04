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
 * Item spawning.
 *
 * =======================================================================
 */
package game

import (
	"fmt"
	"goquake2/shared"
	"math"
	"reflect"
	"strconv"
)

var spawns = map[string]func(ent *edict_t, G *qGame) error{
	"worldspawn": spWorldspawn,
	"light":      spLight,
}

/*
 * Finds the spawn function for
 * the entity and calls it
 */
func (G *qGame) edCallSpawn(ent *edict_t) error {
	//  spawn_t *s;
	//  gitem_t *item;
	//  int i;

	if ent == nil {
		return nil
	}

	if len(ent.Classname) == 0 {
		G.gi.Dprintf("ED_CallSpawn: NULL classname\n")
		//  G_FreeEdict(ent);
		return nil
	}

	/* check item spawn functions */
	//  for (i = 0, item = itemlist; i < game.num_items; i++, item++)
	//  {
	// 	 if (!item->classname)
	// 	 {
	// 		 continue;
	// 	 }

	// 	 if (!strcmp(item->classname, ent->classname))
	// 	 {
	// 		 /* found it */
	// 		 SpawnItem(ent, item);
	// 		 return;
	// 	 }
	//  }

	/* check normal spawn functions */
	if s, ok := spawns[ent.Classname]; ok {
		/* found it */
		return s(ent, G)
	}

	G.gi.Dprintf("%s doesn't have a spawn function\n", ent.Classname)
	return nil
}

/*
 * Takes a key/value pair and sets
 * the binary values in an edict
 */
func (G *qGame) edParseField(key, value string, ent *edict_t) {

	for _, f := range fields {
		if (f.flags&FFL_NOSPAWN) == 0 && f.name == key {
			/* found it */

			var b reflect.Value
			if (f.flags & FFL_ENTITYSTATE) != 0 {
				b = reflect.ValueOf(&ent.s).Elem()
			} else if (f.flags & FFL_SPAWNTEMP) != 0 {
				b = reflect.ValueOf(&G.st).Elem()
			} else {
				b = reflect.ValueOf(ent).Elem()
			}

			switch f.ftype {
			case F_LSTRING:
				b.FieldByName(f.fname).SetString(value)
			case F_VECTOR:
				var vect [3]float32
				fmt.Sscanf(value, "%f %f %f", &vect[0], &vect[1], &vect[2])
				tgt := b.FieldByName(f.fname)
				tgt.Index(0).SetFloat(float64(vect[0]))
				tgt.Index(1).SetFloat(float64(vect[1]))
				tgt.Index(2).SetFloat(float64(vect[2]))
			case F_INT:
				v, _ := strconv.ParseInt(value, 10, 32)
				b.FieldByName(f.fname).SetInt(v)
			case F_FLOAT:
				v, _ := strconv.ParseFloat(value, 32)
				b.FieldByName(f.fname).SetFloat(v)
			case F_ANGLEHACK:
				v, _ := strconv.ParseFloat(value, 32)
				tgt := b.FieldByName(f.fname)
				tgt.Index(0).SetFloat(float64(0))
				tgt.Index(1).SetFloat(float64(v))
				tgt.Index(2).SetFloat(float64(0))
			case F_IGNORE:
			default:
			}

			return
		}
	}

	G.gi.Dprintf("%s is not a field\n", key)
}

/*
 * Parses an edict out of the given string,
 * returning the new position ed should be
 * a properly initialized empty edict.
 */
func (G *qGame) edParseEdict(data string, index int, ent *edict_t) (int, error) {

	if ent == nil {
		return -1, nil
	}

	init := false
	G.st = spawn_temp_t{}

	/* go through all the dictionary pairs */
	for {
		/* parse key */
		var token string
		token, index = shared.COM_Parse(data, index)

		if token[0] == '}' {
			break
		}

		if index < 0 {
			return -1, G.gi.Error("ED_ParseEntity: EOF without closing brace")
		}

		keyname := string(token)

		/* parse value */
		token, index = shared.COM_Parse(data, index)

		if index < 0 {
			return -1, G.gi.Error("ED_ParseEntity: EOF without closing brace")
		}

		if token[0] == '}' {
			return -1, G.gi.Error("ED_ParseEntity: closing brace without data")
		}

		init = true

		/* keynames with a leading underscore are
		used for utility comments, and are
		immediately discarded by quake */
		if keyname[0] == '_' {
			continue
		}

		G.edParseField(keyname, token, ent)
	}

	if !init {
		// 	 memset(ent, 0, sizeof(*ent));
	}

	return index, nil
}

/* =================================================================== */

const single_statusbar = "yb	-24 " +

	/* health */
	"xv	0 " +
	"hnum " +
	"xv	50 " +
	"pic 0 " +

	/* ammo */
	"if 2 " +
	"	xv	100 " +
	"	anum " +
	"	xv	150 " +
	"	pic 2 " +
	"endif " +

	/* armor */
	"if 4 " +
	"	xv	200 " +
	"	rnum " +
	"	xv	250 " +
	"	pic 4 " +
	"endif " +

	/* selected item */
	"if 6 " +
	"	xv	296 " +
	"	pic 6 " +
	"endif " +

	"yb	-50 " +

	/* picked up item */
	"if 7 " +
	"	xv	0 " +
	"	pic 7 " +
	"	xv	26 " +
	"	yb	-42 " +
	"	stat_string 8 " +
	"	yb	-50 " +
	"endif " +

	/* timer */
	"if 9 " +
	"	xv	262 " +
	"	num	2	10 " +
	"	xv	296 " +
	"	pic	9 " +
	"endif " +

	/*  help / weapon icon */
	"if 11 " +
	"	xv	148 " +
	"	pic	11 " +
	"endif "

const dm_statusbar = "yb	-24 " +

	/* health */
	"xv	0 " +
	"hnum " +
	"xv	50 " +
	"pic 0 " +

	/* ammo */
	"if 2 " +
	"	xv	100 " +
	"	anum " +
	"	xv	150 " +
	"	pic 2 " +
	"endif " +

	/* armor */
	"if 4 " +
	"	xv	200 " +
	"	rnum " +
	"	xv	250 " +
	"	pic 4 " +
	"endif " +

	/* selected item */
	"if 6 " +
	"	xv	296 " +
	"	pic 6 " +
	"endif " +

	"yb	-50 " +

	/* picked up item */
	"if 7 " +
	"	xv	0 " +
	"	pic 7 " +
	"	xv	26 " +
	"	yb	-42 " +
	"	stat_string 8 " +
	"	yb	-50 " +
	"endif " +

	/* timer */
	"if 9 " +
	"	xv	246 " +
	"	num	2	10 " +
	"	xv	296 " +
	"	pic	9 " +
	"endif " +

	/*  help / weapon icon */
	"if 11 " +
	"	xv	148 " +
	"	pic	11 " +
	"endif " +

	/*  frags */
	"xr	-50 " +
	"yt 2 " +
	"num 3 14 " +

	/* spectator */
	"if 17 " +
	"xv 0 " +
	"yb -58 " +
	"string2 \"SPECTATOR MODE\" " +
	"endif " +

	/* chase camera */
	"if 16 " +
	"xv 0 " +
	"yb -68 " +
	"string \"Chasing\" " +
	"xv 64 " +
	"stat_string 16 " +
	"endif "

/*
 * Creates a server's entity / program execution context by
 * parsing textual entity definitions out of an ent file.
 */
func (G *qGame) SpawnEntities(mapname, entities, spawnpoint string) error {
	//  edict_t *ent;
	//  int inhibit;
	//  const char *com_token;
	//  int i;
	//  float skill_level;
	//  static qboolean monster_count_city2 = false;
	//  static qboolean monster_count_city3 = false;
	//  static qboolean monster_count_cool1 = false;
	//  static qboolean monster_count_lab = false;

	//  if (!mapname || !entities || !spawnpoint)
	//  {
	// 	 return;
	//  }

	skill_level := math.Floor(float64(G.skill.Float()))

	if skill_level < 0 {
		skill_level = 0
	}

	if skill_level > 3 {
		skill_level = 3
	}

	if float64(G.skill.Float()) != skill_level {
		//  gi.cvar_forceset("skill", va("%f", skill_level));
	}

	//  SaveClientData();

	//  gi.FreeTags(TAG_LEVEL);

	//  memset(&level, 0, sizeof(level));
	//  memset(g_edicts, 0, game.maxentities * sizeof(g_edicts[0]));

	//  Q_strlcpy(level.mapname, mapname, sizeof(level.mapname));
	//  Q_strlcpy(game.spawnpoint, spawnpoint, sizeof(game.spawnpoint));

	/* set client fields on player ents */
	for i := 0; i < G.game.maxclients; i++ {
		G.g_edicts[i+1].client = &G.game.clients[i]
	}

	var ent *edict_t
	inhibit := 0

	/* parse ents */
	index := 0
	var err error
	for index >= 0 && index < len(entities) {
		// 	 /* parse the opening brace */
		var token string
		token, index = shared.COM_Parse(entities, index)
		if index < 0 {
			break
		}

		if token[0] != '{' {
			return G.gi.Error("ED_LoadFromFile: found %s when expecting {", token)
		}

		if ent == nil {
			ent = &G.g_edicts[0]
		} else {
			ent, err = G.gSpawn()
			if err != nil {
				return err
			}
		}

		index, err = G.edParseEdict(entities, index, ent)
		if err != nil {
			return err
		}

		// 	 /* yet another map hack */
		// 	 if (!Q_stricmp(level.mapname, "command") &&
		// 		 !Q_stricmp(ent->classname, "trigger_once") &&
		// 			!Q_stricmp(ent->model, "*27")) {
		// 		 ent->spawnflags &= ~SPAWNFLAG_NOT_HARD;
		// 	 }

		// 	 /*
		// 	  * The 'monsters' count in city3.bsp is wrong.
		// 	  * There're two monsters triggered in a hidden
		// 	  * and unreachable room next to the security
		// 	  * pass.
		// 	  *
		// 	  * We need to make sure that this hack is only
		// 	  * applied once!
		// 	  */
		// 	 if (!Q_stricmp(level.mapname, "city3") && !monster_count_city3)
		// 	 {
		// 		 level.total_monsters = level.total_monsters - 2;
		// 		 monster_count_city3 = true;
		// 	 }

		// 	 /* A slightly other problem in city2.bsp. There's a floater
		// 	  * with missing trigger on the right gallery above the data
		// 	  * spinner console, right before the door to the staircase.
		// 	  */
		// 	 if ((skill->value > 0) && !Q_stricmp(level.mapname, "city2") && !monster_count_city2)
		// 	 {
		// 		 level.total_monsters = level.total_monsters - 1;
		// 		 monster_count_city2 = true;
		// 	 }

		// 	 /*
		// 	  * Nearly the same problem exists in cool1.bsp.
		// 	  * On medium skill a gladiator is spawned in a
		// 	  * crate that's never triggered.
		// 	  */
		// 	 if ((skill->value == 1) && !Q_stricmp(level.mapname, "cool1") && !monster_count_cool1)
		// 	 {
		// 		 level.total_monsters = level.total_monsters - 1;
		// 		 monster_count_cool1 = true;
		// 	 }

		// 	 /*
		// 	  * Nearly the same problem exists in lab.bsp.
		// 	  * On medium skill two parasites are spawned
		// 	  * in a hidden place that never triggers.
		// 	  */
		// 	 if ((skill->value == 1) && !Q_stricmp(level.mapname, "lab") && !monster_count_lab)
		// 	 {
		// 		 level.total_monsters = level.total_monsters - 2;
		// 		 monster_count_lab = true;
		// 	 }

		// 	 /* remove things (except the world) from
		// 		different skill levels or deathmatch */
		// 	 if (ent != g_edicts) {
		// 		 if (deathmatch->value) {
		// 			 if (ent->spawnflags & SPAWNFLAG_NOT_DEATHMATCH) != 0 {
		// 				 G_FreeEdict(ent);
		// 				 inhibit++;
		// 				 continue;
		// 			 }
		// 		 } else {
		// 			 if (((skill->value == SKILL_EASY) &&
		// 				  (ent->spawnflags & SPAWNFLAG_NOT_EASY)) ||
		// 				 ((skill->value == SKILL_MEDIUM) &&
		// 				  (ent->spawnflags & SPAWNFLAG_NOT_MEDIUM)) ||
		// 				 (((skill->value == SKILL_HARD) ||
		// 				   (skill->value == SKILL_HARDPLUS)) &&
		// 				  (ent->spawnflags & SPAWNFLAG_NOT_HARD))
		// 				 ) {
		// 				 G_FreeEdict(ent);
		// 				 inhibit++;
		// 				 continue;
		// 			 }
		// 		 }

		// 		 ent->spawnflags &=
		// 			 ~(SPAWNFLAG_NOT_EASY | SPAWNFLAG_NOT_MEDIUM |
		// 			   SPAWNFLAG_NOT_HARD |
		// 			   SPAWNFLAG_NOT_COOP | SPAWNFLAG_NOT_DEATHMATCH);
		// 	 }

		if err := G.edCallSpawn(ent); err != nil {
			return err
		}
	}

	G.gi.Dprintf("%v entities inhibited.\n", inhibit)

	//  G_FindTeams();

	//  PlayerTrail_Init();
	return nil
}

/*QUAKED worldspawn (0 0 0) ?
 *
 * Only used for the world.
 *  "sky"		environment map name
 *  "skyaxis"	vector axis for rotating sky
 *  "skyrotate"	speed of rotation in degrees/second
 *  "sounds"	music cd track number
 *  "gravity"	800 is default gravity
 *  "message"	text to print at user logon
 */
func spWorldspawn(ent *edict_t, G *qGame) error {
	if ent == nil {
		return nil
	}

	// ent.movetype = MOVETYPE_PUSH
	ent.solid = shared.SOLID_BSP
	ent.inuse = true     /* since the world doesn't use G_Spawn() */
	ent.s.Modelindex = 1 /* world model is always index 1 */

	/* --------------- */

	//  /* reserve some spots for dead
	// 	player bodies for coop / deathmatch */
	//  InitBodyQue();

	//  /* set configstrings for items */
	//  SetItemNames();

	//  if (G.st.Nextmap) {
	// 	 strcpy(level.nextmap, st.nextmap);
	//  }

	//  /* make some data visible to the server */
	//  if (ent->message && ent->message[0]) {
	// 	 gi.configstring(CS_NAME, ent->message);
	// 	 Q_strlcpy(level.level_name, ent->message, sizeof(level.level_name));
	//  } else {
	// 	 Q_strlcpy(level.level_name, level.mapname, sizeof(level.level_name));
	//  }

	if len(G.st.Sky) > 0 {
		G.gi.Configstring(shared.CS_SKY, G.st.Sky)
	} else {
		G.gi.Configstring(shared.CS_SKY, "unit1_")
	}

	G.gi.Configstring(shared.CS_SKYROTATE, fmt.Sprintf("%f", G.st.Skyrotate))

	G.gi.Configstring(shared.CS_SKYAXIS, fmt.Sprintf("%f %f %f",
		G.st.Skyaxis[0], G.st.Skyaxis[1], G.st.Skyaxis[2]))

	//  gi.configstring(CS_CDTRACK, va("%i", ent->sounds));

	G.gi.Configstring(shared.CS_MAXCLIENTS, fmt.Sprintf("%v", G.maxclients.Int()))

	/* status bar program */
	if G.deathmatch.Bool() {
		G.gi.Configstring(shared.CS_STATUSBAR, dm_statusbar)
	} else {
		G.gi.Configstring(shared.CS_STATUSBAR, single_statusbar)
	}

	//  /* --------------- */

	//  /* help icon for statusbar */
	//  gi.imageindex("i_help");
	//  level.pic_health = gi.imageindex("i_health");
	//  gi.imageindex("help");
	//  gi.imageindex("field_3");

	//  if (!st.gravity) {
	// 	 gi.cvar_set("sv_gravity", "800");
	//  } else {
	// 	 gi.cvar_set("sv_gravity", st.gravity);
	//  }

	//  snd_fry = gi.soundindex("player/fry.wav"); /* standing in lava / slime */

	//  PrecacheItem(FindItem("Blaster"));

	//  gi.soundindex("player/lava1.wav");
	//  gi.soundindex("player/lava2.wav");

	//  gi.soundindex("misc/pc_up.wav");
	//  gi.soundindex("misc/talk1.wav");

	//  gi.soundindex("misc/udeath.wav");

	//  /* gibs */
	//  gi.soundindex("items/respawn1.wav");

	//  /* sexed sounds */
	//  gi.soundindex("*death1.wav");
	//  gi.soundindex("*death2.wav");
	//  gi.soundindex("*death3.wav");
	//  gi.soundindex("*death4.wav");
	//  gi.soundindex("*fall1.wav");
	//  gi.soundindex("*fall2.wav");
	//  gi.soundindex("*gurp1.wav"); /* drowning damage */
	//  gi.soundindex("*gurp2.wav");
	//  gi.soundindex("*jump1.wav"); /* player jump */
	//  gi.soundindex("*pain25_1.wav");
	//  gi.soundindex("*pain25_2.wav");
	//  gi.soundindex("*pain50_1.wav");
	//  gi.soundindex("*pain50_2.wav");
	//  gi.soundindex("*pain75_1.wav");
	//  gi.soundindex("*pain75_2.wav");
	//  gi.soundindex("*pain100_1.wav");
	//  gi.soundindex("*pain100_2.wav");

	//  /* sexed models: THIS ORDER MUST MATCH THE DEFINES IN g_local.h
	// 	you can add more, max 19 (pete change)these models are only
	// 	loaded in coop or deathmatch. not singleplayer. */
	//  if (coop->value || deathmatch->value) {
	// 	 gi.modelindex("#w_blaster.md2");
	// 	 gi.modelindex("#w_shotgun.md2");
	// 	 gi.modelindex("#w_sshotgun.md2");
	// 	 gi.modelindex("#w_machinegun.md2");
	// 	 gi.modelindex("#w_chaingun.md2");
	// 	 gi.modelindex("#a_grenades.md2");
	// 	 gi.modelindex("#w_glauncher.md2");
	// 	 gi.modelindex("#w_rlauncher.md2");
	// 	 gi.modelindex("#w_hyperblaster.md2");
	// 	 gi.modelindex("#w_railgun.md2");
	// 	 gi.modelindex("#w_bfg.md2");
	//  }

	//  /* ------------------- */

	//  gi.soundindex("player/gasp1.wav"); /* gasping for air */
	//  gi.soundindex("player/gasp2.wav"); /* head breaking surface, not gasping */

	//  gi.soundindex("player/watr_in.wav"); /* feet hitting water */
	//  gi.soundindex("player/watr_out.wav"); /* feet leaving water */

	//  gi.soundindex("player/watr_un.wav"); /* head going underwater */

	//  gi.soundindex("player/u_breath1.wav");
	//  gi.soundindex("player/u_breath2.wav");

	//  gi.soundindex("items/pkup.wav"); /* bonus item pickup */
	//  gi.soundindex("world/land.wav"); /* landing thud */
	//  gi.soundindex("misc/h2ohit1.wav"); /* landing splash */

	//  gi.soundindex("items/damage.wav");
	//  gi.soundindex("items/protect.wav");
	//  gi.soundindex("items/protect4.wav");
	//  gi.soundindex("weapons/noammo.wav");

	//  gi.soundindex("infantry/inflies1.wav");

	//  sm_meat_index = gi.modelindex("models/objects/gibs/sm_meat/tris.md2");
	//  gi.modelindex("models/objects/gibs/arm/tris.md2");
	//  gi.modelindex("models/objects/gibs/bone/tris.md2");
	//  gi.modelindex("models/objects/gibs/bone2/tris.md2");
	//  gi.modelindex("models/objects/gibs/chest/tris.md2");
	//  gi.modelindex("models/objects/gibs/skull/tris.md2");
	//  gi.modelindex("models/objects/gibs/head2/tris.md2");

	/* Setup light animation tables. 'a'
	is total darkness, 'z' is doublebright. */

	/* 0 normal */
	G.gi.Configstring(shared.CS_LIGHTS+0, "m")

	/* 1 FLICKER (first variety) */
	G.gi.Configstring(shared.CS_LIGHTS+1, "mmnmmommommnonmmonqnmmo")

	/* 2 SLOW STRONG PULSE */
	G.gi.Configstring(shared.CS_LIGHTS+2, "abcdefghijklmnopqrstuvwxyzyxwvutsrqponmlkjihgfedcba")

	/* 3 CANDLE (first variety) */
	G.gi.Configstring(shared.CS_LIGHTS+3, "mmmmmaaaaammmmmaaaaaabcdefgabcdefg")

	/* 4 FAST STROBE */
	G.gi.Configstring(shared.CS_LIGHTS+4, "mamamamamama")

	/* 5 GENTLE PULSE 1 */
	G.gi.Configstring(shared.CS_LIGHTS+5, "jklmnopqrstuvwxyzyxwvutsrqponmlkj")

	/* 6 FLICKER (second variety) */
	G.gi.Configstring(shared.CS_LIGHTS+6, "nmonqnmomnmomomno")

	/* 7 CANDLE (second variety) */
	G.gi.Configstring(shared.CS_LIGHTS+7, "mmmaaaabcdefgmmmmaaaammmaamm")

	/* 8 CANDLE (third variety) */
	G.gi.Configstring(shared.CS_LIGHTS+8, "mmmaaammmaaammmabcdefaaaammmmabcdefmmmaaaa")

	/* 9 SLOW STROBE (fourth variety) */
	G.gi.Configstring(shared.CS_LIGHTS+9, "aaaaaaaazzzzzzzz")

	/* 10 FLUORESCENT FLICKER */
	G.gi.Configstring(shared.CS_LIGHTS+10, "mmamammmmammamamaaamammma")

	/* 11 SLOW PULSE NOT FADE TO BLACK */
	G.gi.Configstring(shared.CS_LIGHTS+11, "abcdefghijklmnopqrrqponmlkjihgfedcba")

	/* styles 32-62 are assigned by the light program for switchable lights */

	/* 63 testing */
	G.gi.Configstring(shared.CS_LIGHTS+63, "a")
	return nil
}
