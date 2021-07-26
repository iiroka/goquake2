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
	// shaderInfo.uniLmScales = -1

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
	// gl3state.uni3DData.transProjMat4 = HMM_Mat4();
	// gl3state.uni3DData.transViewMat4 = HMM_Mat4();
	// gl3state.uni3DData.transModelMat4 = gl3_identityMat4;
	// gl3state.uni3DData.scroll = 0.0f;
	// gl3state.uni3DData.time = 0.0f;
	// gl3state.uni3DData.alpha = 1.0f;
	// // gl3_overbrightbits 0 means "no scaling" which is equivalent to multiplying with 1
	// gl3state.uni3DData.overbrightbits = (gl3_overbrightbits->value <= 0.0f) ? 1.0f : gl3_overbrightbits->value;
	// gl3state.uni3DData.particleFadeFactor = gl3_particle_fade_factor->value;

	gl.GenBuffers(1, &T.gl3state.uni3DUBO)
	gl.BindBuffer(gl.UNIFORM_BUFFER, T.gl3state.uni3DUBO)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, GL3_BINDINGPOINT_UNI3D, T.gl3state.uni3DUBO)
	// glBufferData(GL_UNIFORM_BUFFER, sizeof(gl3state.uni3DData), &gl3state.uni3DData, GL_DYNAMIC_DRAW);

	gl.GenBuffers(1, &T.gl3state.uniLightsUBO)
	gl.BindBuffer(gl.UNIFORM_BUFFER, T.gl3state.uniLightsUBO)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, GL3_BINDINGPOINT_UNILIGHTS, T.gl3state.uniLightsUBO)
	// glBufferData(GL_UNIFORM_BUFFER, sizeof(gl3state.uniLightsData), &gl3state.uniLightsData, GL_DYNAMIC_DRAW);

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
	// if(!initShader3D(&gl3state.si3Dlm, vertexSrc3Dlm, fragmentSrc3Dlm))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for textured 3D rendering with lightmap!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3Dtrans, vertexSrc3D, fragmentSrc3D))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for rendering translucent 3D things!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3DcolorOnly, vertexSrc3D, fragmentSrc3Dcolor))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for flat-colored 3D rendering!\n");
	// 	return false;
	// }
	// /*
	// if(!initShader3D(&gl3state.si3Dlm, vertexSrc3Dlm, fragmentSrc3D))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for blending 3D lightmaps rendering!\n");
	// 	return false;
	// }
	// */
	// if(!initShader3D(&gl3state.si3Dturb, vertexSrc3Dwater, fragmentSrc3Dwater))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for water rendering!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3DlmFlow, vertexSrc3DlmFlow, fragmentSrc3Dlm))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for scrolling textured 3D rendering with lightmap!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3DtransFlow, vertexSrc3Dflow, fragmentSrc3D))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for scrolling textured translucent 3D rendering!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3Dsky, vertexSrc3D, fragmentSrc3Dsky))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for sky rendering!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3Dsprite, vertexSrc3D, fragmentSrc3Dsprite))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for sprite rendering!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3DspriteAlpha, vertexSrc3D, fragmentSrc3DspriteAlpha))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for alpha-tested sprite rendering!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3Dalias, vertexSrcAlias, fragmentSrcAlias))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for rendering textured models!\n");
	// 	return false;
	// }
	// if(!initShader3D(&gl3state.si3DaliasColor, vertexSrcAlias, fragmentSrcAliasColor))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for rendering flat-colored models!\n");
	// 	return false;
	// }

	// const char* particleFrag = fragmentSrcParticles;
	// if(gl3_particle_square->value != 0.0f)
	// {
	// 	particleFrag = fragmentSrcParticlesSquare;
	// }

	// if(!initShader3D(&gl3state.siParticle, vertexSrcParticles, particleFrag))
	// {
	// 	R_Printf(PRINT_ALL, "WARNING: Failed to create shader program for rendering particles!\n");
	// 	return false;
	// }

	// gl3state.currentShaderProgram = 0;

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
