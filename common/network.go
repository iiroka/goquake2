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
 * Low level network code, based upon the BSD socket api.
 *
 * =======================================================================
 */
package common

import (
	"fmt"
	"goquake2/shared"
	"net"
	"strconv"
	"time"
)

func (T *qCommon) NET_GetPacket(sock shared.Netsrc_t) (*shared.Netadr_t, []byte) {
	index := (int(sock) + 1) & 1
	select {
	case c := <-T.loopback[index]:
		a := &shared.Netadr_t{}
		a.Type = shared.NA_LOOPBACK
		return a, c
	default:
		var msg [shared.MAX_MSGLEN]byte
		if T.ip_sockets[sock] != nil {
			T.ip_sockets[sock].SetReadDeadline(time.Now().Add(time.Microsecond))
			r, addr, err := T.ip_sockets[sock].ReadFromUDP(msg[:])
			if err != nil {
				// T.Com_Printf("ERROR receiving %s\n", err.Error())
			}
			if r > 0 {
				println("Received", r, addr)
				a := &shared.Netadr_t{}
				a.Type = shared.NA_IP
				a.Port = uint16(addr.Port)
				a.Ip[0] = addr.IP[0]
				a.Ip[1] = addr.IP[1]
				a.Ip[2] = addr.IP[2]
				a.Ip[3] = addr.IP[3]
				return a, msg[:r]
			}
		}
		return nil, nil
	}
}

func (T *qCommon) NET_SendPacket(sock shared.Netsrc_t, data []byte, to shared.Netadr_t) error {
	// 	int ret;
	// 	struct sockaddr_storage addr;
	// 	int net_socket;
	// 	int addr_size = sizeof(struct sockaddr_in);

	switch to.Type {
	case shared.NA_LOOPBACK:
		index := int(sock)
		T.loopback[index] <- data
		return nil

	case shared.NA_BROADCAST:
		net_socket := T.ip_sockets[sock]
		if net_socket == nil {
			println("NO SOCKET")
			return nil
		}
		_, err := net_socket.Write(data)
		if err != nil {
			T.Com_Printf("NET_SendPacket ERROR: %s to %v\n", err.Error(), to)
			return err
		}
		return nil

	case shared.NA_IP:
		// net_socket := T.ip_sockets[sock]
		// if net_socket == nil {
		// 	println("NO SOCKET")
		// 	return nil
		// }
		// addr := &net.UDPAddr{
		// 	IP:   net.IPv4(to.Ip[0], to.Ip[1], to.Ip[2], to.Ip[3]),
		// 	Port: int(to.Port),
		// }
		return nil

	case shared.NA_IP6,
		shared.NA_MULTICAST6:
		// 			net_socket = ip6_sockets[sock];
		// 			addr_size = sizeof(struct sockaddr_in6);
		// 			if (!net_socket) {
		return nil
	// 			}

	// 		case NA_IPX:
	// 		case NA_BROADCAST_IPX:
	// 			net_socket = ipx_sockets[sock];
	// 			if (!net_socket) {
	// 				return;
	// 			}
	// 			break;

	default:
		return T.Com_Error(shared.ERR_FATAL, "NET_SendPacket: bad address type")
	}

	// 	NetadrToSockadr(&to, &addr);

	// 	/* Re-check the address family. If to.type is NA_IP6 but
	// 	   contains an IPv4 mapped address, NetadrToSockadr will
	// 	   return an AF_INET struct.  If so, switch back to AF_INET
	// 	   socket.*/
	// 	if ((to.type == NA_IP6) && (addr.ss_family == AF_INET))
	// 	{
	// 		net_socket = ip_sockets[sock];
	// 		addr_size = sizeof(struct sockaddr_in);

	// 		if (!net_socket)
	// 		{
	// 			return;
	// 		}
	// 	}

	// 	if (addr.ss_family == AF_INET6)
	// 	{
	// 		struct sockaddr_in6 *s6 = (struct sockaddr_in6 *)&addr;

	// 		/* If multicast socket, must specify scope.
	// 		   So multicast_interface must be specified */
	// 		if (IN6_IS_ADDR_MULTICAST(&s6->sin6_addr))
	// 		{
	// 			struct addrinfo hints;
	// 			struct addrinfo *res;
	// 			char tmp[128];

	// 			if (multicast_interface != NULL)
	// 			{
	// 				int error;
	// 				char mcast_addr[128], mcast_port[10];

	// 				/* Do a getnameinfo/getaddrinfo cycle
	// 				   to calculate the scope_id of the
	// 				   multicast address. getaddrinfo is
	// 				   passed a multicast address of the
	// 				   form ff0x::xxx%multicast_interface */
	// #ifdef SIN6_LEN
	// 				error = getnameinfo((struct sockaddr *)s6, s6->sin6_len, tmp,
	// 						sizeof(tmp), NULL, 0, NI_NUMERICHOST);
	// #else
	// 				error = getnameinfo((struct sockaddr *)s6,
	// 							sizeof(struct sockaddr_in6),
	// 							tmp, sizeof(tmp), NULL, 0, NI_NUMERICHOST);
	// #endif

	// 				if (error)
	// 				{
	// 					Com_Printf("NET_SendPacket: getnameinfo: %s\n",
	// 							gai_strerror(error));
	// 					return;
	// 				}

	// 				Com_sprintf(mcast_addr, sizeof(mcast_addr), "%s%%%s", tmp,
	// 						multicast_interface);
	// 				Com_sprintf(mcast_port, sizeof(mcast_port), "%d",
	// 						ntohs(s6->sin6_port));
	// 				memset(&hints, 0, sizeof(hints));
	// 				hints.ai_family = AF_INET6;
	// 				hints.ai_socktype = SOCK_DGRAM;
	// 				hints.ai_flags = AI_NUMERICHOST;

	// 				error = getaddrinfo(mcast_addr, mcast_port, &hints, &res);

	// 				if (error)
	// 				{
	// 					Com_Printf("NET_SendPacket: getaddrinfo: %s\n",
	// 							gai_strerror(error));
	// 					return;
	// 				}

	// 				/* sockaddr_in6 should now have a valid scope_id. */
	// 				memcpy(s6, res->ai_addr, res->ai_addrlen);
	// 				freeaddrinfo(res);
	// 			}
	// 			else
	// 			{
	// 				Com_Printf("NET_SendPacket: IPv6 multicast destination but +set multicast not specified: %s\n",
	// 						inet_ntop(AF_INET6, &s6->sin6_addr, tmp, sizeof(tmp)));
	// 				return;
	// 			}
	// 		}
	// 	}

	// println("WriteToUDP")
	// r, err := net_socket.WriteToUDP(data, addr)
	// if err != nil {
	// 	T.Com_Printf("NET_SendPacket ERROR: %s to %v\n", err.Error(), to)
	// 	return err
	// }
	// T.Com_Printf("NET_SendPacket %v\n", r)

	// 	ret = sendto(net_socket,
	// 			data,
	// 			length,
	// 			0,
	// 			(struct sockaddr *)&addr,
	// 			addr_size);

	// 	if (ret == -1)
	// 	{
	// 		Com_Printf("NET_SendPacket ERROR: %s to %s\n", NET_ErrorString(),
	// 				NET_AdrToString(to));
	// 	}
	return nil
}

func (T *qCommon) netOpenIP() {

	println("netOpenIP")
	port := T.Cvar_Get("port", strconv.Itoa(shared.PORT_SERVER), shared.CVAR_NOSET)
	ip := T.Cvar_Get("ip", "localhost", shared.CVAR_NOSET)
	host := ip.String
	if len(host) == 0 || host == "localhost" {
		host = "0.0.0.0"
	}

	// if (!ip6_sockets[NS_SERVER])
	// {
	// 	ip6_sockets[NS_SERVER] = NET_Socket(ip->string, port->value,
	// 			NS_SERVER, AF_INET6);
	// }

	// if (!ip6_sockets[NS_CLIENT])
	// {
	// 	ip6_sockets[NS_CLIENT] = NET_Socket(ip->string, PORT_ANY,
	// 			NS_CLIENT, AF_INET6);
	// }

	if T.ip_sockets[shared.NS_SERVER] == nil {
		addr := net.UDPAddr{
			Port: port.Int(),
			IP:   net.ParseIP(host),
		}
		sock, err := net.ListenUDP("udp4", &addr)
		if err != nil {
			fmt.Printf("Failed to open SERVER socket %s\n", err.Error())
		} else {
			T.ip_sockets[shared.NS_SERVER] = sock
		}
	}

	if T.ip_sockets[shared.NS_CLIENT] == nil {
		addr := net.UDPAddr{
			Port: port.Int(),
			IP:   net.IPv4(255, 255, 255, 255),
		}
		sock, err := net.DialUDP("udp4", nil, &addr)
		if err != nil {
			fmt.Printf("Failed to open CLIENT socket %s\n", err.Error())
		} else {
			T.ip_sockets[shared.NS_CLIENT] = sock
		}
	}
}

/*
 * A single player game will only use the loopback code
 */
func (T *qCommon) NET_Config(multiplayer bool) {
	println("NET_Config", multiplayer)
	if !multiplayer {

		/* shut down any existing sockets */
		for i := 0; i < 2; i++ {
			if T.ip_sockets[i] != nil {
				T.ip_sockets[i].Close()
				T.ip_sockets[i] = nil
			}

			// 	if (ip6_sockets[i]) {
			// 		close(ip6_sockets[i]);
			// 		ip6_sockets[i] = 0;
			// 	}

			// 	if (ipx_sockets[i]) {
			// 		close(ipx_sockets[i]);
			// 		ipx_sockets[i] = 0;
			// 	}
		}
	} else {
		/* open sockets */
		T.netOpenIP()
	}
}
