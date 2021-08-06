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
 * The PCX file format
 *
 * =======================================================================
 */
package shared

import "strings"

/* PCX files are used for as many images as possible */

type pcx_t struct {
	manufacturer   uint8
	version        uint8
	encoding       uint8
	bits_per_pixel uint8
	xmin           uint16
	ymin           uint16
	xmax           uint16
	ymax           uint16
	hres           uint16
	vres           uint16
	// unsigned short xmin, ymin, xmax, ymax;
	// unsigned short hres, vres;
	// unsigned char palette[48];
	// char reserved;
	color_planes   uint8
	bytes_per_line uint16
	palette_type   uint16
	// unsigned short bytes_per_line;
	// unsigned short palette_type;
	// char filler[58];
	// unsigned char data;   /* unbounded */
}

const pcx_size = 128

func readPcx(data []byte) pcx_t {
	p := pcx_t{}
	p.manufacturer = data[0]
	p.version = data[1]
	p.encoding = data[2]
	p.bits_per_pixel = data[3]
	p.xmin = ReadUint16(data[4:])
	p.ymin = ReadUint16(data[6:])
	p.xmax = ReadUint16(data[8:])
	p.ymax = ReadUint16(data[10:])
	p.hres = ReadUint16(data[12:])
	p.vres = ReadUint16(data[104:])
	return p
}

func LoadPCX(ri Refimport_t, origname string, loadPic, loadPal bool) ([]byte, []byte, int, int) {

	filename := origname

	/* Add the extension */
	if !strings.HasSuffix(filename, "pcx") {
		filename = filename + ".pcx"
	}

	var pic []byte = nil
	var palette []byte = nil

	/* load the file */
	raw, err := ri.LoadFile(filename)
	if err != nil {
		return nil, nil, -1, -1
	}
	if raw == nil || len(raw) < pcx_size {
		ri.Com_VPrintf(PRINT_DEVELOPER, "Bad pcx file %s\n", filename)
		return nil, nil, -1, -1
	}

	/* parse the PCX file */
	pcx := readPcx(raw)

	pcx_width := int(pcx.xmax - pcx.xmin)
	pcx_height := int(pcx.ymax - pcx.ymin)

	if (pcx.manufacturer != 0x0a) || (pcx.version != 5) ||
		(pcx.encoding != 1) || (pcx.bits_per_pixel != 8) ||
		(pcx_width >= 4096) || (pcx_height >= 4096) {
		ri.Com_VPrintf(PRINT_ALL, "Bad pcx file %s\n", filename)
		return nil, nil, -1, -1
	}

	// full_size := (pcx_height + 1) * (pcx_width + 1)
	// out = malloc(full_size);
	// if (!out)
	// {
	// 	R_Printf(PRINT_ALL, "Can't allocate\n");
	// 	ri.FS_FreeFile(pcx);
	// 	return;
	// }

	// *pic = out;

	// pix = out;

	if loadPal {
		palette = make([]byte, 768)
		copy(palette, raw[len(raw)-768:])
	}

	if loadPic {
		full_size := (pcx_height + 1) * (pcx_width + 1)
		pic = make([]byte, full_size)
		src_i := pcx_size
		pix_i := 0
		for y := 0; y <= pcx_height; y++ {
			for x := 0; x <= pcx_width; {
				// 		if (raw - (byte *)pcx > len) {
				// 			// no place for read
				// 			image_issues = true;
				// 			x = pcx_width;
				// 			break;
				// 		}
				dataByte := raw[src_i]
				src_i++
				runLength := 1

				if (dataByte & 0xC0) == 0xC0 {
					runLength = int(dataByte & 0x3F)
					// 			if (raw - (byte *)pcx > len) {
					// 				// no place for read
					// 				image_issues = true;
					// 				x = pcx_width;
					// 				break;
					// 			}
					dataByte = raw[src_i]
					src_i++
				}

				for runLength > 0 {
					// 			if ((*pic + full_size) <= (pix + x))
					// 			{
					// 				// no place for write
					// 				image_issues = true;
					// 				x += runLength;
					// 				runLength = 0;
					// 			}
					// 			else
					// 			{
					pic[pix_i+x] = dataByte
					x++
					// 			}
					runLength--
				}
			}
			pix_i += pcx_width + 1
		}

		// if (raw - (byte *)pcx > len)
		// {
		// 	R_Printf(PRINT_DEVELOPER, "PCX file %s was malformed", filename);
		// 	free(*pic);
		// 	*pic = NULL;
		// }
		// else if(pcx_width == 319 && pcx_height == 239
		// 		&& Q_strcasecmp(origname, "pics/quit.pcx") == 0
		// 		&& Com_BlockChecksum(pcx, len) == 3329419434u)
		// {
		// 	// it's the quit screen, and the baseq2 one (identified by checksum)
		// 	// so fix it
		// 	fixQuitScreen(*pic);
		// }

		// if (image_issues)
		// {
		// 	R_Printf(PRINT_ALL, "PCX file %s has possible size issues.\n", filename);
		// }
	}

	// ri.FS_FreeFile(pcx);
	return pic, palette, pcx_width + 1, pcx_height + 1
}
