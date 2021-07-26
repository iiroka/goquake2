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
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307,
 * USA.
 *
 * =======================================================================
 *
 * The header file for the upper level key event processing
 *
 * =======================================================================
 */
package client

/* Max length of a console command line. 1024
 * chars allow for a vertical resolution of
 * 8192 pixel which should be enough for the
 * years to come. */
const MAXCMDLINE = 1024

/* number of console command lines saved in history,
 * must be a power of two, because we use & (NUM_KEY_LINES-1)
 * instead of % so -1 wraps to NUM_KEY_LINES-1 */
const NUM_KEY_LINES = 32
