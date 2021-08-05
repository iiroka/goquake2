/*
 * Copyright (C) 1997-2001 Id Software, Inc.
 * Copyright (C) 2016-2017 Daniel Gibson
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
 * Texture handling for OpenGL3
 *
 * =======================================================================
 */
package gl3

import (
	"fmt"
	"goquake2/shared"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v3.2-core/gl"
)

type glmode_t struct {
	name     string
	minimize int32
	maximize int32
}

var modes = []glmode_t{
	{"GL_NEAREST", gl.NEAREST, gl.NEAREST},
	{"GL_LINEAR", gl.LINEAR, gl.LINEAR},
	{"GL_NEAREST_MIPMAP_NEAREST", gl.NEAREST_MIPMAP_NEAREST, gl.NEAREST},
	{"GL_LINEAR_MIPMAP_NEAREST", gl.LINEAR_MIPMAP_NEAREST, gl.LINEAR},
	{"GL_NEAREST_MIPMAP_LINEAR", gl.NEAREST_MIPMAP_LINEAR, gl.NEAREST},
	{"GL_LINEAR_MIPMAP_LINEAR", gl.LINEAR_MIPMAP_LINEAR, gl.LINEAR},
}

func (T *qGl3) textureMode(str string) {
	// const int num_modes = sizeof(modes)/sizeof(modes[0]);
	// int i;

	index := -1
	for i, m := range modes {
		if m.name == str {
			index = i
			break
		}
	}

	if index < 0 {
		T.rPrintf(shared.PRINT_ALL, "bad filter name '%s' (probably from gl_texturemode)\n", str)
		return
	}

	T.gl_filter_min = modes[index].minimize
	T.gl_filter_max = modes[index].maximize

	/* clamp selected anisotropy */
	if T.gl3config.anisotropic {
		if T.gl_anisotropic.Int() > int(T.gl3config.max_anisotropy) {
			T.ri.Cvar_Set("r_anisotropic", fmt.Sprintf("%v", T.gl3config.max_anisotropy))
		}
	} else {
		T.ri.Cvar_Set("r_anisotropic", "0.0")
	}

	// gl3image_t *glt;

	// const char* nolerplist = gl_nolerp_list->string;

	/* change all the existing texture objects */
	for i := 0; i < T.numgl3textures; i++ {
		glt := &T.gl3textures[i]
		// 	if (nolerplist != NULL && strstr(nolerplist, glt->name) != NULL)
		// 	{
		// 		continue; /* those (by default: font and crosshairs) always only use GL_NEAREST */
		// 	}

		T.selectTMU(gl.TEXTURE0)
		T.bind(glt.texnum)
		if (glt.itype != it_pic) && (glt.itype != it_sky) { /* mipmapped texture */
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, T.gl_filter_min)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, T.gl_filter_max)

			// 		/* Set anisotropic filter if supported and enabled */
			// 		if (gl3config.anisotropic && gl_anisotropic->value)
			// 		{
			// 			glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAX_ANISOTROPY_EXT, max(gl_anisotropic->value, 1.f));
			// 		}
		} else { /* texture has no mipmaps */
			// we can't use gl_filter_min which might be GL_*_MIPMAP_*
			// also, there's no anisotropic filtering for textures w/o mipmaps
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, T.gl_filter_max)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, T.gl_filter_max)
		}
	}
}

func (T *qGl3) bind(texnum uint32) {
	// extern gl3image_t *draw_chars;

	// if (gl_nobind->value && draw_chars) { /* performance evaluation option */
	// 	texnum = draw_chars->texnum;
	// }

	if T.gl3state.currenttexture == texnum {
		return
	}

	T.gl3state.currenttexture = texnum
	T.selectTMU(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texnum)
}

func (T *qGl3) bindLightmap(lightmapnum int) {

	if lightmapnum < 0 || lightmapnum >= MAX_LIGHTMAPS {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Invalid lightmapnum %v used!\n", lightmapnum)
		return
	}

	if T.gl3state.currentlightmap == lightmapnum {
		return
	}

	T.gl3state.currentlightmap = lightmapnum
	lmindex := lightmapnum * MAX_LIGHTMAPS_PER_SURFACE
	for i := 0; i < MAX_LIGHTMAPS_PER_SURFACE; i++ {
		// this assumes that GL_TEXTURE<i+1> = GL_TEXTURE<i> + 1
		// at least for GL_TEXTURE0 .. GL_TEXTURE31 that's true
		T.selectTMU(gl.TEXTURE1 + uint32(i))
		gl.BindTexture(gl.TEXTURE_2D, T.gl3state.lightmap_textureIDs[lmindex+i])
	}
}

/*
 * Returns has_alpha
 */
func (T *qGl3) upload32(data unsafe.Pointer, width, height int, mipmap bool) bool {

	c := width * height
	samples := gl3_solid_format
	comp := gl3_tex_solid_format

	scan := uintptr(data)
	for i := 0; i < c; i++ {
		if (scan & 0xFF000000) != 0xFF000000 {
			samples = gl3_alpha_format
			comp = gl3_tex_alpha_format
			break
		}
		scan++
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, int32(comp), int32(width), int32(height),
		0, gl.RGBA, gl.UNSIGNED_BYTE, data)

	res := (samples == gl3_alpha_format)

	if mipmap {
		// TODO: some hardware may require mipmapping disabled for NPOT textures!
		gl.GenerateMipmap(gl.TEXTURE_2D)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, T.gl_filter_min)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, T.gl_filter_max)
	} else { // if the texture has no mipmaps, we can't use gl_filter_min which might be GL_*_MIPMAP_*
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, T.gl_filter_max)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, T.gl_filter_max)
	}

	if mipmap && T.gl3config.anisotropic && T.gl_anisotropic.Bool() {
		// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAX_ANISOTROPY_EXT, max(T.gl_anisotropic.Int(), 1))
	}

	return res
}

/*
 * Returns has_alpha
 */
func (T *qGl3) upload8(data []byte, width, height int, mipmap, is_sky bool) bool {
	//  unsigned trans[512 * 256];
	//  int i, s;
	//  int p;

	s := width * height

	//  if (s > sizeof(trans) / 4)
	//  {
	// 	 ri.Sys_Error(ERR_DROP, "GL3_Upload8: too large");
	//  }
	trans := make([]uint32, s)

	for i := 0; i < s; i++ {
		p := data[i]
		trans[i] = T.d_8to24table[p]

		/* transparent, so scan around for
		another color to avoid alpha fringes */
		if p == 255 {
			// 		 if ((i > width) && (data[i - width] != 255))
			// 		 {
			// 			 p = data[i - width];
			// 		 }
			// 		 else if ((i < s - width) && (data[i + width] != 255))
			// 		 {
			// 			 p = data[i + width];
			// 		 }
			// 		 else if ((i > 0) && (data[i - 1] != 255))
			// 		 {
			// 			 p = data[i - 1];
			// 		 }
			// 		 else if ((i < s - 1) && (data[i + 1] != 255))
			// 		 {
			// 			 p = data[i + 1];
			// 		 }
			// 		 else
			// 		 {
			// 			 p = 0;
			// 		 }

			// 		 /* copy rgb components */
			// 		 ((byte *)&trans[i])[0] = ((byte *)&d_8to24table[p])[0];
			// 		 ((byte *)&trans[i])[1] = ((byte *)&d_8to24table[p])[1];
			// 		 ((byte *)&trans[i])[2] = ((byte *)&d_8to24table[p])[2];
		}
	}

	return T.upload32(gl.Ptr(trans), width, height, mipmap)
}

/*
 * This is also used as an entry point for the generated r_notexture
 */
func (T *qGl3) loadPic(name string, pic []byte, width, realwidth,
	height, realheight int, itype imagetype_t, bits int) *gl3image_t {
	//  gl3image_t *image = NULL;
	//  GLuint texNum=0;
	//  int i;

	nolerp := false

	if T.gl_nolerp_list != nil && len(T.gl_nolerp_list.String) > 0 {
		nolerp = strings.Contains(T.gl_nolerp_list.String, name)
	}
	/* find a free gl3image_t */
	var image *gl3image_t
	for i := 0; i < T.numgl3textures; i++ {
		if T.gl3textures[i].texnum == 0 {
			image = &T.gl3textures[i]
			break
		}
	}

	if image == nil {
		if T.numgl3textures == MAX_GL3TEXTURES {
			T.ri.Sys_Error(shared.ERR_DROP, "MAX_GLTEXTURES")
		}

		image = &T.gl3textures[T.numgl3textures]
		T.numgl3textures++
	}

	//  if (strlen(name) >= sizeof(image->name))
	//  {
	// 	 ri.Sys_Error(ERR_DROP, "GL3_LoadPic: \"%s\" is too long", name);
	//  }

	image.name = name
	image.registration_sequence = T.registration_sequence

	image.width = width
	image.height = height
	image.itype = itype

	//  if ((type == it_skin) && (bits == 8)) {
	// 	 FloodFillSkin(pic, width, height);
	//  }

	// image->scrap = false; // TODO: reintroduce scrap? would allow optimizations in 2D rendering..

	var texNum uint32
	gl.GenTextures(1, &texNum)

	image.texnum = texNum

	T.selectTMU(gl.TEXTURE0)
	T.bind(texNum)

	if bits == 8 {
		image.has_alpha = T.upload8(pic, width, height,
			(image.itype != it_pic && image.itype != it_sky),
			image.itype == it_sky)
	} else {
		image.has_alpha = T.upload32(gl.Ptr(pic), width, height,
			(image.itype != it_pic && image.itype != it_sky))
	}

	if realwidth != 0 && realheight != 0 {
		if (realwidth <= image.width) && (realheight <= image.height) {
			image.width = realwidth
			image.height = realheight
		} else {
			T.rPrintf(shared.PRINT_DEVELOPER,
				"Warning, image '%s' has hi-res replacement smaller than the original! (%d x %d) < (%d x %d)\n",
				name, image.width, image.height, realwidth, realheight)
		}
	}

	image.sl = 0
	image.sh = 1
	image.tl = 0
	image.th = 1

	if nolerp {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	}
	return image
}

func (T *qGl3) loadWal(origname string, itype imagetype_t) *gl3image_t {
	// miptex_t *mt;
	// int width, height, ofs, size;
	// gl3image_t *image;
	// char name[256];

	name := origname

	/* Add the extension */
	if !strings.HasSuffix(name, "wal") {
		name += ".wal"
	}

	buf, err := T.ri.LoadFile(name)
	if err != nil || buf == nil {
		T.rPrintf(shared.PRINT_ALL, "LoadWal: can't load %s\n", name)
		return T.gl3_notexture
	}

	// if (size < sizeof(miptex_t))
	// {
	// 	R_Printf(PRINT_ALL, "LoadWal: can't load %s, small header\n", name);
	// 	ri.FS_FreeFile((void *)mt);
	// 	return gl3_notexture;
	// }

	mt := shared.Miptex(buf)
	if (mt.Offsets[0] <= 0) || (mt.Width <= 0) || (mt.Height <= 0) ||
		(((len(buf) - int(mt.Offsets[0])) / int(mt.Height)) < int(mt.Width)) {
		T.rPrintf(shared.PRINT_ALL, "LoadWal: can't load %s, small body\n", name)
		return T.gl3_notexture
	}

	return T.loadPic(name, buf[mt.Offsets[0]:], int(mt.Width), 0, int(mt.Height), 0, itype, 8)
}

/*
 * Finds or loads the given image
 */
func (T *qGl3) findImage(name string, itype imagetype_t) *gl3image_t {
	//  gl3image_t *image;
	//  int i, len;
	//  byte *pic;
	//  int width, height;
	//  char *ptr;
	//  char namewe[256];
	//  int realwidth = 0, realheight = 0;
	//  const char* ext;

	if len(name) == 0 {
		return nil
	}

	//  ext = COM_FileExtension(name);
	//  if(!ext[0])
	//  {
	// 	 /* file has no extension */
	// 	 return NULL;
	//  }

	//  len = strlen(name);

	//  /* Remove the extension */
	//  memset(namewe, 0, 256);
	//  memcpy(namewe, name, len - (strlen(ext) + 1));

	//  if (len < 5)
	//  {
	// 	 return NULL;
	//  }

	//  /* fix backslashes */
	//  while ((ptr = strchr(name, '\\')))
	//  {
	// 	 *ptr = '/';
	//  }

	/* look for it */
	for i := 0; i < T.numgl3textures; i++ {
		if T.gl3textures[i].name == name {
			T.gl3textures[i].registration_sequence = T.registration_sequence
			return &T.gl3textures[i]
		}
	}

	//  /* load the pic from disk */
	//  pic = NULL;

	if strings.HasSuffix(name, ".pcx") {
		// 	 if (gl_retexturing->value)
		// 	 {
		// 		 GetPCXInfo(name, &realwidth, &realheight);
		// 		 if(realwidth == 0)
		// 		 {
		// 			 /* No texture found */
		// 			 return NULL;
		// 		 }

		// 		 /* try to load a tga, png or jpg (in that order/priority) */
		// 		 if (  LoadSTB(namewe, "tga", &pic, &width, &height)
		// 			|| LoadSTB(namewe, "png", &pic, &width, &height)
		// 			|| LoadSTB(namewe, "jpg", &pic, &width, &height) )
		// 		 {
		// 			 /* upload tga or png or jpg */
		// 			 image = GL3_LoadPic(name, pic, width, realwidth, height,
		// 					 realheight, type, 32);
		// 		 }
		// 		 else
		// 		 {
		// 			 /* PCX if no TGA/PNG/JPEG available (exists always) */
		// 			 LoadPCX(name, &pic, NULL, &width, &height);

		// 			 if (!pic)
		// 			 {
		// 				 /* No texture found */
		// 				 return NULL;
		// 			 }

		// 			 /* Upload the PCX */
		// 			 image = GL3_LoadPic(name, pic, width, 0, height, 0, type, 8);
		// 		 }
		// 	 }
		// 	 else /* gl_retexture is not set */
		// 	 {
		pic, _, width, height := shared.LoadPCX(T.ri, name, true, false)

		if pic == nil {
			return nil
		}

		return T.loadPic(name, pic, width, 0, height, 0, itype, 8)
		// 	 }
	} else if strings.HasSuffix(name, ".wal") {
		//  else if (strcmp(ext, "wal") == 0 || strcmp(ext, "m8") == 0)
		//  {
		// 	 if (gl_retexturing->value)
		// 	 {
		// 		 /* Get size of the original texture */
		// 		 if (strcmp(ext, "m8") == 0)
		// 		 {
		// 			 GetM8Info(name, &realwidth, &realheight);
		// 		 }
		// 		 else
		// 		 {
		// 			 GetWalInfo(name, &realwidth, &realheight);
		// 		 }

		// 		 if(realwidth == 0)
		// 		 {
		// 			 /* No texture found */
		// 			 return NULL;
		// 		 }

		// 		 /* try to load a tga, png or jpg (in that order/priority) */
		// 		 if (  LoadSTB(namewe, "tga", &pic, &width, &height)
		// 			|| LoadSTB(namewe, "png", &pic, &width, &height)
		// 			|| LoadSTB(namewe, "jpg", &pic, &width, &height) )
		// 		 {
		// 			 /* upload tga or png or jpg */
		// 			 image = GL3_LoadPic(name, pic, width, realwidth, height, realheight, type, 32);
		// 		 }
		// 		 else if (strcmp(ext, "m8") == 0)
		// 		 {
		// 			 image = LoadM8(namewe, type);
		// 		 }
		// 		 else
		// 		 {
		// 			 /* WAL if no TGA/PNG/JPEG available (exists always) */
		// 			 image = LoadWal(namewe, type);
		// 		 }

		// 		 if (!image)
		// 		 {
		// 			 /* No texture found */
		// 			 return NULL;
		// 		 }
		// 	 }
		// 	 else if (strcmp(ext, "m8") == 0)
		// 	 {
		// 		 image = LoadM8(name, type);

		// 		 if (!image)
		// 		 {
		// 			 /* No texture found */
		// 			 return NULL;
		// 		 }
		// 	 }
		// 	 else /* gl_retexture is not set */
		// 	 {
		return T.loadWal(name, itype)

		// 		 if (!image)
		// 		 {
		// 			 /* No texture found */
		// 			 return NULL;
		// 		 }
		// 	 }
		//  }
		//  else if (strcmp(ext, "tga") == 0 || strcmp(ext, "png") == 0 || strcmp(ext, "jpg") == 0)
		//  {
		// 	 char tmp_name[256];

		// 	 realwidth = 0;
		// 	 realheight = 0;

		// 	 strcpy(tmp_name, namewe);
		// 	 strcat(tmp_name, ".wal");
		// 	 GetWalInfo(tmp_name, &realwidth, &realheight);

		// 	 if (realwidth == 0 || realheight == 0) {
		// 		 strcpy(tmp_name, namewe);
		// 		 strcat(tmp_name, ".m8");
		// 		 GetM8Info(tmp_name, &realwidth, &realheight);
		// 	 }

		// 	 if (realwidth == 0 || realheight == 0) {
		// 		 /* It's a sky or model skin. */
		// 		 strcpy(tmp_name, namewe);
		// 		 strcat(tmp_name, ".pcx");
		// 		 GetPCXInfo(tmp_name, &realwidth, &realheight);
		// 	 }

		// 	 /* TODO: not sure if not having realwidth/heigth is bad - a tga/png/jpg
		// 	  * was requested, after all, so there might be no corresponding wal/pcx?
		// 	  * if (realwidth == 0 || realheight == 0) return NULL;
		// 	  */

		// 	 if(LoadSTB(name, ext, &pic, &width, &height))
		// 	 {
		// 		 image = GL3_LoadPic(name, pic, width, realwidth, height, realheight, type, 32);
		// 	 } else {
		// 		 return NULL;
		// 	 }
		//  }
	}

	//  if (pic)
	//  {
	// 	 free(pic);
	//  }

	return nil
}

func (T *qGl3) RegisterSkin(name string) interface{} {
	return T.findImage(name, it_skin)
}
