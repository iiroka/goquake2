/*
 * Copyright (C) 1997-2001 Id Software, Inc.
 * Copyright (C) 2016-2017 Daniel Gibson
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
 * Lightmaps and dynamic lighting
 *
 * =======================================================================
 */
package gl3

import "goquake2/shared"

/*
 * Combine and scale multiple lightmaps into the floating format in blocklights
 */
func (T *qGl3) buildLightMap(surf *msurface_t, offsetInLMbuf, stride int) error {
	//  int smax, tmax;
	//  int r, g, b, a, max;
	//  int i, j, size, map, numMaps;
	//  byte *lightmap;

	if (surf.texinfo.flags &
		(shared.SURF_SKY | shared.SURF_TRANS33 | shared.SURF_TRANS66 | shared.SURF_WARP)) != 0 {
		return T.ri.Sys_Error(shared.ERR_DROP, "GL3_BuildLightMap called for non-lit surface")
	}

	smax := (int(surf.extents[0]) >> 4) + 1
	tmax := (int(surf.extents[1]) >> 4) + 1
	size := smax * tmax

	stride -= (smax << 2)

	if size > 34*34*3 {
		return T.ri.Sys_Error(shared.ERR_DROP, "Bad s_blocklights size")
	}

	// count number of lightmaps surf actually has
	numMaps := MAX_LIGHTMAPS_PER_SURFACE
	for i := range surf.styles {
		if surf.styles[i] == 255 {
			numMaps = i
			break
		}
	}

	if surf.samples == nil {
		// no lightmap samples? set at least one lightmap to fullbright, rest to 0 as normal

		if numMaps == 0 {
			numMaps = 1 // make sure at least one lightmap is set to fullbright
		}

		for mmap := 0; mmap < MAX_LIGHTMAPS_PER_SURFACE; mmap++ {
			// we always create 4 (MAX_LIGHTMAPS_PER_SURFACE) lightmaps.
			// if surf has less (numMaps < 4), the remaining ones are zeroed out.
			// this makes sure that all 4 lightmap textures in gl3state.lightmap_textureIDs[i] have the same layout
			// and the shader can use the same texture coordinates for all of them

			c := 0
			if mmap < numMaps {
				c = 255
			}
			dest_i := offsetInLMbuf

			for i := 0; i < tmax; i++ {
				for j := 0; i < 4*smax; j++ {
					T.gl3_lms.lightmap_buffers[mmap][dest_i+i] = byte(c)
				}
				dest_i += 4*smax + stride
			}
		}

		return nil
	}

	/* add all the lightmaps */

	// Note: dynamic lights aren't handled here anymore, they're handled in the shader

	// as we don't apply scale here anymore, nor blend the numMaps lightmaps together,
	// the code has gotten a lot easier and we can copy directly from surf->samples to dest
	// without converting to float first etc

	lightmap := surf.samples
	lm_i := 0

	for mmap := 0; mmap < numMaps; mmap++ {
		// 	 byte* dest = gl3_lms.lightmap_buffers[map] + offsetInLMbuf;
		dest_i := offsetInLMbuf
		idxInLightmap := 0
		for i := 0; i < tmax; i++ {
			// 	 for (i = 0; i < tmax; i++, dest += stride)
			// 	 {
			for j := 0; j < smax; j++ {
				r := lightmap[idxInLightmap*3+lm_i+0]
				g := lightmap[idxInLightmap*3+lm_i+1]
				b := lightmap[idxInLightmap*3+lm_i+2]

				/* determine the brightest of the three color components */
				var max byte
				if r > g {
					max = r
				} else {
					max = g
				}
				if b > max {
					max = b
				}

				/* alpha is ONLY used for the mono lightmap case. For this
				reason we set it to the brightest of the color components
				so that things don't get too dim. */
				a := max

				T.gl3_lms.lightmap_buffers[mmap][dest_i+0] = r
				T.gl3_lms.lightmap_buffers[mmap][dest_i+1] = g
				T.gl3_lms.lightmap_buffers[mmap][dest_i+2] = b
				T.gl3_lms.lightmap_buffers[mmap][dest_i+3] = a

				dest_i += 4
				idxInLightmap++
			}
			//  }
			dest_i += stride
		}

		lm_i += size + 3
	}

	for mmap := numMaps; mmap < MAX_LIGHTMAPS_PER_SURFACE; mmap++ {
		// like above, fill up remaining lightmaps with 0
		dest_i := offsetInLMbuf

		for i := 0; i < tmax; i++ {
			for j := 0; j < 4*smax; j++ {
				T.gl3_lms.lightmap_buffers[mmap][dest_i+i] = 0
			}
			dest_i += 4*smax + stride
		}
	}
	return nil
}
