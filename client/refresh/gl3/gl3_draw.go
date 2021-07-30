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
 * Drawing of all images that are not textures
 *
 * =======================================================================
 */
package gl3

import (
	"fmt"
	"goquake2/shared"
	"log"

	"github.com/go-gl/gl/v3.2-core/gl"
)

func (T *qGl3) drawInitLocal() error {
	/* load console characters */
	T.draw_chars = T.findImage("pics/conchars.pcx", it_pic)
	if T.draw_chars == nil {
		T.ri.Sys_Error(shared.ERR_FATAL, "Couldn't load pics/conchars.pcx")
	}

	// set up attribute layout for 2D textured rendering
	gl.GenVertexArrays(1, &T.vao2D)
	gl.BindVertexArray(T.vao2D)

	gl.GenBuffers(1, &T.vbo2D)
	T.bindVBO(T.vbo2D)

	T.useProgram(T.gl3state.si2D.shaderProgram)

	gl.EnableVertexAttribArray(GL3_ATTRIB_POSITION)
	// Note: the glVertexAttribPointer() configuration is stored in the VAO, not the shader or sth
	//       (that's why I use one VAO per 2D shader)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_POSITION, 2, gl.FLOAT, false, 4*4, 0)

	gl.EnableVertexAttribArray(GL3_ATTRIB_TEXCOORD)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_TEXCOORD, 2, gl.FLOAT, false, 4*4, 2*4)

	// set up attribute layout for 2D flat color rendering

	gl.GenVertexArrays(1, &T.vao2Dcolor)
	gl.BindVertexArray(T.vao2Dcolor)

	T.bindVBO(T.vbo2D) // yes, both VAOs share the same VBO

	T.useProgram(T.gl3state.si2Dcolor.shaderProgram)

	gl.EnableVertexAttribArray(GL3_ATTRIB_POSITION)
	gl.VertexAttribPointerWithOffset(GL3_ATTRIB_POSITION, 2, gl.FLOAT, false, 2*4, 0)

	T.bindVAO(0)
	return nil
}

// bind the texture before calling this
func (T *qGl3) drawTexturedRectangle(x, y, w, h, sl, tl, sh, th float32) {
	/*
	 *  x,y+h      x+w,y+h
	 * sl,th--------sh,th
	 *  |             |
	 *  |             |
	 *  |             |
	 * sl,tl--------sh,tl
	 *  x,y        x+w,y
	 */

	vBuf := []float32{
		//  X,   Y,   S,  T
		x, y + h, sl, th,
		x, y, sl, tl,
		x + w, y + h, sh, th,
		x + w, y, sh, tl,
	}

	T.bindVAO(T.vao2D)

	// Note: while vao2D "remembers" its vbo for drawing, binding the vao does *not*
	//       implicitly bind the vbo, so I need to explicitly bind it before glBufferData()
	T.bindVBO(T.vbo2D)
	gl.BufferData(gl.ARRAY_BUFFER, len(vBuf)*4, gl.Ptr(vBuf), gl.STREAM_DRAW)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

	//glMultiDrawArrays(mode, first, count, drawcount) ??
}

func (T *qGl3) drawFindPic(name string) *gl3image_t {
	if (name[0] != '/') && (name[0] != '\\') {
		fullname := fmt.Sprintf("pics/%s.pcx", name)
		return T.findImage(fullname, it_pic)
	} else {
		return T.findImage(name[1:], it_pic)
	}
}

func (T *qGl3) DrawStretchPic(x, y, w, h int, name string) {
	img := T.drawFindPic(name)
	if img == nil {
		T.rPrintf(shared.PRINT_ALL, "Can't find pic: %s\n", name)
		return
	}

	T.useProgram(T.gl3state.si2D.shaderProgram)
	T.bind(img.texnum)

	T.drawTexturedRectangle(float32(x), float32(y), float32(w), float32(h), img.sl, img.tl, img.sh, img.th)
}

/*
 * This repeats a 64*64 tile graphic to fill
 * the screen around a sized down
 * refresh window.
 */
func (T *qGl3) DrawTileClear(x, y, w, h int, name string) {
	img := T.drawFindPic(name)
	if img == nil {
		T.rPrintf(shared.PRINT_ALL, "Can't find pic: %s\n", name)
		return
	}

	T.useProgram(T.gl3state.si2D.shaderProgram)
	T.bind(img.texnum)

	T.drawTexturedRectangle(float32(x), float32(y), float32(w), float32(h), float32(x)/64.0, float32(y)/64.0, float32(x+w)/64.0, float32(y+h)/64.0)
}

/*
 * Fills a box of pixels with a single color
 */
func (T *qGl3) DrawFill(x, y, w, h, c int) {
	//  union
	//  {
	// 	 unsigned c;
	// 	 byte v[4];
	//  } color;
	//  int i;

	if c < 0 || c > 255 {
		log.Fatal("Draw_Fill: bad color")
	}

	color := T.d_8to24table[c]

	vBuf := []float32{
		//  X,   Y
		float32(x), float32(y + h),
		float32(x), float32(y),
		float32(x + w), float32(y + h),
		float32(x + w), float32(y)}

	//  for(i=0; i<3; ++i)
	//  {
	// 	 gl3state.uniCommonData.color.Elements[i] = color.v[i] * (1.0f/255.0f);
	//  }
	//  gl3state.uniCommonData.color.A = 1.0f;
	T.gl3state.uniCommonData.setColor(1.0, float32((color>>16)&0xFF)/255,
		float32((color>>8)&0xFF)/255, float32((color>>0)&0xFF)/255)

	T.updateUBOCommon()

	T.useProgram(T.gl3state.si2Dcolor.shaderProgram)
	T.bindVAO(T.vao2Dcolor)

	T.bindVBO(T.vbo2D)
	gl.BufferData(gl.ARRAY_BUFFER, len(vBuf)*4, gl.Ptr(vBuf), gl.STREAM_DRAW)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

func (T *qGl3) drawGetPalette() error {

	/* get the palette */
	_, pal, _, _ := shared.LoadPCX(T.ri, "pics/colormap.pcx", false, true)
	if pal == nil {
		return T.ri.Sys_Error(shared.ERR_FATAL, "Couldn't load pics/colormap.pcx")
	}

	T.d_8to24table = make([]uint32, 256)
	for i := 0; i < 256; i++ {
		r := pal[i*3+0]
		g := pal[i*3+1]
		b := pal[i*3+2]

		v := uint32((255)<<24) + (uint32(r) << 0) + (uint32(g) << 8) + (uint32(b) << 16)
		T.d_8to24table[i] = v
	}

	T.d_8to24table[255] = T.d_8to24table[255] & 0xFFFFFF /* 255 is transparent */

	return nil
}
