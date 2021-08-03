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
 * The Quake II CVAR subsystem. Implements dynamic variable handling.
 *
 * =======================================================================
 */
package common

import (
	"goquake2/shared"
	"strings"
)

/* An ugly hack to rewrite CVARs loaded from config.cfg */
var replacements = map[string]string{
	"cd_shuffle":                "ogg_shuffle",
	"cl_anglekicks":             "cl_kickangles",
	"cl_drawfps":                "cl_showfps",
	"gl_drawentities":           "r_drawentities",
	"gl_drawworld":              "r_drawworld",
	"gl_fullbright":             "r_fullbright",
	"gl_lerpmodels":             "r_lerpmodels",
	"gl_lightlevel":             "r_lightlevel",
	"gl_norefresh":              "r_norefresh",
	"gl_novis":                  "r_novis",
	"gl_speeds":                 "r_speeds",
	"gl_clear":                  "r_clear",
	"gl_consolescale":           "r_consolescale",
	"gl_hudscale":               "r_hudscale",
	"gl_menuscale":              "r_scale",
	"gl_customheight":           "r_customheight",
	"gl_customwidth":            "r_customheight",
	"gl_dynamic":                "gl1_dynamic",
	"gl_farsee":                 "r_farsee",
	"gl_flashblend":             "gl1_flashblend",
	"gl_lockpvs":                "r_lockpvs",
	"gl_maxfps":                 "vid_maxfps",
	"gl_mode":                   "r_mode",
	"gl_modulate":               "r_modulate",
	"gl_overbrightbits":         "gl1_overbrightbits",
	"gl_palettedtextures":       "gl1_palettedtextures",
	"gl_particle_min_size":      "gl1_particle_min_size",
	"gl_particle_max_size":      "gl1_particle_max_size",
	"gl_particle_size":          "gl1_particle_size",
	"gl_particle_att_a":         "gl1_particle_att_a",
	"gl_particle_att_b":         "gl1_particle_att_b",
	"gl_particle_att_c":         "gl1_particle_att_c",
	"gl_picmip":                 "gl1_picmip",
	"gl_pointparameters":        "gl1_pointparameters",
	"gl_polyblend":              "gl1_polyblend",
	"gl_round_down":             "gl1_round_down",
	"gl_saturatelightning":      "gl1_saturatelightning",
	"gl_stencilshadows":         "gl1_stencilshadows",
	"gl_stereo":                 "gl1_stereo",
	"gl_stereo_separation":      "gl1_stereo_separation",
	"gl_stereo_anaglyph_colors": "gl1_stereo_anaglyph_colors",
	"gl_stereo_convergence":     "gl1_stereo_convergence",
	"gl_swapinterval":           "r_vsync",
	"gl_texturealphamode":       "gl1_texturealphamode",
	"gl_texturesolidmode":       "gl1_texturesolidmode",
	"gl_ztrick":                 "gl1_ztrick",
	"gl_msaa_samples":           "r_msaa_samples",
	"gl_nolerp_list":            "r_nolerp_list",
	"gl_retexturing":            "r_retexturing",
	"gl_shadows":                "r_shadows",
	"gl_anisotropic":            "r_anisotropic",
	"intensity":                 "gl1_intensity",
}

func cvarInfoValidate(s string) bool {
	return !strings.ContainsAny(s, "\\\":")
}

func (T *qCommon) cvarFindVar(name string) *shared.CvarT {

	/* An ugly hack to rewrite changed CVARs */
	replacement, ok := replacements[name]
	if ok {
		v, ok2 := T.cvarVars[replacement]
		if ok2 {
			return v
		}
	} else {
		v, ok2 := T.cvarVars[name]
		if ok2 {
			return v
		}
	}
	return nil
}

func (T *qCommon) Cvar_VariableBool(var_name string) bool {
	v := T.cvarFindVar(var_name)
	if v == nil {
		return false
	}
	return v.Bool()
}

func (T *qCommon) Cvar_VariableInt(var_name string) int {
	v := T.cvarFindVar(var_name)
	if v == nil {
		return 0
	}
	return v.Int()
}

func (T *qCommon) Cvar_VariableString(var_name string) string {
	v := T.cvarFindVar(var_name)
	if v == nil {
		return ""
	}
	return v.String
}

/*
 * If the variable already exists, the value will not be set
 * The flags will be or'ed in if the variable exists.
 */
func (T *qCommon) Cvar_Get(var_name, var_value string, flags int) *shared.CvarT {

	if (flags & (shared.CVAR_USERINFO | shared.CVAR_SERVERINFO)) != 0 {
		if !cvarInfoValidate(var_name) {
			T.Com_Printf("invalid info cvar name\n")
			return nil
		}
	}

	v := T.cvarFindVar(var_name)

	if v != nil {
		v.Flags |= flags

		// 	 if (!var_value)
		// 	 {
		// 		 var->default_string = CopyString("");
		// 	 }
		// 	 else
		// 	 {
		// 		 var->default_string = CopyString(var_value);
		// 	 }

		return v
	}

	//  if (!var_value)
	//  {
	// 	 return NULL;
	//  }

	if (flags & (shared.CVAR_USERINFO | shared.CVAR_SERVERINFO)) != 0 {
		if !cvarInfoValidate(var_value) {
			T.Com_Printf("invalid info cvar value\n")
			return nil
		}
	}

	//  // if $game is the default one ("baseq2"), then use "" instead because
	//  // other code assumes this behavior (e.g. FS_BuildGameSpecificSearchPath())
	//  if(strcmp(var_name, "game") == 0 && strcmp(var_value, BASEDIRNAME) == 0)
	//  {
	// 	 var_value = "";
	//  }

	v = &shared.CvarT{}
	v.Name = var_name
	v.String = var_value
	v.DefaultString = var_value
	v.LatchedString = nil
	v.Modified = true
	v.Flags = flags

	T.cvarVars[var_name] = v

	return v
}

func (T *qCommon) Cvar_Set2(var_name, value string, force bool) *shared.CvarT {

	v := T.cvarFindVar(var_name)
	if v == nil {
		return T.Cvar_Get(var_name, value, 0)
	}

	if (v.Flags & (shared.CVAR_USERINFO | shared.CVAR_SERVERINFO)) != 0 {
		if !cvarInfoValidate(value) {
			T.Com_Printf("invalid info cvar value\n")
			return v
		}
	}

	// if $game is the default one ("baseq2"), then use "" instead because
	// other code assumes this behavior (e.g. FS_BuildGameSpecificSearchPath())
	// if(strcmp(var_name, "game") == 0 && strcmp(value, BASEDIRNAME) == 0) {
	// 	value = "";
	// }

	if !force {
		if (v.Flags & shared.CVAR_NOSET) != 0 {
			T.Com_Printf("%s is write protected.\n", var_name)
			return v
		}

		if (v.Flags & shared.CVAR_LATCH) != 0 {
			if v.LatchedString != nil {
				if value == *v.LatchedString {
					return v
				}

				v.LatchedString = nil
			} else {
				if value == v.String {
					return v
				}
			}

			if T.ServerState() != 0 {
				T.Com_Printf("%v will be changed for next game.\n", var_name)
				v.LatchedString = &value
			} else {
				v.String = string(value)

				// if (!strcmp(var->name, "game")) {
				// 	FS_BuildGameSpecificSearchPath(var->string);
				// }
			}

			return v
		}
	} else {
		v.LatchedString = nil
	}

	if value == v.String {
		return v
	}

	v.Modified = true

	if (v.Flags & shared.CVAR_USERINFO) != 0 {
		T.UserinfoModified = true
	}

	v.String = string(value)

	return v
}

func (T *qCommon) Cvar_ForceSet(var_name, value string) *shared.CvarT {
	return T.Cvar_Set2(var_name, value, true)
}

func (T *qCommon) Cvar_Set(var_name, value string) *shared.CvarT {
	return T.Cvar_Set2(var_name, value, false)
}

func (T *qCommon) Cvar_FullSet(var_name, value string, flags int) *shared.CvarT {

	v := T.cvarFindVar(var_name)
	if v == nil {
		return T.Cvar_Get(var_name, value, flags)
	}

	v.Modified = true

	if (v.Flags & shared.CVAR_USERINFO) != 0 {
		T.UserinfoModified = true
	}

	// if $game is the default one ("baseq2"), then use "" instead because
	// other code assumes this behavior (e.g. FS_BuildGameSpecificSearchPath())
	// if(strcmp(var_name, "game") == 0 && strcmp(value, BASEDIRNAME) == 0)
	// {
	// 	value = "";
	// }

	v.String = string(value)
	v.Flags = flags

	return v
}

func (T *qCommon) cvarBitInfo(bit int) string {
	// static char info[MAX_INFO_STRING];
	// cvar_t *var;

	info := ""

	for key, val := range T.cvarVars {
		if (val.Flags & bit) != 0 {
			info += "\\"
			info += key
			info += "\\"
			info += val.String
		}
	}

	return info
}

/*
 * returns an info string containing
 * all the CVAR_USERINFO cvars
 */
func (T *qCommon) Cvar_Userinfo() string {
	return T.cvarBitInfo(shared.CVAR_USERINFO)
}

/*
 * returns an info string containing
 * all the CVAR_SERVERINFO cvars
 */
func (T *qCommon) Cvar_Serverinfo() string {
	return T.cvarBitInfo(shared.CVAR_SERVERINFO)
}

func (T *qCommon) Cvar_ClearUserinfoModified() {
	T.UserinfoModified = false
}

/*
 * Handles variable inspection and changing from the console
 */
func (T *qCommon) cvar_Command(args []string) bool {

	/* check variables */
	v := T.cvarFindVar(args[0])
	if v == nil {
		return false
	}

	/* perform a variable print or set */
	if len(args) == 1 {
		T.Com_Printf("\"%s\" is \"%s\"\n", v.Name, v.String)
		return true
	}

	/* Another evil hack: The user has just changed 'game' trough
	the console. We reset userGivenGame to that value, otherwise
	we would revert to the initialy given game at disconnect. */
	//  if (strcmp(v->name, "game") == 0) {
	// 	 Q_strlcpy(userGivenGame, Cmd_Argv(1), sizeof(userGivenGame));
	//  }

	T.Cvar_Set(v.Name, args[1])
	return true
}

/*
 * Allows setting and defining of arbitrary cvars from console
 */
func cvar_Set_f(args []string, arg interface{}) error {
	//  char *firstarg;
	//  int c, i;

	//  c = Cmd_Argc();
	T := arg.(*qCommon)

	if (len(args) != 3) && (len(args) != 4) {
		T.Com_Printf("usage: set <variable> <value> [u / s]\n")
		return nil
	}

	firstarg := args[1]

	/* An ugly hack to rewrite changed CVARs */
	replacement, ok := replacements[firstarg]
	if ok {
		firstarg = replacement
	}

	if len(args) == 4 {
		var flags = 0

		if args[3] == "u" {
			flags = shared.CVAR_USERINFO
		} else if args[3] == "s" {
			flags = shared.CVAR_SERVERINFO
		} else {
			T.Com_Printf("flags can only be 'u' or 's'\n")
			return nil
		}

		T.Cvar_FullSet(firstarg, args[2], flags)
	} else {
		T.Cvar_Set(firstarg, args[2])
	}
	return nil
}

/*
 * Reads in all archived cvars
 */
func (T *qCommon) cvarInit() {
	// Cmd_AddCommand("cvarlist", Cvar_List_f)
	// Cmd_AddCommand("dec", Cvar_Inc_f)
	// Cmd_AddCommand("inc", Cvar_Inc_f)
	// Cmd_AddCommand("reset", Cvar_Reset_f)
	// Cmd_AddCommand("resetall", Cvar_ResetAll_f)
	T.Cmd_AddCommand("set", cvar_Set_f, T)
	// Cmd_AddCommand("toggle", Cvar_Toggle_f)
}
