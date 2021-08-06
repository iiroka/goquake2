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
 * Misc. utility functions for the game logic.
 *
 * =======================================================================
 */
package game

import (
	"fmt"
	"reflect"
)

/*
 * Searches all active entities for the next
 * one that holds the matching string at fieldofs
 * (use the FOFS() macro) in the structure.
 *
 * Searches beginning at the edict after from, or
 * the beginning. If NULL, NULL will be returned
 * if the end of the list is reached.
 */
func (G *qGame) gFind(from *edict_t, fname, match string) *edict_t {

	var index int = 0
	if from != nil {
		index = from.index + 1
	}

	if len(match) == 0 {
		return nil
	}

	for ; index < G.num_edicts; index++ {
		if !G.g_edicts[index].inuse {
			continue
		}

		b := reflect.ValueOf(&G.g_edicts[index]).Elem()
		f := b.FieldByName(fname)

		if !f.IsValid() || f.Kind() != reflect.String {
			continue
		}

		s := f.String()
		if s == match {
			return &G.g_edicts[index]
		}
	}

	return nil
}

/*
 * This is just a convenience function
 * for printing vectors
 */
func vtos(v []float32) string {
	return fmt.Sprintf("(%v %v %v)", int(v[0]), int(v[1]), int(v[2]))
}

func G_InitEdict(e *edict_t, index int) {
	e.inuse = true
	e.Classname = "noclass"
	e.gravity = 1.0
	e.s.Number = index
}

/*
 * Either finds a free edict, or allocates a
 * new one.  Try to avoid reusing an entity
 * that was recently freed, because it can
 * cause the client to think the entity
 * morphed into something else instead of
 * being removed and recreated, which can
 * cause interpolated angles and bad trails.
 */
func (G *qGame) gSpawn() (*edict_t, error) {
	//  int i;
	//  edict_t *e;

	//  e = &g_edicts[(int)maxclients->value + 1];

	for i := G.maxclients.Int() + 1; i < G.num_edicts; i++ {
		e := &G.g_edicts[i]
		/* the first couple seconds of
		server time can involve a lot of
		freeing and allocating, so relax
		the replacement policy */
		if !e.inuse && ((e.freetime < 2) || (G.level.time-e.freetime > 0.5)) {
			G_InitEdict(e, i)
			return e, nil
		}
	}

	if G.num_edicts == G.game.maxentities {
		return nil, G.gi.Error("ED_Alloc: no free edicts")
	}

	e := &G.g_edicts[G.num_edicts]
	G_InitEdict(e, G.num_edicts)
	G.num_edicts++
	return e, nil
}
