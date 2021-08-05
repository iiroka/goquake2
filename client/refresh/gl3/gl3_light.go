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

const DLIGHT_CUTOFF = 64

// bit: 1 << i for light number i, will be or'ed into msurface_t::dlightbits if surface is affected by this light
func (T *qGl3) markLights(light *shared.Dlight_t, bit int, anode mnode_or_leaf) {

	if anode.Contents() != -1 {
		return
	}

	node := anode.(*mnode_t)
	splitplane := node.plane
	dist := shared.DotProduct(light.Origin[:], splitplane.Normal[:]) - splitplane.Dist

	if dist > light.Intensity-DLIGHT_CUTOFF {
		T.markLights(light, bit, node.children[0])
		return
	}

	if dist < -light.Intensity+DLIGHT_CUTOFF {
		T.markLights(light, bit, node.children[1])
		return
	}

	/* mark the polygons */
	for i := 0; i < int(node.numsurfaces); i++ {
		surf := T.gl3_worldmodel.surfaces[int(node.firstsurface)+i]
		if surf.dlightframe != T.r_dlightframecount {
			surf.dlightbits = 0
			surf.dlightframe = T.r_dlightframecount
		}

		dist = shared.DotProduct(light.Origin[:], surf.plane.Normal[:]) - surf.plane.Dist
		var sidebit int
		if dist >= 0 {
			sidebit = 0
		} else {
			sidebit = SURF_PLANEBACK
		}

		if (surf.flags & SURF_PLANEBACK) != sidebit {
			continue
		}

		surf.dlightbits |= bit
	}

	T.markLights(light, bit, node.children[0])
	T.markLights(light, bit, node.children[1])
}

func (T *qGl3) pushDlights() {

	/* because the count hasn't advanced yet for this frame */
	T.r_dlightframecount = T.gl3_framecount + 1

	T.gl3state.uniLightsData.setNumDynLights(len(T.gl3_newrefdef.Dlights))

	for i, l := range T.gl3_newrefdef.Dlights {
		udl := &T.gl3state.uniLightsData.dynLights[i]
		T.markLights(&T.gl3_newrefdef.Dlights[i], 1<<i, &T.gl3_worldmodel.nodes[0])

		udl.setOrigin(l.Origin[:])
		udl.setColor(l.Color[:])
		udl.setIntensity(l.Intensity)
	}

	// assert(MAX_DLIGHTS == 32 && "If MAX_DLIGHTS changes, remember to adjust the uniform buffer definition in the shader!")

	// if i < MAX_DLIGHTS {
	// 	memset(&gl3state.uniLightsData.dynLights[i], 0, (MAX_DLIGHTS-i)*sizeof(gl3state.uniLightsData.dynLights[0]))
	// }
	for i := 4 + len(T.gl3_newrefdef.Dlights)*gl3UniDynLight_Size; i < gl3UniLights_Size; i++ {
		T.gl3state.uniLightsData.data[i] = 0
	}

	T.updateUBOLights()
}

func (T *qGl3) lightPoint(p, color []float32) {
	// vec3_t end;
	// float r;
	// int lnum;
	// dlight_t *dl;
	// vec3_t dist;
	// float add;

	color[0] = 1.0
	color[1] = 1.0
	color[2] = 1.0
	// if (!gl3_worldmodel.lightdata || !currententity) {
	// 	color[0] = color[1] = color[2] = 1.0;
	// 	return;
	// }

	// end[0] = p[0];
	// end[1] = p[1];
	// end[2] = p[2] - 2048;

	// // TODO: don't just aggregate the color, but also save position of brightest+nearest light
	// //       for shadow position and maybe lighting on model?

	// r = RecursiveLightPoint(gl3_worldmodel->nodes, p, end);

	// if (r == -1)
	// {
	// 	VectorCopy(vec3_origin, color);
	// }
	// else
	// {
	// 	VectorCopy(pointcolor, color);
	// }

	// /* add dynamic lights */
	// dl = gl3_newrefdef.dlights;

	// for (lnum = 0; lnum < gl3_newrefdef.num_dlights; lnum++, dl++)
	// {
	// 	VectorSubtract(currententity->origin,
	// 			dl->origin, dist);
	// 	add = dl->intensity - VectorLength(dist);
	// 	add *= (1.0f / 256.0f);

	// 	if (add > 0)
	// 	{
	// 		VectorMA(color, add, dl->color, color);
	// 	}
	// }

	// VectorScale(color, r_modulate->value, color);
}

/*
 * Combine and scale multiple lightmaps into the floating format in blocklights
 */
func (T *qGl3) buildLightMap(surf *msurface_t, offsetInLMbuf, stride int) error {

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
				for j := 0; j < 4*smax; j++ {
					T.gl3_lms.lightmap_buffers[mmap][dest_i+j] = byte(c)
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

	for mmap := 0; mmap < numMaps; mmap++ {
		dest := T.gl3_lms.lightmap_buffers[mmap][offsetInLMbuf:]
		dest_i := 0
		lightmap := surf.samples[mmap*size*3:]
		idxInLightmap := 0
		for i := 0; i < tmax; i++ {
			for j := 0; j < smax; j++ {

				r := lightmap[idxInLightmap*3+0]
				g := lightmap[idxInLightmap*3+1]
				b := lightmap[idxInLightmap*3+2]

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
				if a < 255 {
					a = 255
				}

				dest[dest_i+0] = 255 // r
				dest[dest_i+1] = 255 // g
				dest[dest_i+2] = 255 // b
				dest[dest_i+3] = a   // a

				dest_i += 4
				idxInLightmap++
			}
			dest_i += stride
		}
	}

	for mmap := numMaps; mmap < MAX_LIGHTMAPS_PER_SURFACE; mmap++ {
		// like above, fill up remaining lightmaps with 0
		dest_i := offsetInLMbuf

		for i := 0; i < tmax; i++ {
			for j := 0; j < 4*smax; j++ {
				T.gl3_lms.lightmap_buffers[mmap][dest_i+j] = 255
			}
			dest_i += 4*smax + stride
		}
	}
	return nil
}
