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
 * This is the main header file shared between client, renderer, server
 * and the game. Do NOT edit this file unless you know what you're
 * doing. Changes here may break the client <-> renderer <-> server
 * <-> game API, leading to problems with mods!
 *
 * =======================================================================
 */
package shared

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
)

const (
	/* angle indexes */
	PITCH = 0 /* up / down */
	YAW   = 1 /* left / right */
	ROLL  = 2 /* fall over */

	/* per-level limits */
	MAX_CLIENTS     = 256  /* absolute limit */
	MAX_EDICTS      = 1024 /* must change protocol to increase more */
	MAX_LIGHTSTYLES = 256
	MAX_MODELS      = 256 /* these are sent over the net as bytes */
	MAX_SOUNDS      = 256 /* so they cannot be blindly increased */
	MAX_IMAGES      = 256
	MAX_ITEMS       = 256
	MAX_GENERAL     = (MAX_CLIENTS * 2) /* general config strings */

	ERR_FATAL      = 0 /* exit the entire game with a popup window */
	ERR_DROP       = 1 /* print to console and disconnect from game */
	ERR_DISCONNECT = 2 /* don't kill server */

	PRINT_ALL       = 0
	PRINT_DEVELOPER = 1 /* only print when "developer 1" */
	PRINT_ALERT     = 2
)

/*
 * ==========================================================
 *
 * CVARS (console variables)
 *
 * ==========================================================
 */

const (
	CVAR_ARCHIVE    = 1 /* set to cause it to be saved to vars.rc */
	CVAR_USERINFO   = 2 /* added to userinfo  when changed */
	CVAR_SERVERINFO = 4 /* added to serverinfo when changed */
	CVAR_NOSET      = 8 /* don't allow change from console at all, */
	/* but can be set from the command line */
	CVAR_LATCH = 16 /* save changes until server restart */
)

/* nothing outside the Cvar_*() functions should modify these fields! */
type CvarT struct {
	Name          string
	String        string
	LatchedString *string /* for CVAR_LATCH vars */
	Flags         int
	Modified      bool /* set each time the cvar is changed */
	/* Added by YQ2. Must be at the end to preserve ABI. */
	DefaultString string
}

/* pmove_state_t is the information necessary for client side movement */
/* prediction */
type Pmtype_t int

const (
	/* can accelerate and turn */
	PM_NORMAL    Pmtype_t = 0
	PM_SPECTATOR Pmtype_t = 1
	/* no acceleration or turning */
	PM_DEAD   Pmtype_t = 2
	PM_GIB    Pmtype_t = 3 /* different bounding box */
	PM_FREEZE Pmtype_t = 4
)

/* pmove->pm_flags */
const (
	PMF_DUCKED         = 1
	PMF_JUMP_HELD      = 2
	PMF_ON_GROUND      = 4
	PMF_TIME_WATERJUMP = 8  /* pm_time is waterjump */
	PMF_TIME_LAND      = 16 /* pm_time is time before rejump */
	PMF_TIME_TELEPORT  = 32 /* pm_time is non-moving time */
	PMF_NO_PREDICTION  = 64 /* temporarily disables prediction (used for grappling hook) */
)

/* plane_t structure */
type Cplane_t struct {
	Normal   [3]float32
	Dist     float32
	Type     byte /* for fast side tests */
	Signbits byte /* signx + (signy<<1) + (signz<<2) */
}

/* this structure needs to be communicated bit-accurate/
 * from the server to the client to guarantee that
 * prediction stays in sync, so no floats are used.
 * if any part of the game code modifies this struct, it
 * will result in a prediction error of some degree. */
type Pmove_state_t struct {
	Pm_type Pmtype_t

	Origin       [3]int16 /* 12.3 */
	Velocity     [3]int16 /* 12.3 */
	Pm_flags     uint8    /* ducked, jump_held, etc */
	Pm_time      uint8    /* each unit = 8 ms */
	Gravity      int16
	Delta_angles [3]int16 /* add to command angles to get view direction
	 * changed by spawns, rotating objects, and teleporters */
}

func (T *Pmove_state_t) Copy(other Pmove_state_t) {
	T.Pm_type = other.Pm_type
	for i := 0; i < 3; i++ {
		T.Origin[i] = other.Origin[i]
		T.Velocity[i] = other.Velocity[i]
		T.Delta_angles[i] = other.Delta_angles[i]
	}
	T.Pm_flags = other.Pm_flags
	T.Pm_time = other.Pm_time
	T.Gravity = other.Gravity
}

const (
	/* entity_state_t->effects
	 * Effects are things handled on the client side (lights, particles,
	 * frame animations)  that happen constantly on the given entity.
	 * An entity that has effects will be sent to the client even if
	 * it has a zero index model. */
	EF_ROTATE           = 0x00000001 /* rotate (bonus items) */
	EF_GIB              = 0x00000002 /* leave a trail */
	EF_BLASTER          = 0x00000008 /* redlight + trail */
	EF_ROCKET           = 0x00000010 /* redlight + trail */
	EF_GRENADE          = 0x00000020
	EF_HYPERBLASTER     = 0x00000040
	EF_BFG              = 0x00000080
	EF_COLOR_SHELL      = 0x00000100
	EF_POWERSCREEN      = 0x00000200
	EF_ANIM01           = 0x00000400 /* automatically cycle between frames 0 and 1 at 2 hz */
	EF_ANIM23           = 0x00000800 /* automatically cycle between frames 2 and 3 at 2 hz */
	EF_ANIM_ALL         = 0x00001000 /* automatically cycle through all frames at 2hz */
	EF_ANIM_ALLFAST     = 0x00002000 /* automatically cycle through all frames at 10hz */
	EF_FLIES            = 0x00004000
	EF_QUAD             = 0x00008000
	EF_PENT             = 0x00010000
	EF_TELEPORTER       = 0x00020000 /* particle fountain */
	EF_FLAG1            = 0x00040000
	EF_FLAG2            = 0x00080000
	EF_IONRIPPER        = 0x00100000
	EF_GREENGIB         = 0x00200000
	EF_BLUEHYPERBLASTER = 0x00400000
	EF_SPINNINGLIGHTS   = 0x00800000
	EF_PLASMA           = 0x01000000
	EF_TRAP             = 0x02000000
	EF_TRACKER          = 0x04000000
	EF_DOUBLE           = 0x08000000
	EF_SPHERETRANS      = 0x10000000
	EF_TAGTRAIL         = 0x20000000
	EF_HALF_DAMAGE      = 0x40000000
	EF_TRACKERTRAIL     = 0x80000000

	/* entity_state_t->renderfx flags */
	RF_MINLIGHT       = 1  /* allways have some light (viewmodel) */
	RF_VIEWERMODEL    = 2  /* don't draw through eyes, only mirrors */
	RF_WEAPONMODEL    = 4  /* only draw through eyes */
	RF_FULLBRIGHT     = 8  /* allways draw full intensity */
	RF_DEPTHHACK      = 16 /* for view weapon Z crunching */
	RF_TRANSLUCENT    = 32
	RF_FRAMELERP      = 64
	RF_BEAM           = 128
	RF_CUSTOMSKIN     = 256 /* skin is an index in image_precache */
	RF_GLOW           = 512 /* pulse lighting for bonus items */
	RF_SHELL_RED      = 1024
	RF_SHELL_GREEN    = 2048
	RF_SHELL_BLUE     = 4096
	RF_NOSHADOW       = 8192       /* don't draw a shadow */
	RF_IR_VISIBLE     = 0x00008000 /* 32768 */
	RF_SHELL_DOUBLE   = 0x00010000 /* 65536 */
	RF_SHELL_HALF_DAM = 0x00020000
	RF_USE_DISGUISE   = 0x00040000

	/* player_state_t->refdef flags */
	RDF_UNDERWATER   = 1 /* warp the screen as apropriate */
	RDF_NOWORLDMODEL = 2 /* used for player configuration screen */
	RDF_IRGOGGLES    = 4
	RDF_UVGOGGLES    = 8

	/* muzzle flashes / player effects */
	MZ_BLASTER          = 0
	MZ_MACHINEGUN       = 1
	MZ_SHOTGUN          = 2
	MZ_CHAINGUN1        = 3
	MZ_CHAINGUN2        = 4
	MZ_CHAINGUN3        = 5
	MZ_RAILGUN          = 6
	MZ_ROCKET           = 7
	MZ_GRENADE          = 8
	MZ_LOGIN            = 9
	MZ_LOGOUT           = 10
	MZ_RESPAWN          = 11
	MZ_BFG              = 12
	MZ_SSHOTGUN         = 13
	MZ_HYPERBLASTER     = 14
	MZ_ITEMRESPAWN      = 15
	MZ_IONRIPPER        = 16
	MZ_BLUEHYPERBLASTER = 17
	MZ_PHALANX          = 18
	MZ_SILENCED         = 128 /* bit flag ORed with one of the above numbers */
	MZ_ETF_RIFLE        = 30
	MZ_UNUSED           = 31
	MZ_SHOTGUN2         = 32
	MZ_HEATBEAM         = 33
	MZ_BLASTER2         = 34
	MZ_TRACKER          = 35
	MZ_NUKE1            = 36
	MZ_NUKE2            = 37
	MZ_NUKE4            = 38
	MZ_NUKE8            = 39

	/* monster muzzle flashes */
	MZ2_TANK_BLASTER_1     = 1
	MZ2_TANK_BLASTER_2     = 2
	MZ2_TANK_BLASTER_3     = 3
	MZ2_TANK_MACHINEGUN_1  = 4
	MZ2_TANK_MACHINEGUN_2  = 5
	MZ2_TANK_MACHINEGUN_3  = 6
	MZ2_TANK_MACHINEGUN_4  = 7
	MZ2_TANK_MACHINEGUN_5  = 8
	MZ2_TANK_MACHINEGUN_6  = 9
	MZ2_TANK_MACHINEGUN_7  = 10
	MZ2_TANK_MACHINEGUN_8  = 11
	MZ2_TANK_MACHINEGUN_9  = 12
	MZ2_TANK_MACHINEGUN_10 = 13
	MZ2_TANK_MACHINEGUN_11 = 14
	MZ2_TANK_MACHINEGUN_12 = 15
	MZ2_TANK_MACHINEGUN_13 = 16
	MZ2_TANK_MACHINEGUN_14 = 17
	MZ2_TANK_MACHINEGUN_15 = 18
	MZ2_TANK_MACHINEGUN_16 = 19
	MZ2_TANK_MACHINEGUN_17 = 20
	MZ2_TANK_MACHINEGUN_18 = 21
	MZ2_TANK_MACHINEGUN_19 = 22
	MZ2_TANK_ROCKET_1      = 23
	MZ2_TANK_ROCKET_2      = 24
	MZ2_TANK_ROCKET_3      = 25

	MZ2_INFANTRY_MACHINEGUN_1  = 26
	MZ2_INFANTRY_MACHINEGUN_2  = 27
	MZ2_INFANTRY_MACHINEGUN_3  = 28
	MZ2_INFANTRY_MACHINEGUN_4  = 29
	MZ2_INFANTRY_MACHINEGUN_5  = 30
	MZ2_INFANTRY_MACHINEGUN_6  = 31
	MZ2_INFANTRY_MACHINEGUN_7  = 32
	MZ2_INFANTRY_MACHINEGUN_8  = 33
	MZ2_INFANTRY_MACHINEGUN_9  = 34
	MZ2_INFANTRY_MACHINEGUN_10 = 35
	MZ2_INFANTRY_MACHINEGUN_11 = 36
	MZ2_INFANTRY_MACHINEGUN_12 = 37
	MZ2_INFANTRY_MACHINEGUN_13 = 38

	MZ2_SOLDIER_BLASTER_1    = 39
	MZ2_SOLDIER_BLASTER_2    = 40
	MZ2_SOLDIER_SHOTGUN_1    = 41
	MZ2_SOLDIER_SHOTGUN_2    = 42
	MZ2_SOLDIER_MACHINEGUN_1 = 43
	MZ2_SOLDIER_MACHINEGUN_2 = 44

	MZ2_GUNNER_MACHINEGUN_1 = 45
	MZ2_GUNNER_MACHINEGUN_2 = 46
	MZ2_GUNNER_MACHINEGUN_3 = 47
	MZ2_GUNNER_MACHINEGUN_4 = 48
	MZ2_GUNNER_MACHINEGUN_5 = 49
	MZ2_GUNNER_MACHINEGUN_6 = 50
	MZ2_GUNNER_MACHINEGUN_7 = 51
	MZ2_GUNNER_MACHINEGUN_8 = 52
	MZ2_GUNNER_GRENADE_1    = 53
	MZ2_GUNNER_GRENADE_2    = 54
	MZ2_GUNNER_GRENADE_3    = 55
	MZ2_GUNNER_GRENADE_4    = 56

	MZ2_CHICK_ROCKET_1 = 57

	MZ2_FLYER_BLASTER_1 = 58
	MZ2_FLYER_BLASTER_2 = 59

	MZ2_MEDIC_BLASTER_1 = 60

	MZ2_GLADIATOR_RAILGUN_1 = 61

	MZ2_HOVER_BLASTER_1 = 62

	MZ2_ACTOR_MACHINEGUN_1 = 63

	MZ2_SUPERTANK_MACHINEGUN_1 = 64
	MZ2_SUPERTANK_MACHINEGUN_2 = 65
	MZ2_SUPERTANK_MACHINEGUN_3 = 66
	MZ2_SUPERTANK_MACHINEGUN_4 = 67
	MZ2_SUPERTANK_MACHINEGUN_5 = 68
	MZ2_SUPERTANK_MACHINEGUN_6 = 69
	MZ2_SUPERTANK_ROCKET_1     = 70
	MZ2_SUPERTANK_ROCKET_2     = 71
	MZ2_SUPERTANK_ROCKET_3     = 72

	MZ2_BOSS2_MACHINEGUN_L1 = 73
	MZ2_BOSS2_MACHINEGUN_L2 = 74
	MZ2_BOSS2_MACHINEGUN_L3 = 75
	MZ2_BOSS2_MACHINEGUN_L4 = 76
	MZ2_BOSS2_MACHINEGUN_L5 = 77
	MZ2_BOSS2_ROCKET_1      = 78
	MZ2_BOSS2_ROCKET_2      = 79
	MZ2_BOSS2_ROCKET_3      = 80
	MZ2_BOSS2_ROCKET_4      = 81

	MZ2_FLOAT_BLASTER_1 = 82

	MZ2_SOLDIER_BLASTER_3    = 83
	MZ2_SOLDIER_SHOTGUN_3    = 84
	MZ2_SOLDIER_MACHINEGUN_3 = 85
	MZ2_SOLDIER_BLASTER_4    = 86
	MZ2_SOLDIER_SHOTGUN_4    = 87
	MZ2_SOLDIER_MACHINEGUN_4 = 88
	MZ2_SOLDIER_BLASTER_5    = 89
	MZ2_SOLDIER_SHOTGUN_5    = 90
	MZ2_SOLDIER_MACHINEGUN_5 = 91
	MZ2_SOLDIER_BLASTER_6    = 92
	MZ2_SOLDIER_SHOTGUN_6    = 93
	MZ2_SOLDIER_MACHINEGUN_6 = 94
	MZ2_SOLDIER_BLASTER_7    = 95
	MZ2_SOLDIER_SHOTGUN_7    = 96
	MZ2_SOLDIER_MACHINEGUN_7 = 97
	MZ2_SOLDIER_BLASTER_8    = 98
	MZ2_SOLDIER_SHOTGUN_8    = 99
	MZ2_SOLDIER_MACHINEGUN_8 = 100

	MZ2_MAKRON_BFG          = 101
	MZ2_MAKRON_BLASTER_1    = 102
	MZ2_MAKRON_BLASTER_2    = 103
	MZ2_MAKRON_BLASTER_3    = 104
	MZ2_MAKRON_BLASTER_4    = 105
	MZ2_MAKRON_BLASTER_5    = 106
	MZ2_MAKRON_BLASTER_6    = 107
	MZ2_MAKRON_BLASTER_7    = 108
	MZ2_MAKRON_BLASTER_8    = 109
	MZ2_MAKRON_BLASTER_9    = 110
	MZ2_MAKRON_BLASTER_10   = 111
	MZ2_MAKRON_BLASTER_11   = 112
	MZ2_MAKRON_BLASTER_12   = 113
	MZ2_MAKRON_BLASTER_13   = 114
	MZ2_MAKRON_BLASTER_14   = 115
	MZ2_MAKRON_BLASTER_15   = 116
	MZ2_MAKRON_BLASTER_16   = 117
	MZ2_MAKRON_BLASTER_17   = 118
	MZ2_MAKRON_RAILGUN_1    = 119
	MZ2_JORG_MACHINEGUN_L1  = 120
	MZ2_JORG_MACHINEGUN_L2  = 121
	MZ2_JORG_MACHINEGUN_L3  = 122
	MZ2_JORG_MACHINEGUN_L4  = 123
	MZ2_JORG_MACHINEGUN_L5  = 124
	MZ2_JORG_MACHINEGUN_L6  = 125
	MZ2_JORG_MACHINEGUN_R1  = 126
	MZ2_JORG_MACHINEGUN_R2  = 127
	MZ2_JORG_MACHINEGUN_R3  = 128
	MZ2_JORG_MACHINEGUN_R4  = 129
	MZ2_JORG_MACHINEGUN_R5  = 130
	MZ2_JORG_MACHINEGUN_R6  = 131
	MZ2_JORG_BFG_1          = 132
	MZ2_BOSS2_MACHINEGUN_R1 = 133
	MZ2_BOSS2_MACHINEGUN_R2 = 134
	MZ2_BOSS2_MACHINEGUN_R3 = 135
	MZ2_BOSS2_MACHINEGUN_R4 = 136
	MZ2_BOSS2_MACHINEGUN_R5 = 137

	MZ2_CARRIER_MACHINEGUN_L1 = 138
	MZ2_CARRIER_MACHINEGUN_R1 = 139
	MZ2_CARRIER_GRENADE       = 140
	MZ2_TURRET_MACHINEGUN     = 141
	MZ2_TURRET_ROCKET         = 142
	MZ2_TURRET_BLASTER        = 143
	MZ2_STALKER_BLASTER       = 144
	MZ2_DAEDALUS_BLASTER      = 145
	MZ2_MEDIC_BLASTER_2       = 146
	MZ2_CARRIER_RAILGUN       = 147
	MZ2_WIDOW_DISRUPTOR       = 148
	MZ2_WIDOW_BLASTER         = 149
	MZ2_WIDOW_RAIL            = 150
	MZ2_WIDOW_PLASMABEAM      = 151
	MZ2_CARRIER_MACHINEGUN_L2 = 152
	MZ2_CARRIER_MACHINEGUN_R2 = 153
	MZ2_WIDOW_RAIL_LEFT       = 154
	MZ2_WIDOW_RAIL_RIGHT      = 155
	MZ2_WIDOW_BLASTER_SWEEP1  = 156
	MZ2_WIDOW_BLASTER_SWEEP2  = 157
	MZ2_WIDOW_BLASTER_SWEEP3  = 158
	MZ2_WIDOW_BLASTER_SWEEP4  = 159
	MZ2_WIDOW_BLASTER_SWEEP5  = 160
	MZ2_WIDOW_BLASTER_SWEEP6  = 161
	MZ2_WIDOW_BLASTER_SWEEP7  = 162
	MZ2_WIDOW_BLASTER_SWEEP8  = 163
	MZ2_WIDOW_BLASTER_SWEEP9  = 164
	MZ2_WIDOW_BLASTER_100     = 165
	MZ2_WIDOW_BLASTER_90      = 166
	MZ2_WIDOW_BLASTER_80      = 167
	MZ2_WIDOW_BLASTER_70      = 168
	MZ2_WIDOW_BLASTER_60      = 169
	MZ2_WIDOW_BLASTER_50      = 170
	MZ2_WIDOW_BLASTER_40      = 171
	MZ2_WIDOW_BLASTER_30      = 172
	MZ2_WIDOW_BLASTER_20      = 173
	MZ2_WIDOW_BLASTER_10      = 174
	MZ2_WIDOW_BLASTER_0       = 175
	MZ2_WIDOW_BLASTER_10L     = 176
	MZ2_WIDOW_BLASTER_20L     = 177
	MZ2_WIDOW_BLASTER_30L     = 178
	MZ2_WIDOW_BLASTER_40L     = 179
	MZ2_WIDOW_BLASTER_50L     = 180
	MZ2_WIDOW_BLASTER_60L     = 181
	MZ2_WIDOW_BLASTER_70L     = 182
	MZ2_WIDOW_RUN_1           = 183
	MZ2_WIDOW_RUN_2           = 184
	MZ2_WIDOW_RUN_3           = 185
	MZ2_WIDOW_RUN_4           = 186
	MZ2_WIDOW_RUN_5           = 187
	MZ2_WIDOW_RUN_6           = 188
	MZ2_WIDOW_RUN_7           = 189
	MZ2_WIDOW_RUN_8           = 190
	MZ2_CARRIER_ROCKET_1      = 191
	MZ2_CARRIER_ROCKET_2      = 192
	MZ2_CARRIER_ROCKET_3      = 193
	MZ2_CARRIER_ROCKET_4      = 194
	MZ2_WIDOW2_BEAMER_1       = 195
	MZ2_WIDOW2_BEAMER_2       = 196
	MZ2_WIDOW2_BEAMER_3       = 197
	MZ2_WIDOW2_BEAMER_4       = 198
	MZ2_WIDOW2_BEAMER_5       = 199
	MZ2_WIDOW2_BEAM_SWEEP_1   = 200
	MZ2_WIDOW2_BEAM_SWEEP_2   = 201
	MZ2_WIDOW2_BEAM_SWEEP_3   = 202
	MZ2_WIDOW2_BEAM_SWEEP_4   = 203
	MZ2_WIDOW2_BEAM_SWEEP_5   = 204
	MZ2_WIDOW2_BEAM_SWEEP_6   = 205
	MZ2_WIDOW2_BEAM_SWEEP_7   = 206
	MZ2_WIDOW2_BEAM_SWEEP_8   = 207
	MZ2_WIDOW2_BEAM_SWEEP_9   = 208
	MZ2_WIDOW2_BEAM_SWEEP_10  = 209
	MZ2_WIDOW2_BEAM_SWEEP_11  = 210
)

const (
	SPLASH_UNKNOWN     = 0
	SPLASH_SPARKS      = 1
	SPLASH_BLUE_WATER  = 2
	SPLASH_BROWN_WATER = 3
	SPLASH_SLIME       = 4
	SPLASH_LAVA        = 5
	SPLASH_BLOOD       = 6

	/* sound channels:
	 * channel 0 never willingly overrides
	 * other channels (1-7) allways override
	 * a playing sound on that channel */
	CHAN_AUTO   = 0
	CHAN_WEAPON = 1
	CHAN_VOICE  = 2
	CHAN_ITEM   = 3
	CHAN_BODY   = 4
	/* modifier flags */
	CHAN_NO_PHS_ADD = 8  /* send to all clients, not just ones in PHS (ATTN 0 will also do this) */
	CHAN_RELIABLE   = 16 /* send by reliable message, not datagram */

	/* sound attenuation values */
	ATTN_NONE   = 0 /* full volume the entire level */
	ATTN_NORM   = 1
	ATTN_IDLE   = 2
	ATTN_STATIC = 3 /* diminish very rapidly with distance */

	/* player_state->stats[] indexes */
	STAT_HEALTH_ICON   = 0
	STAT_HEALTH        = 1
	STAT_AMMO_ICON     = 2
	STAT_AMMO          = 3
	STAT_ARMOR_ICON    = 4
	STAT_ARMOR         = 5
	STAT_SELECTED_ICON = 6
	STAT_PICKUP_ICON   = 7
	STAT_PICKUP_STRING = 8
	STAT_TIMER_ICON    = 9
	STAT_TIMER         = 10
	STAT_HELPICON      = 11
	STAT_SELECTED_ITEM = 12
	STAT_LAYOUTS       = 13
	STAT_FRAGS         = 14
	STAT_FLASHES       = 15 /* cleared each frame, 1 = health, 2 = armor */
	STAT_CHASE         = 16
	STAT_SPECTATOR     = 17

	MAX_STATS = 32

	/* dmflags->value flags */
	DF_NO_HEALTH        = 0x00000001 /* 1 */
	DF_NO_ITEMS         = 0x00000002 /* 2 */
	DF_WEAPONS_STAY     = 0x00000004 /* 4 */
	DF_NO_FALLING       = 0x00000008 /* 8 */
	DF_INSTANT_ITEMS    = 0x00000010 /* 16 */
	DF_SAME_LEVEL       = 0x00000020 /* 32 */
	DF_SKINTEAMS        = 0x00000040 /* 64 */
	DF_MODELTEAMS       = 0x00000080 /* 128 */
	DF_NO_FRIENDLY_FIRE = 0x00000100 /* 256 */
	DF_SPAWN_FARTHEST   = 0x00000200 /* 512 */
	DF_FORCE_RESPAWN    = 0x00000400 /* 1024 */
	DF_NO_ARMOR         = 0x00000800 /* 2048 */
	DF_ALLOW_EXIT       = 0x00001000 /* 4096 */
	DF_INFINITE_AMMO    = 0x00002000 /* 8192 */
	DF_QUAD_DROP        = 0x00004000 /* 16384 */
	DF_FIXED_FOV        = 0x00008000 /* 32768 */
	DF_QUADFIRE_DROP    = 0x00010000 /* 65536 */
	DF_NO_MINES         = 0x00020000
	DF_NO_STACK_DOUBLE  = 0x00040000
	DF_NO_NUKES         = 0x00080000
	DF_NO_SPHERES       = 0x00100000

	ROGUE_VERSION_STRING = "08/21/1998 Beta 2 for Ensemble"
)

/*
 * ==========================================================
 *
 * ELEMENTS COMMUNICATED ACROSS THE NET
 *
 * ==========================================================
 */

//   ANGLE2SHORT(x) ((int)((x) * 65536 / 360) & 65535)
func SHORT2ANGLE(x int) float32 {
	return (float32(x) * (360.0 / 65536.0))
}

const (
	/* config strings are a general means of communication from
	 * the server to all connected clients. Each config string
	 * can be at most MAX_QPATH characters. */
	CS_NAME      = 0
	CS_CDTRACK   = 1
	CS_SKY       = 2
	CS_SKYAXIS   = 3 /* %f %f %f format */
	CS_SKYROTATE = 4
	CS_STATUSBAR = 5 /* display program string */

	CS_AIRACCEL    = 29 /* air acceleration control */
	CS_MAXCLIENTS  = 30
	CS_MAPCHECKSUM = 31 /* for catching cheater maps */

	CS_MODELS         = 32
	CS_SOUNDS         = (CS_MODELS + MAX_MODELS)
	CS_IMAGES         = (CS_SOUNDS + MAX_SOUNDS)
	CS_LIGHTS         = (CS_IMAGES + MAX_IMAGES)
	CS_ITEMS          = (CS_LIGHTS + MAX_LIGHTSTYLES)
	CS_PLAYERSKINS    = (CS_ITEMS + MAX_ITEMS)
	CS_GENERAL        = (CS_PLAYERSKINS + MAX_CLIENTS)
	MAX_CONFIGSTRINGS = (CS_GENERAL + MAX_GENERAL)
)

/* ============================================== */

/* entity_state_t->event values
 * entity events are for effects that take place reletive
 * to an existing entities origin.  Very network efficient.
 * All muzzle flashes really should be converted to events... */
const (
	EV_NONE            = 0
	EV_ITEM_RESPAWN    = 1
	EV_FOOTSTEP        = 2
	EV_FALLSHORT       = 3
	EV_FALL            = 4
	EV_FALLFAR         = 5
	EV_PLAYER_TELEPORT = 6
	EV_OTHER_TELEPORT  = 7
)

/* entity_state_t is the information conveyed from the server
 * in an update message about entities that the client will
 * need to render in some way */
type Entity_state_t struct {
	Number int /* edict index */

	Origin                                [3]float32
	Angles                                [3]float32
	Old_origin                            [3]float32 /* for lerping */
	Modelindex                            int
	Modelindex2, Modelindex3, Modelindex4 int /* weapons, CTF flags, etc */
	Frame                                 int
	Skinnum                               int
	Effects                               uint
	Renderfx                              int
	Solid                                 int /* for client side prediction, 8*(bits 0-4) is x/y radius */
	/* 8*(bits 5-9) is z down distance, 8(bits10-15) is z up */
	/* gi.linkentity sets this properly */
	Sound int /* for looping sounds, to guarantee shutoff */
	Event int /* impulse events -- muzzle flashes, footsteps, etc */
	/* events only go out for a single frame, they */
	/* are automatically cleared each frame */
}

func (T *Entity_state_t) Copy(other Entity_state_t) {
	T.Number = other.Number
	for i := 0; i < 3; i++ {
		T.Origin[i] = other.Origin[i]
		T.Angles[i] = other.Angles[i]
		T.Old_origin[i] = other.Old_origin[i]
	}
	T.Modelindex = other.Modelindex
	T.Modelindex2 = other.Modelindex2
	T.Modelindex3 = other.Modelindex3
	T.Modelindex4 = other.Modelindex4
	T.Frame = other.Frame
	T.Skinnum = other.Skinnum
	T.Effects = other.Effects
	T.Renderfx = other.Renderfx
	T.Solid = other.Solid
	T.Sound = other.Sound
	T.Event = other.Event
}

/* ============================================== */

/* player_state_t is the information needed in addition to pmove_state_t
 * to rendered a view.  There will only be 10 player_state_t sent each second,
 * but the number of pmove_state_t changes will be reletive to client
 * frame rates */
type Player_state_t struct {
	Pmove Pmove_state_t /* for prediction */

	Viewangles  [3]float32 /* for fixed views */
	Viewoffset  [3]float32 /* add to pmovestate->origin */
	Kick_angles [3]float32 /* add to view direction to get render angles */
	/* set by weapon kicks, pain effects, etc */

	Gunangles [3]float32
	Gunoffset [3]float32
	Gunindex  int
	Gunframe  int

	Blend   [4]float32 /* rgba full screen effect */
	Fov     float32    /* horizontal field of view */
	Rdflags int        /* refdef flags */

	Stats [MAX_STATS]int16 /* fast status bar updates */
}

func (T *Player_state_t) Copy(other Player_state_t) {
	T.Pmove.Copy(other.Pmove)
	for i := 0; i < 3; i++ {
		T.Viewangles[i] = other.Viewangles[i]
		T.Viewoffset[i] = other.Viewoffset[i]
		T.Kick_angles[i] = other.Kick_angles[i]
		T.Gunangles[i] = other.Gunangles[i]
		T.Gunoffset[i] = other.Gunoffset[i]
	}
	T.Gunindex = other.Gunindex
	T.Gunframe = other.Gunframe
	for i := 0; i < 4; i++ {
		T.Blend[i] = other.Blend[i]
	}
	T.Fov = other.Fov
	T.Rdflags = other.Rdflags
	for i := 0; i < MAX_STATS; i++ {
		T.Stats[i] = other.Stats[i]
	}
}

func (T *CvarT) Bool() bool {
	v, e := strconv.ParseFloat(T.String, 32)
	if e == nil && v != 0.0 {
		return true
	}
	return false
}

func (T *CvarT) Int() int {
	v, e := strconv.ParseFloat(T.String, 32)
	if e == nil {
		return int(v)
	}
	return 0
}

func (T *CvarT) Float() float32 {
	v, e := strconv.ParseFloat(T.String, 32)
	if e == nil {
		return float32(v)
	}
	return 0.0
}

func DEG2RAD(a float32) float32 {
	return (a * math.Pi) / 180.0
}

/* ============================================================================ */

func RotatePointAroundVector(dst, dir, point []float32, degrees float32) {

	vf := []float32{dir[0], dir[1], dir[2]}

	vr := perpendicularVector(dir)
	vup := make([]float32, 3)
	CrossProduct(vr, vf, vup)

	var m [3][3]float32
	m[0][0] = vr[0]
	m[1][0] = vr[1]
	m[2][0] = vr[2]

	m[0][1] = vup[0]
	m[1][1] = vup[1]
	m[2][1] = vup[2]

	m[0][2] = vf[0]
	m[1][2] = vf[1]
	m[2][2] = vf[2]

	var im [3][3]float32
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			im[i][j] = m[i][j]
		}
	}

	im[0][1] = m[1][0]
	im[0][2] = m[2][0]
	im[1][0] = m[0][1]
	im[1][2] = m[2][1]
	im[2][0] = m[0][2]
	im[2][1] = m[1][2]

	var zrot [3][3]float32
	zrot[2][2] = 1.0

	zrot[0][0] = float32(math.Cos(float64(DEG2RAD(degrees))))
	zrot[0][1] = float32(math.Sin(float64(DEG2RAD(degrees))))
	zrot[1][0] = float32(-math.Sin(float64(DEG2RAD(degrees))))
	zrot[1][1] = float32(math.Cos(float64(DEG2RAD(degrees))))

	var tmpmat [3][3]float32
	R_ConcatRotations(m, zrot, tmpmat)
	var rot [3][3]float32
	R_ConcatRotations(tmpmat, im, rot)

	for i := 0; i < 3; i++ {
		dst[i] = rot[i][0]*point[0] + rot[i][1]*point[1] + rot[i][2]*
			point[2]
	}
}

func AngleVectors(angles, forward, right, up []float32) {

	angle := float64(angles[YAW]) * (math.Pi * 2 / 360)
	sy := float32(math.Sin(angle))
	cy := float32(math.Cos(angle))
	angle = float64(angles[PITCH]) * (math.Pi * 2 / 360)
	sp := float32(math.Sin(angle))
	cp := float32(math.Cos(angle))
	angle = float64(angles[ROLL]) * (math.Pi * 2 / 360)
	sr := float32(math.Sin(angle))
	cr := float32(math.Cos(angle))

	if forward != nil {
		forward[0] = cp * cy
		forward[1] = cp * sy
		forward[2] = -sp
	}

	if right != nil {
		right[0] = (-1*sr*sp*cy + -1*cr*-sy)
		right[1] = (-1*sr*sp*sy + -1*cr*cy)
		right[2] = -1 * sr * cp
	}

	if up != nil {
		up[0] = (cr*sp*cy + -sr*-sy)
		up[1] = (cr*sp*sy + -sr*cy)
		up[2] = cr * cp
	}
}

func projectPointOnPlane(p, normal []float32) []float32 {
	// float d;
	// vec3_t n;
	// float inv_denom;

	inv_denom := 1.0 / DotProduct(normal, normal)

	d := DotProduct(normal, p) * inv_denom

	n := []float32{normal[0] * inv_denom, normal[1] * inv_denom, normal[2] * inv_denom}

	return []float32{p[0] - d*n[0], p[1] - d*n[1], p[2] - d*n[2]}
}

/* assumes "src" is normalized */
func perpendicularVector(src []float32) []float32 {

	/* find the smallest magnitude axially aligned vector */
	pos := 0
	minelem := 1.0
	for i := 0; i < 3; i++ {
		if math.Abs(float64(src[i])) < minelem {
			pos = i
			minelem = math.Abs(float64(src[i]))
		}
	}

	tempvec := []float32{0, 0, 0}
	tempvec[pos] = 1.0

	/* project the point onto the plane defined by src */
	dst := projectPointOnPlane(tempvec, src)

	/* normalize the result */
	VectorNormalize(dst)
	return dst
}

func R_ConcatRotations(in1, in2, out [3][3]float32) {
	out[0][0] = in1[0][0]*in2[0][0] + in1[0][1]*in2[1][0] +
		in1[0][2]*in2[2][0]
	out[0][1] = in1[0][0]*in2[0][1] + in1[0][1]*in2[1][1] +
		in1[0][2]*in2[2][1]
	out[0][2] = in1[0][0]*in2[0][2] + in1[0][1]*in2[1][2] +
		in1[0][2]*in2[2][2]
	out[1][0] = in1[1][0]*in2[0][0] + in1[1][1]*in2[1][0] +
		in1[1][2]*in2[2][0]
	out[1][1] = in1[1][0]*in2[0][1] + in1[1][1]*in2[1][1] +
		in1[1][2]*in2[2][1]
	out[1][2] = in1[1][0]*in2[0][2] + in1[1][1]*in2[1][2] +
		in1[1][2]*in2[2][2]
	out[2][0] = in1[2][0]*in2[0][0] + in1[2][1]*in2[1][0] +
		in1[2][2]*in2[2][0]
	out[2][1] = in1[2][0]*in2[0][1] + in1[2][1]*in2[1][1] +
		in1[2][2]*in2[2][1]
	out[2][2] = in1[2][0]*in2[0][2] + in1[2][1]*in2[1][2] +
		in1[2][2]*in2[2][2]
}

func LerpAngle(a2, a1, frac float32) float32 {
	if a1-a2 > 180 {
		a1 -= 360
	}

	if a1-a2 < -180 {
		a1 += 360
	}

	return a2 + frac*(a1-a2)
}

/*
 * Returns 1, 2, or 1 + 2
 */
func BoxOnPlaneSide(emins, emaxs []float32, p *Cplane_t) int {
	//  float dist1, dist2;
	//  int sides;

	/* fast axial cases */
	if p.Type < 3 {
		if p.Dist <= emins[p.Type] {
			return 1
		}

		if p.Dist >= emaxs[p.Type] {
			return 2
		}

		return 3
	}

	/* general case */
	var dist1, dist2 float32
	switch p.Signbits {
	case 0:
		dist1 = p.Normal[0]*emaxs[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emaxs[2]
		dist2 = p.Normal[0]*emins[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emins[2]
		break
	case 1:
		dist1 = p.Normal[0]*emins[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emaxs[2]
		dist2 = p.Normal[0]*emaxs[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emins[2]
		break
	case 2:
		dist1 = p.Normal[0]*emaxs[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emaxs[2]
		dist2 = p.Normal[0]*emins[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emins[2]
		break
	case 3:
		dist1 = p.Normal[0]*emins[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emaxs[2]
		dist2 = p.Normal[0]*emaxs[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emins[2]
		break
	case 4:
		dist1 = p.Normal[0]*emaxs[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emins[2]
		dist2 = p.Normal[0]*emins[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emaxs[2]
		break
	case 5:
		dist1 = p.Normal[0]*emins[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emins[2]
		dist2 = p.Normal[0]*emaxs[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emaxs[2]
		break
	case 6:
		dist1 = p.Normal[0]*emaxs[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emins[2]
		dist2 = p.Normal[0]*emins[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emaxs[2]
		break
	case 7:
		dist1 = p.Normal[0]*emins[0] + p.Normal[1]*emins[1] +
			p.Normal[2]*emins[2]
		dist2 = p.Normal[0]*emaxs[0] + p.Normal[1]*emaxs[1] +
			p.Normal[2]*emaxs[2]
		break
	default:
		break
	}

	sides := 0

	if dist1 >= p.Dist {
		sides = 1
	}

	if dist2 < p.Dist {
		sides |= 2
	}

	return sides
}

func VectorNormalize(v []float32) float32 {

	dlength := float64(v[0])*float64(v[0]) + float64(v[1])*float64(v[1]) + float64(v[2])*float64(v[2])
	length := float32(math.Sqrt(dlength))

	if length != 0.0 {
		ilength := 1 / length
		v[0] *= ilength
		v[1] *= ilength
		v[2] *= ilength
	}

	return length
}

func VectorAdd(veca, vecb, out []float32) {
	out[0] = veca[0] + vecb[0]
	out[1] = veca[1] + vecb[1]
	out[2] = veca[2] + vecb[2]
}

func VectorSubtract(veca, vecb, out []float32) {
	out[0] = veca[0] - vecb[0]
	out[1] = veca[1] - vecb[1]
	out[2] = veca[2] - vecb[2]
}

func VectorScaled(in []float32, scale float32) []float32 {
	return []float32{
		in[0] * scale,
		in[1] * scale,
		in[2] * scale}
}

func VectorLength(v []float32) float32 {

	var length float32 = 0

	for i := 0; i < 3; i++ {
		length += v[i] * v[i]
	}

	return float32(math.Sqrt(float64(length)))
}

func DotProduct(v1, v2 []float32) float32 {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
}

func CrossProduct(v1, v2, cross []float32) {
	cross[0] = v1[1]*v2[2] - v1[2]*v2[1]
	cross[1] = v1[2]*v2[0] - v1[0]*v2[2]
	cross[2] = v1[0]*v2[1] - v1[1]*v2[0]
}

/*
 * Parse a token out of a string
 */
func COM_Parse(data string, index int) (string, int) {

	if index < 0 || index >= len(data) {
		return "", -1
	}

	skipwhite := true

	for skipwhite {
		skipwhite = false
		for index < len(data) && data[index] <= ' ' {
			index++
		}
		if index >= len(data) {
			return "", -1
		}

		/* skip // comments */
		if (data[index] == '/') && (data[index+1] == '/') {
			for index < len(data) && data[index] != '\n' {
				index++
			}
			skipwhite = true
		}
	}

	var token strings.Builder

	/* handle quoted strings specially */
	if data[index] == '"' {
		index++

		for {
			if index >= len(data) || data[index] == '"' {
				return token.String(), index + 1
			}
			token.WriteByte(data[index])
			index++
		}
	}

	/* parse a regular word */
	for {
		if index >= len(data) || data[index] <= ' ' {
			return token.String(), index
		}
		token.WriteByte(data[index])
		index++
	}
}

/*
 * =====================================================================
 *
 * INFO STRINGS
 *
 * =====================================================================
 */

/*
 * Searches the string for the given
 * key and returns the associated value,
 * or an empty string.
 */
func Info_ValueForKey(s, key string) string {

	split := strings.Split(s, "\\")
	index := 0
	for index < len(split)-1 {
		if split[index] == key {
			return split[index+1]
		}
		index += 2
	}

	return ""
}

/*
 * Generate a pseudorandom
 * integer >0.
 */
func Randk() int {
	return int(rand.Uint32())
}

type QCommon interface {
	Init() error
	IsDedicated() bool
	SetServerState(state int)
	ServerState() int
	Curtime() int
	Sys_Milliseconds() int
	QPort() int

	Com_VPrintf(print_level int, format string, a ...interface{})
	Com_Printf(format string, a ...interface{})
	Com_DPrintf(format string, a ...interface{})
	Com_Error(code int, format string, a ...interface{}) error

	Cvar_Get(var_name, var_value string, flags int) *CvarT
	Cvar_Set(var_name, value string) *CvarT
	Cvar_FullSet(var_name, value string, flags int) *CvarT
	Cvar_VariableBool(var_name string) bool
	Cvar_VariableInt(var_name string) int
	Cvar_VariableString(var_name string) string
	Cvar_Userinfo() string
	Cvar_ClearUserinfoModified()

	Cbuf_AddText(text string)
	Cbuf_Execute() error
	Cmd_AddCommand(cmd_name string, function func([]string, interface{}) error, arg interface{})
	Cmd_TokenizeString(text string, macroExpand bool) []string

	Netchan_OutOfBandPrint(net_socket Netsrc_t, adr Netadr_t, format string, a ...interface{}) error

	NET_GetPacket(sock Netsrc_t) (*Netadr_t, []byte)
	NET_SendPacket(sock Netsrc_t, data []byte, to Netadr_t) error

	FS_FOpenFile(name string, gamedir_only bool) (QFileHandle, error)
	LoadFile(path string) ([]byte, error)
}

type QClient interface {
	Init(common QCommon) error
	Frame(packetdelta, renderdelta, timedelta int, packetframe, renderframe bool) error
	IsVSyncActive() bool
	GetRefreshRate() int
}

type QServer interface {
	Init(common QCommon) error
	Frame(usec int) error
}
