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
 * OpenGL3 refresher: Handling shaders
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

func (T *qGl3) compileShader(shaderType uint32, cshaderSrc string, shaderSrc2 *string) uint32 {
	shader := gl.CreateShader(shaderType)

	var csources **uint8
	var free func()
	// const char* sources[2] = { shaderSrc, shaderSrc2 };
	// int numSources = shaderSrc2 != NULL ? 2 : 1;
	numSources := 1
	if shaderSrc2 != nil {
		csources, free = gl.Strs(cshaderSrc, *shaderSrc2)
		numSources = 2
	} else {
		csources, free = gl.Strs(cshaderSrc)
	}

	gl.ShaderSource(shader, int32(numSources), csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status != gl.TRUE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		shaderTypeStr := ""
		switch shaderType {
		case gl.VERTEX_SHADER:
			shaderTypeStr = "Vertex"
		case gl.FRAGMENT_SHADER:
			shaderTypeStr = "Fragment"
		case gl.GEOMETRY_SHADER:
			shaderTypeStr = "Geometry"
			/* not supported in OpenGL3.2 and we're unlikely to need/use them anyway
			case GL_COMPUTE_SHADER:  shaderTypeStr = "Compute"; break;
			case GL_TESS_CONTROL_SHADER:    shaderTypeStr = "TessControl"; break;
			case GL_TESS_EVALUATION_SHADER: shaderTypeStr = "TessEvaluation"; break;
			*/
		}
		T.rPrintf(shared.PRINT_ALL, "ERROR: Compiling %s Shader failed: %s\n", shaderTypeStr, log)
		gl.DeleteShader(shader)

		return 0
	}

	return shader
}

func (T *qGl3) createShaderProgram(shaders []uint32) uint32 {
	// int i=0;
	shaderProgram := gl.CreateProgram()
	if shaderProgram == 0 {
		T.rPrintf(shared.PRINT_ALL, "ERROR: Couldn't create a new Shader Program!\n")
		return 0
	}

	for _, sh := range shaders {
		gl.AttachShader(shaderProgram, sh)
	}

	// make sure all shaders use the same attribute locations for common attributes
	// (so the same VAO can easily be used with different shaders)
	gl.BindAttribLocation(shaderProgram, GL3_ATTRIB_POSITION, gl.Str("position\x00"))
	gl.BindAttribLocation(shaderProgram, GL3_ATTRIB_TEXCOORD, gl.Str("texCoord\x00"))
	gl.BindAttribLocation(shaderProgram, GL3_ATTRIB_LMTEXCOORD, gl.Str("lmTexCoord\x00"))
	gl.BindAttribLocation(shaderProgram, GL3_ATTRIB_COLOR, gl.Str("vertColor\x00"))
	gl.BindAttribLocation(shaderProgram, GL3_ATTRIB_NORMAL, gl.Str("normal\x00"))
	gl.BindAttribLocation(shaderProgram, GL3_ATTRIB_LIGHTFLAGS, gl.Str("lightFlags\x00"))

	// the following line is not necessary/implicit (as there's only one output)
	// glBindFragDataLocation(shaderProgram, 0, "outColor"); XXX would this even be here?

	gl.LinkProgram(shaderProgram)

	var status int32
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status != gl.TRUE {

		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))

		T.rPrintf(shared.PRINT_ALL, "ERROR: Linking shader program failed: %s\n", log)

		gl.DeleteProgram(shaderProgram)

		return 0
	}

	// for(i=0; i<numShaders; ++i)
	// {
	// 	// after linking, they don't need to be attached anymore.
	// 	// no idea  why they even are, if they don't have to..
	// 	glDetachShader(shaderProgram, shaders[i]);
	// }

	return shaderProgram
}

const (
	GL3_BINDINGPOINT_UNICOMMON = 0
	GL3_BINDINGPOINT_UNI2D     = 1
	GL3_BINDINGPOINT_UNI3D     = 2
	GL3_BINDINGPOINT_UNILIGHTS = 3
)

func (T *qGl3) initShader2D(shaderInfo *gl3ShaderInfo_t, vertSrc, fragSrc string) bool {
	// GLuint prog = 0;

	// if(shaderInfo->shaderProgram != 0) {
	// 	R_Printf(PRINT_ALL, "WARNING: calling initShader2D for gl3ShaderInfo_t that already has a shaderProgram!\n");
	// 	glDeleteProgram(shaderInfo->shaderProgram);
	// }

	//shaderInfo->uniColor = shaderInfo->uniProjMatrix = shaderInfo->uniModelViewMatrix = -1;
	shaderInfo.shaderProgram = 0
	shaderInfo.uniLmScales = -1

	shaders2D := make([]uint32, 2)
	shaders2D[0] = T.compileShader(gl.VERTEX_SHADER, vertSrc, nil)
	if shaders2D[0] == 0 {
		return false
	}

	shaders2D[1] = T.compileShader(gl.FRAGMENT_SHADER, fragSrc, nil)
	if shaders2D[1] == 0 {
		gl.DeleteShader(shaders2D[0])
		return false
	}

	prog := T.createShaderProgram(shaders2D)

	// I think the shaders aren't needed anymore once they're linked into the program
	gl.DeleteShader(shaders2D[0])
	gl.DeleteShader(shaders2D[1])

	if prog == 0 {
		return false
	}

	shaderInfo.shaderProgram = prog
	T.useProgram(prog)

	// Bind the buffer object to the uniform blocks
	blockIndex := gl.GetUniformBlockIndex(prog, gl.Str("uniCommon\x00"))
	if blockIndex != gl.INVALID_INDEX {
		var blockSize int32
		gl.GetActiveUniformBlockiv(prog, blockIndex, gl.UNIFORM_BLOCK_DATA_SIZE, &blockSize)
		if int(blockSize) != len(T.gl3state.uniCommonData.data)*4 {
			T.rPrintf(shared.PRINT_ALL, "WARNING: OpenGL driver disagrees with us about UBO size of 'uniCommon': %v vs %v\n",
				blockSize, len(T.gl3state.uniCommonData.data)*4)

			gl.DeleteProgram(prog)
			return false
		}

		gl.UniformBlockBinding(prog, blockIndex, GL3_BINDINGPOINT_UNICOMMON)
	} else {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Couldn't find uniform block index 'uniCommon'\n")
		gl.DeleteProgram(prog)
		return false
	}

	blockIndex = gl.GetUniformBlockIndex(prog, gl.Str("uni2D\x00"))
	if blockIndex != gl.INVALID_INDEX {
		var blockSize int32
		gl.GetActiveUniformBlockiv(prog, blockIndex, gl.UNIFORM_BLOCK_DATA_SIZE, &blockSize)
		if int(blockSize) != len(T.gl3state.uni2DData.data)*4 {
			T.rPrintf(shared.PRINT_ALL, "WARNING: OpenGL driver disagrees with us about UBO size of 'uni2D'\n")
			gl.DeleteProgram(prog)
			return false
		}

		gl.UniformBlockBinding(prog, blockIndex, GL3_BINDINGPOINT_UNI2D)
	} else {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Couldn't find uniform block index 'uni2D'\n")
		gl.DeleteProgram(prog)
		return false
	}

	return true
}

func (T *qGl3) initShader3D(shaderInfo *gl3ShaderInfo_t, vertSrc, fragSrc string) bool {
	// GLuint shaders3D[2] = {0};
	// GLuint prog = 0;
	// int i=0;

	// if(shaderInfo->shaderProgram != 0)
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: calling initShader3D for gl3ShaderInfo_t that already has a shaderProgram!\n");
	// 	glDeleteProgram(shaderInfo->shaderProgram);
	// }

	shaderInfo.shaderProgram = 0
	shaderInfo.uniLmScales = -1

	shaders3D := make([]uint32, 2)
	shaders3D[0] = T.compileShader(gl.VERTEX_SHADER, vertexCommon3D, &vertSrc)
	if shaders3D[0] == 0 {
		return false
	}

	shaders3D[1] = T.compileShader(gl.FRAGMENT_SHADER, fragmentCommon3D, &fragSrc)
	if shaders3D[1] == 0 {
		gl.DeleteShader(shaders3D[0])
		return false
	}

	prog := T.createShaderProgram(shaders3D)
	if prog == 0 {
		gl.DeleteShader(shaders3D[0])
		gl.DeleteShader(shaders3D[1])
		return false
	}

	T.useProgram(prog)

	// Bind the buffer object to the uniform blocks
	blockIndex := gl.GetUniformBlockIndex(prog, gl.Str("uniCommon\x00"))
	if blockIndex != gl.INVALID_INDEX {
		var blockSize int32
		gl.GetActiveUniformBlockiv(prog, blockIndex, gl.UNIFORM_BLOCK_DATA_SIZE, &blockSize)
		if int(blockSize) != len(T.gl3state.uniCommonData.data)*4 {
			T.rPrintf(shared.PRINT_ALL, "WARNING: OpenGL driver disagrees with us about UBO size of 'uniCommon': %v vs %v\n",
				blockSize, len(T.gl3state.uniCommonData.data)*4)

			gl.DeleteShader(shaders3D[0])
			gl.DeleteShader(shaders3D[1])
			gl.DeleteProgram(prog)
			return false
		}

		gl.UniformBlockBinding(prog, blockIndex, GL3_BINDINGPOINT_UNICOMMON)
	} else {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Couldn't find uniform block index 'uniCommon'\n")
		gl.DeleteShader(shaders3D[0])
		gl.DeleteShader(shaders3D[1])
		gl.DeleteProgram(prog)
		return false
	}

	blockIndex = gl.GetUniformBlockIndex(prog, gl.Str("uni3D\x00"))
	if blockIndex != gl.INVALID_INDEX {
		var blockSize int32
		gl.GetActiveUniformBlockiv(prog, blockIndex, gl.UNIFORM_BLOCK_DATA_SIZE, &blockSize)
		if int(blockSize) != len(T.gl3state.uni3DData.data)*4 {
			T.rPrintf(shared.PRINT_ALL, "WARNING: OpenGL driver disagrees with us about UBO size of 'uni3D': %v vs %v\n",
				blockSize, len(T.gl3state.uni3DData.data)*4)

			gl.DeleteShader(shaders3D[0])
			gl.DeleteShader(shaders3D[1])
			gl.DeleteProgram(prog)
			return false
		}

		gl.UniformBlockBinding(prog, blockIndex, GL3_BINDINGPOINT_UNI3D)
	} else {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Couldn't find uniform block index 'uni3D'\n")
		gl.DeleteShader(shaders3D[0])
		gl.DeleteShader(shaders3D[1])
		gl.DeleteProgram(prog)
		return false
	}

	blockIndex = gl.GetUniformBlockIndex(prog, gl.Str("uniLights\x00"))
	if blockIndex != gl.INVALID_INDEX {
		var blockSize int32
		gl.GetActiveUniformBlockiv(prog, blockIndex, gl.UNIFORM_BLOCK_DATA_SIZE, &blockSize)
		if int(blockSize) != len(T.gl3state.uniLightsData.data)*4 {
			T.rPrintf(shared.PRINT_ALL, "WARNING: OpenGL driver disagrees with us about UBO size of 'uniLights'\n")
			T.rPrintf(shared.PRINT_ALL, "         OpenGL says %d, we say %d\n", blockSize, len(T.gl3state.uniLightsData.data)*4)
			gl.DeleteShader(shaders3D[0])
			gl.DeleteShader(shaders3D[1])
			gl.DeleteProgram(prog)
			return false
		}

		gl.UniformBlockBinding(prog, blockIndex, GL3_BINDINGPOINT_UNILIGHTS)
	}
	// else: as uniLights is only used in the LM shaders, it's ok if it's missing

	// make sure texture is GL_TEXTURE0
	texLoc := gl.GetUniformLocation(prog, gl.Str("tex\x00"))
	if texLoc != -1 {
		gl.Uniform1i(texLoc, 0)
	}

	// ..  and the 4 lightmap texture use GL_TEXTURE1..4
	// char lmName[10] = "lightmapX";
	for i := 0; i < 4; i++ {
		lmName := fmt.Sprintf("lightmap%v\x00", i)
		lmLoc := gl.GetUniformLocation(prog, gl.Str(lmName))
		if lmLoc != -1 {
			gl.Uniform1i(lmLoc, int32(i+1)) // lightmap0 belongs to GL_TEXTURE1, lightmap1 to GL_TEXTURE2 etc
		}
	}

	lmScalesLoc := gl.GetUniformLocation(prog, gl.Str("lmScales\x00"))
	shaderInfo.uniLmScales = lmScalesLoc
	if lmScalesLoc != -1 {
		for j := 1; j < 4; j++ {
			shaderInfo.lmScales[0][j] = 1.0
		}

		for i := 1; i < 4; i++ {
			for j := 1; j < 4; j++ {
				shaderInfo.lmScales[i][j] = 0.0
			}
		}

		// gl.Uniform4fv(lmScalesLoc, 4, &shaderInfo.lmScales[:])
	}

	shaderInfo.shaderProgram = prog

	// I think the shaders aren't needed anymore once they're linked into the program
	// glDeleteShader(shaders3D[0]);
	// glDeleteShader(shaders3D[1]);

	return true

	// err_cleanup:

	// 	glDeleteShader(shaders3D[0]);
	// 	glDeleteShader(shaders3D[1]);

	// 	if(prog != 0)  glDeleteProgram(prog);

	// 	return false;
}

func (T *qGl3) initUBOs() {
	T.gl3state.uniCommonData.data = make([]float32, 8)
	T.gl3state.uniCommonData.setGamma(float32(1.0 / T.vid_gamma.Float()))
	T.gl3state.uniCommonData.setIntensity(float32(T.gl3_intensity.Float()))
	T.gl3state.uniCommonData.setIntensity2D(float32(T.gl3_intensity_2D.Float()))
	T.gl3state.uniCommonData.setColor(1, 1, 1, 1)

	gl.GenBuffers(1, &T.gl3state.uniCommonUBO)
	gl.BindBuffer(gl.UNIFORM_BUFFER, T.gl3state.uniCommonUBO)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, GL3_BINDINGPOINT_UNICOMMON, T.gl3state.uniCommonUBO)
	gl.BufferData(gl.UNIFORM_BUFFER, len(T.gl3state.uniCommonData.data)*4, gl.Ptr(T.gl3state.uniCommonData.data), gl.DYNAMIC_DRAW)

	// the matrix will be set to something more useful later, before being used
	T.gl3state.uni2DData.data = make([]float32, 16)

	gl.GenBuffers(1, &T.gl3state.uni2DUBO)
	gl.BindBuffer(gl.UNIFORM_BUFFER, T.gl3state.uni2DUBO)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, GL3_BINDINGPOINT_UNI2D, T.gl3state.uni2DUBO)
	gl.BufferData(gl.UNIFORM_BUFFER, len(T.gl3state.uni2DData.data)*4, gl.Ptr(T.gl3state.uni2DData.data), gl.DYNAMIC_DRAW)

	// the matrices will be set to something more useful later, before being used
	T.gl3state.uni3DData.data = make([]float32, gl3Uni3D_Size)
	// gl3state.uni3DData.transProjMat4 = HMM_Mat4();
	// gl3state.uni3DData.transViewMat4 = HMM_Mat4();
	T.gl3state.uni3DData.setTransModelMat4(gl3_identityMat4)
	T.gl3state.uni3DData.setScroll(0.0)
	T.gl3state.uni3DData.setTime(0.0)
	T.gl3state.uni3DData.setAlpha(1.0)
	// gl3_overbrightbits 0 means "no scaling" which is equivalent to multiplying with 1
	if T.gl3_overbrightbits.Float() <= 0.0 {
		T.gl3state.uni3DData.setOverbrightbits(1.0)
	} else {
		T.gl3state.uni3DData.setOverbrightbits(T.gl3_overbrightbits.Float())
	}
	T.gl3state.uni3DData.setParticleFadeFactor(T.gl3_particle_fade_factor.Float())

	gl.GenBuffers(1, &T.gl3state.uni3DUBO)
	gl.BindBuffer(gl.UNIFORM_BUFFER, T.gl3state.uni3DUBO)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, GL3_BINDINGPOINT_UNI3D, T.gl3state.uni3DUBO)
	gl.BufferData(gl.UNIFORM_BUFFER, len(T.gl3state.uni3DData.data)*4, gl.Ptr(T.gl3state.uni3DData.data), gl.DYNAMIC_DRAW)

	T.gl3state.uniLightsData.data = make([]uint32, gl3UniLights_Size)
	T.gl3state.uniLightsData.initialize()

	gl.GenBuffers(1, &T.gl3state.uniLightsUBO)
	gl.BindBuffer(gl.UNIFORM_BUFFER, T.gl3state.uniLightsUBO)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, GL3_BINDINGPOINT_UNILIGHTS, T.gl3state.uniLightsUBO)
	gl.BufferData(gl.UNIFORM_BUFFER, len(T.gl3state.uniLightsData.data)*4, gl.Ptr(T.gl3state.uniLightsData.data), gl.DYNAMIC_DRAW)

	T.gl3state.currentUBO = T.gl3state.uniLightsUBO
}

func (T *qGl3) createShaders() bool {
	if !T.initShader2D(&T.gl3state.si2D, vertexSrc2D, fragmentSrc2D) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for textured 2D rendering!\n")
		return false
	}
	if !T.initShader2D(&T.gl3state.si2Dcolor, vertexSrc2Dcolor, fragmentSrc2Dcolor) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for color-only 2D rendering!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3Dlm, vertexSrc3Dlm, fragmentSrc3Dlm) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for textured 3D rendering with lightmap!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3Dtrans, vertexSrc3D, fragmentSrc3D) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for rendering translucent 3D things!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3DcolorOnly, vertexSrc3D, fragmentSrc3Dcolor) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for flat-colored 3D rendering!\n")
		return false
	}
	/*
		if(!initShader3D(&gl3state.si3Dlm, vertexSrc3Dlm, fragmentSrc3D)) {
			R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for blending 3D lightmaps rendering!\n");
			return false;
		}
	*/
	if !T.initShader3D(&T.gl3state.si3Dturb, vertexSrc3Dwater, fragmentSrc3Dwater) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for water rendering!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3DlmFlow, vertexSrc3DlmFlow, fragmentSrc3Dlm) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for scrolling textured 3D rendering with lightmap!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3DtransFlow, vertexSrc3Dflow, fragmentSrc3D) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for scrolling textured translucent 3D rendering!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3Dsky, vertexSrc3D, fragmentSrc3Dsky) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for sky rendering!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3Dsprite, vertexSrc3D, fragmentSrc3Dsprite) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for sprite rendering!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3DspriteAlpha, vertexSrc3D, fragmentSrc3DspriteAlpha) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for alpha-tested sprite rendering!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3Dalias, vertexSrcAlias, fragmentSrcAlias) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for rendering textured models!\n")
		return false
	}
	if !T.initShader3D(&T.gl3state.si3DaliasColor, vertexSrcAlias, fragmentSrcAliasColor) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for rendering flat-colored models!\n")
		return false
	}

	particleFrag := fragmentSrcParticles
	if T.gl3_particle_square.Float() != 0.0 {
		particleFrag = fragmentSrcParticlesSquare
	}

	if !T.initShader3D(&T.gl3state.siParticle, vertexSrcParticles, particleFrag) {
		T.rPrintf(shared.PRINT_ALL, "WARNING: Failed to create shader program for rendering particles!\n")
		return false
	}

	T.gl3state.currentShaderProgram = 0

	return true
}

func (T *qGl3) initShaders() bool {
	T.initUBOs()
	return T.createShaders()
}

func (T *qGl3) updateUBO(ubo uint32, size int, data unsafe.Pointer) {
	if T.gl3state.currentUBO != ubo {
		T.gl3state.currentUBO = ubo
		gl.BindBuffer(gl.UNIFORM_BUFFER, ubo)
	}

	// http://docs.gl/gl3/glBufferSubData says  "When replacing the entire data store,
	// consider using glBufferSubData rather than completely recreating the data store
	// with glBufferData. This avoids the cost of reallocating the data store."
	// no idea why glBufferData() doesn't just do that when size doesn't change, but whatever..
	// however, it also says glBufferSubData() might cause a stall so I DON'T KNOW!
	// on Linux/nvidia, by just looking at the fps, glBufferData() and glBufferSubData() make no difference
	// TODO: STREAM instead DYNAMIC?

	// #if 1
	// this seems to be reasonably fast everywhere.. glMapBuffer() seems to be a bit faster on OSX though..
	gl.BufferData(gl.UNIFORM_BUFFER, size, data, gl.DYNAMIC_DRAW)
	// #elif 0
	// 	// on OSX this is super slow (200fps instead of 470-500), BUT it is as fast as glBufferData() when orphaning first
	// 	// nvidia/linux-blob doesn't care about this vs glBufferData()
	// 	// AMD open source linux (R3 370) is also slower here (not as bad as OSX though)
	// 	// intel linux doesn't seem to care either (maybe 3% faster, but that might be imagination)
	// 	// AMD Windows legacy driver (Radeon HD 6950) doesn't care, all 3 alternatives seem to be equally fast
	// 	//glBufferData(GL_UNIFORM_BUFFER, size, NULL, GL_DYNAMIC_DRAW); // orphan
	// 	glBufferSubData(GL_UNIFORM_BUFFER, 0, size, data);
	// #else
	// 	// with my current nvidia-driver (GTX 770, 375.39), the following *really* makes it slower. (<140fps instead of ~850)
	// 	// on OSX (Intel Haswell Iris Pro, OSX 10.11) this is fastest (~500fps instead of ~470)
	// 	// on Linux/intel (Ivy Bridge HD-4000, Linux 4.4) this might be a tiny bit faster than the alternatives..
	// 	glBufferData(GL_UNIFORM_BUFFER, size, NULL, GL_DYNAMIC_DRAW); // orphan
	// 	GLvoid* ptr = glMapBuffer(GL_UNIFORM_BUFFER, GL_WRITE_ONLY);
	// 	memcpy(ptr, data, size);
	// 	glUnmapBuffer(GL_UNIFORM_BUFFER);
	// #endif

	// TODO: another alternative: glMapBufferRange() and each time update a different part
	//       of buffer asynchronously (GL_MAP_UNSYNCHRONIZED_BIT) => ringbuffer style
	//       when starting again from the beginning, synchronization must happen I guess..
	//       also, orphaning might be necessary
	//       and somehow make sure the new range is used by the UBO => glBindBufferRange()
	//  see http://git.quintin.ninja/mjones/Dolphin/blob/4a463f4588e2968c499236458c5712a489622633/Source/Plugins/Plugin_VideoOGL/Src/ProgramShaderCache.cpp#L207
	//   or https://github.com/dolphin-emu/dolphin/blob/master/Source/Core/VideoBackends/OGL/ProgramShaderCache.cpp
}

func (T *qGl3) updateUBOCommon() {
	T.updateUBO(T.gl3state.uniCommonUBO, len(T.gl3state.uniCommonData.data)*4, gl.Ptr(T.gl3state.uniCommonData.data))
}

func (T *qGl3) updateUBO2D() {
	T.updateUBO(T.gl3state.uni2DUBO, len(T.gl3state.uni2DData.data)*5, gl.Ptr(T.gl3state.uni2DData.data))
}

// ############## shaders for 2D rendering (HUD, menus, console, videos, ..) #####################

const vertexSrc2D = `#version 150

in vec2 position; // GL3_ATTRIB_POSITION
in vec2 texCoord; // GL3_ATTRIB_TEXCOORD

// for UBO shared between 2D shaders
layout (std140) uniform uni2D
{
	mat4 trans;
};

out vec2 passTexCoord;

void main()
{
	gl_Position = trans * vec4(position, 0.0, 1.0);
	passTexCoord = texCoord;
}
` + "\x00"

const fragmentSrc2D = `#version 150

in vec2 passTexCoord;

// for UBO shared between all shaders (incl. 2D)
layout (std140) uniform uniCommon
{
	float gamma;
	float intensity;
	float intensity2D; // for HUD, menu etc

	vec4 color;
};

uniform sampler2D tex;

out vec4 outColor;

void main()
{
	vec4 texel = texture(tex, passTexCoord);
	// the gl1 renderer used glAlphaFunc(GL_GREATER, 0.666);
	// and glEnable(GL_ALPHA_TEST); for 2D rendering
	// this should do the same
	if(texel.a <= 0.666)
		discard;

	// apply gamma correction and intensity
	texel.rgb *= intensity2D;
	outColor.rgb = pow(texel.rgb, vec3(gamma));
	outColor.a = texel.a; // I think alpha shouldn't be modified by gamma and intensity
}
` + "\x00"

// 2D color only rendering, GL3_Draw_Fill(), GL3_Draw_FadeScreen()
const vertexSrc2Dcolor = `#version 150

in vec2 position; // GL3_ATTRIB_POSITION

// for UBO shared between 2D shaders
layout (std140) uniform uni2D
{
	mat4 trans;
};

void main()
{
	gl_Position = trans * vec4(position, 0.0, 1.0);
}
` + "\x00"

const fragmentSrc2Dcolor = `#version 150

// for UBO shared between all shaders (incl. 2D)
layout (std140) uniform uniCommon
{
	float gamma;
	float intensity;
	float intensity2D; // for HUD, menus etc

	vec4 color;
};

out vec4 outColor;

void main()
{
	vec3 col = color.rgb * intensity2D;
	outColor.rgb = pow(col, vec3(gamma));
	outColor.a = color.a;
}
` + "\x00"

// ############## shaders for 3D rendering #####################

const vertexCommon3D = `#version 150

	in vec3 position;   // GL3_ATTRIB_POSITION
	in vec2 texCoord;   // GL3_ATTRIB_TEXCOORD
	in vec2 lmTexCoord; // GL3_ATTRIB_LMTEXCOORD
	in vec4 vertColor;  // GL3_ATTRIB_COLOR
	in vec3 normal;     // GL3_ATTRIB_NORMAL
	in uint lightFlags; // GL3_ATTRIB_LIGHTFLAGS

	out vec2 passTexCoord;

	// for UBO shared between all 3D shaders
	layout (std140) uniform uni3D
	{
		mat4 transProj;
		mat4 transView;
		mat4 transModel;

		float scroll; // for SURF_FLOWING
		float time;
		float alpha;
		float overbrightbits;
		float particleFadeFactor;
		float _pad_1; // AMDs legacy windows driver needs this, otherwise uni3D has wrong size
		float _pad_2;
		float _pad_3;
	};
	` + "\x00"

const fragmentCommon3D = `#version 150

	in vec2 passTexCoord;

	out vec4 outColor;

	// for UBO shared between all shaders (incl. 2D)
	layout (std140) uniform uniCommon
	{
		float gamma; // this is 1.0/vid_gamma
		float intensity;
		float intensity2D; // for HUD, menus etc

		vec4 color; // really?

	};
	// for UBO shared between all 3D shaders
	layout (std140) uniform uni3D
	{
		mat4 transProj;
		mat4 transView;
		mat4 transModel;

		float scroll; // for SURF_FLOWING
		float time;
		float alpha;
		float overbrightbits;
		float particleFadeFactor;
		float _pad_1; // AMDs legacy windows driver needs this, otherwise uni3D has wrong size
		float _pad_2;
		float _pad_3;
	};
	` + "\x00"

const vertexSrc3D = `

	// it gets attributes and uniforms from vertexCommon3D

	void main()
	{
		passTexCoord = texCoord;
		gl_Position = transProj * transView * transModel * vec4(position, 1.0);
	}
	` + "\x00"

const vertexSrc3Dflow = `

	// it gets attributes and uniforms from vertexCommon3D

	void main()
	{
		passTexCoord = texCoord + vec2(scroll, 0);
		gl_Position = transProj * transView * transModel * vec4(position, 1.0);
	}
	` + "\x00"

const vertexSrc3Dlm = `

	// it gets attributes and uniforms from vertexCommon3D

	out vec2 passLMcoord;
	out vec3 passWorldCoord;
	out vec3 passNormal;
	flat out uint passLightFlags;

	void main()
	{
		passTexCoord = texCoord;
		passLMcoord = lmTexCoord;
		vec4 worldCoord = transModel * vec4(position, 1.0);
		passWorldCoord = worldCoord.xyz;
		vec4 worldNormal = transModel * vec4(normal, 0.0f);
		passNormal = normalize(worldNormal.xyz);
		passLightFlags = lightFlags;

		gl_Position = transProj * transView * worldCoord;
	}
	` + "\x00"

const vertexSrc3DlmFlow = `

	// it gets attributes and uniforms from vertexCommon3D

	out vec2 passLMcoord;
	out vec3 passWorldCoord;
	out vec3 passNormal;
	flat out uint passLightFlags;

	void main()
	{
		passTexCoord = texCoord + vec2(scroll, 0);
		passLMcoord = lmTexCoord;
		vec4 worldCoord = transModel * vec4(position, 1.0);
		passWorldCoord = worldCoord.xyz;
		vec4 worldNormal = transModel * vec4(normal, 0.0f);
		passNormal = normalize(worldNormal.xyz);
		passLightFlags = lightFlags;

		gl_Position = transProj * transView * worldCoord;
	}
	` + "\x00"

const fragmentSrc3D = `

	// it gets attributes and uniforms from fragmentCommon3D

	uniform sampler2D tex;

	void main()
	{
		vec4 texel = texture(tex, passTexCoord);

		// apply intensity and gamma
		texel.rgb *= intensity;
		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a*alpha; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const fragmentSrc3Dwater = `

	// it gets attributes and uniforms from fragmentCommon3D

	uniform sampler2D tex;

	void main()
	{
		vec2 tc = passTexCoord;
		tc.s += sin( passTexCoord.t*0.125 + time ) * 4;
		tc.s += scroll;
		tc.t += sin( passTexCoord.s*0.125 + time ) * 4;
		tc *= 1.0/64.0; // do this last

		vec4 texel = texture(tex, tc);

		// apply intensity and gamma
		texel.rgb *= intensity*0.5;
		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a*alpha; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const fragmentSrc3Dlm = `

	// it gets attributes and uniforms from fragmentCommon3D

	struct DynLight { // gl3UniDynLight in C
		vec3 lightOrigin;
		float _pad;
		//vec3 lightColor;
		//float lightIntensity;
		vec4 lightColor; // .a is intensity; this way it also works on OSX...
		// (otherwise lightIntensity always contained 1 there)
	};

	layout (std140) uniform uniLights
	{
		DynLight dynLights[32];
		uint numDynLights;
		uint _pad1; uint _pad2; uint _pad3; // FFS, AMD!
	};

	uniform sampler2D tex;

	uniform sampler2D lightmap0;
	uniform sampler2D lightmap1;
	uniform sampler2D lightmap2;
	uniform sampler2D lightmap3;

	uniform vec4 lmScales[4];

	in vec2 passLMcoord;
	in vec3 passWorldCoord;
	in vec3 passNormal;
	flat in uint passLightFlags;

	void main()
	{
		vec4 texel = texture(tex, passTexCoord);

		// apply intensity
		texel.rgb *= intensity;

		// apply lightmap
		vec4 lmTex = texture(lightmap0, passLMcoord) * lmScales[0];
		lmTex     += texture(lightmap1, passLMcoord) * lmScales[1];
		lmTex     += texture(lightmap2, passLMcoord) * lmScales[2];
		lmTex     += texture(lightmap3, passLMcoord) * lmScales[3];

		if(passLightFlags != 0u)
		{
			// TODO: or is hardcoding 32 better?
			for(uint i=0u; i<numDynLights; ++i)
			{
				// I made the following up, it's probably not too cool..
				// it basically checks if the light is on the right side of the surface
				// and, if it is, sets intensity according to distance between light and pixel on surface

				// dyn light number i does not affect this plane, just skip it
				if((passLightFlags & (1u << i)) == 0u)  continue;

				float intens = dynLights[i].lightColor.a;

				vec3 lightToPos = dynLights[i].lightOrigin - passWorldCoord;
				float distLightToPos = length(lightToPos);
				float fact = max(0, intens - distLightToPos - 52);

				// move the light source a bit further above the surface
				// => helps if the lightsource is so close to the surface (e.g. grenades, rockets)
				//    that the dot product below would return 0
				// (light sources that are below the surface are filtered out by lightFlags)
				lightToPos += passNormal*32.0;

				// also factor in angle between light and point on surface
				fact *= max(0, dot(passNormal, normalize(lightToPos)));


				lmTex.rgb += dynLights[i].lightColor.rgb * fact * (1.0/256.0);
			}
		}

		lmTex.rgb *= overbrightbits;
		outColor = lmTex*texel;
		outColor.rgb = pow(outColor.rgb, vec3(gamma)); // apply gamma correction to result

		outColor.a = 1; // lightmaps aren't used with translucent surfaces
	}
	` + "\x00"

const fragmentSrc3Dcolor = `

	// it gets attributes and uniforms from fragmentCommon3D

	void main()
	{
		vec4 texel = color;

		// apply gamma correction and intensity
		// texel.rgb *= intensity; TODO: use intensity here? (this is used for beams)
		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a*alpha; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const fragmentSrc3Dsky = `

	// it gets attributes and uniforms from fragmentCommon3D

	uniform sampler2D tex;

	void main()
	{
		vec4 texel = texture(tex, passTexCoord);

		// TODO: something about GL_BLEND vs GL_ALPHATEST etc

		// apply gamma correction
		// texel.rgb *= intensity; // TODO: really no intensity for sky?
		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a*alpha; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const fragmentSrc3Dsprite = `

	// it gets attributes and uniforms from fragmentCommon3D

	uniform sampler2D tex;

	void main()
	{
		vec4 texel = texture(tex, passTexCoord);

		// apply gamma correction and intensity
		texel.rgb *= intensity;
		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a*alpha; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const fragmentSrc3DspriteAlpha = `

	// it gets attributes and uniforms from fragmentCommon3D

	uniform sampler2D tex;

	void main()
	{
		vec4 texel = texture(tex, passTexCoord);

		if(texel.a <= 0.666)
			discard;

		// apply gamma correction and intensity
		texel.rgb *= intensity;
		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a*alpha; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const vertexSrc3Dwater = `

	// it gets attributes and uniforms from vertexCommon3D
	void main()
	{
		passTexCoord = texCoord;

		gl_Position = transProj * transView * transModel * vec4(position, 1.0);
	}
	` + "\x00"

const vertexSrcAlias = `

	// it gets attributes and uniforms from vertexCommon3D

	out vec4 passColor;

	void main()
	{
		passColor = vertColor*overbrightbits;
		passTexCoord = texCoord;
		gl_Position = transProj * transView * transModel * vec4(position, 1.0);
	}
	` + "\x00"

const fragmentSrcAlias = `

	// it gets attributes and uniforms from fragmentCommon3D

	uniform sampler2D tex;

	in vec4 passColor;

	void main()
	{
		vec4 texel = texture(tex, passTexCoord);

		// apply gamma correction and intensity
		texel.rgb *= intensity;
		texel.a *= alpha; // is alpha even used here?
		texel *= min(vec4(1.5), passColor);

		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const fragmentSrcAliasColor = `

	// it gets attributes and uniforms from fragmentCommon3D

	in vec4 passColor;

	void main()
	{
		vec4 texel = passColor;

		// apply gamma correction and intensity
		// texel.rgb *= intensity; // TODO: color-only rendering probably shouldn't use intensity?
		texel.a *= alpha; // is alpha even used here?
		outColor.rgb = pow(texel.rgb, vec3(gamma));
		outColor.a = texel.a; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const vertexSrcParticles = `

	// it gets attributes and uniforms from vertexCommon3D

	out vec4 passColor;

	void main()
	{
		passColor = vertColor;
		gl_Position = transProj * transView * transModel * vec4(position, 1.0);

		// abusing texCoord for pointSize, pointDist for particles
		float pointDist = texCoord.y*0.1; // with factor 0.1 it looks good.

		gl_PointSize = texCoord.x/pointDist;
	}
	` + "\x00"

const fragmentSrcParticles = `

	// it gets attributes and uniforms from fragmentCommon3D

	in vec4 passColor;

	void main()
	{
		vec2 offsetFromCenter = 2.0*(gl_PointCoord - vec2(0.5, 0.5)); // normalize so offset is between 0 and 1 instead 0 and 0.5
		float distSquared = dot(offsetFromCenter, offsetFromCenter);
		if(distSquared > 1.0) // this makes sure the particle is round
			discard;

		vec4 texel = passColor;

		// apply gamma correction and intensity
		//texel.rgb *= intensity; TODO: intensity? Probably not?
		outColor.rgb = pow(texel.rgb, vec3(gamma));

		// I want the particles to fade out towards the edge, the following seems to look nice
		texel.a *= min(1.0, particleFadeFactor*(1.0 - distSquared));

		outColor.a = texel.a; // I think alpha shouldn't be modified by gamma and intensity
	}
	` + "\x00"

const fragmentSrcParticlesSquare = `

	// it gets attributes and uniforms from fragmentCommon3D

	in vec4 passColor;

	void main()
	{
		// outColor = passColor;
		// so far we didn't use gamma correction for square particles, but this way
		// uniCommon is referenced so hopefully Intels Ivy Bridge HD4000 GPU driver
		// for Windows stops shitting itself (see https://github.com/yquake2/yquake2/issues/391)
		outColor.rgb = pow(passColor.rgb, vec3(gamma));
		outColor.a = passColor.a;
	}
	` + "\x00"
