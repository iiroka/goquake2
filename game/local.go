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

const FRAMETIME = 0.1

const (
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

	/* armor types */
	ARMOR_NONE   = 0
	ARMOR_JACKET = 1
	ARMOR_COMBAT = 2
	ARMOR_BODY   = 3
	ARMOR_SHARD  = 4
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

	//    edict_t *sight_client; /* changed once each frame for coop games */

	//    edict_t *sight_entity;
	//    int sight_entity_framenum;
	//    edict_t *sound_entity;
	//    int sound_entity_framenum;
	//    edict_t *sound2_entity;
	//    int sound2_entity_framenum;

	//    int pic_health;

	//    int total_secrets;
	//    int found_secrets;

	//    int total_goals;
	//    int found_goals;

	//    int total_monsters;
	//    int killed_monsters;

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

	//    int lip;
	//    int distance;
	//    int height;
	Noise string
	//    float pausetime;
	//    char *item;
	//    char *gravity;

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
	// int aiflags;
	nextframe int
	scale     float32

	stand func(self *edict_t, G *qGame)
	// void (*idle)(edict_t *self);
	// void (*search)(edict_t *self);
	// void (*walk)(edict_t *self);
	// void (*run)(edict_t *self);
	// void (*dodge)(edict_t *self, edict_t *other, float eta);
	// void (*attack)(edict_t *self);
	// void (*melee)(edict_t *self);
	// void (*sight)(edict_t *self, edict_t *other);
	// qboolean (*checkattack)(edict_t *self);

	// float pausetime;
	// float attack_finished;

	// vec3_t saved_goal;
	// float search_time;
	// float trail_time;
	// vec3_t last_sighting;
	// int attack_state;
	// int lefty;
	// float idle_time;
	// int linkcount;

	// int power_armor_type;
	// int power_armor_power;
}

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
	// int score       /* for calculating total unit score in coop games */

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
	// int power_cubes /* used for tracking the cubes in coop games */
	// int score       /* for calculating total unit score in coop games */
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

/* this structure is cleared on each PutClientInServer(),
   except for 'client->pers' */
type gclient_t struct {
	/* known to server */
	ps   shared.Player_state_t /* communicated by server to clients */
	ping int

	/* private to game */
	pers client_persistant_t
	resp client_respawn_t
	// pmove_state_t old_pmove; /* for detecting out-of-pmove changes */

	// qboolean showscores; /* set layout stat */
	// qboolean showinventory; /* set layout stat */
	// qboolean showhelp;
	// qboolean showhelpicon;

	// int ammo_index;

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

	// weaponstate_t weaponstate;
	// vec3_t kick_angles; /* weapon kicks */
	// vec3_t kick_origin;
	// float v_dmg_roll, v_dmg_pitch, v_dmg_time; /* damage kicks */
	// float fall_time, fall_value; /* for view drop on fall */
	// float damage_alpha;
	// float bonus_alpha;
	// vec3_t damage_blend;
	v_angle [3]float32 /* aiming direction */
	// float bobtime; /* so off-ground doesn't change it */
	// vec3_t oldviewangles;
	// vec3_t oldvelocity;

	// float next_drown_time;
	// int old_waterlevel;
	// int breather_sound;

	// int machinegun_shots; /* for weapon raising */

	// /* animation vars */
	// int anim_end;
	// int anim_priority;
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

	// float pickup_msg_time;

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
	// pmove_state_t old_pmove; /* for detecting out-of-pmove changes */
	// qboolean showscores; /* set layout stat */
	// qboolean showinventory; /* set layout stat */
	// qboolean showhelp;
	// qboolean showhelpicon;
	// int ammo_index;
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
	// vec3_t kick_angles; /* weapon kicks */
	// vec3_t kick_origin;
	// float v_dmg_roll, v_dmg_pitch, v_dmg_time; /* damage kicks */
	// float fall_time, fall_value; /* for view drop on fall */
	// float damage_alpha;
	// float bonus_alpha;
	// vec3_t damage_blend;
	// float bobtime; /* so off-ground doesn't change it */
	// vec3_t oldviewangles;
	// vec3_t oldvelocity;
	// float next_drown_time;
	// int old_waterlevel;
	// int breather_sound;
	// int machinegun_shots; /* for weapon raising */
	// int anim_end;
	// int anim_priority;
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
		G.v_angle[i] = other.v_angle[i]
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
	Target     string
	Targetname string
	// char *killtarget;
	// char *team;
	// char *pathtarget;
	// char *deathtarget;
	// char *combattarget;
	// edict_t *target_ent;

	Speed, Accel, Decel float32
	// vec3_t movedir;
	// vec3_t pos1, pos2;

	velocity [3]float32
	// vec3_t avelocity;
	// int mass;
	// float air_finished;
	gravity float32 /* per entity gravity multiplier (1.0 is normal)
	   use for lowgrav artifact, flares */

	// edict_t *goalentity;
	// edict_t *movetarget;
	// float yaw_speed;
	// float ideal_yaw;

	nextthink float32
	// void (*prethink)(edict_t *ent);
	// void (*think)(edict_t *self);
	// void (*blocked)(edict_t *self, edict_t *other);
	// void (*touch)(edict_t *self, edict_t *other, cplane_t *plane,
	// 		csurface_t *surf);
	// void (*use)(edict_t *self, edict_t *other, edict_t *activator);
	// void (*pain)(edict_t *self, edict_t *other, float kick, int damage);
	// void (*die)(edict_t *self, edict_t *inflictor, edict_t *attacker,
	// 		int damage, vec3_t point);

	// float touch_debounce_time;
	// float pain_debounce_time;
	// float damage_debounce_time;
	// float fly_sound_debounce_time;	/* now also used by insane marines to store pain sound timeout */
	// float last_move_time;

	health     int
	max_health int
	gib_health int
	deadflag   int

	// float show_hostile;
	// float powerarmor_time;

	// char *map; /* target_changelevel */

	viewheight int /* height above origin where eyesight is determined */
	// int takedamage;
	Dmg int
	// int radius_dmg;
	// float dmg_radius;
	Sounds int /* make this a spawntemp var? */
	// int count;

	// edict_t *chain;
	// edict_t *enemy;
	// edict_t *oldenemy;
	// edict_t *activator;
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

	// int watertype;
	// int waterlevel;

	// vec3_t move_origin;
	// vec3_t move_angles;

	// /* move this to clientinfo? */
	// int light_level;

	Style int /* also used as areaportal number */

	// gitem_t *item; /* for bonus items */

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
}

func QGameCreate(gi shared.Game_import_t) shared.Game_export_t {
	g := &qGame{}
	g.gi = gi
	return g
}
