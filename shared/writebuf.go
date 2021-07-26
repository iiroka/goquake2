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

import "log"

type QWritebuf struct {
	Allowoverflow bool /* if false, do a Com_Error */
	Overflowed    bool /* set to true if the buffer size failed */
	data          []byte
	Cursize       int
}

func QWritebufCreate(size int) *QWritebuf {
	wb := &QWritebuf{}
	wb.data = make([]byte, size)
	return wb
}

func (buf *QWritebuf) Data() []byte {
	return buf.data[0:buf.Cursize]
}

func (buf *QWritebuf) Clear() {
	buf.Cursize = 0
	buf.Overflowed = false
}

func (buf *QWritebuf) getSpace(length int) []byte {

	if buf.Cursize+length > len(buf.data) {
		if !buf.Allowoverflow {
			log.Fatal("SZ_GetSpace: overflow without allowoverflow set")
		}

		if length > len(buf.data) {
			log.Fatalf("SZ_GetSpace: %v is > full buffer size", length)
		}

		buf.Clear()
		buf.Overflowed = true
		println("SZ_GetSpace: overflow\n")
	}

	data := buf.data[buf.Cursize:]
	buf.Cursize += length

	return data
}

func (sb *QWritebuf) WriteChar(c int) {

	buf := sb.getSpace(1)
	buf[0] = byte(c)
}

func (sb *QWritebuf) WriteByte(c int) {

	buf := sb.getSpace(1)
	buf[0] = byte(c & 0xFF)
}

func (sb *QWritebuf) WriteLong(c int) {

	buf := sb.getSpace(4)
	buf[0] = byte(c & 0xff)
	buf[1] = byte((c >> 8) & 0xff)
	buf[2] = byte((c >> 16) & 0xff)
	buf[3] = byte(c >> 24)
}

func (sb *QWritebuf) WriteShort(c int) {

	buf := sb.getSpace(2)
	buf[0] = byte(c & 0xff)
	buf[1] = byte(c >> 8)
}

func (sb *QWritebuf) Write(data []byte) {
	buf := sb.getSpace(len(data))
	copy(buf, data)
}

func (sb *QWritebuf) WriteString(s string) {
	if len(s) == 0 {
		sb.WriteChar(0)
	} else {
		sb.Write([]byte(s))
		sb.WriteChar(0)
	}
}
