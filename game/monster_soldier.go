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
 * Soldier aka "Guard". This is the most complex enemy in Quake 2, since
 * it uses all AI features (dodging, sight, crouching, etc) and comes
 * in a myriad of variants.
 *
 * =======================================================================
 */
package game

import (
	"goquake2/game/soldier"
	"goquake2/shared"
)

func soldier_idle(self *edict_t, G *qGame) {
	if self == nil || G == nil {
		return
	}

	// if (random() > 0.8) {
	// 	gi.sound(self, CHAN_VOICE, sound_idle, 1, ATTN_IDLE, 0);
	// }
}

func soldier_cock(self *edict_t, G *qGame) {
	if self == nil || G == nil {
		return
	}

	// if (self->s.frame == FRAME_stand322)
	// {
	// 	gi.sound(self, CHAN_WEAPON, sound_cock, 1, ATTN_IDLE, 0);
	// }
	// else
	// {
	// 	gi.sound(self, CHAN_WEAPON, sound_cock, 1, ATTN_NORM, 0);
	// }
}

var soldier_frames_stand1 = []mframe_t{
	{ai_stand, 0, soldier_idle},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},

	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},

	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil}}

var soldier_move_stand1 = mmove_t{
	soldier.FRAME_stand101,
	soldier.FRAME_stand130,
	soldier_frames_stand1,
	nil,
}

var soldier_frames_stand3 = []mframe_t{
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},

	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},

	{ai_stand, 0, nil},
	{ai_stand, 0, soldier_cock},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},

	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil},
	{ai_stand, 0, nil}}

var soldier_move_stand3 = mmove_t{
	soldier.FRAME_stand301,
	soldier.FRAME_stand339,
	soldier_frames_stand3,
	nil,
}

func soldier_stand(self *edict_t, G *qGame) {
	if self == nil || G == nil {
		return
	}

	if (self.monsterinfo.currentmove == &soldier_move_stand3) ||
		(shared.Frandk() < 0.8) {
		self.monsterinfo.currentmove = &soldier_move_stand1
	} else {
		self.monsterinfo.currentmove = &soldier_move_stand3
	}
}

func soldier_walk1_random(self *edict_t, G *qGame) {
	if self == nil || G == nil {
		return
	}

	if shared.Frandk() > 0.1 {
		self.monsterinfo.nextframe = soldier.FRAME_walk101
	}
}

var soldier_frames_walk1 = []mframe_t{
	{ai_walk, 3, nil},
	{ai_walk, 6, nil},
	{ai_walk, 2, nil},
	{ai_walk, 2, nil},
	{ai_walk, 2, nil},
	{ai_walk, 1, nil},
	{ai_walk, 6, nil},
	{ai_walk, 5, nil},
	{ai_walk, 3, nil},
	{ai_walk, -1, soldier_walk1_random},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
	{ai_walk, 0, nil},
}

var soldier_move_walk1 = mmove_t{
	soldier.FRAME_walk101,
	soldier.FRAME_walk133,
	soldier_frames_walk1,
	nil,
}

var soldier_frames_walk2 = []mframe_t{
	{ai_walk, 4, nil},
	{ai_walk, 4, nil},
	{ai_walk, 9, nil},
	{ai_walk, 8, nil},
	{ai_walk, 5, nil},
	{ai_walk, 1, nil},
	{ai_walk, 3, nil},
	{ai_walk, 7, nil},
	{ai_walk, 6, nil},
	{ai_walk, 7, nil},
}

var soldier_move_walk2 = mmove_t{
	soldier.FRAME_walk209,
	soldier.FRAME_walk218,
	soldier_frames_walk2,
	nil,
}

func soldier_walk(self *edict_t, G *qGame) {
	if self == nil || G == nil {
		return
	}

	if shared.Frandk() < 0.5 {
		self.monsterinfo.currentmove = &soldier_move_walk1
	} else {
		self.monsterinfo.currentmove = &soldier_move_walk2
	}
}

func (G *qGame) spMonsterSoldierX(self *edict_t) {
	if self == nil {
		return
	}

	soldier_move_stand1.endfunc = soldier_stand
	soldier_move_stand3.endfunc = soldier_stand

	self.s.Modelindex = G.gi.Modelindex("models/monsters/soldier/tris.md2")
	self.monsterinfo.scale = soldier.MODEL_SCALE
	self.mins = [3]float32{-16, -16, -24}
	self.maxs = [3]float32{16, 16, 32}
	self.movetype = MOVETYPE_STEP
	self.solid = shared.SOLID_BBOX

	// sound_idle = gi.soundindex("soldier/solidle1.wav");
	// sound_sight1 = gi.soundindex("soldier/solsght1.wav");
	// sound_sight2 = gi.soundindex("soldier/solsrch1.wav");
	// sound_cock = gi.soundindex("infantry/infatck3.wav");

	self.Mass = 100

	// self->pain = soldier_pain;
	// self->die = soldier_die;

	self.monsterinfo.stand = soldier_stand
	self.monsterinfo.walk = soldier_walk
	// self->monsterinfo.run = soldier_run;
	// self->monsterinfo.dodge = soldier_dodge;
	// self->monsterinfo.attack = soldier_attack;
	// self->monsterinfo.melee = NULL;
	// self->monsterinfo.sight = soldier_sight;

	G.gi.Linkentity(self)

	self.monsterinfo.stand(self, G)

	G.walkmonster_start(self)
}

/*
 * QUAKED monster_soldier (1 .5 0) (-16 -16 -24) (16 16 32) Ambush Trigger_Spawn Sight
 */
func spMonsterSoldier(self *edict_t, G *qGame) error {
	if self == nil {
		return nil
	}

	if G.deathmatch.Bool() {
		G.gFreeEdict(self)
		return nil
	}

	G.spMonsterSoldierX(self)

	// sound_pain = gi.soundindex("soldier/solpain1.wav");
	// sound_death = gi.soundindex("soldier/soldeth1.wav");
	G.gi.Soundindex("soldier/solatck1.wav")

	self.s.Skinnum = 2
	self.Health = 30
	self.gib_health = -30
	return nil
}
