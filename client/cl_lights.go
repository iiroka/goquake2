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
 * This file implements all client side lighting
 *
 * =======================================================================
 */
package client

import "goquake2/shared"

type clightstyle_t struct {
	length int
	value  [3]float32
	mmap   [shared.MAX_QPATH]float32
}

func (T *qClient) clearLightStyles() {
	for i := range T.cl_lightstyle {
		T.cl_lightstyle[i].length = 0
		T.cl_lightstyle[i].value[0] = 0
		T.cl_lightstyle[i].value[1] = 0
		T.cl_lightstyle[i].value[2] = 0
		for j := range T.cl_lightstyle[i].mmap {
			T.cl_lightstyle[i].mmap[j] = 0
		}
	}
	T.lastofs = -1
}

func (T *qClient) runLightStyles() {

	ofs := T.cl.time / 100

	if ofs == T.lastofs {
		return
	}

	T.lastofs = ofs

	for i, ls := range T.cl_lightstyle {
		var v float32
		if ls.length == 0 {
			v = 0
		} else if ls.length == 1 {
			v = ls.mmap[0]
		} else {
			v = ls.mmap[ofs%ls.length]
		}
		T.cl_lightstyle[i].value[0] = v
		T.cl_lightstyle[i].value[1] = v
		T.cl_lightstyle[i].value[2] = v
	}
}

func (T *qClient) setLightstyle(i int) {

	s := T.cl.configstrings[i+shared.CS_LIGHTS]

	T.cl_lightstyle[i].length = len(s)

	for k, ch := range s {
		T.cl_lightstyle[i].mmap[k] = float32(ch-'a') / float32('m'-'a')
	}
}

func (T *qClient) addLightStyles() {

	for i, ls := range T.cl_lightstyle {
		T.addLightStyle(i, ls.value[0], ls.value[1], ls.value[2])
	}
}
