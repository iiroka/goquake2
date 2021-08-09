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
 * The "camera" through which the player looks into the game.
 *
 * =======================================================================
 */
package game

import "goquake2/shared"

/*
 * Called for each player at the end of
 * the server frame and right after spawning
 */
func (G *qGame) clientEndServerFrame(ent *edict_t) {
	//  float bobtime;
	//  int i;

	if ent == nil {
		return
	}

	G.current_player = ent
	G.current_client = ent.client

	/* If the origin or velocity have changed since ClientThink(),
	update the pmove values. This will happen when the client
	is pushed by a bmodel or kicked by an explosion.
	If it wasn't updated here, the view position would lag a frame
	behind the body position when pushed -- "sinking into plats" */
	for i := 0; i < 3; i++ {
		G.current_client.ps.Pmove.Origin[i] = int16(ent.s.Origin[i] * 8.0)
		G.current_client.ps.Pmove.Velocity[i] = int16(ent.velocity[i] * 8.0)
	}

	/* If the end of unit layout is displayed, don't give
	the player any normal movement attributes */
	//  if (level.intermissiontime) {
	// 	 current_client->ps.blend[3] = 0;
	// 	 current_client->ps.fov = 90;
	// 	 G_SetStats(ent);
	// 	 return;
	//  }

	shared.AngleVectors(ent.client.v_angle[:], G.player_view_forward[:], G.player_view_right[:], G.player_view_up[:])

	/* burn from lava, etc */
	//  P_WorldEffects();

	/* set model angles from view angles so other things in
	the world can tell which direction you are looking */
	if ent.client.v_angle[shared.PITCH] > 180 {
		ent.s.Angles[shared.PITCH] = (-360 + ent.client.v_angle[shared.PITCH]) / 3
	} else {
		ent.s.Angles[shared.PITCH] = ent.client.v_angle[shared.PITCH] / 3
	}

	ent.s.Angles[shared.YAW] = ent.client.v_angle[shared.YAW]
	ent.s.Angles[shared.ROLL] = 0
	//  ent->s.angles[ROLL] = SV_CalcRoll(ent->s.angles, ent->velocity) * 4;

	/* calculate speed and cycle to be used for
	all cyclic walking effects */
	//  xyspeed = sqrt(
	// 		 ent->velocity[0] * ent->velocity[0] + ent->velocity[1] *
	// 		 ent->velocity[1]);

	//  if (xyspeed < 5) {
	// 	 bobmove = 0;
	// 	 current_client->bobtime = 0; /* start at beginning of cycle again */
	//  }
	//  else if (ent->groundentity)
	//  {
	// 	 /* so bobbing only cycles when on ground */
	// 	 if (xyspeed > 210)
	// 	 {
	// 		 bobmove = 0.25;
	// 	 }
	// 	 else if (xyspeed > 100)
	// 	 {
	// 		 bobmove = 0.125;
	// 	 }
	// 	 else
	// 	 {
	// 		 bobmove = 0.0625;
	// 	 }
	//  }

	//  bobtime = (current_client->bobtime += bobmove);

	//  if (current_client->ps.pmove.pm_flags & PMF_DUCKED)
	//  {
	// 	 bobtime *= 4;
	//  }

	//  bobcycle = (int)bobtime;
	//  bobfracsin = fabs(sin(bobtime * M_PI));

	//  /* detect hitting the floor */
	//  P_FallingDamage(ent);

	//  /* apply all the damage taken this frame */
	//  P_DamageFeedback(ent);

	//  /* determine the view offsets */
	//  SV_CalcViewOffset(ent);

	//  /* determine the gun offsets */
	//  SV_CalcGunOffset(ent);

	//  /* determine the full screen color blend
	// 	must be after viewoffset, so eye contents
	// 	can be accurately determined */
	//  SV_CalcBlend(ent);

	//  /* chase cam stuff */
	//  if (ent->client->resp.spectator)
	//  {
	// 	 G_SetSpectatorStats(ent);
	//  }
	//  else
	//  {
	// 	 G_SetStats(ent);
	//  }

	//  G_CheckChaseStats(ent);

	//  G_SetClientEvent(ent);

	//  G_SetClientEffects(ent);

	//  G_SetClientSound(ent);

	//  G_SetClientFrame(ent);

	copy(ent.client.oldvelocity[:], ent.velocity[:])
	copy(ent.client.oldviewangles[:], ent.client.ps.Viewangles[:])

	/* clear weapon kicks */
	copy(ent.client.kick_origin[:], []float32{0, 0, 0})
	copy(ent.client.kick_angles[:], []float32{0, 0, 0})

	//  if (!(level.framenum & 31))
	//  {
	// 	 /* if the scoreboard is up, update it */
	// 	 if (ent->client->showscores)
	// 	 {
	// 		 DeathmatchScoreboardMessage(ent, ent->enemy);
	// 		 gi.unicast(ent, false);
	// 	 }

	// 	 /* if the help computer is up, update it */
	// 	 if (ent->client->showhelp)
	// 	 {
	// 		 ent->client->pers.helpchanged = 0;
	// 		 HelpComputerMessage(ent);
	// 		 gi.unicast(ent, false);
	// 	 }
	//  }

	//  /* if the inventory is up, update it */
	//  if (ent->client->showinventory)
	//  {
	// 	 InventoryMessage(ent);
	// 	 gi.unicast(ent, false);
	//  }
}
