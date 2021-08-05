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
 * This file implements the console
 *
 * =======================================================================
 */
package client

const NUM_CON_TIMES = 4
const CON_TEXTSIZE = 32768

type console_t struct {
	initialized bool

	text []rune
	// char	text[CON_TEXTSIZE];
	current int /* line where next message will be printed */
	x       int /* offset in current line for next print */
	display int /* bottom of console displays this line */

	ormask int /* high bit mask for colored characters */

	linewidth  int /* characters across screen */
	totallines int /* total lines in console scrollback */

	cursorspeed float32

	vislines int

	// float	times[NUM_CON_TIMES]; /* cls.realtime time the line was generated */
}

func (T *qClient) drawStringScaled(x, y int, s string, factor float32) {
	for _, c := range s {
		T.Draw_CharScaled(x, y, int(c), factor)
		x += int(8 * factor)
	}
}

/*
 * If the line width has changed, reformat the buffer.
 */
func (T *qClient) conCheckResize() {
	scale := T.scrGetConsoleScale()

	/* We need to clamp the line width to MAXCMDLINE - 2,
	otherwise we may overflow the text buffer if the
	vertical resultion / 8 (one char == 8 pixels) is
	bigger then MAXCMDLINE.
	MAXCMDLINE - 2 because 1 for the prompt and 1 for
	the terminating \0. */
	width := int((float32(T.viddef.width)/scale)/8) - 2
	if width > MAXCMDLINE-2 {
		width = MAXCMDLINE - 2
	}

	if width == T.con.linewidth {
		return
	}

	/* video hasn't been initialized yet */
	if width < 1 {
		width = 38
		T.con.linewidth = width
		T.con.totallines = CON_TEXTSIZE / T.con.linewidth
		for i := 0; i < CON_TEXTSIZE; i++ {
			T.con.text[i] = ' '
		}
	} else {
		oldwidth := T.con.linewidth
		T.con.linewidth = width
		oldtotallines := T.con.totallines
		T.con.totallines = CON_TEXTSIZE / T.con.linewidth
		numlines := oldtotallines

		if T.con.totallines < numlines {
			numlines = T.con.totallines
		}

		numchars := oldwidth

		if T.con.linewidth < numchars {
			numchars = T.con.linewidth
		}

		tbuf := make([]rune, CON_TEXTSIZE)
		for i := 0; i < CON_TEXTSIZE; i++ {
			tbuf[i] = T.con.text[i]
			T.con.text[i] = ' '
		}

		for i := 0; i < numlines; i++ {
			for j := 0; j < numchars; j++ {
				T.con.text[(T.con.totallines-1-i)*T.con.linewidth+j] =
					tbuf[((T.con.current-i+oldtotallines)%
						oldtotallines)*oldwidth+j]
			}
		}

		// Con_ClearNotify()
	}

	T.con.current = T.con.totallines - 1
	T.con.display = T.con.current
}

func (T *qClient) conInit() {
	T.con.text = make([]rune, CON_TEXTSIZE)
	T.con.linewidth = -1

	T.conCheckResize()

	T.common.Com_Printf("Console initialized.\n")

	/* register our commands */
	//  con_notifytime = Cvar_Get("con_notifytime", "3", 0);

	//  Cmd_AddCommand("toggleconsole", Con_ToggleConsole_f);
	//  Cmd_AddCommand("togglechat", Con_ToggleChat_f);
	//  Cmd_AddCommand("messagemode", Con_MessageMode_f);
	//  Cmd_AddCommand("messagemode2", Con_MessageMode2_f);
	//  Cmd_AddCommand("clear", Con_Clear_f);
	//  Cmd_AddCommand("condump", Con_Dump_f);
	T.con.initialized = true
}

/*
 * Draws the console with the solid background
 */
func (T *qClient) conDrawConsole(frac float32) {
	// 	 int i, j, x, y, n;
	// 	 int rows;
	// 	 int verLen;
	// 	 char *text;
	// 	 int row;
	// 	 int lines;
	// 	 float scale;
	// 	 char version[48];
	// 	 char dlbar[1024];
	// 	 char timebuf[48];
	// 	 char tmpbuf[48];

	// 	 time_t t;
	// 	 struct tm *today;

	// scale := T.scrGetConsoleScale()
	lines := int(float32(T.viddef.height) * frac)

	if lines <= 0 {
		return
	}

	if lines > T.viddef.height {
		lines = T.viddef.height
	}

	/* draw the background */
	T.Draw_StretchPic(0, -T.viddef.height+lines, T.viddef.width,
		T.viddef.height, "conback")
	T.scrAddDirtyPoint(0, 0)
	T.scrAddDirtyPoint(T.viddef.width-1, lines-1)

	// 	 Com_sprintf(version, sizeof(version), "Yamagi Quake II v%s", YQ2VERSION);

	// 	 verLen = strlen(version);

	// 	 for (x = 0; x < verLen; x++)
	// 	 {
	// 		 Draw_CharScaled(viddef.width - ((verLen*8+5) * scale) + x * 8 * scale, lines - 35 * scale, 128 + version[x], scale);
	// 	 }

	// 	 t = time(NULL);
	// 	 today = localtime(&t);
	// 	 strftime(timebuf, sizeof(timebuf), "%H:%M:%S - %m/%d/%Y", today);

	// 	 Com_sprintf(tmpbuf, sizeof(tmpbuf), "%s", timebuf);

	// 	 for (x = 0; x < 21; x++)
	// 	 {
	// 		 Draw_CharScaled(viddef.width - (173 * scale) + x * 8 * scale, lines - 25 * scale, 128 + tmpbuf[x], scale);
	// 	 }

	// 	 /* draw the text */
	// 	 con.vislines = lines;

	// 	 rows = (lines - 22) >> 3; /* rows of text to draw */
	// 	 y = (lines - 30 * scale) / scale;

	// 	 /* draw from the bottom up */
	// 	 if (con.display != con.current)
	// 	 {
	// 		 /* draw arrows to show the buffer is backscrolled */
	// 		 for (x = 0; x < con.linewidth; x += 4)
	// 		 {
	// 			 Draw_CharScaled(((x + 1) << 3) * scale, y * scale, '^', scale);
	// 		 }

	// 		 y -= 8;
	// 		 rows--;
	// 	 }

	// 	 row = con.display;

	// 	 for (i = 0; i < rows; i++, y -= 8, row--)
	// 	 {
	// 		 if (row < 0)
	// 		 {
	// 			 break;
	// 		 }

	// 		 if (con.current - row >= con.totallines)
	// 		 {
	// 			 break; /* past scrollback wrap point */
	// 		 }

	// 		 text = con.text + (row % con.totallines) * con.linewidth;

	// 		 for (x = 0; x < con.linewidth; x++)
	// 		 {
	// 			 Draw_CharScaled(((x + 1) << 3) * scale, y * scale, text[x], scale);
	// 		 }
	// 	 }

	// 	 /* draw the download bar, figure out width */
	//  #ifdef USE_CURL
	// 	 if (cls.downloadname[0] && (cls.download || cls.downloadposition))
	//  #else
	// 	 if (cls.download)
	//  #endif
	// 	 {
	// 		 if ((text = strrchr(cls.downloadname, '/')) != NULL)
	// 		 {
	// 			 text++;
	// 		 }

	// 		 else
	// 		 {
	// 			 text = cls.downloadname;
	// 		 }

	// 		 x = con.linewidth - ((con.linewidth * 7) / 40);
	// 		 y = x - strlen(text) - 8;
	// 		 i = con.linewidth / 3;

	// 		 if (strlen(text) > i)
	// 		 {
	// 			 y = x - i - 11;
	// 			 memcpy(dlbar, text, i);
	// 			 dlbar[i] = 0;
	// 			 strcat(dlbar, "...");
	// 		 }
	// 		 else
	// 		 {
	// 			 strcpy(dlbar, text);
	// 		 }

	// 		 strcat(dlbar, ": ");
	// 		 i = strlen(dlbar);
	// 		 dlbar[i++] = '\x80';

	// 		 /* where's the dot gone? */
	// 		 if (cls.downloadpercent == 0)
	// 		 {
	// 			 n = 0;
	// 		 }

	// 		 else
	// 		 {
	// 			 n = y * cls.downloadpercent / 100;
	// 		 }

	// 		 for (j = 0; j < y; j++)
	// 		 {
	// 			 if (j == n)
	// 			 {
	// 				 dlbar[i++] = '\x83';
	// 			 }

	// 			 else
	// 			 {
	// 				 dlbar[i++] = '\x81';
	// 			 }
	// 		 }

	// 		 dlbar[i++] = '\x82';
	// 		 dlbar[i] = 0;

	// 		 sprintf(dlbar + strlen(dlbar), " %02d%%", cls.downloadpercent);

	// 		 /* draw it */
	// 		 y = con.vislines - 12;

	// 		 for (i = 0; i < strlen(dlbar); i++)
	// 		 {
	// 			 Draw_CharScaled(((i + 1) << 3) * scale, y * scale, dlbar[i], scale);
	// 		 }
	// 	 }

	// 	 /* draw the input prompt, user text, and cursor if desired */
	// 	 Con_DrawInput();
}
