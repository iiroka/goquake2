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

/* these are the key numbers that should be passed to Key_Event
   they must be matched by the low level key event processing! */
const (
	K_TAB    = 9
	K_ENTER  = 13
	K_ESCAPE = 27
	// Note: ASCII keys are generally valid but don't get constants here,
	// just use 'a' (yes, lowercase) or '2' or whatever, however there are
	// some special cases when writing/parsing configs (space or quotes or
	// also ; and $ have a special meaning there so we use e.g. "SPACE" instead),
	// see keynames[] in cl_keyboard.c
	K_SPACE = 32

	K_BACKSPACE = 127

	K_COMMAND  = 128 // "Windows Key"
	K_CAPSLOCK = 129
	K_POWER    = 130
	K_PAUSE    = 131

	K_UPARROW    = 132
	K_DOWNARROW  = 133
	K_LEFTARROW  = 134
	K_RIGHTARROW = 135

	K_ALT   = 136
	K_CTRL  = 137
	K_SHIFT = 138
	K_INS   = 139
	K_DEL   = 140
	K_PGDN  = 141
	K_PGUP  = 142
	K_HOME  = 143
	K_END   = 144

	K_F1  = 145
	K_F2  = 146
	K_F3  = 147
	K_F4  = 148
	K_F5  = 149
	K_F6  = 150
	K_F7  = 151
	K_F8  = 152
	K_F9  = 153
	K_F10 = 154
	K_F11 = 155
	K_F12 = 156
	K_F13 = 157
	K_F14 = 158
	K_F15 = 159

	K_KP_HOME       = 160
	K_KP_UPARROW    = 161
	K_KP_PGUP       = 162
	K_KP_LEFTARROW  = 163
	K_KP_5          = 164
	K_KP_RIGHTARROW = 165
	K_KP_END        = 166
	K_KP_DOWNARROW  = 167
	K_KP_PGDN       = 168
	K_KP_ENTER      = 169
	K_KP_INS        = 170
	K_KP_DEL        = 171
	K_KP_SLASH      = 172
	K_KP_MINUS      = 173
	K_KP_PLUS       = 174
	K_KP_NUMLOCK    = 175
	K_KP_STAR       = 176
	K_KP_EQUALS     = 177

	K_MOUSE1 = 178
	K_MOUSE2 = 179
	K_MOUSE3 = 180
	K_MOUSE4 = 181
	K_MOUSE5 = 182

	K_MWHEELDOWN = 183
	K_MWHEELUP   = 184

	K_JOY1  = 185
	K_JOY2  = 186
	K_JOY3  = 187
	K_JOY4  = 188
	K_JOY5  = 189
	K_JOY6  = 190
	K_JOY7  = 191
	K_JOY8  = 192
	K_JOY9  = 193
	K_JOY10 = 194
	K_JOY11 = 195
	K_JOY12 = 196
	K_JOY13 = 197
	K_JOY14 = 198
	K_JOY15 = 199
	K_JOY16 = 200
	K_JOY17 = 201
	K_JOY18 = 202
	K_JOY19 = 203
	K_JOY20 = 204
	K_JOY21 = 205
	K_JOY22 = 206
	K_JOY23 = 207
	K_JOY24 = 208
	K_JOY25 = 209
	K_JOY26 = 210
	K_JOY27 = 211
	K_JOY28 = 212
	K_JOY29 = 213
	K_JOY30 = 214
	K_JOY31 = 215
	K_JOY32 = 216

	K_HAT_UP    = 217
	K_HAT_RIGHT = 218
	K_HAT_DOWN  = 219
	K_HAT_LEFT  = 220

	K_TRIG_LEFT  = 221
	K_TRIG_RIGHT = 222

	// add other joystick/controller keys before this one
	// and adjust it accordingly also remember to add corresponding _ALT key below!
	K_JOY_LAST_REGULAR = K_TRIG_RIGHT

	/* Can't be mapped to any action (=> not regular) */
	K_JOY_BACK = 223

	K_JOY1_ALT  = 224
	K_JOY2_ALT  = 225
	K_JOY3_ALT  = 226
	K_JOY4_ALT  = 227
	K_JOY5_ALT  = 228
	K_JOY6_ALT  = 229
	K_JOY7_ALT  = 230
	K_JOY8_ALT  = 231
	K_JOY9_ALT  = 232
	K_JOY10_ALT = 233
	K_JOY11_ALT = 234
	K_JOY12_ALT = 235
	K_JOY13_ALT = 236
	K_JOY14_ALT = 237
	K_JOY15_ALT = 238
	K_JOY16_ALT = 239
	K_JOY17_ALT = 240
	K_JOY18_ALT = 241
	K_JOY19_ALT = 242
	K_JOY20_ALT = 243
	K_JOY21_ALT = 244
	K_JOY22_ALT = 245
	K_JOY23_ALT = 246
	K_JOY24_ALT = 247
	K_JOY25_ALT = 248
	K_JOY26_ALT = 249
	K_JOY27_ALT = 250
	K_JOY28_ALT = 251
	K_JOY29_ALT = 252
	K_JOY30_ALT = 253
	K_JOY31_ALT = 254
	K_JOY32_ALT = 255

	K_HAT_UP_ALT    = 256
	K_HAT_RIGHT_ALT = 257
	K_HAT_DOWN_ALT  = 258
	K_HAT_LEFT_ALT  = 259

	K_TRIG_LEFT_ALT  = 260
	K_TRIG_RIGHT_ALT = 261

	// add other joystick/controller keys before this one and adjust it accordingly
	K_JOY_LAST_REGULAR_ALT = K_TRIG_RIGHT_ALT

	K_SUPER     = 262 // TODO: what is this? SDL doesn't seem to know it..
	K_COMPOSE   = 263
	K_MODE      = 264
	K_HELP      = 265
	K_PRINT     = 266
	K_SYSREQ    = 267
	K_SCROLLOCK = 268
	K_MENU      = 269
	K_UNDO      = 270

	// The following are mapped from SDL_Scancodes used as a *fallback* for keys
	// whose SDL_KeyCode we don't have a K_ constant for like German Umlaut keys.
	// The scancode name corresponds to the key at that position on US-QWERTY keyboards
	// *not* the one in the local layout (e.g. German 'Ö' key is K_SC_SEMICOLON)
	// !!! NOTE: if you add a scancode here make sure to also add it to:
	// 1. keynames[] in cl_keyboard.c
	// 2. IN_TranslateScancodeToQ2Key() in input/sdl.c
	K_SC_A = 271
	K_SC_B = 272
	K_SC_C = 273
	K_SC_D = 274
	K_SC_E = 275
	K_SC_F = 276
	K_SC_G = 277
	K_SC_H = 278
	K_SC_I = 279
	K_SC_J = 280
	K_SC_K = 281
	K_SC_L = 282
	K_SC_M = 283
	K_SC_N = 284
	K_SC_O = 285
	K_SC_P = 286
	K_SC_Q = 287
	K_SC_R = 288
	K_SC_S = 289
	K_SC_T = 290
	K_SC_U = 291
	K_SC_V = 292
	K_SC_W = 293
	K_SC_X = 294
	K_SC_Y = 295
	K_SC_Z = 296
	// leaving out SDL_SCANCODE_1 ... _0 we handle them separately already
	// also return escape backspace tab space already handled as keycodes
	K_SC_MINUS        = 297
	K_SC_EQUALS       = 298
	K_SC_LEFTBRACKET  = 299
	K_SC_RIGHTBRACKET = 299
	K_SC_BACKSLASH    = 300
	K_SC_NONUSHASH    = 301
	K_SC_SEMICOLON    = 302
	K_SC_APOSTROPHE   = 303
	K_SC_GRAVE        = 304
	K_SC_COMMA        = 305
	K_SC_PERIOD       = 306
	K_SC_SLASH        = 307
	// leaving out lots of key incl. from keypad we already handle them as normal keys
	K_SC_NONUSBACKSLASH     = 308
	K_SC_INTERNATIONAL1     = 309 /**< used on Asian keyboards see footnotes in USB doc */
	K_SC_INTERNATIONAL2     = 310
	K_SC_INTERNATIONAL3     = 311 /**< Yen */
	K_SC_INTERNATIONAL4     = 312
	K_SC_INTERNATIONAL5     = 313
	K_SC_INTERNATIONAL6     = 314
	K_SC_INTERNATIONAL7     = 315
	K_SC_INTERNATIONAL8     = 316
	K_SC_INTERNATIONAL9     = 317
	K_SC_THOUSANDSSEPARATOR = 318
	K_SC_DECIMALSEPARATOR   = 319
	K_SC_CURRENCYUNIT       = 320
	K_SC_CURRENCYSUBUNIT    = 321

	// hardcoded pseudo-key to open the console emitted when pressing the "console key"
	// (SDL_SCANCODE_GRAVE the one between Esc 1 and Tab) on layouts that don't
	// have a relevant char there (unlike Brazilian which has quotes there which you
	// want to be able to type in the console) - the user can't bind this key.
	K_CONSOLE = 322

	K_LAST = 323
)
