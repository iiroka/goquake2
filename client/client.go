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
 * Main header for the client
 *
 * =======================================================================
 */
package client

import (
	"goquake2/shared"

	"github.com/veandco/go-sdl2/sdl"
)

const MAX_CLIENTWEAPONMODELS = 20
const CMD_BACKUP = 256 /* allow a lot of command backups for very fast systems */

/* the cl_parse_entities must be large enough to hold UPDATE_BACKUP frames of
   entities, so that when a delta compressed message arives from the server
   it can be un-deltad from the original */
const MAX_PARSE_ENTITIES = 1024

const PARTICLE_GRAVITY = 40
const INSTANT_PARTICLE = -10000.0

type frame_t struct {
	valid          bool /* cleared if delta parsing was invalid */
	serverframe    int
	servertime     int /* server time the message is valid for (in msec) */
	deltaframe     int
	areabits       [shared.MAX_MAP_AREAS / 8]byte /* portalarea visibility bits */
	playerstate    shared.Player_state_t
	num_entities   int
	parse_entities int /* non-masked index into cl_parse_entities array */
}

func (T *frame_t) copy(other frame_t) {
	T.valid = other.valid
	T.serverframe = other.serverframe
	T.servertime = other.servertime
	T.deltaframe = other.deltaframe
	copy(T.areabits[:], other.areabits[:])
	T.playerstate.Copy(other.playerstate)
	T.num_entities = other.num_entities
	T.parse_entities = other.parse_entities
}

type centity_t struct {
	baseline shared.Entity_state_t /* delta from this if not from a previous frame */
	current  shared.Entity_state_t
	prev     shared.Entity_state_t /* will always be valid, but might just be a copy of current */

	serverframe int /* if not current, this ent isn't in the frame */

	trailcount  int       /* for diminishing grenade trails */
	lerp_origin []float32 /* for trails (variable hz) */

	fly_stoptime int
}

type clientinfo_t struct {
	name  string
	cinfo string

	skin interface{}

	icon     interface{}
	iconname string

	model interface{}

	weaponmodel [MAX_CLIENTWEAPONMODELS]interface{}
}

/* the client_state_t structure is wiped
   completely at every server map change */
type client_state_t struct {
	timeoutcount int

	timedemo_frames int
	timedemo_start  int

	refresh_prepped bool /* false if on new level or new ref dll */
	sound_prepped   bool /* ambient sounds can start */
	force_refdef    bool /* vid has changed, so we can't use a paused refdef */

	parse_entities int /* index (not anded off) into cl_parse_entities[] */

	cmd               shared.Usercmd_t
	cmds              [CMD_BACKUP]shared.Usercmd_t /* each mesage will send several old cmds */
	cmd_time          [CMD_BACKUP]int              /* time sent, for calculating pings */
	predicted_origins [CMD_BACKUP][3]int16         /* for debug comparing against server */

	predicted_step      float32 /* for stair up smoothing */
	predicted_step_time uint

	predicted_origin [3]float32 /* generated by CL_PredictMovement */
	predicted_angles [3]float32
	prediction_error [3]float32

	frame         frame_t /* received from server */
	surpressCount int     /* number of messages rate supressed */
	frames        [shared.UPDATE_BACKUP]frame_t

	/* the client maintains its own idea of view angles, which are
	sent to the server each frame.  It is cleared to 0 upon entering each level.
	the server sends a delta each frame which is added to the locally
	tracked view angles to account for standing on rotating objects,
	and teleport direction changes */
	viewangles [3]float32

	time     int     /* this is the time value that the client is rendering at. always <= cls.realtime */
	lerpfrac float32 /* between oldframe and frame */

	refdef shared.Refdef_t

	/* set when refdef.angles is set */
	v_forward [3]float32
	v_right   [3]float32
	v_up      [3]float32

	/* transient data from server */
	layout    string /* general 2D overlay */
	inventory [shared.MAX_ITEMS]int

	/* non-gameserver infornamtion */
	cinematic_file shared.QFileHandle
	cinematictime  int /* cls.realtime for first cinematic frame */
	cinematicframe int
	//    unsigned char	cinematicpalette[768];
	//    qboolean	cinematicpalette_active;

	/* server state information */
	attractloop bool /* running the attract loop, any key will menu */
	servercount int  /* server identification for prespawns */
	gamedir     string
	playernum   int

	configstrings []string
	//    char		configstrings[MAX_CONFIGSTRINGS][MAX_QPATH];

	//    /* locally derived information from server state */

	model_draw []interface{}
	//    struct model_s	*model_draw[MAX_MODELS];

	model_clip [shared.MAX_MODELS]*shared.Cmodel_t

	//    struct sfx_s	*sound_precache[MAX_SOUNDS];

	//    struct image_s	*image_precache[MAX_IMAGES];

	clientinfo     [shared.MAX_CLIENTS]clientinfo_t
	baseclientinfo clientinfo_t
}

/* the client_static_t structure is persistant through
   an arbitrary number of server connections */
type connstate_t int

const (
	ca_uninitialized connstate_t = 0
	ca_disconnected  connstate_t = 1 /* not talking to a server */
	ca_connecting    connstate_t = 2 /* sending request packets to the server */
	ca_connected     connstate_t = 3 /* netchan_t established, waiting for svc_serverdata */
	ca_active        connstate_t = 4 /* game views should be displayed */
)

type keydest_t int

const (
	key_game    keydest_t = 0
	key_console keydest_t = 1
	key_message keydest_t = 2
	key_menu    keydest_t = 3
)

type client_static_t struct {
	state    connstate_t
	key_dest keydest_t

	framecount int
	realtime   int     /* always increasing, no clamping, etc */
	rframetime float32 /* seconds since last render frame */
	nframetime float32 /* network frame time */

	/* screen rendering information */
	disable_screen float32 /* showing loading plaque between levels */
	/* or changing rendering dlls */

	/* if time gets > 30 seconds ahead, break it */
	disable_servercount int /* when we receive a frame and cl.servercount */
	/* > cls.disable_servercount, clear disable_screen */

	/* connection information */
	servername   string  /* name of server from original connect */
	connect_time float32 /* for connection retransmits */

	quakePort int /* a 16 bit value that allows quake servers */
	/* to work around address translating routers */
	netchan        shared.Netchan_t
	serverProtocol int /* in case we are doing some kind of version hack */

	challenge int /* from the server to use for connecting */

	forcePacket bool /* Forces a package to be send at the next frame. */

	// 	FILE		*download; /* file transfer from server */
	// 	char		downloadtempname[MAX_OSPATH];
	// 	char		downloadname[MAX_OSPATH];
	// 	int			downloadnumber;
	// 	dltype_t	downloadtype;
	// 	size_t		downloadposition;
	// 	int			downloadpercent;

	// 	/* demo recording info must be here, so it isn't cleared on level change */
	// 	qboolean	demorecording;
	// 	qboolean	demowaiting; /* don't record until a non-delta message is received */
	// 	FILE		*demofile;

	// #ifdef USE_CURL
	// 	/* http downloading */
	// 	dlqueue_t  downloadQueue; /* queues with files to download. */
	// 	dlhandle_t HTTPHandles[MAX_HTTP_HANDLES]; /* download handles. */
	// 	char	   downloadServer[512]; /* URL prefix to dowload from .*/
	// 	char	   downloadServerRetry[512]; /* retry count. */
	// 	char	   downloadReferer[32]; /* referer string. */
	// #endif
}

type dirty_t struct {
	x1, y1, x2, y2 int
}

type cparticle_t struct {
	next *cparticle_t

	time float32

	org      [3]float32
	vel      [3]float32
	accel    [3]float32
	color    float32
	colorvel float32
	alpha    float32
	alphavel float32
}

type cdlight_t struct {
	key      int /* so entities can reuse same entry */
	color    [3]float32
	origin   [3]float32
	radius   float32
	die      float32 /* stop lighting after this time */
	decay    float32 /* drop this each second */
	minlight float32 /* don't add when contributing less */
}

type kbutton_t struct {
	down     [2]int /* key nums holding it down */
	downtime uint   /* msec timestamp */
	msec     uint   /* msec down this frame */
	state    int
}

type qClient struct {
	common shared.QCommon
	input  QInput

	vid_gamma      *shared.CvarT
	vid_fullscreen *shared.CvarT
	vid_renderer   *shared.CvarT
	viddef         viddef_t

	re shared.Refexport_t
	ri shared.Refimport_t

	vid_displayrefreshrate *shared.CvarT
	vid_displayindex       *shared.CvarT
	vid_rate               *shared.CvarT
	num_displays           int
	window                 *sdl.Window
	last_display           int
	glimp_refreshRate      int

	rcon_client_password *shared.CvarT
	rcon_address         *shared.CvarT

	cl_noskins       *shared.CvarT
	cl_footsteps     *shared.CvarT
	cl_timeout       *shared.CvarT
	cl_predict       *shared.CvarT
	cl_showfps       *shared.CvarT
	cl_gun           *shared.CvarT
	cl_add_particles *shared.CvarT
	cl_add_lights    *shared.CvarT
	cl_add_entities  *shared.CvarT
	cl_add_blend     *shared.CvarT
	cl_kickangles    *shared.CvarT

	cl_shownet   *shared.CvarT
	cl_showmiss  *shared.CvarT
	cl_showclamp *shared.CvarT

	cl_paused     *shared.CvarT
	cl_loadpaused *shared.CvarT

	cl_lightlevel       *shared.CvarT
	cl_r1q2_lightstyle  *shared.CvarT
	cl_limitsparksounds *shared.CvarT

	/* userinfo */
	name           *shared.CvarT
	skin           *shared.CvarT
	rate           *shared.CvarT
	fov            *shared.CvarT
	horplus        *shared.CvarT
	windowed_mouse *shared.CvarT
	msg            *shared.CvarT
	hand           *shared.CvarT
	gender         *shared.CvarT
	gender_auto    *shared.CvarT

	gl1_stereo             *shared.CvarT
	gl1_stereo_separation  *shared.CvarT
	gl1_stereo_convergence *shared.CvarT

	cl_vwep *shared.CvarT

	cl  client_state_t
	cls client_static_t

	cl_upspeed       *shared.CvarT
	cl_forwardspeed  *shared.CvarT
	cl_sidespeed     *shared.CvarT
	cl_yawspeed      *shared.CvarT
	cl_pitchspeed    *shared.CvarT
	cl_run           *shared.CvarT
	cl_anglespeedkey *shared.CvarT

	// screen
	scr_con_current float32 /* aproaches scr_conlines at scr_conspeed */
	scr_conlines    float32 /* 0.0 to 1.0 lines of console to display */

	scr_initialized bool /* ready to draw */

	scr_draw_loading int

	scr_vrect vrect_t /* position of render window on screen */

	scr_viewsize   *shared.CvarT
	scr_conspeed   *shared.CvarT
	scr_centertime *shared.CvarT
	scr_showturtle *shared.CvarT
	scr_showpause  *shared.CvarT

	scr_netgraph    *shared.CvarT
	scr_timegraph   *shared.CvarT
	scr_debuggraph  *shared.CvarT
	scr_graphheight *shared.CvarT
	scr_graphscale  *shared.CvarT
	scr_graphshift  *shared.CvarT
	scr_drawall     *shared.CvarT

	scr_dirty     dirty_t
	scr_old_dirty [2]dirty_t

	r_hudscale     *shared.CvarT /* named for consistency with R1Q2 */
	r_consolescale *shared.CvarT
	r_menuscale    *shared.CvarT

	cl_entities       []centity_t
	cl_parse_entities []shared.Entity_state_t

	// console
	con console_t

	r_entities    []shared.Entity_t
	r_particles   []shared.Particle_t
	r_lightstyles [shared.MAX_LIGHTSTYLES]shared.Lightstyle_t
	r_dlights     []shared.Dlight_t

	num_cl_weaponmodels int
	cl_weaponmodels     [MAX_CLIENTWEAPONMODELS]string

	frame_msec         int
	old_sys_frame_time int

	// 	struct sfx_s *cl_sfx_ric1;
	// struct sfx_s *cl_sfx_ric2;
	// struct sfx_s *cl_sfx_ric3;
	// struct sfx_s *cl_sfx_lashit;
	// struct sfx_s *cl_sfx_spark5;
	// struct sfx_s *cl_sfx_spark6;
	// struct sfx_s *cl_sfx_spark7;
	// struct sfx_s *cl_sfx_railg;
	// struct sfx_s *cl_sfx_rockexp;
	// struct sfx_s *cl_sfx_grenexp;
	// struct sfx_s *cl_sfx_watrexp;
	// struct sfx_s *cl_sfx_plasexp;
	// struct sfx_s *cl_sfx_footsteps[4];

	cl_mod_explode          interface{}
	cl_mod_smoke            interface{}
	cl_mod_flash            interface{}
	cl_mod_parasite_segment interface{}
	cl_mod_grapple_cable    interface{}
	cl_mod_parasite_tip     interface{}
	cl_mod_explo4           interface{}
	cl_mod_bfg_explo        interface{}
	cl_mod_powerscreen      interface{}
	cl_mod_plasmaexplo      interface{}

	// cl_sfx_lightning        interface{}
	// cl_sfx_disrexp          interface{}
	cl_mod_lightning        interface{}
	cl_mod_heatbeam         interface{}
	cl_mod_monster_heatbeam interface{}
	cl_mod_explo4_big       interface{}

	active_particles *cparticle_t
	free_particles   *cparticle_t
	particles        [shared.MAX_PARTICLES]cparticle_t

	cl_lightstyle [shared.MAX_LIGHTSTYLES]clightstyle_t
	lastofs       int

	cl_dlights [shared.MAX_DLIGHTS]cdlight_t

	// keyboard
	key_lines   [NUM_KEY_LINES]string
	key_linepos int
	anykeydown  int

	edit_line    int
	history_line int

	key_waiting int
	keybindings [K_LAST]string
	consolekeys [K_LAST]bool /* if true, can't be rebound while in console */
	menubound   [K_LAST]bool /* if true, can't be rebound while in menu */
	key_repeats [K_LAST]int  /* if > 1, it is autorepeating */
	keydown     [K_LAST]bool

	menu MenuStr

	precache_check         int
	precache_spawncount    int
	precache_tex           int
	precache_model_skin    int
	precache_model         []byte
	allow_download         *shared.CvarT
	allow_download_players *shared.CvarT
	allow_download_models  *shared.CvarT
	allow_download_sounds  *shared.CvarT
	allow_download_maps    *shared.CvarT

	cl_explosions [MAX_EXPLOSIONS]explosion_t

	in_klook                                          kbutton_t
	in_left, in_right, in_forward, in_back            kbutton_t
	in_lookup, in_lookdown, in_moveleft, in_moveright kbutton_t
	in_strafe, in_speed, in_use, in_attack            kbutton_t
	in_up, in_down                                    kbutton_t

	in_impulse int

	gun_frame int
	gun_model interface{}
}

func CreateClient() shared.QClient {
	q := &qClient{}
	q.cl.configstrings = make([]string, shared.MAX_CONFIGSTRINGS)
	q.cl_entities = make([]centity_t, shared.MAX_EDICTS)
	q.cl.model_draw = make([]interface{}, shared.MAX_MODELS)
	q.cl_parse_entities = make([]shared.Entity_state_t, MAX_PARSE_ENTITIES)
	return q
}

func (T *qClient) SetCommon(common shared.QCommon) {
	T.common = common
}

func (T *qClient) IsAttractloop() bool {
	return T.cl.attractloop
}
