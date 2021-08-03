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
 * Client / Server interactions
 *
 * =======================================================================
 */
package common

import (
	"fmt"
	"goquake2/shared"
	"log"
)

/*
 * Both client and server can use this, and it will output
 * to the apropriate place.
 */
func (T *qCommon) Com_VPrintf(print_level int, format string, a ...interface{}) {
	//  if((print_level == PRINT_DEVELOPER) && (!developer || !developer->value))
	//  {
	// 	 return; /* don't confuse non-developers with techie stuff... */
	//  }
	//  else
	//  {
	// 	 int i;
	// 	 char msg[MAXPRINTMSG];

	// 	 int msgLen = vsnprintf(msg, MAXPRINTMSG, fmt, argptr);
	// 	 if (msgLen >= MAXPRINTMSG || msgLen < 0) {
	// 		 msgLen = MAXPRINTMSG-1;
	// 		 msg[msgLen] = '\0';
	// 	 }

	// 	 if (rd_target)
	// 	 {
	// 		 if ((msgLen + strlen(rd_buffer)) > (rd_buffersize - 1))
	// 		 {
	// 			 rd_flush(rd_target, rd_buffer);
	// 			 *rd_buffer = 0;
	// 		 }

	// 		 strcat(rd_buffer, msg);
	// 		 return;
	// 	 }

	//  #ifndef DEDICATED_ONLY
	// 	 Con_Print(msg);
	//  #endif

	// 	 // remove unprintable characters
	// 	 for(i=0; i<msgLen; ++i)
	// 	 {
	// 		 char c = msg[i];
	// 		 if(c < ' ' && (c < '\t' || c > '\r'))
	// 		 {
	// 			 switch(c)
	// 			 {
	// 				 // no idea if the following two are ever sent here, but in conchars.pcx they look like this
	// 				 // so do the replacements.. won't hurt I guess..
	// 				 case 0x10:
	// 					 msg[i] = '[';
	// 					 break;
	// 				 case 0x11:
	// 					 msg[i] = ']';
	// 					 break;
	// 				 // horizontal line chars
	// 				 case 0x1D:
	// 				 case 0x1F:
	// 					 msg[i] = '-';
	// 					 break;
	// 				 case 0x1E:
	// 					 msg[i] = '=';
	// 					 break;
	// 				 default: // just replace all other unprintable chars with space, should be good enough
	// 					 msg[i] = ' ';
	// 			 }
	// 		 }
	// 	 }

	// 	 /* also echo to debugging console */
	// 	 Sys_ConsoleOutput(msg);

	// 	 /* logfile */
	// 	 if (logfile_active && logfile_active->value)
	// 	 {
	// 		 char name[MAX_OSPATH];

	// 		 if (!logfile)
	// 		 {
	// 			 Com_sprintf(name, sizeof(name), "%s/qconsole.log", FS_Gamedir());

	// 			 if (logfile_active->value > 2)
	// 			 {
	// 				 logfile = Q_fopen(name, "a");
	// 			 }

	// 			 else
	// 			 {
	// 				 logfile = Q_fopen(name, "w");
	// 			 }
	// 		 }

	// 		 if (logfile)
	// 		 {
	// 			 fprintf(logfile, "%s", msg);
	// 		 }

	// 		 if (logfile_active->value > 1)
	// 		 {
	// 			 fflush(logfile);  /* force it to save every time */
	// 		 }
	// 	 }
	//  }
	fmt.Printf(format, a...)
}

/*
 * Both client and server can use this, and it will output
 * to the apropriate place.
 */
func (T *qCommon) Com_Printf(format string, a ...interface{}) {
	T.Com_VPrintf(shared.PRINT_ALL, format, a...)
}

/*
 * A Com_Printf that only shows up if the "developer" cvar is set
 */
func (T *qCommon) Com_DPrintf(format string, a ...interface{}) {
	T.Com_VPrintf(shared.PRINT_DEVELOPER, format, a...)
}

type AbortFrame struct{}

func (m *AbortFrame) Error() string {
	return "abortframe"
}

/*
 * Both client and server can use this, and it will
 * do the apropriate things.
 */
func (T *qCommon) Com_Error(code int, format string, a ...interface{}) error {

	if T.recursive {
		log.Fatalf("recursive error after: %v", T.msg)
	}

	T.recursive = true

	T.msg = fmt.Sprintf(format, a...)

	if code == shared.ERR_DISCONNECT {
		// CL_Drop()
		T.recursive = false
		return &AbortFrame{}
	} else if code == shared.ERR_DROP {
		T.Com_Printf("********************\nERROR: %s\n********************\n", T.msg)
		// SV_Shutdown(va("Server crashed: %s\n", msg), false)
		// CL_Drop()
		T.recursive = false
		return &AbortFrame{}
	} else {
		// SV_Shutdown(va("Server fatal crashed: %s\n", msg), false)
		// CL_Shutdown()
	}

	// if logfile {
	// 	fclose(logfile)
	// 	logfile = NULL
	// }

	log.Fatal(T.msg)
	T.recursive = false
	return nil
}

/*
 * Both client and server can use this, and it will
 * do the apropriate things.
 */
func (T *qCommon) Com_Quit() {
	T.Com_Printf("\n----------- shutting down ----------\n")
	//  SV_Shutdown("Server quit\n", false);
	//  Sys_Quit();
	T.running = false
}

func (T *qCommon) Showpackets() bool {
	return T.showpackets.Bool()
}
