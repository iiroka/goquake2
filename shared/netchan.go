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
 * The low level, platform independant network code
 *
 * =======================================================================
 */
package shared

type Netchan_t struct {
	common      QCommon
	fatal_error bool

	sock Netsrc_t

	dropped int /* between last packet and previous */

	last_received int /* for timeouts */
	LastSent      int /* for retransmits */

	remote_address Netadr_t
	Qport          int /* qport value to write when transmitting */

	/* sequencing variables */
	incoming_sequence              int
	incoming_acknowledged          int
	incoming_reliable_acknowledged int /* single bit */

	incoming_reliable_sequence int /* single bit, maintained local */

	outgoing_sequence      int
	reliable_sequence      int /* single bit */
	last_reliable_sequence int /* sequence number of last send */

	/* reliable staging and holding areas */
	Message *QWritebuf /* writing buffer to send to server */
	// byte message_buf[MAX_MSGLEN - 16];          /* leave space for header */

	/* message is copied to this buffer when it is first transfered */
	reliable_length int
	reliable_buf    []byte
	// byte reliable_buf[MAX_MSGLEN - 16];         /* unacked reliable message */
}

/*
 * called to open a channel to a remote system
 */
func (ch *Netchan_t) Setup(common QCommon, sock Netsrc_t, adr Netadr_t, qport int) {
	ch.common = common
	ch.fatal_error = false
	ch.sock = sock
	ch.dropped = 0
	ch.last_received = common.Curtime()
	ch.LastSent = 0
	ch.remote_address = adr
	ch.Qport = qport
	ch.incoming_sequence = 0
	ch.incoming_acknowledged = 0
	ch.incoming_reliable_acknowledged = 0
	ch.incoming_reliable_sequence = 0
	ch.outgoing_sequence = 1
	ch.reliable_sequence = 0
	ch.last_reliable_sequence = 0
	ch.reliable_length = 0
	ch.reliable_buf = make([]byte, MAX_MSGLEN-16)
	ch.Message = QWritebufCreate(MAX_MSGLEN - 16)
	ch.Message.Allowoverflow = true
}

/*
 * Returns true if the last reliable message has acked
 */
func (ch *Netchan_t) canReliable() bool {
	if ch.reliable_length > 0 {
		return false /* waiting for ack */
	}

	return true
}

func (ch *Netchan_t) needReliable() bool {

	/* if the remote side dropped the last reliable message, resend it */
	send_reliable := false

	if (ch.incoming_acknowledged > ch.last_reliable_sequence) &&
		(ch.incoming_reliable_acknowledged != ch.reliable_sequence) {
		send_reliable = true
	}

	/* if the reliable transmit buffer is empty, copy the current message out */
	if ch.reliable_length == 0 && ch.Message.Cursize > 0 {
		send_reliable = true
	}

	return send_reliable
}

/*
 * tries to send an unreliable message to a connection, and handles the
 * transmition / retransmition of the reliable messages.
 *
 * A 0 length will still generate a packet and deal with the reliable messages.
 */
func (ch *Netchan_t) Transmit(data []byte) {
	//  sizebuf_t send;
	//  byte send_buf[MAX_MSGLEN];
	//  qboolean send_reliable;
	//  unsigned w1, w2;

	/* check for message overflow */
	if ch.Message.Overflowed {
		ch.fatal_error = true
		ch.common.Com_Printf("%v:Outgoing message overflow\n",
			ch.remote_address)
		return
	}

	send_reliable := ch.needReliable()

	if ch.reliable_length == 0 && ch.Message.Cursize > 0 {
		copy(ch.reliable_buf, ch.Message.Data())
		ch.reliable_length = ch.Message.Cursize
		ch.Message.Cursize = 0
		ch.reliable_sequence ^= 1
	}

	/* write the packet header */
	send := QWritebufCreate(MAX_MSGLEN)

	w1 := (ch.outgoing_sequence & 0x7FFFFFFF)
	if send_reliable {
		w1 |= 0x80000000
	}
	w2 := (ch.incoming_sequence & 0x7FFFFFFF) | (ch.incoming_reliable_sequence << 31)

	ch.outgoing_sequence++
	ch.LastSent = ch.common.Curtime()

	send.WriteLong(w1)
	send.WriteLong(w2)

	/* send the qport if we are a client */
	if ch.sock == NS_CLIENT {
		send.WriteShort(ch.common.QPort())
	}

	/* copy the reliable message to the packet first */
	if send_reliable {
		send.Write(ch.reliable_buf[0:ch.reliable_length])
		ch.last_reliable_sequence = ch.outgoing_sequence
	}

	/* add the unreliable part if space is available */
	if data != nil {
		if len(send.data)-send.Cursize >= len(data) {
			send.Write(data)
		} else {
			ch.common.Com_Printf("Netchan_Transmit: dumped unreliable\n")
		}
	}

	/* send the datagram */
	ch.common.NET_SendPacket(ch.sock, send.Data(), ch.remote_address)

	//  if (showpackets->value)
	//  {
	// 	 if (send_reliable)
	// 	 {
	// 		 Com_Printf("send %4i : s=%i reliable=%i ack=%i rack=%i\n",
	// 				 send.cursize, chan->outgoing_sequence - 1,
	// 				 chan->reliable_sequence, chan->incoming_sequence,
	// 				 chan->incoming_reliable_sequence);
	// 	 }
	// 	 else
	// 	 {
	// 		 Com_Printf("send %4i : s=%i ack=%i rack=%i\n",
	// 				 send.cursize, chan->outgoing_sequence - 1,
	// 				 chan->incoming_sequence,
	// 				 chan->incoming_reliable_sequence);
	// 	 }
	//  }
}

/*
 * called when the current net_message is from remote_address
 * modifies net_message so that it points to the packet payload
 */
func (ch *Netchan_t) Process(msg *QReadbuf) bool {
	//  unsigned sequence, sequence_ack;
	//  unsigned reliable_ack, reliable_message;

	/* get sequence numbers */
	msg.BeginReading()
	sequence := msg.ReadLong()
	sequence_ack := msg.ReadLong()

	/* read the qport if we are a server */
	if ch.sock == NS_SERVER {
		msg.ReadShort()
	}

	reliable_message := (sequence >> 31) & 1
	reliable_ack := (sequence_ack >> 31) & 1

	sequence &= 0x7FFFFFFF
	sequence_ack &= 0x7FFFFFFF

	//  if (showpackets->value)
	//  {
	// 	 if (reliable_message)
	// 	 {
	// 		 Com_Printf("recv %4i : s=%i reliable=%i ack=%i rack=%i\n",
	// 				 msg->cursize, sequence,
	// 				 chan->incoming_reliable_sequence ^ 1,
	// 				 sequence_ack, reliable_ack);
	// 	 }
	// 	 else
	// 	 {
	// 		 Com_Printf("recv %4i : s=%i ack=%i rack=%i\n",
	// 				 msg->cursize, sequence, sequence_ack,
	// 				 reliable_ack);
	// 	 }
	//  }

	/* discard stale or duplicated packets */
	if sequence <= ch.incoming_sequence {
		// 	 if (showdrop->value)
		// 	 {
		// 		 Com_Printf("%s:Out of order packet %i at %i\n",
		// 				 NET_AdrToString(chan->remote_address),
		// 				 sequence, chan->incoming_sequence);
		// 	 }

		return false
	}

	/* dropped packets don't keep the message from being used */
	ch.dropped = sequence - (ch.incoming_sequence + 1)

	if ch.dropped > 0 {
		// 	 if (showdrop->value)
		// 	 {
		// 		 Com_Printf("%s:Dropped %i packets at %i\n",
		// 				 NET_AdrToString(chan->remote_address),
		// 				 chan->dropped, sequence);
		// 	 }
	}

	/* if the current outgoing reliable message has been acknowledged
	 * clear the buffer to make way for the next */
	if reliable_ack == ch.reliable_sequence {
		ch.reliable_length = 0 /* it has been received */
	}

	/* if this message contains a reliable message, bump incoming_reliable_sequence */
	ch.incoming_sequence = sequence
	ch.incoming_acknowledged = sequence_ack
	ch.incoming_reliable_acknowledged = reliable_ack

	if reliable_message != 0 {
		ch.incoming_reliable_sequence ^= 1
	}

	/* the message can now be read from the current message pointer */
	ch.last_received = ch.common.Curtime()

	return true
}
