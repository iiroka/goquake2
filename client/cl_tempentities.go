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
 * This file implements all temporary (dynamic created) entities
 *
 * =======================================================================
 */
package client

import (
	"goquake2/shared"
	"math"
)

type exptype_t int

const (
	ex_free      exptype_t = 0
	ex_explosion exptype_t = 1
	ex_misc      exptype_t = 2
	ex_flash     exptype_t = 3
	ex_mflash    exptype_t = 4
	ex_poly      exptype_t = 5
	ex_poly2     exptype_t = 6
)

type explosion_t struct {
	etype exptype_t
	ent   shared.Entity_t

	frames     int
	light      float32
	lightcolor [3]float32
	start      float32
	baseframe  int
}

const MAX_EXPLOSIONS = 64

func (T *qClient) registerTEntModels() {
	T.cl_mod_explode, _ = T.R_RegisterModel("models/objects/explode/tris.md2")
	T.cl_mod_smoke, _ = T.R_RegisterModel("models/objects/smoke/tris.md2")
	T.cl_mod_flash, _ = T.R_RegisterModel("models/objects/flash/tris.md2")
	T.cl_mod_parasite_segment, _ = T.R_RegisterModel("models/monsters/parasite/segment/tris.md2")
	T.cl_mod_grapple_cable, _ = T.R_RegisterModel("models/ctf/segment/tris.md2")
	T.cl_mod_parasite_tip, _ = T.R_RegisterModel("models/monsters/parasite/tip/tris.md2")
	T.cl_mod_explo4, _ = T.R_RegisterModel("models/objects/r_explode/tris.md2")
	T.cl_mod_bfg_explo, _ = T.R_RegisterModel("sprites/s_bfg2.sp2")
	T.cl_mod_powerscreen, _ = T.R_RegisterModel("models/items/armor/effect/tris.md2")

	T.R_RegisterModel("models/objects/laser/tris.md2")
	T.R_RegisterModel("models/objects/grenade2/tris.md2")
	T.R_RegisterModel("models/weapons/v_machn/tris.md2")
	T.R_RegisterModel("models/weapons/v_handgr/tris.md2")
	T.R_RegisterModel("models/weapons/v_shotg2/tris.md2")
	T.R_RegisterModel("models/objects/gibs/bone/tris.md2")
	T.R_RegisterModel("models/objects/gibs/sm_meat/tris.md2")
	T.R_RegisterModel("models/objects/gibs/bone2/tris.md2")

	// T.Draw_FindPic("w_machinegun")
	// T.Draw_FindPic("a_bullets")
	// T.Draw_FindPic("i_health")
	// T.Draw_FindPic("a_grenades")

	T.cl_mod_explo4_big, _ = T.R_RegisterModel("models/objects/r_explode2/tris.md2")
	T.cl_mod_lightning, _ = T.R_RegisterModel("models/proj/lightning/tris.md2")
	T.cl_mod_heatbeam, _ = T.R_RegisterModel("models/proj/beam/tris.md2")
	T.cl_mod_monster_heatbeam, _ = T.R_RegisterModel("models/proj/widowbeam/tris.md2")
}

func (T *qClient) clearTEnts() {
	// memset(cl_beams, 0, sizeof(cl_beams));
	for i := range T.cl_explosions {
		T.cl_explosions[i].etype = ex_free
		T.cl_explosions[i].start = 0
	}
	// memset(cl_lasers, 0, sizeof(cl_lasers));

	// memset(cl_playerbeams, 0, sizeof(cl_playerbeams));
	// memset(cl_sustains, 0, sizeof(cl_sustains));
}

func (T *qClient) allocExplosion() *explosion_t {
	// int i;
	// float time;
	// int index;

	for i, exp := range T.cl_explosions {
		if exp.etype == ex_free {
			T.cl_explosions[i] = explosion_t{}
			return &T.cl_explosions[i]
		}
	}

	/* find the oldest explosion */
	time := float32(T.cl.time)
	index := 0

	for i, exp := range T.cl_explosions {
		if exp.start < time {
			time = exp.start
			index = i
		}
	}

	T.cl_explosions[index] = explosion_t{}
	return &T.cl_explosions[index]
}

func (T *qClient) smokeAndFlash(origin []float32) {

	ex := T.allocExplosion()
	copy(ex.ent.Origin[:], origin)
	ex.etype = ex_misc
	ex.frames = 4
	ex.ent.Flags = shared.RF_TRANSLUCENT
	ex.start = float32(T.cl.frame.servertime) - 100.0
	ex.ent.Model = T.cl_mod_smoke

	ex = T.allocExplosion()
	copy(ex.ent.Origin[:], origin)
	ex.etype = ex_flash
	ex.ent.Flags = shared.RF_FULLBRIGHT
	ex.frames = 2
	ex.start = float32(T.cl.frame.servertime) - 100.0
	ex.ent.Model = T.cl_mod_flash
}

func (T *qClient) parseTEnt(msg *shared.QReadbuf) error {
	// int type;
	// vec3_t pos, pos2, dir;
	// explosion_t *ex;
	// int cnt;
	// int color;
	// int r;
	// int ent;
	// int magnitude;

	mtype := msg.ReadByte()

	switch mtype {
	case shared.TE_BLOOD: /* bullet hitting flesh */
		pos := msg.ReadPos()
		dir := msg.ReadDir()
		T.particleEffect(pos, dir, 0xe8, 60)

	case shared.TE_GUNSHOT, /* bullet hitting wall */
		shared.TE_SPARKS,
		shared.TE_BULLET_SPARKS:
		pos := msg.ReadPos()
		dir := msg.ReadDir()

		if mtype == shared.TE_GUNSHOT {
			T.particleEffect(pos, dir, 0, 40)
		} else {
			T.particleEffect(pos, dir, 0xe0, 6)
		}

	// 	if (type != TE_SPARKS) {
	// 		CL_SmokeAndFlash(pos);
	// 		/* impact sound */
	// 		cnt = randk() & 15;

	// 		if (cnt == 1)
	// 		{
	// 			S_StartSound(pos, 0, 0, cl_sfx_ric1, 1, ATTN_NORM, 0);
	// 		}
	// 		else if (cnt == 2)
	// 		{
	// 			S_StartSound(pos, 0, 0, cl_sfx_ric2, 1, ATTN_NORM, 0);
	// 		}
	// 		else if (cnt == 3)
	// 		{
	// 			S_StartSound(pos, 0, 0, cl_sfx_ric3, 1, ATTN_NORM, 0);
	// 		}
	// 	}

	case shared.TE_SCREEN_SPARKS,
		shared.TE_SHIELD_SPARKS:
		pos := msg.ReadPos()
		dir := msg.ReadDir()

		if mtype == shared.TE_SCREEN_SPARKS {
			T.particleEffect(pos, dir, 0xd0, 40)
		} else {
			T.particleEffect(pos, dir, 0xb0, 40)
		}

	// 	if (cl_limitsparksounds->value)
	// 	{
	// 		num_power_sounds++;

	// 		/* If too many of these sounds are started in one frame
	// 		 * (for example if the player shoots with the super
	// 		 * shotgun into the power screen of a Brain) things get
	// 		 * too loud and OpenAL is forced to scale the volume of
	// 		 * several other sounds and the background music down.
	// 		 * That leads to a noticable and annoying drop in the
	// 		 * overall volume.
	// 		 *
	// 		 * Work around that by limiting the number of sounds
	// 		 * started.
	// 		 * 16 was choosen by empirical testing.
	// 		 *
	// 		 * This was fixed in openal-soft 0.19.0. We're keeping
	// 		 * the work around hidden behind a cvar and no longer
	// 		 * limited to OpenAL because a) some Linux distros may
	// 		 * still ship older openal-soft versions and b) some
	// 		 * player may like the changed behavior.
	// 		 */
	// 		if (num_power_sounds < 16)
	// 		{
	// 			S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 		}
	// 	}
	// 	else
	// 	{
	// 		S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 	}

	// 	break;

	case shared.TE_SHOTGUN: /* bullet hitting wall */
		pos := msg.ReadPos()
		dir := msg.ReadDir()
		T.particleEffect(pos, dir, 0, 20)
	// 	CL_SmokeAndFlash(pos);

	// case TE_SPLASH: /* bullet hitting water */
	// 	cnt = MSG_ReadByte(&net_message);
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	r = MSG_ReadByte(&net_message);

	// 	if (r > 6)
	// 	{
	// 		color = 0x00;
	// 	}
	// 	else
	// 	{
	// 		color = splash_color[r];
	// 	}

	// 	CL_ParticleEffect(pos, dir, color, cnt);

	// 	if (r == SPLASH_SPARKS)
	// 	{
	// 		r = randk() & 3;

	// 		if (r == 0)
	// 		{
	// 			S_StartSound(pos, 0, 0, cl_sfx_spark5, 1, ATTN_STATIC, 0);
	// 		}
	// 		else if (r == 1)
	// 		{
	// 			S_StartSound(pos, 0, 0, cl_sfx_spark6, 1, ATTN_STATIC, 0);
	// 		}
	// 		else
	// 		{
	// 			S_StartSound(pos, 0, 0, cl_sfx_spark7, 1, ATTN_STATIC, 0);
	// 		}
	// 	}

	// 	break;

	// case TE_LASER_SPARKS:
	// 	cnt = MSG_ReadByte(&net_message);
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	color = MSG_ReadByte(&net_message);
	// 	CL_ParticleEffect2(pos, dir, color, cnt);
	// 	break;

	// case TE_BLUEHYPERBLASTER:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadPos(&net_message, dir);
	// 	CL_BlasterParticles(pos, dir);
	// 	break;

	case shared.TE_BLASTER: /* blaster hitting wall */
		pos := msg.ReadPos()
		dir := msg.ReadDir()
		T.blasterParticles(pos, dir)

		ex := T.allocExplosion()
		copy(ex.ent.Origin[:], pos)
		ex.ent.Angles[0] = float32(math.Acos(float64(dir[2])) / math.Pi * 180)

		if dir[0] != 0 {
			ex.ent.Angles[1] = float32(math.Atan2(float64(dir[1]), float64(dir[0])) / math.Pi * 180)
		} else if dir[1] > 0 {
			ex.ent.Angles[1] = 90
		} else if dir[1] < 0 {
			ex.ent.Angles[1] = 270
		} else {
			ex.ent.Angles[1] = 0
		}

		ex.etype = ex_misc
		ex.ent.Flags = 0
		ex.start = float32(T.cl.frame.servertime) - 100.0
		ex.light = 150
		ex.lightcolor[0] = 1
		ex.lightcolor[1] = 1
		ex.ent.Model = T.cl_mod_explode
		ex.frames = 4
	// 	S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 	break;

	// case TE_RAILTRAIL: /* railgun effect */
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadPos(&net_message, pos2);
	// 	CL_RailTrail(pos, pos2);
	// 	S_StartSound(pos2, 0, 0, cl_sfx_railg, 1, ATTN_NORM, 0);
	// 	break;

	case shared.TE_EXPLOSION2,
		shared.TE_GRENADE_EXPLOSION,
		shared.TE_GRENADE_EXPLOSION_WATER:
		pos := msg.ReadPos()
		ex := T.allocExplosion()
		copy(ex.ent.Origin[:], pos)
		ex.etype = ex_poly
		ex.ent.Flags = shared.RF_FULLBRIGHT | shared.RF_NOSHADOW
		ex.start = float32(T.cl.frame.servertime) - 100.0
		ex.light = 350
		ex.lightcolor[0] = 1.0
		ex.lightcolor[1] = 0.5
		ex.lightcolor[2] = 0.5
		ex.ent.Model = T.cl_mod_explo4
		ex.frames = 19
		ex.baseframe = 30
		ex.ent.Angles[1] = float32(shared.Randk() % 360)
		T.explosionParticles(pos)

	// 	if (type == TE_GRENADE_EXPLOSION_WATER) {
	// 		S_StartSound(pos, 0, 0, cl_sfx_watrexp, 1, ATTN_NORM, 0);
	// 	} else {
	// 		S_StartSound(pos, 0, 0, cl_sfx_grenexp, 1, ATTN_NORM, 0);
	// 	}

	// 	break;

	// case TE_PLASMA_EXPLOSION:
	// 	MSG_ReadPos(&net_message, pos);
	// 	ex = CL_AllocExplosion();
	// 	VectorCopy(pos, ex->ent.origin);
	// 	ex->type = ex_poly;
	// 	ex->ent.flags = RF_FULLBRIGHT | RF_NOSHADOW;
	// 	ex->start = cl.frame.servertime - 100.0f;
	// 	ex->light = 350;
	// 	ex->lightcolor[0] = 1.0;
	// 	ex->lightcolor[1] = 0.5;
	// 	ex->lightcolor[2] = 0.5;
	// 	ex->ent.angles[1] = (float)(randk() % 360);
	// 	ex->ent.model = cl_mod_explo4;

	// 	if (frandk() < 0.5)
	// 	{
	// 		ex->baseframe = 15;
	// 	}

	// 	ex->frames = 15;
	// 	EXPLOSION_PARTICLES(pos);
	// 	S_StartSound(pos, 0, 0, cl_sfx_rockexp, 1, ATTN_NORM, 0);
	// 	break;

	case shared.TE_EXPLOSION1_BIG,
		shared.TE_EXPLOSION1_NP,
		shared.TE_EXPLOSION1,
		shared.TE_ROCKET_EXPLOSION,
		shared.TE_ROCKET_EXPLOSION_WATER:
		pos := msg.ReadPos()
		ex := T.allocExplosion()
		copy(ex.ent.Origin[:], pos)
		ex.etype = ex_poly
		ex.ent.Flags = shared.RF_FULLBRIGHT | shared.RF_NOSHADOW
		ex.start = float32(T.cl.frame.servertime) - 100.0
		ex.light = 350
		ex.lightcolor[0] = 1.0
		ex.lightcolor[1] = 0.5
		ex.lightcolor[2] = 0.5
		ex.ent.Angles[1] = float32(shared.Randk() % 360)

		if mtype != shared.TE_EXPLOSION1_BIG {
			ex.ent.Model = T.cl_mod_explo4
		} else {
			ex.ent.Model = T.cl_mod_explo4_big
		}

		if shared.Frandk() < 0.5 {
			ex.baseframe = 15
		}

		ex.frames = 15

		if (mtype != shared.TE_EXPLOSION1_BIG) && (mtype != shared.TE_EXPLOSION1_NP) {
			T.explosionParticles(pos)
		}

	// 	if (type == TE_ROCKET_EXPLOSION_WATER)
	// 	{
	// 		S_StartSound(pos, 0, 0, cl_sfx_watrexp, 1, ATTN_NORM, 0);
	// 	}
	// 	else
	// 	{
	// 		S_StartSound(pos, 0, 0, cl_sfx_rockexp, 1, ATTN_NORM, 0);
	// 	}

	// 	break;

	// case TE_BFG_EXPLOSION:
	// 	MSG_ReadPos(&net_message, pos);
	// 	ex = CL_AllocExplosion();
	// 	VectorCopy(pos, ex->ent.origin);
	// 	ex->type = ex_poly;
	// 	ex->ent.flags = RF_FULLBRIGHT | RF_NOSHADOW;
	// 	ex->start = cl.frame.servertime - 100.0f;
	// 	ex->light = 350;
	// 	ex->lightcolor[0] = 0.0;
	// 	ex->lightcolor[1] = 1.0;
	// 	ex->lightcolor[2] = 0.0;
	// 	ex->ent.model = cl_mod_bfg_explo;
	// 	ex->ent.flags |= RF_TRANSLUCENT;
	// 	ex->ent.alpha = 0.30f;
	// 	ex->frames = 4;
	// 	break;

	// case TE_BFG_BIGEXPLOSION:
	// 	MSG_ReadPos(&net_message, pos);
	// 	CL_BFGExplosionParticles(pos);
	// 	break;

	// case TE_BFG_LASER:
	// 	CL_ParseLaser(0xd0d1d2d3);
	// 	break;

	// case TE_BUBBLETRAIL:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadPos(&net_message, pos2);
	// 	CL_BubbleTrail(pos, pos2);
	// 	break;

	// case TE_PARASITE_ATTACK:
	// case TE_MEDIC_CABLE_ATTACK:
	// 	CL_ParseBeam(cl_mod_parasite_segment);
	// 	break;

	// case TE_BOSSTPORT: /* boss teleporting to station */
	// 	MSG_ReadPos(&net_message, pos);
	// 	CL_BigTeleportParticles(pos);
	// 	S_StartSound(pos, 0, 0, S_RegisterSound(
	// 				"misc/bigtele.wav"), 1, ATTN_NONE, 0);
	// 	break;

	// case TE_GRAPPLE_CABLE:
	// 	CL_ParseBeam2(cl_mod_grapple_cable);
	// 	break;

	// case TE_WELDING_SPARKS:
	// 	cnt = MSG_ReadByte(&net_message);
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	color = MSG_ReadByte(&net_message);
	// 	CL_ParticleEffect2(pos, dir, color, cnt);

	// 	ex = CL_AllocExplosion();
	// 	VectorCopy(pos, ex->ent.origin);
	// 	ex->type = ex_flash;
	// 	ex->ent.flags = RF_BEAM;
	// 	ex->start = cl.frame.servertime - 0.1f;
	// 	ex->light = 100 + (float)(randk() % 75);
	// 	ex->lightcolor[0] = 1.0f;
	// 	ex->lightcolor[1] = 1.0f;
	// 	ex->lightcolor[2] = 0.3f;
	// 	ex->ent.model = cl_mod_flash;
	// 	ex->frames = 2;
	// 	break;

	// case TE_GREENBLOOD:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	CL_ParticleEffect2(pos, dir, 0xdf, 30);
	// 	break;

	// case TE_TUNNEL_SPARKS:
	// 	cnt = MSG_ReadByte(&net_message);
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	color = MSG_ReadByte(&net_message);
	// 	CL_ParticleEffect3(pos, dir, color, cnt);
	// 	break;

	// case TE_BLASTER2:
	// case TE_FLECHETTE:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);

	// 	if (type == TE_BLASTER2)
	// 	{
	// 		CL_BlasterParticles2(pos, dir, 0xd0);
	// 	}
	// 	else
	// 	{
	// 		CL_BlasterParticles2(pos, dir, 0x6f);
	// 	}

	// 	ex = CL_AllocExplosion();
	// 	VectorCopy(pos, ex->ent.origin);
	// 	ex->ent.angles[0] = (float)acos(dir[2]) / M_PI * 180;

	// 	if (dir[0])
	// 	{
	// 		ex->ent.angles[1] = (float)atan2(dir[1], dir[0]) / M_PI * 180;
	// 	}
	// 	else if (dir[1] > 0)
	// 	{
	// 		ex->ent.angles[1] = 90;
	// 	}

	// 	else if (dir[1] < 0)
	// 	{
	// 		ex->ent.angles[1] = 270;
	// 	}
	// 	else
	// 	{
	// 		ex->ent.angles[1] = 0;
	// 	}

	// 	ex->type = ex_misc;
	// 	ex->ent.flags = RF_FULLBRIGHT | RF_TRANSLUCENT;

	// 	if (type == TE_BLASTER2)
	// 	{
	// 		ex->ent.skinnum = 1;
	// 	}
	// 	else /* flechette */
	// 	{
	// 		ex->ent.skinnum = 2;
	// 	}

	// 	ex->start = cl.frame.servertime - 100.0f;
	// 	ex->light = 150;

	// 	if (type == TE_BLASTER2)
	// 	{
	// 		ex->lightcolor[1] = 1;
	// 	}
	// 	else
	// 	{
	// 		/* flechette */
	// 		ex->lightcolor[0] = 0.19f;
	// 		ex->lightcolor[1] = 0.41f;
	// 		ex->lightcolor[2] = 0.75f;
	// 	}

	// 	ex->ent.model = cl_mod_explode;
	// 	ex->frames = 4;
	// 	S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 	break;

	// case TE_LIGHTNING:
	// 	ent = CL_ParseLightning(cl_mod_lightning);
	// 	S_StartSound(NULL, ent, CHAN_WEAPON, cl_sfx_lightning,
	// 		1, ATTN_NORM, 0);
	// 	break;

	// case TE_DEBUGTRAIL:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadPos(&net_message, pos2);
	// 	CL_DebugTrail(pos, pos2);
	// 	break;

	// case TE_PLAIN_EXPLOSION:
	// 	MSG_ReadPos(&net_message, pos);

	// 	ex = CL_AllocExplosion();
	// 	VectorCopy(pos, ex->ent.origin);
	// 	ex->type = ex_poly;
	// 	ex->ent.flags = RF_FULLBRIGHT | RF_NOSHADOW;
	// 	ex->start = cl.frame.servertime - 100.0f;
	// 	ex->light = 350;
	// 	ex->lightcolor[0] = 1.0;
	// 	ex->lightcolor[1] = 0.5;
	// 	ex->lightcolor[2] = 0.5;
	// 	ex->ent.angles[1] = randk() % 360;
	// 	ex->ent.model = cl_mod_explo4;

	// 	if (frandk() < 0.5)
	// 	{
	// 		ex->baseframe = 15;
	// 	}

	// 	ex->frames = 15;

	// 	S_StartSound(pos, 0, 0, cl_sfx_rockexp, 1, ATTN_NORM, 0);

	// 	break;

	// case TE_FLASHLIGHT:
	// 	MSG_ReadPos(&net_message, pos);
	// 	ent = MSG_ReadShort(&net_message);
	// 	CL_Flashlight(ent, pos);
	// 	break;

	// case TE_FORCEWALL:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadPos(&net_message, pos2);
	// 	color = MSG_ReadByte(&net_message);
	// 	CL_ForceWall(pos, pos2, color);
	// 	break;

	// case TE_HEATBEAM:
	// 	CL_ParsePlayerBeam(cl_mod_heatbeam);
	// 	break;

	// case TE_MONSTER_HEATBEAM:
	// 	CL_ParsePlayerBeam(cl_mod_monster_heatbeam);
	// 	break;

	// case TE_HEATBEAM_SPARKS:
	// 	cnt = 50;
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	r = 8;
	// 	magnitude = 60;
	// 	color = r & 0xff;
	// 	CL_ParticleSteamEffect(pos, dir, color, cnt, magnitude);
	// 	S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 	break;

	// case TE_HEATBEAM_STEAM:
	// 	cnt = 20;
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	color = 0xe0;
	// 	magnitude = 60;
	// 	CL_ParticleSteamEffect(pos, dir, color, cnt, magnitude);
	// 	S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 	break;

	// case TE_STEAM:
	// 	CL_ParseSteam();
	// 	break;

	// case TE_BUBBLETRAIL2:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadPos(&net_message, pos2);
	// 	CL_BubbleTrail2(pos, pos2, 8);
	// 	S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 	break;

	// case TE_MOREBLOOD:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	CL_ParticleEffect(pos, dir, 0xe8, 250);
	// 	break;

	// case TE_CHAINFIST_SMOKE:
	// 	dir[0] = 0;
	// 	dir[1] = 0;
	// 	dir[2] = 1;
	// 	MSG_ReadPos(&net_message, pos);
	// 	CL_ParticleSmokeEffect(pos, dir, 0, 20, 20);
	// 	break;

	// case TE_ELECTRIC_SPARKS:
	// 	MSG_ReadPos(&net_message, pos);
	// 	MSG_ReadDir(&net_message, dir);
	// 	CL_ParticleEffect(pos, dir, 0x75, 40);
	// 	S_StartSound(pos, 0, 0, cl_sfx_lashit, 1, ATTN_NORM, 0);
	// 	break;

	// case TE_TRACKER_EXPLOSION:
	// 	MSG_ReadPos(&net_message, pos);
	// 	CL_ColorFlash(pos, 0, 150, -1, -1, -1);
	// 	CL_ColorExplosionParticles(pos, 0, 1);
	// 	S_StartSound(pos, 0, 0, cl_sfx_disrexp, 1, ATTN_NORM, 0);
	// 	break;

	// case TE_TELEPORT_EFFECT:
	// case TE_DBALL_GOAL:
	// 	MSG_ReadPos(&net_message, pos);
	// 	CL_TeleportParticles(pos);
	// 	break;

	// case TE_WIDOWBEAMOUT:
	// 	CL_ParseWidow();
	// 	break;

	// case TE_NUKEBLAST:
	// 	CL_ParseNuke();
	// 	break;

	// case TE_WIDOWSPLASH:
	// 	MSG_ReadPos(&net_message, pos);
	// 	CL_WidowSplash(pos);
	// 	break;

	default:
		return T.common.Com_Error(shared.ERR_DROP, "CL_ParseTEnt: bad type %v", mtype)
	}
	return nil
}

func (T *qClient) addExplosions() {

	for _, ex := range T.cl_explosions {
		if ex.etype == ex_free {
			continue
		}

		frac := (float32(T.cl.time) - ex.start) / 100.0
		f := int(math.Floor(float64(frac)))

		ent := &ex.ent

		switch ex.etype {
		case ex_mflash:

			if f >= ex.frames-1 {
				ex.etype = ex_free
			}

			break
		case ex_misc:

			if f >= ex.frames-1 {
				ex.etype = ex_free
				break
			}

			ent.Alpha = 1.0 - float32(frac)/float32(ex.frames-1)
			break
		case ex_flash:

			if f >= 1 {
				ex.etype = ex_free
				break
			}

			ent.Alpha = 1.0
			break
		case ex_poly:

			if f >= ex.frames-1 {
				ex.etype = ex_free
				break
			}

			ent.Alpha = (16.0 - float32(f)) / 16.0

			if f < 10 {
				ent.Skinnum = (f >> 1)

				if ent.Skinnum < 0 {
					ent.Skinnum = 0
				}
			} else {
				ent.Flags |= shared.RF_TRANSLUCENT

				if f < 13 {
					ent.Skinnum = 5
				} else {
					ent.Skinnum = 6
				}
			}

			break
		case ex_poly2:

			if f >= ex.frames-1 {
				ex.etype = ex_free
				break
			}

			ent.Alpha = (5.0 - float32(f)) / 5.0
			ent.Skinnum = 0
			ent.Flags |= shared.RF_TRANSLUCENT
			break
		default:
			break
		}

		if ex.etype == ex_free {
			continue
		}

		if ex.light != 0 {
			T.addLight(ent.Origin[:], ex.light*ent.Alpha,
				ex.lightcolor[0], ex.lightcolor[1], ex.lightcolor[2])
		}

		copy(ent.Oldorigin[:], ent.Origin[:])

		if f < 0 {
			f = 0
		}

		ent.Frame = ex.baseframe + f + 1
		ent.Oldframe = ex.baseframe + f
		ent.Backlerp = 1.0 - T.cl.lerpfrac

		T.addEntity(*ent)
	}
}

func (T *qClient) addTEnts() {
	// CL_AddBeams();
	// CL_AddPlayerBeams();
	T.addExplosions()
	// CL_AddLasers();
	// CL_ProcessSustain();
}
