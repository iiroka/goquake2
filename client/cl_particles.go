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
 * This file implements all generic particle stuff
 *
 * =======================================================================
 */
package client

import "goquake2/shared"

func (T *qClient) clearParticles() {

	T.free_particles = &T.particles[0]
	T.active_particles = nil

	for i := 0; i < len(T.particles)-1; i++ {
		T.particles[i].next = &T.particles[i+1]
	}

	T.particles[len(T.particles)-1].next = nil
}

func (T *qClient) particleEffect(org, dir []float32, color, count int) {

	for i := 0; i < count; i++ {
		if T.free_particles == nil {
			return
		}

		p := T.free_particles
		T.free_particles = p.next
		p.next = T.active_particles
		T.active_particles = p

		p.time = float32(T.cl.time)
		p.color = float32(color + (shared.Randk() & 7))
		d := shared.Randk() & 31

		for j := 0; j < 3; j++ {
			p.org[j] = org[j] + float32((shared.Randk()&7)-4) + float32(d)*dir[j]
			p.vel[j] = shared.Crandk() * 20
		}

		p.accel[0] = 0
		p.accel[1] = 0
		p.accel[2] = -PARTICLE_GRAVITY + 0.2
		p.alpha = 1.0

		p.alphavel = -1.0 / (0.5 + shared.Frandk()*0.3)
	}
}

func (T *qClient) addParticles() {

	var tail *cparticle_t = nil
	var next *cparticle_t
	var active *cparticle_t = nil
	for p := T.active_particles; p != nil; p = next {
		next = p.next

		var time, alpha float32
		if p.alphavel != INSTANT_PARTICLE {
			time = (float32(T.cl.time) - p.time) * 0.001
			alpha = p.alpha + time*p.alphavel

			if alpha <= 0 {
				/* faded out */
				p.next = T.free_particles
				T.free_particles = p
				continue
			}
		} else {
			time = 0.0
			alpha = p.alpha
		}

		p.next = nil

		if tail == nil {
			active = p
			tail = p
		} else {
			tail.next = p
			tail = p
		}

		if alpha > 1.0 {
			alpha = 1
		}

		color := int(p.color)
		time2 := time * time

		var org [3]float32
		org[0] = p.org[0] + p.vel[0]*time + p.accel[0]*time2
		org[1] = p.org[1] + p.vel[1]*time + p.accel[1]*time2
		org[2] = p.org[2] + p.vel[2]*time + p.accel[2]*time2

		T.addParticle(org, color, alpha)

		if p.alphavel == INSTANT_PARTICLE {
			p.alphavel = 0.0
			p.alpha = 0.0
		}
	}

	T.active_particles = active
}
