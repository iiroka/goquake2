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
 * Main header file for the game module.
 *
 * =======================================================================
 */
package game

import "goquake2/shared"

const (
	/* ================================================================== */

	/* view pitching times */
	DAMAGE_TIME = 0.5
	FALL_TIME   = 0.3

	/* these are set with checkboxes on each entity in the map editor */
	SPAWNFLAG_NOT_EASY       = 0x00000100
	SPAWNFLAG_NOT_MEDIUM     = 0x00000200
	SPAWNFLAG_NOT_HARD       = 0x00000400
	SPAWNFLAG_NOT_DEATHMATCH = 0x00000800
	SPAWNFLAG_NOT_COOP       = 0x00001000

	FL_FLY           = 0x00000001
	FL_SWIM          = 0x00000002 /* implied immunity to drowining */
	FL_IMMUNE_LASER  = 0x00000004
	FL_INWATER       = 0x00000008
	FL_GODMODE       = 0x00000010
	FL_NOTARGET      = 0x00000020
	FL_IMMUNE_SLIME  = 0x00000040
	FL_IMMUNE_LAVA   = 0x00000080
	FL_PARTIALGROUND = 0x00000100 /* not all corners are valid */
	FL_WATERJUMP     = 0x00000200 /* player jumping out of water */
	FL_TEAMSLAVE     = 0x00000400 /* not the first on the team */
	FL_NO_KNOCKBACK  = 0x00000800
	FL_POWER_ARMOR   = 0x00001000 /* power armor (if any) is active */
	FL_COOP_TAKEN    = 0x00002000 /* Another client has already taken it */
	FL_RESPAWN       = 0x80000000 /* used for item respawning */

	FRAMETIME = 0.1

	/* memory tags to allow dynamic memory to be cleaned up */
	TAG_GAME  = 765 /* clear when unloading the dll */
	TAG_LEVEL = 766 /* clear when loading a new level */

	MELEE_DISTANCE  = 80
	BODY_QUEUE_SIZE = 8

	AMMO_BULLETS  = 0
	AMMO_SHELLS   = 1
	AMMO_ROCKETS  = 2
	AMMO_GRENADES = 3
	AMMO_CELLS    = 4
	AMMO_SLUGS    = 5

	/* deadflag */
	DEAD_NO          = 0
	DEAD_DYING       = 1
	DEAD_DEAD        = 2
	DEAD_RESPAWNABLE = 3

	/* range */
	RANGE_MELEE = 0
	RANGE_NEAR  = 1
	RANGE_MID   = 2
	RANGE_FAR   = 3

	/* monster ai flags */
	AI_STAND_GROUND      = 0x00000001
	AI_TEMP_STAND_GROUND = 0x00000002
	AI_SOUND_TARGET      = 0x00000004
	AI_LOST_SIGHT        = 0x00000008
	AI_PURSUIT_LAST_SEEN = 0x00000010
	AI_PURSUE_NEXT       = 0x00000020
	AI_PURSUE_TEMP       = 0x00000040
	AI_HOLD_FRAME        = 0x00000080
	AI_GOOD_GUY          = 0x00000100
	AI_BRUTAL            = 0x00000200
	AI_NOSTEP            = 0x00000400
	AI_DUCKED            = 0x00000800
	AI_COMBAT_POINT      = 0x00001000
	AI_MEDIC             = 0x00002000
	AI_RESURRECTING      = 0x00004000

	/* armor types */
	ARMOR_NONE   = 0
	ARMOR_JACKET = 1
	ARMOR_COMBAT = 2
	ARMOR_BODY   = 3
	ARMOR_SHARD  = 4
)

type weaponstate_t int

const (
	WEAPON_READY      weaponstate_t = 0
	WEAPON_ACTIVATING weaponstate_t = 1
	WEAPON_DROPPING   weaponstate_t = 2
	WEAPON_FIRING     weaponstate_t = 3
)

/* edict->movetype values */
type movetype_t int

const (
	MOVETYPE_NONE   movetype_t = 0 /* never moves */
	MOVETYPE_NOCLIP movetype_t = 1 /* origin and angles change with no interaction */
	MOVETYPE_PUSH   movetype_t = 2 /* no clip to world, push on box contact */
	MOVETYPE_STOP   movetype_t = 3 /* no clip to world, stops on box contact */

	MOVETYPE_WALK       movetype_t = 4 /* gravity */
	MOVETYPE_STEP       movetype_t = 5 /* gravity, special edge handling */
	MOVETYPE_FLY        movetype_t = 6
	MOVETYPE_TOSS       movetype_t = 7 /* gravity */
	MOVETYPE_FLYMISSILE movetype_t = 8 /* extra size to monsters */
	MOVETYPE_BOUNCE     movetype_t = 9
)

type gitem_armor_t struct {
	base_count        int
	max_count         int
	normal_protection float32
	energy_protection float32
	armor             int
}

const (
	IT_WEAPON      = 1 /* use makes active weapon */
	IT_AMMO        = 2
	IT_ARMOR       = 4
	IT_STAY_COOP   = 8
	IT_KEY         = 16
	IT_POWERUP     = 32
	IT_INSTANT_USE = 64 /* item is insta-used on pickup if dmflag is set */

	/* gitem_t->weapmodel for weapons indicates model index */
	WEAP_BLASTER         = 1
	WEAP_SHOTGUN         = 2
	WEAP_SUPERSHOTGUN    = 3
	WEAP_MACHINEGUN      = 4
	WEAP_CHAINGUN        = 5
	WEAP_GRENADES        = 6
	WEAP_GRENADELAUNCHER = 7
	WEAP_ROCKETLAUNCHER  = 8
	WEAP_HYPERBLASTER    = 9
	WEAP_RAILGUN         = 10
	WEAP_BFG             = 11
)

type gitem_t struct {
	classname         string /* spawning name */
	pickup            func(ent, other *edict_t, G *qGame) bool
	use               func(ent *edict_t, item *gitem_t, G *qGame)
	drop              func(ent *edict_t, item *gitem_t, G *qGame)
	weaponthink       func(ent *edict_t, G *qGame)
	pickup_sound      string
	world_model       string
	world_model_flags int
	view_model        string

	/* client side info */
	icon        string
	pickup_name string /* for printing on pickup */
	count_width int    /* number of digits to display by icon */

	quantity int    /* for ammo how much, for weapons how much is used per shot */
	ammo     string /* for weapons */
	flags    int    /* IT_* flags */

	weapmodel int /* weapon model index (for weapons) */

	info interface{}
	tag  int

	precaches string /* string of all models, sounds, and images this item will use */
}

/* this structure is left intact through an entire game
   it should be initialized at dll load time, and read/written to
   the server.ssv file for savegames */
type game_locals_t struct {
	helpmessage1 string
	helpmessage2 string
	helpchanged  int /* flash F1 icon if non 0, play sound
	   and increment only if 1, 2, or 3 */

	clients []gclient_t /* [maxclients] */

	/* can't store spawnpoint in level, because
	it would get overwritten by the savegame
	restore */
	spawnpoint string /* needed for coop respawns */

	/* store latched cvars here that we want to get at often */
	maxclients  int
	maxentities int

	/* cross level triggers */
	serverflags int

	/* items */
	num_items int

	autosaved bool
}

/* this structure is cleared as each map is entered
   it is read/written to the level.sav file for savegames */
type level_locals_t struct {
	framenum int
	time     float32

	level_name string /* the descriptive name (Outer Base, etc) */
	mapname    string /* the server name (base1, etc) */
	nextmap    string /* go here when fraglimit is hit */

	/* intermission state */
	intermissiontime float32 /* time the intermission was started */
	//    char *changemap;
	//    int exitintermission;
	//    vec3_t intermission_origin;
	//    vec3_t intermission_angle;

	sight_client *edict_t /* changed once each frame for coop games */

	sight_entity          *edict_t
	sight_entity_framenum int
	sound_entity          *edict_t
	sound_entity_framenum int
	//    edict_t *sound2_entity;
	//    int sound2_entity_framenum;

	pic_health int

	//    int total_secrets;
	//    int found_secrets;

	//    int total_goals;
	//    int found_goals;

	total_monsters  int
	killed_monsters int

	current_entity *edict_t /* entity running from G_RunFrame */
	//    int body_que; /* dead bodies */

	//    int power_cubes; /* ugly necessity for coop */
}

/* spawn_temp_t is only used to hold entity field values that
   can be set from the editor, but aren't actualy present
   in edict_t during gameplay */
type spawn_temp_t struct {
	/* world vars */
	Sky       string
	Skyrotate float32
	Skyaxis   [3]float32
	Nextmap   string

	Lip int
	//    int distance;
	//    int height;
	Noise     string
	pausetime float32
	//    char *item;
	Ggravity string

	//    float minyaw;
	//    float maxyaw;
	//    float minpitch;
	//    float maxpitch;
}

type mframe_t struct {
	aifunc    func(self *edict_t, dist float32, G *qGame)
	dist      float32
	thinkfunc func(self *edict_t, G *qGame)
}

type mmove_t struct {
	firstframe int
	lastframe  int
	frame      []mframe_t
	endfunc    func(self *edict_t, G *qGame)
}

type monsterinfo_t struct {
	currentmove *mmove_t
	aiflags     int
	nextframe   int
	scale       float32

	stand  func(self *edict_t, G *qGame)
	idle   func(self *edict_t, G *qGame)
	search func(self *edict_t, G *qGame)
	walk   func(self *edict_t, G *qGame)
	// void (*run)(edict_t *self);
	// void (*dodge)(edict_t *self, edict_t *other, float eta);
	// void (*attack)(edict_t *self);
	// void (*melee)(edict_t *self);
	// void (*sight)(edict_t *self, edict_t *other);
	// qboolean (*checkattack)(edict_t *self);

	pausetime float32
	// float attack_finished;

	// vec3_t saved_goal;
	// float search_time;
	// float trail_time;
	// vec3_t last_sighting;
	// int attack_state;
	// int lefty;
	idle_time float32
	linkcount int

	// int power_armor_type;
	// int power_armor_power;
}

const (
	/* Easier handling of AI skill levels */
	SKILL_EASY     = 0
	SKILL_MEDIUM   = 1
	SKILL_HARD     = 2
	SKILL_HARDPLUS = 3
)

/* ============================================================================ */

const (
	/* client_t->anim_priority */
	ANIM_BASIC   = 0 /* stand / run */
	ANIM_WAVE    = 1
	ANIM_JUMP    = 2
	ANIM_PAIN    = 3
	ANIM_ATTACK  = 4
	ANIM_DEATH   = 5
	ANIM_REVERSE = 6
)

/* client data that stays across multiple level loads */
type client_persistant_t struct {
	userinfo string
	netname  string
	hand     int

	connected bool /* a loadgame will leave valid entities that
	   just don't have a connection yet */

	/* values saved and restored
	   from edicts when changing levels */
	health     int
	max_health int
	savedFlags int

	selected_item int
	inventory     [shared.MAX_ITEMS]int

	/* ammo capacities */
	max_bullets  int
	max_shells   int
	max_rockets  int
	max_grenades int
	max_cells    int
	max_slugs    int

	weapon     *gitem_t
	lastweapon *gitem_t

	// int power_cubes /* used for tracking the cubes in coop games */
	score int /* for calculating total unit score in coop games */

	// int game_helpchanged
	// int helpchanged

	spectator bool /* client is a spectator */
}

func (G *client_persistant_t) copy(other client_persistant_t) {
	G.userinfo = other.userinfo
	G.netname = other.netname
	G.hand = other.hand
	G.connected = other.connected
	G.health = other.health
	G.max_health = other.max_health
	G.savedFlags = other.savedFlags

	G.selected_item = other.selected_item
	for i := range G.inventory {
		G.inventory[i] = other.inventory[i]
	}
	G.max_bullets = other.max_bullets
	G.max_shells = other.max_shells
	G.max_rockets = other.max_rockets
	G.max_grenades = other.max_grenades
	G.max_cells = other.max_cells
	G.max_slugs = other.max_slugs
	G.weapon = other.weapon
	G.lastweapon = other.lastweapon
	// G.power_cubes = other.power_cubes
	G.score = other.score
	// int game_helpchanged
	// int helpchanged
	G.spectator = other.spectator
}

/* client data that stays across deathmatch respawns */
type client_respawn_t struct {
	coop_respawn client_persistant_t /* what to set client->pers to on a respawn */
	enterframe   int                 /* level.framenum the client entered the game */
	score        int                 /* frags, etc */
	cmd_angles   [3]float32          /* angles sent over in the last command */

	spectator bool /* client is a spectator */
}

func (R *client_respawn_t) copy(other client_respawn_t) {
	R.coop_respawn.copy(other.coop_respawn)
	R.enterframe = other.enterframe
	R.score = other.score
	for i := range R.cmd_angles {
		R.cmd_angles[i] = other.cmd_angles[i]
	}
	R.spectator = other.spectator
}

/* this structure is cleared on each PutClientInServer(),
   except for 'client->pers' */
type gclient_t struct {
	/* known to server */
	ps   shared.Player_state_t /* communicated by server to clients */
	ping int

	/* private to game */
	pers      client_persistant_t
	resp      client_respawn_t
	old_pmove shared.Pmove_state_t /* for detecting out-of-pmove changes */

	// qboolean showscores; /* set layout stat */
	// qboolean showinventory; /* set layout stat */
	// qboolean showhelp;
	// qboolean showhelpicon;

	ammo_index int

	buttons         int
	oldbuttons      int
	latched_buttons int

	weapon_thunk bool

	newweapon *gitem_t

	/* sum up damage over an entire frame, so
	   shotgun blasts give a single big kick */
	// int damage_armor; /* damage absorbed by armor */
	// int damage_parmor; /* damage absorbed by power armor */
	// int damage_blood; /* damage taken out of health */
	// int damage_knockback; /* impact damage */
	// vec3_t damage_from; /* origin for vector calculation */

	// float killer_yaw; /* when dead, look at killer */

	weaponstate                         weaponstate_t
	kick_angles                         [3]float32 /* weapon kicks */
	kick_origin                         [3]float32
	v_dmg_roll, v_dmg_pitch, v_dmg_time float32 /* damage kicks */
	fall_time, fall_value               float32 /* for view drop on fall */
	damage_alpha                        float32
	bonus_alpha                         float32
	damage_blend                        [3]float32
	v_angle                             [3]float32 /* aiming direction */
	bobtime                             float32    /* so off-ground doesn't change it */
	oldviewangles                       [3]float32
	oldvelocity                         [3]float32

	// float next_drown_time;
	// int old_waterlevel;
	// int breather_sound;

	machinegun_shots int /* for weapon raising */

	/* animation vars */
	anim_end      int
	anim_priority int
	// qboolean anim_duck;
	// qboolean anim_run;

	// /* powerup timers */
	// float quad_framenum;
	// float invincible_framenum;
	// float breather_framenum;
	// float enviro_framenum;

	// qboolean grenade_blew_up;
	// float grenade_time;
	// int silencer_shots;
	// int weapon_sound;

	pickup_msg_time float32

	// float flood_locktill; /* locked from talking */
	// float flood_when[10]; /* when messages were said */
	// int flood_whenhead; /* head pointer for when said */

	// float respawn_time; /* can respawn when time > this */

	chase_target *edict_t /* player we are chasing */
	// qboolean update_chase; /* need to update chase info? */
}

func (G *gclient_t) Ps() *shared.Player_state_t {
	return &G.ps
}

func (G *gclient_t) Ping() int {
	return G.ping
}

func (G *gclient_t) copy(other gclient_t) {
	/* known to server */
	G.ps.Copy(other.ps)
	G.ping = other.ping
	G.pers.copy(other.pers)
	// resp client_respawn_t
	G.old_pmove.Copy(other.old_pmove) /* for detecting out-of-pmove changes */
	// qboolean showscores; /* set layout stat */
	// qboolean showinventory; /* set layout stat */
	// qboolean showhelp;
	// qboolean showhelpicon;
	G.ammo_index = other.ammo_index
	G.buttons = other.buttons
	G.oldbuttons = other.oldbuttons
	G.latched_buttons = other.latched_buttons
	G.weapon_thunk = other.weapon_thunk
	G.newweapon = other.newweapon
	// int damage_armor; /* damage absorbed by armor */
	// int damage_parmor; /* damage absorbed by power armor */
	// int damage_blood; /* damage taken out of health */
	// int damage_knockback; /* impact damage */
	// vec3_t damage_from; /* origin for vector calculation */
	// float killer_yaw; /* when dead, look at killer */
	// weaponstate_t weaponstate;
	// float v_dmg_roll, v_dmg_pitch, v_dmg_time; /* damage kicks */
	// float fall_time, fall_value; /* for view drop on fall */
	// float damage_alpha;
	// float bonus_alpha;
	// vec3_t damage_blend;
	G.bobtime = other.bobtime
	// float next_drown_time;
	// int old_waterlevel;
	// int breather_sound;
	G.machinegun_shots = other.machinegun_shots
	G.anim_end = other.anim_end
	G.anim_priority = other.anim_priority
	// qboolean anim_duck;
	// qboolean anim_run;
	// float quad_framenum;
	// float invincible_framenum;
	// float breather_framenum;
	// float enviro_framenum;
	// qboolean grenade_blew_up;
	// float grenade_time;
	// int silencer_shots;
	// int weapon_sound;
	// float pickup_msg_time;
	// float flood_locktill; /* locked from talking */
	// float flood_when[10]; /* when messages were said */
	// int flood_whenhead; /* head pointer for when said */
	// float respawn_time; /* can respawn when time > this */
	// edict_t *chase_target; /* player we are chasing */
	// qboolean update_chase; /* need to update chase info? */

	for i := 0; i < 3; i++ {
		G.kick_angles[i] = other.kick_angles[i]
		G.kick_origin[i] = other.kick_origin[i]
		G.v_angle[i] = other.v_angle[i]
		G.oldviewangles[i] = other.oldviewangles[i]
		G.oldvelocity[i] = other.oldviewangles[i]
	}
}

type edict_t struct {
	index  int
	s      shared.Entity_state_t
	client *gclient_t /* NULL if not a player
	   the server expects the first part
	   of gclient_s to be a player_state_t
	   but the rest of it is opaque */

	inuse     bool
	linkcount int

	area shared.Link_t /* linked to a division node or leaf */

	num_clusters      int /* if -1, use headnode instead */
	clusternums       [shared.MAX_ENT_CLUSTERS]int
	headnode          int /* unused if num_clusters != -1 */
	areanum, areanum2 int

	/* ================================ */

	svflags              int
	mins, maxs           [3]float32
	absmin, absmax, size [3]float32
	solid                shared.Solid_t
	// int clipmask;
	owner *edict_t

	// /* DO NOT MODIFY ANYTHING ABOVE THIS, THE SERVER */
	// /* EXPECTS THE FIELDS IN THAT ORDER! */

	/* ================================ */
	movetype movetype_t
	flags    int

	Model    string
	freetime float32 /* sv.time when the object was freed */

	/* only used locally in game, not by server */
	Message    string
	Classname  string
	Spawnflags int

	ftimestamp float32

	// float angle; /* set in qe3, -1 = up, -2 = down */
	Target       string
	Targetname   string
	Killtarget   string
	Team         string
	Pathtarget   string
	Deathtarget  string
	Combattarget string
	// edict_t *target_ent;

	Speed, Accel, Decel float32
	// vec3_t movedir;
	// vec3_t pos1, pos2;

	velocity  [3]float32
	avelocity [3]float32
	Mass      int
	// float air_finished;
	gravity float32 /* per entity gravity multiplier (1.0 is normal)
	   use for lowgrav artifact, flares */

	goalentity *edict_t
	movetarget *edict_t
	yaw_speed  float32
	ideal_yaw  float32

	nextthink float32
	prethink  func(self *edict_t, G *qGame)
	think     func(self *edict_t, G *qGame)
	// void (*blocked)(edict_t *self, edict_t *other);
	touch func(self, other *edict_t, plane *shared.Cplane_t, surf *shared.Csurface_t, G *qGame)
	use   func(self, other, activator *edict_t, G *qGame)
	// void (*pain)(edict_t *self, edict_t *other, float kick, int damage);
	// void (*die)(edict_t *self, edict_t *inflictor, edict_t *attacker,
	// 		int damage, vec3_t point);

	touch_debounce_time float32
	// float pain_debounce_time;
	// float damage_debounce_time;
	// float fly_sound_debounce_time;	/* now also used by insane marines to store pain sound timeout */
	// float last_move_time;

	Health     int
	max_health int
	gib_health int
	deadflag   int

	// float show_hostile;
	// float powerarmor_time;

	Map string /* target_changelevel */

	viewheight int /* height above origin where eyesight is determined */
	// int takedamage;
	Dmg int
	// int radius_dmg;
	// float dmg_radius;
	Sounds int /* make this a spawntemp var? */
	count  int

	chain                  *edict_t
	enemy                  *edict_t
	oldenemy               *edict_t
	activator              *edict_t
	groundentity           *edict_t
	groundentity_linkcount int
	// edict_t *teamchain;
	// edict_t *teammaster;

	// edict_t *mynoise; /* can go in client only */
	// edict_t *mynoise2;

	noise_index  int
	noise_index2 int
	Volume       float32
	Attenuation  float32

	/* timing variables */
	Wait   float32
	Delay  float32 /* before firing targets */
	Random float32

	// float last_sound_time;

	watertype  int
	waterlevel int

	// vec3_t move_origin;
	// vec3_t move_angles;

	// /* move this to clientinfo? */
	// int light_level;

	Style int /* also used as areaportal number */

	item *gitem_t /* for bonus items */

	/* common data blocks */
	// moveinfo_t moveinfo;
	monsterinfo monsterinfo_t
}

func (G *edict_t) S() *shared.Entity_state_t {
	return &G.s
}

func (G *edict_t) Client() shared.Gclient_s {
	return G.client
}

func (G *edict_t) Area() *shared.Link_t {
	return &G.area
}

func (G *edict_t) Inuse() bool {
	return G.inuse
}

func (G *edict_t) Linkcount() int {
	return G.linkcount
}

func (G *edict_t) SetLinkcount(v int) {
	G.linkcount = v
}

func (G *edict_t) Svflags() int {
	return G.svflags
}

func (G *edict_t) Mins() []float32 {
	return G.mins[:]
}

func (G *edict_t) Maxs() []float32 {
	return G.maxs[:]
}

func (G *edict_t) Absmin() []float32 {
	return G.absmin[:]
}

func (G *edict_t) Absmax() []float32 {
	return G.absmax[:]
}

func (G *edict_t) Size() []float32 {
	return G.size[:]
}

func (G *edict_t) Solid() shared.Solid_t {
	return G.solid
}

func (G *edict_t) NumClusters() int {
	return G.num_clusters
}

func (G *edict_t) SetNumClusters(v int) {
	G.num_clusters = v
}

func (G *edict_t) Clusternums() []int {
	return G.clusternums[:]
}

func (G *edict_t) Headnode() int {
	return G.headnode
}

func (G *edict_t) SetHeadnode(v int) {
	G.headnode = v
}

func (G *edict_t) Areanum() int {
	return G.areanum
}

func (G *edict_t) SetAreanum(v int) {
	G.areanum = v
}

func (G *edict_t) Areanum2() int {
	return G.areanum2
}

func (G *edict_t) SetAreanum2(v int) {
	G.areanum2 = v
}

func (G *edict_t) Owner() shared.Edict_s {
	return G.owner
}

/* item spawnflags */
const ITEM_TRIGGER_SPAWN = 0x00000001
const ITEM_NO_TOUCH = 0x00000002

/* 6 bits reserved for editor flags */
/* 8 bits used as power cube id bits for coop games */
const DROPPED_ITEM = 0x00010000
const DROPPED_PLAYER_ITEM = 0x00020000
const ITEM_TARGETS_USED = 0x00040000

/* fields are needed for spawning from the entity
   string and saving / loading games */
const FFL_SPAWNTEMP = 1
const FFL_NOSPAWN = 2
const FFL_ENTITYSTATE = 4

type fieldtype_t int

const (
	F_INT       fieldtype_t = 0
	F_FLOAT     fieldtype_t = 1
	F_LSTRING   fieldtype_t = 2
	F_GSTRING   fieldtype_t = 3 /* string on disk, pointer in memory, TAG_GAME */
	F_VECTOR    fieldtype_t = 3
	F_ANGLEHACK fieldtype_t = 4
	F_IGNORE    fieldtype_t = 5
)

type field_t struct {
	name  string
	fname string
	ftype fieldtype_t
	flags int
	// short save_ver;
}

type qGame struct {
	gi         shared.Game_import_t
	game       game_locals_t
	level      level_locals_t
	st         spawn_temp_t
	num_edicts int

	g_edicts []edict_t

	deathmatch            *shared.CvarT
	coop                  *shared.CvarT
	coop_pickup_weapons   *shared.CvarT
	coop_elevator_delay   *shared.CvarT
	dmflags               *shared.CvarT
	skill                 *shared.CvarT
	fraglimit             *shared.CvarT
	timelimit             *shared.CvarT
	password              *shared.CvarT
	spectator_password    *shared.CvarT
	needpass              *shared.CvarT
	maxclients            *shared.CvarT
	maxspectators         *shared.CvarT
	maxentities           *shared.CvarT
	g_select_empty        *shared.CvarT
	dedicated             *shared.CvarT
	g_footsteps           *shared.CvarT
	g_fix_triggered       *shared.CvarT
	g_commanderbody_nogod *shared.CvarT

	filterban *shared.CvarT

	sv_maxvelocity *shared.CvarT
	sv_gravity     *shared.CvarT

	sv_rollspeed *shared.CvarT
	sv_rollangle *shared.CvarT
	gun_x        *shared.CvarT
	gun_y        *shared.CvarT
	gun_z        *shared.CvarT

	run_pitch *shared.CvarT
	run_roll  *shared.CvarT
	bob_up    *shared.CvarT
	bob_pitch *shared.CvarT
	bob_roll  *shared.CvarT

	sv_cheats *shared.CvarT

	flood_msgs      *shared.CvarT
	flood_persecond *shared.CvarT
	flood_waitdelay *shared.CvarT

	sv_maplist *shared.CvarT

	gib_on *shared.CvarT

	aimfix *shared.CvarT

	pm_passent *edict_t

	current_player         *edict_t
	current_client         *gclient_t
	player_view_forward    [3]float32
	player_view_right      [3]float32
	player_view_up         [3]float32
	player_view_xyspeed    float32
	player_view_bobmove    float32
	player_view_bobcycle   int
	player_view_bobfracsin float32
}

func QGameCreate(gi shared.Game_import_t) shared.Game_export_t {
	g := &qGame{}
	g.gi = gi
	return g
}
