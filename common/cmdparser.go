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
 * This file implements the Quake II command processor. Every command
 * which is send via the command line at startup, via the console and
 * via rcon is processed here and send to the apropriate subsystem.
 *
 * =======================================================================
 */
package common

import (
	"fmt"
	"goquake2/shared"
	"strings"
)

const aliasLoopCount = 16

/*
 * Adds command text at the end of the buffer
 */
func (T *qCommon) Cbuf_AddText(text string) {
	T.cmd_text += text
}

/*
 * Adds command text immediately after the current command
 * Adds a \n to the text
 */
func (T *qCommon) Cbuf_InsertText(text string) {
	T.cmd_text = text + "\n" + T.cmd_text
}

func (T *qCommon) Cbuf_Execute() error {

	if T.cmd_wait > 0 {
		// make sure that "wait" in scripts waits for ~16.66ms (1 frame at 60fps)
		// regardless of framerate
		// if (Sys_Milliseconds() - cmd_wait <= 16) {
		// 	return;
		// }
		T.cmd_wait = 0
	}

	T.alias_count = 0 /* don't allow infinite alias loops */

	for len(T.cmd_text) > 0 {
		/* find a \n or ; line break */

		quotes := 0

		index := len(T.cmd_text)
		for i := 0; i < len(T.cmd_text); i++ {
			if T.cmd_text[i] == '"' {
				quotes++
			}

			if (quotes&1) == 0 && (T.cmd_text[i] == ';') {
				index = i
				break /* don't break if inside a quoted string */
			}

			if T.cmd_text[i] == '\n' {
				index = i
				break
			}
		}

		line := T.cmd_text[0:index]

		/* delete the text from the command buffer and move remaining
		   commands down this is necessary because commands (exec,
		   alias) can insert data at the beginning of the text buffer */
		if index == len(T.cmd_text) {
			T.cmd_text = ""
		} else {
			T.cmd_text = T.cmd_text[index+1:]
		}

		/* execute the command line */
		err := T.Cmd_ExecuteString(line)
		if err != nil {
			return err
		}

		if T.cmd_wait > 0 {
			/* skip out while text still remains in buffer,
			   leaving it for after we're done waiting */
			break
		}
	}
	return nil
}

/*
 * Parses the given string into command line tokens.
 * $Cvars will be expanded unless they are in a quoted token
 */
func (T *qCommon) Cmd_TokenizeString(text string, macroExpand bool) []string {
	//  int i;
	//  const char *com_token;

	//  /* clear the args from the last string */
	//  for (i = 0; i < cmd_argc; i++)
	//  {
	// 	 Z_Free(cmd_argv[i]);
	//  }

	var cmd_args []string
	//  cmd_argc = 0;
	//  cmd_args[0] = 0;

	//  /* macro expand the text */
	//  if (macroExpand)
	//  {
	// 	 text = Cmd_MacroExpandString(text);
	//  }

	if len(text) == 0 {
		return cmd_args
	}

	index := 0
	for {
		/* skip whitespace up to a /n */
		for index < len(text) && text[index] <= ' ' && text[index] != '\n' {
			index++
		}

		if index >= len(text) || text[index] == '\n' {
			/* a newline seperates commands in the buffer */
			return cmd_args
		}

		token, index2 := shared.COM_Parse(text, index)

		if index2 < 0 {
			return cmd_args
		}
		index = index2
		cmd_args = append(cmd_args, token)
	}
}

func (T *qCommon) Cmd_AddCommand(cmd_name string, function func([]string, interface{}) error, arg interface{}) {

	/* fail if the command is a variable name */
	// if (Cvar_VariableString(cmd_name)[0]) {
	// 	Cmd_RemoveCommand(cmd_name);
	// }

	// /* fail if the command already exists */
	// for (cmd = cmd_functions; cmd; cmd = cmd->next)
	// {
	// 	if (!strcmp(cmd_name, cmd->name))
	// 	{
	// 		Com_Printf("Cmd_AddCommand: %s already defined\n", cmd_name);
	// 		return;
	// 	}
	// }

	T.cmd_functions[strings.ToLower(cmd_name)] = xcommand_t{function: function, param: arg}
}

/*
 * A complete command line has been parsed, so try to execute it
 */
func (T *qCommon) Cmd_ExecuteString(text string) error {
	// cmd_function_t * cmd
	// cmdalias_t * a

	args := T.Cmd_TokenizeString(text, true)

	/* execute the command line */
	if len(args) == 0 {
		return nil /* no tokens */
	}

	// if Cmd_Argc() > 1 && Q_strcasecmp(cmd_argv[0], "exec") == 0 && Q_strcasecmp(cmd_argv[1], "yq2.cfg") == 0 {
	// 	/* exec yq2.cfg is done directly after exec default.cfg, see Qcommon_Init() */
	// 	doneWithDefaultCfg = true
	// }

	/* check functions */
	f, ok := T.cmd_functions[strings.ToLower(args[0])]
	if ok {
		if f.function != nil {
			return f.function(args, f.param)
		} else {
			/* forward to server command */
			return T.Cmd_ExecuteString(fmt.Sprintf("cmd %s", text))
		}
	}

	/* check alias */
	a, ok := T.cmd_alias[strings.ToLower(args[0])]
	if ok {
		T.alias_count++
		if T.alias_count == aliasLoopCount {
			T.Com_Printf("ALIAS_LOOP_COUNT\n")
			return nil
		}

		T.Cbuf_InsertText(a)
		return nil
	}

	//  /* check cvars */
	//  if (Cvar_Command())
	//  {
	// 	 return;
	//  }

	/* send it as a server command if we are connected */
	// Cmd_ForwardToServer()

	fmt.Printf("Unknown command \"%v\"\n", args[0])
	return nil
}

/*
 * Creates a new command that executes
 * a command string (possibly ; seperated)
 */
func cmd_Alias_f(args []string, arg interface{}) error {

	T := arg.(*qCommon)

	if len(args) == 1 {
		T.Com_Printf("Current alias commands:\n")

		for k, v := range T.cmd_alias {
			T.Com_Printf("%v : %v\n", k, v)
		}

		return nil
	}

	/* copy the rest of the command line */
	var cmd strings.Builder

	for i := 2; i < len(args); i++ {
		cmd.WriteString(args[i])

		if i != (len(args) - 1) {
			cmd.WriteRune(' ')
		}
	}

	cmd.WriteRune('\n')

	T.cmd_alias[strings.ToLower(args[1])] = cmd.String()
	return nil
}

/*
 * Execute a script file
 */
func cmd_Exec_f(args []string, arg interface{}) error {

	T := arg.(*qCommon)

	if len(args) != 2 {
		T.Com_Printf("exec <filename> : execute a script file\n")
		return nil
	}

	bfr, err := T.LoadFile(args[1])
	if bfr == nil {
		T.Com_Printf("couldn't exec %s\n", args[1])
		return err
	}

	T.Com_Printf("execing %s.\n", args[1])

	T.Cbuf_InsertText(string(bfr))

	return nil
}

func (T *qCommon) cmdInit() {
	/* register our commands */
	// Cmd_AddCommand("cmdlist", Cmd_List_f)
	T.Cmd_AddCommand("exec", cmd_Exec_f, T)
	// Cmd_AddCommand("vstr", Cmd_Vstr_f)
	// Cmd_AddCommand("echo", Cmd_Echo_f)
	T.Cmd_AddCommand("alias", cmd_Alias_f, T)
	// Cmd_AddCommand("wait", Cmd_Wait_f)
}
