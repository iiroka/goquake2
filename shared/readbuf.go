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
 * Movement message (forward, backward, left, right, etc) handling.
 *
 * =======================================================================
 */
package shared

import "strings"

type QReadbuf struct {
	data      []byte
	readcount int
}

func QReadbufCreate(data []byte) *QReadbuf {
	return &QReadbuf{data, 0}
}

func (msg *QReadbuf) Size() int {
	return len(msg.data)
}

func (msg *QReadbuf) Count() int {
	return msg.readcount
}

func (msg *QReadbuf) IsEmpty() bool {
	return msg.readcount >= len(msg.data)
}

func (msg *QReadbuf) IsOver() bool {
	return msg.readcount > len(msg.data)
}

func (msg *QReadbuf) BeginReading() {
	msg.readcount = 0
}

func (msg *QReadbuf) ReadByte() int {

	var c int
	if msg.readcount+1 > len(msg.data) {
		c = -1
	} else {
		c = int(uint8(msg.data[msg.readcount]))
	}
	msg.readcount += 1
	return c
}

func (msg *QReadbuf) ReadChar() int {

	var c int
	if msg.readcount+1 > len(msg.data) {
		c = -1
	} else {
		c = int(int8(uint8(msg.data[msg.readcount])))
	}
	msg.readcount += 1
	return c
}

func (msg *QReadbuf) ReadShort() int {

	var c int
	if msg.readcount+2 > len(msg.data) {
		c = -1
	} else {
		c = int(int16(uint32(msg.data[msg.readcount]) |
			(uint32(msg.data[msg.readcount+1]) << 8)))
	}
	msg.readcount += 2
	return c
}

func (msg *QReadbuf) ReadLong() int {

	var c int
	if msg.readcount+4 > len(msg.data) {
		c = -1
	} else {
		c = int(int32(uint32(msg.data[msg.readcount]) |
			(uint32(msg.data[msg.readcount+1]) << 8) |
			(uint32(msg.data[msg.readcount+2]) << 16) |
			(uint32(msg.data[msg.readcount+3]) << 24)))
	}
	msg.readcount += 4
	return c
}

func (msg *QReadbuf) ReadString() string {

	var r strings.Builder
	for {
		c := msg.ReadByte()
		if (c == -1) || (c == 0) {
			break
		}

		r.WriteByte(byte(c))
	}
	return r.String()
}

func (msg *QReadbuf) ReadStringLine() string {

	var r strings.Builder
	for {
		c := msg.ReadByte()
		if (c == -1) || (c == 0) || (c == '\n') {
			break
		}

		r.WriteByte(byte(c))
	}
	return r.String()
}

func (msg *QReadbuf) ReadCoord() float32 {
	return float32(msg.ReadShort()) * 0.125
}

func (msg *QReadbuf) ReadPos() []float32 {
	return []float32{
		float32(msg.ReadShort()) * 0.125,
		float32(msg.ReadShort()) * 0.125,
		float32(msg.ReadShort()) * 0.125,
	}
}

func (msg *QReadbuf) ReadAngle() float32 {
	return float32(msg.ReadChar()) * 1.40625
}

func (msg *QReadbuf) ReadAngle16() float32 {
	return SHORT2ANGLE(msg.ReadShort())
}

func (msg *QReadbuf) ReadData(data []byte, len int) {

	for i := 0; i < len; i++ {
		data[i] = byte(msg.ReadByte())
	}
}
