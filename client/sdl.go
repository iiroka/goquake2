/*
 * Copyright (C) 2010 Yamagi Burmeister
 * Copyright (C) 1997-2005 Id Software, Inc.
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
 * Joystick threshold code is partially based on http://ioquake3.org code.
 *
 * =======================================================================
 *
 * This is the Quake II input system backend, implemented with SDL.
 *
 * =======================================================================
 */
package client

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

type QInput struct {
	client *qClient

	// The last time input events were processed.
	// Used throughout the client.
	Sys_frame_time int
}

/* ------------------------------------------------------------------ */

/*
 * This creepy function translates SDL keycodes into
 * the id Tech 2 engines interal representation.
 */
func inTranslateSDLtoQ2Key(keysym int) int {
	key := 0

	/* These must be translated */
	switch keysym {
	case sdl.K_TAB:
		key = K_TAB
		break
	case sdl.K_RETURN:
		key = K_ENTER
		break
	case sdl.K_ESCAPE:
		key = K_ESCAPE
		break
	case sdl.K_BACKSPACE:
		key = K_BACKSPACE
		break
	case sdl.K_LGUI,
		sdl.K_RGUI:
		key = K_COMMAND // Win key
		break
	case sdl.K_CAPSLOCK:
		key = K_CAPSLOCK
		break
	case sdl.K_POWER:
		key = K_POWER
		break
	case sdl.K_PAUSE:
		key = K_PAUSE
		break

	case sdl.K_UP:
		key = K_UPARROW
		break
	case sdl.K_DOWN:
		key = K_DOWNARROW
		break
	case sdl.K_LEFT:
		key = K_LEFTARROW
		break
	case sdl.K_RIGHT:
		key = K_RIGHTARROW
		break

	case sdl.K_RALT,
		sdl.K_LALT:
		key = K_ALT
		break
	case sdl.K_LCTRL,
		sdl.K_RCTRL:
		key = K_CTRL
		break
	case sdl.K_LSHIFT,
		sdl.K_RSHIFT:
		key = K_SHIFT
		break
	case sdl.K_INSERT:
		key = K_INS
		break
	case sdl.K_DELETE:
		key = K_DEL
		break
	case sdl.K_PAGEDOWN:
		key = K_PGDN
		break
	case sdl.K_PAGEUP:
		key = K_PGUP
		break
	case sdl.K_HOME:
		key = K_HOME
		break
	case sdl.K_END:
		key = K_END
		break

	case sdl.K_F1:
		key = K_F1
		break
	case sdl.K_F2:
		key = K_F2
		break
	case sdl.K_F3:
		key = K_F3
		break
	case sdl.K_F4:
		key = K_F4
		break
	case sdl.K_F5:
		key = K_F5
		break
	case sdl.K_F6:
		key = K_F6
		break
	case sdl.K_F7:
		key = K_F7
		break
	case sdl.K_F8:
		key = K_F8
		break
	case sdl.K_F9:
		key = K_F9
		break
	case sdl.K_F10:
		key = K_F10
		break
	case sdl.K_F11:
		key = K_F11
		break
	case sdl.K_F12:
		key = K_F12
		break
	case sdl.K_F13:
		key = K_F13
		break
	case sdl.K_F14:
		key = K_F14
		break
	case sdl.K_F15:
		key = K_F15
		break

	case sdl.K_KP_7:
		key = K_KP_HOME
		break
	case sdl.K_KP_8:
		key = K_KP_UPARROW
		break
	case sdl.K_KP_9:
		key = K_KP_PGUP
		break
	case sdl.K_KP_4:
		key = K_KP_LEFTARROW
		break
	case sdl.K_KP_5:
		key = K_KP_5
		break
	case sdl.K_KP_6:
		key = K_KP_RIGHTARROW
		break
	case sdl.K_KP_1:
		key = K_KP_END
		break
	case sdl.K_KP_2:
		key = K_KP_DOWNARROW
		break
	case sdl.K_KP_3:
		key = K_KP_PGDN
		break
	case sdl.K_KP_ENTER:
		key = K_KP_ENTER
		break
	case sdl.K_KP_0:
		key = K_KP_INS
		break
	case sdl.K_KP_PERIOD:
		key = K_KP_DEL
		break
	case sdl.K_KP_DIVIDE:
		key = K_KP_SLASH
		break
	case sdl.K_KP_MINUS:
		key = K_KP_MINUS
		break
	case sdl.K_KP_PLUS:
		key = K_KP_PLUS
		break
	case sdl.K_NUMLOCKCLEAR:
		key = K_KP_NUMLOCK
		break
	case sdl.K_KP_MULTIPLY:
		key = K_KP_STAR
		break
	case sdl.K_KP_EQUALS:
		key = K_KP_EQUALS
		break

	// TODO: K_SUPER ? Win Key is already K_COMMAND

	case sdl.K_APPLICATION:
		key = K_COMPOSE
		break
	case sdl.K_MODE:
		key = K_MODE
		break
	case sdl.K_HELP:
		key = K_HELP
		break
	case sdl.K_PRINTSCREEN:
		key = K_PRINT
		break
	case sdl.K_SYSREQ:
		key = K_SYSREQ
		break
	case sdl.K_SCROLLLOCK:
		key = K_SCROLLOCK
		break
	case sdl.K_MENU:
		key = K_MENU
		break
	case sdl.K_UNDO:
		key = K_UNDO
		break

	default:
		break
	}

	return key
}

/* ------------------------------------------------------------------ */

/*
 * Updates the input queue state. Called every
 * frame by the client and does nearly all the
 * input magic.
 */
func (T *QInput) Update() {
	//  qboolean want_grab;
	//  SDL_Event event;
	//  unsigned int key;

	//  static char last_hat = SDL_HAT_CENTERED;
	//  static qboolean left_trigger = false;
	//  static qboolean right_trigger = false;

	//  static int consoleKeyCode = 0;

	/* Get and process an event */
	for {
		event := sdl.PollEvent()
		if event == nil {
			break
		}

		switch event.GetType() {
		case sdl.MOUSEWHEEL:
			// 			 Key_Event((event.wheel.y > 0 ? K_MWHEELUP : K_MWHEELDOWN), true, true);
			// 			 Key_Event((event.wheel.y > 0 ? K_MWHEELUP : K_MWHEELDOWN), false, true);
			// 			 break;

		case sdl.MOUSEBUTTONDOWN,
			sdl.MOUSEBUTTONUP:
		// 			 switch (event.button.button)
		// 			 {
		// 				 case SDL_BUTTON_LEFT:
		// 					 key = K_MOUSE1;
		// 					 break;
		// 				 case SDL_BUTTON_MIDDLE:
		// 					 key = K_MOUSE3;
		// 					 break;
		// 				 case SDL_BUTTON_RIGHT:
		// 					 key = K_MOUSE2;
		// 					 break;
		// 				 case SDL_BUTTON_X1:
		// 					 key = K_MOUSE4;
		// 					 break;
		// 				 case SDL_BUTTON_X2:
		// 					 key = K_MOUSE5;
		// 					 break;
		// 				 default:
		// 					 return;
		// 			 }

		// 			 Key_Event(key, (event.type == SDL_MOUSEBUTTONDOWN), true);
		// 			 break;

		case sdl.MOUSEMOTION:
			// 			 if (cls.key_dest == key_game && (int) cl_paused->value == 0)
			// 			 {
			// 				 mouse_x += event.motion.xrel;
			// 				 mouse_y += event.motion.yrel;
			// 			 }
			// 			 break;

			// 		 case SDL_TEXTINPUT:
			// 		 {
			// 			 int c = event.text.text[0];
			// 			 // also make sure we don't get the char that corresponds to the
			// 			 // "console key" (like "^" or "`") as text input
			// 			 if ((c >= ' ') && (c <= '~') && c != consoleKeyCode)
			// 			 {
			// 				 Char_Event(c);
			// 			 }
			// 		 }

			// 			 break;

		case sdl.KEYDOWN,
			sdl.KEYUP:
			// 		 {
			down := (event.GetType() == sdl.KEYDOWN)

			/* workaround for AZERTY-keyboards, which don't have 1, 2, ..., 9, 0 in first row:
			 * always map those physical keys (scancodes) to those keycodes anyway
			 * see also https://bugzilla.libsdl.org/show_bug.cgi?id=3188 */
			kevent := event.(*sdl.KeyboardEvent)
			sc := kevent.Keysym.Scancode

			if sc >= sdl.SCANCODE_1 && sc <= sdl.SCANCODE_0 {
				/* Note that the SDL_SCANCODEs are SDL_SCANCODE_1, _2, ..., _9, SDL_SCANCODE_0
				 * while in ASCII it's '0', '1', ..., '9' => handle 0 and 1-9 separately
				 * (quake2 uses the ASCII values for those keys) */
				key := int('0') /* implicitly handles SDL_SCANCODE_0 */

				if sc <= sdl.SCANCODE_9 {
					key = int('1') + int(sc-sdl.SCANCODE_1)
				}

				T.client.KeyEvent(key, down, false)
			} else {
				kc := kevent.Keysym.Sym
				if sc == sdl.SCANCODE_GRAVE && kc != '\'' && kc != '"' {
					// special case/hack: open the console with the "console key"
					// (beneath Esc, left of 1, above Tab)
					// but not if the keycode for this is a quote (like on Brazilian
					// keyboards) - otherwise you couldn't type them in the console
					if (kevent.Keysym.Mod & (sdl.KMOD_CAPS | sdl.KMOD_SHIFT | sdl.KMOD_ALT | sdl.KMOD_CTRL | sdl.KMOD_GUI)) == 0 {
						// also, only do this if no modifiers like shift or AltGr or whatever are pressed
						// so kc will most likely be the ascii char generated by this and can be ignored
						// in case SDL_TEXTINPUT above (so we don't get ^ or whatever as text in console)
						// (can't just check for mod == 0 because numlock is a KMOD too)
						// 						 Key_Event(K_CONSOLE, down, true);
						// 						 consoleKeyCode = kc;
					}
				} else if (kc >= sdl.K_SPACE) && (kc < sdl.K_DELETE) {
					T.client.KeyEvent(int(kc), down, false)
				} else {
					key := inTranslateSDLtoQ2Key(int(kc))
					if key == 0 {
						// fallback to scancodes if we don't know the keycode
						key = inTranslateSDLtoQ2Key(int(sc))
					}
					if key != 0 {
						T.client.KeyEvent(key, down, true)
					} else {
						T.client.common.Com_DPrintf("Pressed unknown key with SDL_Keycode %x, SDL_Scancode %d.\n", kc, int(sc))
					}
				}
			}

		// 			 break;
		// 		 }

		case sdl.WINDOWEVENT:
			// wevent := event.(*sdl.WindowEvent)
			// println("WINDOWEVENT", wevent.Event)
			// 			 if (event.window.event == SDL_WINDOWEVENT_FOCUS_LOST ||
			// 				 event.window.event == SDL_WINDOWEVENT_FOCUS_GAINED)
			// 			 {
			// 				 Key_MarkAllUp();
			// 			 }
			// 			 else if (event.window.event == SDL_WINDOWEVENT_MOVED)
			// 			 {
			// 				 // make sure GLimp_GetRefreshRate() will query from SDL again - the window might
			// 				 // be on another display now!
			// 				 glimp_refreshRate = -1;
			// 			 }

			// 		 case SDL_CONTROLLERBUTTONUP:
			// 		 case SDL_CONTROLLERBUTTONDOWN: /* Handle Controller Back button */
			// 		 {
			// 			 qboolean down = (event.type == SDL_CONTROLLERBUTTONDOWN);

			// 			 if (event.cbutton.button == SDL_CONTROLLER_BUTTON_BACK)
			// 			 {
			// 				 Key_Event(K_JOY_BACK, down, true);
			// 			 }

			// 			 break;
			// 		 }

			// 		 case SDL_CONTROLLERAXISMOTION:  /* Handle Controller Motion */
			// 		 {
			// 			 char *direction_type;
			// 			 float threshold = 0;
			// 			 float fix_value = 0;
			// 			 int axis_value = event.caxis.value;

			// 			 switch (event.caxis.axis)
			// 			 {
			// 				 /* left/right */
			// 				 case SDL_CONTROLLER_AXIS_LEFTX:
			// 					 direction_type = joy_axis_leftx->string;
			// 					 threshold = joy_axis_leftx_threshold->value;
			// 					 break;

			// 				 /* top/bottom */
			// 				 case SDL_CONTROLLER_AXIS_LEFTY:
			// 					 direction_type = joy_axis_lefty->string;
			// 					 threshold = joy_axis_lefty_threshold->value;
			// 					 break;

			// 				 /* second left/right */
			// 				 case SDL_CONTROLLER_AXIS_RIGHTX:
			// 					 direction_type = joy_axis_rightx->string;
			// 					 threshold = joy_axis_rightx_threshold->value;
			// 					 break;

			// 				 /* second top/bottom */
			// 				 case SDL_CONTROLLER_AXIS_RIGHTY:
			// 					 direction_type = joy_axis_righty->string;
			// 					 threshold = joy_axis_righty_threshold->value;
			// 					 break;

			// 				 case SDL_CONTROLLER_AXIS_TRIGGERLEFT:
			// 					 direction_type = joy_axis_triggerleft->string;
			// 					 threshold = joy_axis_triggerleft_threshold->value;
			// 					 break;

			// 				 case SDL_CONTROLLER_AXIS_TRIGGERRIGHT:
			// 					 direction_type = joy_axis_triggerright->string;
			// 					 threshold = joy_axis_triggerright_threshold->value;
			// 					 break;

			// 				 default:
			// 					 direction_type = "none";
			// 			 }

			// 			 if (threshold > 0.9)
			// 			 {
			// 				 threshold = 0.9;
			// 			 }

			// 			 if (axis_value < 0 && (axis_value > (32768 * threshold)))
			// 			 {
			// 				 axis_value = 0;
			// 			 }
			// 			 else if (axis_value > 0 && (axis_value < (32768 * threshold)))
			// 			 {
			// 				 axis_value = 0;
			// 			 }

			// 			 // Smoothly ramp from dead zone to maximum value (from ioquake)
			// 			 // https://github.com/ioquake/ioq3/blob/master/code/sdl/sdl_input.c
			// 			 fix_value = ((float) abs(axis_value) / 32767.0f - threshold) / (1.0f - threshold);

			// 			 if (fix_value < 0.0f)
			// 			 {
			// 				 fix_value = 0.0f;
			// 			 }

			// 			 // Apply expo
			// 			 fix_value = pow(fix_value, joy_expo->value);

			// 			 axis_value = (int) (32767 * ((axis_value < 0) ? -fix_value : fix_value));

			// 			 if (cls.key_dest == key_game && (int) cl_paused->value == 0)
			// 			 {
			// 				 if (strcmp(direction_type, "sidemove") == 0)
			// 				 {
			// 					 joystick_sidemove = axis_value * joy_sidesensitivity->value;

			// 					 // We need to be twice faster because with joystic we run...
			// 					 joystick_sidemove *= cl_sidespeed->value * 2.0f;
			// 				 }
			// 				 else if (strcmp(direction_type, "forwardmove") == 0)
			// 				 {
			// 					 joystick_forwardmove = axis_value * joy_forwardsensitivity->value;

			// 					 // We need to be twice faster because with joystic we run...
			// 					 joystick_forwardmove *= cl_forwardspeed->value * 2.0f;
			// 				 }
			// 				 else if (strcmp(direction_type, "yaw") == 0)
			// 				 {
			// 					 joystick_yaw = axis_value * joy_yawsensitivity->value;
			// 					 joystick_yaw *= cl_yawspeed->value;
			// 				 }
			// 				 else if (strcmp(direction_type, "pitch") == 0)
			// 				 {
			// 					 joystick_pitch = axis_value * joy_pitchsensitivity->value;
			// 					 joystick_pitch *= cl_pitchspeed->value;
			// 				 }
			// 				 else if (strcmp(direction_type, "updown") == 0)
			// 				 {
			// 					 joystick_up = axis_value * joy_upsensitivity->value;
			// 					 joystick_up *= cl_upspeed->value;
			// 				 }
			// 			 }

			// 			 if (strcmp(direction_type, "triggerleft") == 0)
			// 			 {
			// 				 qboolean new_left_trigger = abs(axis_value) > (32767 / 4);

			// 				 if (new_left_trigger != left_trigger)
			// 				 {
			// 					 left_trigger = new_left_trigger;
			// 					 Key_Event(K_TRIG_LEFT, left_trigger, true);
			// 				 }
			// 			 }
			// 			 else if (strcmp(direction_type, "triggerright") == 0)
			// 			 {
			// 				 qboolean new_right_trigger = abs(axis_value) > (32767 / 4);

			// 				 if (new_right_trigger != right_trigger)
			// 				 {
			// 					 right_trigger = new_right_trigger;
			// 					 Key_Event(K_TRIG_RIGHT, right_trigger, true);
			// 				 }
			// 			 }

			// 			 break;
			// 		 }

			// 		 // Joystick can have more buttons than on general game controller
			// 		 // so try to map not free buttons
			// 		 case SDL_JOYBUTTONUP:
			// 		 case SDL_JOYBUTTONDOWN:
			// 		 {
			// 			 qboolean down = (event.type == SDL_JOYBUTTONDOWN);

			// 			 // Ignore back button, we don't need event for such button
			// 			 if (back_button_id == event.jbutton.button)
			// 			 {
			// 				 return;
			// 			 }

			// 			 if (event.jbutton.button <= (K_JOY32 - K_JOY1))
			// 			 {
			// 				 Key_Event(event.jbutton.button + K_JOY1, down, true);
			// 			 }

			// 			 break;
			// 		 }

			// 		 case SDL_JOYHATMOTION:
			// 		 {
			// 			 if (last_hat != event.jhat.value)
			// 			 {
			// 				 char diff = last_hat ^event.jhat.value;
			// 				 int i;

			// 				 for (i = 0; i < 4; i++)
			// 				 {
			// 					 if (diff & (1 << i))
			// 					 {
			// 						 // check that we have button up for some bit
			// 						 if (last_hat & (1 << i))
			// 						 {
			// 							 Key_Event(i + K_HAT_UP, false, true);
			// 						 }

			// 						 /* check that we have button down for some bit */
			// 						 if (event.jhat.value & (1 << i))
			// 						 {
			// 							 Key_Event(i + K_HAT_UP, true, true);
			// 						 }
			// 					 }
			// 				 }

			// 				 last_hat = event.jhat.value;
			// 			 }

			// 			 break;
			// 		 }

		case sdl.QUIT:
			T.client.common.Com_Quit()

		default:
			fmt.Printf("SDLEvent 0x%x\n", event.GetType())
		}
	}

	//  /* Grab and ungrab the mouse if the console or the menu is opened */
	//  if (in_grab->value == 3)
	//  {
	// 	 want_grab = windowed_mouse->value;
	//  }
	//  else
	//  {
	// 	 want_grab = (vid_fullscreen->value || in_grab->value == 1 ||
	// 		 (in_grab->value == 2 && windowed_mouse->value));
	//  }

	//  // calling GLimp_GrabInput() each frame is a bit ugly but simple and should work.
	//  // The called SDL functions return after a cheap check, if there's nothing to do.
	//  GLimp_GrabInput(want_grab);

	// We need to save the frame time so other subsystems
	// know the exact time of the last input events.
	T.Sys_frame_time = T.client.common.Sys_Milliseconds()
}

/*
 * Initializes the backend
 */
func (T *QInput) Init(client *qClient) {

	T.client = client

	T.client.common.Com_Printf("------- input initialization -------\n")

	//  mouse_x = mouse_y = 0;
	//  joystick_yaw = joystick_pitch = joystick_forwardmove = joystick_sidemove = 0;

	//  exponential_speedup = Cvar_Get("exponential_speedup", "0", CVAR_ARCHIVE);
	//  freelook = Cvar_Get("freelook", "1", CVAR_ARCHIVE);
	//  in_grab = Cvar_Get("in_grab", "2", CVAR_ARCHIVE);
	//  lookstrafe = Cvar_Get("lookstrafe", "0", CVAR_ARCHIVE);
	//  m_filter = Cvar_Get("m_filter", "0", CVAR_ARCHIVE);
	//  m_up = Cvar_Get("m_up", "1", CVAR_ARCHIVE);
	//  m_forward = Cvar_Get("m_forward", "1", CVAR_ARCHIVE);
	//  m_pitch = Cvar_Get("m_pitch", "0.022", CVAR_ARCHIVE);
	//  m_side = Cvar_Get("m_side", "0.8", CVAR_ARCHIVE);
	//  m_yaw = Cvar_Get("m_yaw", "0.022", CVAR_ARCHIVE);
	//  sensitivity = Cvar_Get("sensitivity", "3", CVAR_ARCHIVE);

	//  joy_haptic_magnitude = Cvar_Get("joy_haptic_magnitude", "0.0", CVAR_ARCHIVE);

	//  joy_yawsensitivity = Cvar_Get("joy_yawsensitivity", "1.0", CVAR_ARCHIVE);
	//  joy_pitchsensitivity = Cvar_Get("joy_pitchsensitivity", "1.0", CVAR_ARCHIVE);
	//  joy_forwardsensitivity = Cvar_Get("joy_forwardsensitivity", "1.0", CVAR_ARCHIVE);
	//  joy_sidesensitivity = Cvar_Get("joy_sidesensitivity", "1.0", CVAR_ARCHIVE);
	//  joy_upsensitivity = Cvar_Get("joy_upsensitivity", "1.0", CVAR_ARCHIVE);
	//  joy_expo = Cvar_Get("joy_expo", "2.0", CVAR_ARCHIVE);

	//  joy_axis_leftx = Cvar_Get("joy_axis_leftx", "sidemove", CVAR_ARCHIVE);
	//  joy_axis_lefty = Cvar_Get("joy_axis_lefty", "forwardmove", CVAR_ARCHIVE);
	//  joy_axis_rightx = Cvar_Get("joy_axis_rightx", "yaw", CVAR_ARCHIVE);
	//  joy_axis_righty = Cvar_Get("joy_axis_righty", "pitch", CVAR_ARCHIVE);
	//  joy_axis_triggerleft = Cvar_Get("joy_axis_triggerleft", "triggerleft", CVAR_ARCHIVE);
	//  joy_axis_triggerright = Cvar_Get("joy_axis_triggerright", "triggerright", CVAR_ARCHIVE);

	//  joy_axis_leftx_threshold = Cvar_Get("joy_axis_leftx_threshold", "0.15", CVAR_ARCHIVE);
	//  joy_axis_lefty_threshold = Cvar_Get("joy_axis_lefty_threshold", "0.15", CVAR_ARCHIVE);
	//  joy_axis_rightx_threshold = Cvar_Get("joy_axis_rightx_threshold", "0.15", CVAR_ARCHIVE);
	//  joy_axis_righty_threshold = Cvar_Get("joy_axis_righty_threshold", "0.15", CVAR_ARCHIVE);
	//  joy_axis_triggerleft_threshold = Cvar_Get("joy_axis_triggerleft_threshold", "0.15", CVAR_ARCHIVE);
	//  joy_axis_triggerright_threshold = Cvar_Get("joy_axis_triggerright_threshold", "0.15", CVAR_ARCHIVE);

	//  windowed_mouse = Cvar_Get("windowed_mouse", "1", CVAR_USERINFO | CVAR_ARCHIVE);

	//  Cmd_AddCommand("+mlook", IN_MLookDown);
	//  Cmd_AddCommand("-mlook", IN_MLookUp);

	//  Cmd_AddCommand("+joyaltselector", IN_JoyAltSelectorDown);
	//  Cmd_AddCommand("-joyaltselector", IN_JoyAltSelectorUp);

	//  SDL_StartTextInput();

	//  /* Joystick init */
	//  if (!SDL_WasInit(SDL_INIT_GAMECONTROLLER | SDL_INIT_HAPTIC))
	//  {
	// 	 if (SDL_Init(SDL_INIT_GAMECONTROLLER | SDL_INIT_HAPTIC) == -1)
	// 	 {
	// 		 Com_Printf ("Couldn't init SDL joystick: %s.\n", SDL_GetError ());
	// 	 }
	// 	 else
	// 	 {
	// 		 Com_Printf ("%i joysticks were found.\n", SDL_NumJoysticks());

	// 		 if (SDL_NumJoysticks() > 0) {
	// 			 for (int i = 0; i < SDL_NumJoysticks(); i++) {
	// 				 joystick = SDL_JoystickOpen(i);

	// 				 Com_Printf ("The name of the joystick is '%s'\n", SDL_JoystickName(joystick));
	// 				 Com_Printf ("Number of Axes: %d\n", SDL_JoystickNumAxes(joystick));
	// 				 Com_Printf ("Number of Buttons: %d\n", SDL_JoystickNumButtons(joystick));
	// 				 Com_Printf ("Number of Balls: %d\n", SDL_JoystickNumBalls(joystick));
	// 				 Com_Printf ("Number of Hats: %d\n", SDL_JoystickNumHats(joystick));

	// 				 joystick_haptic = SDL_HapticOpenFromJoystick(joystick);

	// 				 if (joystick_haptic == NULL)
	// 				 {
	// 					 Com_Printf("Most likely joystick isn't haptic.\n");
	// 				 }
	// 				 else
	// 				 {
	// 					 IN_Haptic_Effects_Info();
	// 				 }

	// 				 if(SDL_IsGameController(i))
	// 				 {
	// 					 SDL_GameControllerButtonBind backBind;
	// 					 controller = SDL_GameControllerOpen(i);

	// 					 Com_Printf ("Controller settings: %s\n", SDL_GameControllerMapping(controller));
	// 					 Com_Printf ("Controller axis: \n");
	// 					 Com_Printf (" * leftx = %s\n", joy_axis_leftx->string);
	// 					 Com_Printf (" * lefty = %s\n", joy_axis_lefty->string);
	// 					 Com_Printf (" * rightx = %s\n", joy_axis_rightx->string);
	// 					 Com_Printf (" * righty = %s\n", joy_axis_righty->string);
	// 					 Com_Printf (" * triggerleft = %s\n", joy_axis_triggerleft->string);
	// 					 Com_Printf (" * triggerright = %s\n", joy_axis_triggerright->string);

	// 					 Com_Printf ("Controller thresholds: \n");
	// 					 Com_Printf (" * leftx = %f\n", joy_axis_leftx_threshold->value);
	// 					 Com_Printf (" * lefty = %f\n", joy_axis_lefty_threshold->value);
	// 					 Com_Printf (" * rightx = %f\n", joy_axis_rightx_threshold->value);
	// 					 Com_Printf (" * righty = %f\n", joy_axis_righty_threshold->value);
	// 					 Com_Printf (" * triggerleft = %f\n", joy_axis_triggerleft_threshold->value);
	// 					 Com_Printf (" * triggerright = %f\n", joy_axis_triggerright_threshold->value);

	// 					 backBind = SDL_GameControllerGetBindForButton(controller, SDL_CONTROLLER_BUTTON_BACK);

	// 					 if (backBind.bindType == SDL_CONTROLLER_BINDTYPE_BUTTON)
	// 					 {
	// 						 back_button_id = backBind.value.button;
	// 						 Com_Printf ("\nBack button JOY%d will be unbindable.\n", back_button_id+1);
	// 					 }

	// 					 break;
	// 				 }
	// 				 else
	// 				 {
	// 					 char joystick_guid[256] = {0};

	// 					 SDL_JoystickGUID guid;
	// 					 guid = SDL_JoystickGetDeviceGUID(i);

	// 					 SDL_JoystickGetGUIDString(guid, joystick_guid, 255);

	// 					 Com_Printf ("To use joystick as game controller please set SDL_GAMECONTROLLERCONFIG:\n");
	// 					 Com_Printf ("e.g.: SDL_GAMECONTROLLERCONFIG='%s,%s,leftx:a0,lefty:a1,rightx:a2,righty:a3,back:b1,...\n", joystick_guid, SDL_JoystickName(joystick));
	// 				 }
	// 			 }
	// 		 }
	// 		 else
	// 		 {
	// 			 joystick_haptic = SDL_HapticOpenFromMouse();

	// 			 if (joystick_haptic == NULL)
	// 			 {
	// 				 Com_Printf("Most likely mouse isn't haptic.\n");
	// 			 }
	// 			 else
	// 			 {
	// 				 IN_Haptic_Effects_Info();
	// 			 }
	// 		 }
	// 	 }
	//  }

	T.client.common.Com_Printf("------------------------------------\n\n")
}
