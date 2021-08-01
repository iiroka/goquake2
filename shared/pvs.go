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
 * The PVS Decompress
 *
 * =======================================================================
 */
package shared

import "math"

/*
 ===================
 Mod_DecompressVis
 ===================
*/
func Mod_DecompressVis(in []byte, offset, row int) []byte {
	//  YQ2_ALIGNAS_TYPE(int) static byte decompressed[MAX_MAP_LEAFS / 8];
	//  int c;
	//  byte *out;
	decompressed := make([]byte, row)
	index := 0

	if in == nil {
		/* no vis info, so make all visible */
		for row > 0 {
			decompressed[index] = 0xff
			index++
			row--
		}

		return decompressed
	}

	for index < row {
		if in[offset] != 0 {
			decompressed[index] = in[offset]
			index++
			offset++
			continue
		}

		c := in[offset+1]
		offset += 2

		for c > 0 {
			decompressed[index] = 0
			index++
			c--
		}
	}

	return decompressed
}

func Mod_RadiusFromBounds(mins, maxs []float32) float32 {

	var corner [3]float32
	for i := 0; i < 3; i++ {
		if math.Abs(float64(mins[i])) > math.Abs(float64(maxs[i])) {
			corner[i] = float32(math.Abs(float64(mins[i])))
		} else {
			corner[i] = float32(math.Abs(float64(maxs[i])))
		}
	}

	return VectorLength(corner[:])
}
