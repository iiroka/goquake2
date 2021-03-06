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
 * Local header for the OpenGL3 refresher.
 *
 * =======================================================================
 */
package gl3

import (
	"goquake2/shared"
	"math"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

const MAX_LBM_HEIGHT = 480

// Hold the video state.
type viddef_t struct {
	height int
	width  int
}

type gl3config_t struct {
	renderer_string     string
	vendor_string       string
	version_string      string
	glsl_version_string string

	major_version int
	minor_version int

	// ----

	anisotropic  bool // is GL_EXT_texture_filter_anisotropic supported?
	debug_output bool // is GL_ARB_debug_output supported?
	stencil      bool // Do we have a stencil buffer?

	useBigVBO bool // workaround for AMDs windows driver for fewer calls to glBufferData()

	// ----

	max_anisotropy float32
}

type gl3ShaderInfo_t struct {
	shaderProgram uint32
	uniLmScales   int32
	lmScales      [16]float32
}

type gl3UniCommon_t struct {
	// gamma       float32
	// intensity   float32
	// intensity2D float32 // for HUD, menus etc

	// // entries of std140 UBOs are aligned to multiples of their own size
	// // so we'll need to pad accordingly for following vec4
	// _padding float32

	// color [4]float32
	data []float32
}

func (T *gl3UniCommon_t) setGamma(v float32) {
	T.data[0] = v
}

func (T *gl3UniCommon_t) setIntensity(v float32) {
	T.data[1] = v
}

func (T *gl3UniCommon_t) setIntensity2D(v float32) {
	T.data[2] = v
}

func (T *gl3UniCommon_t) setColor(a float32, r float32, g float32, b float32) {
	T.data[4] = r
	T.data[5] = g
	T.data[6] = b
	T.data[7] = a
}

type gl3Uni2D_t struct {
	// hmm_mat4 transMat4;
	data []float32
}

type gl3Uni3D_t struct {
	// hmm_mat4 transProjMat4;
	// hmm_mat4 transViewMat4;
	// hmm_mat4 transModelMat4;

	// GLfloat scroll; // for SURF_FLOWING
	// GLfloat time; // for warping surfaces like water & possibly other things
	// GLfloat alpha; // for translucent surfaces (water, glass, ..)
	// GLfloat overbrightbits; // gl3_overbrightbits, applied to lightmaps (and elsewhere to models)
	// GLfloat particleFadeFactor; // gl3_particle_fade_factor, higher => less fading out towards edges

	// 	GLfloat _padding[3]; // again, some padding to ensure this has right size
	data []float32
}

const gl3Uni3D_Size = 3*16 + 5 + 3

func (T *gl3Uni3D_t) setTransProjMat4(v []float32) {
	copy(T.data[:16], v)
}

func (T *gl3Uni3D_t) getTransProjMat4() []float32 {
	r := make([]float32, 16)
	copy(r, T.data[:])
	return r
}

func (T *gl3Uni3D_t) setTransViewMat4(v []float32) {
	copy(T.data[16:2*16], v)
}

func (T *gl3Uni3D_t) setTransModelMat4(v []float32) {
	copy(T.data[2*16:3*16], v)
}

func (T *gl3Uni3D_t) getTransModelMat4() []float32 {
	r := make([]float32, 16)
	copy(r, T.data[2*16:])
	return r
}

func (T *gl3Uni3D_t) getScroll() float32 {
	return T.data[3*16]
}

func (T *gl3Uni3D_t) setScroll(v float32) {
	T.data[3*16] = v
}

func (T *gl3Uni3D_t) setTime(v float32) {
	T.data[3*16+1] = v
}

func (T *gl3Uni3D_t) getAlpha() float32 {
	return T.data[3*16+2]
}

func (T *gl3Uni3D_t) setAlpha(v float32) {
	T.data[3*16+2] = v
}

func (T *gl3Uni3D_t) setOverbrightbits(v float32) {
	T.data[3*16+3] = v
}

func (T *gl3Uni3D_t) setParticleFadeFactor(v float32) {
	T.data[3*16+4] = v
}

// extern const hmm_mat4 gl3_identityMat4;

type gl3UniDynLight_t struct {
	// vec3_t origin;
	// GLfloat _padding;
	// vec3_t color;
	// GLfloat intensity;
	data []uint32
}

const gl3UniDynLight_Size = 8

func (T *gl3UniDynLight_t) setOrigin(v []float32) {
	T.data[0] = math.Float32bits(v[0])
	T.data[1] = math.Float32bits(v[1])
	T.data[2] = math.Float32bits(v[2])
}

func (T *gl3UniDynLight_t) setColor(v []float32) {
	T.data[4] = math.Float32bits(v[0])
	T.data[5] = math.Float32bits(v[1])
	T.data[6] = math.Float32bits(v[2])
}

func (T *gl3UniDynLight_t) setIntensity(v float32) {
	T.data[7] = math.Float32bits(v)
}

type gl3UniLights_t struct {
	dynLights []gl3UniDynLight_t
	// GLuint numDynLights;
	// GLfloat _padding[3];
	data []uint32
}

const gl3UniLights_Size = 4 + shared.MAX_DLIGHTS*gl3UniDynLight_Size

func (T *gl3UniLights_t) initialize() {
	T.dynLights = make([]gl3UniDynLight_t, shared.MAX_DLIGHTS)
	for i := 0; i < shared.MAX_DLIGHTS; i++ {
		T.dynLights[i].data = T.data[4+i*gl3UniDynLight_Size:]
	}
}

func (T *gl3UniLights_t) setNumDynLights(v int) {
	T.data[0] = uint32(v)
}

const (
	// width and height used to be 128, so now we should be able to get the same lightmap data
	// that used 32 lightmaps before into one, so 4 lightmaps should be enough
	BLOCK_WIDTH               = 1024
	BLOCK_HEIGHT              = 512
	LIGHTMAP_BYTES            = 4
	MAX_LIGHTMAPS             = 4
	MAX_LIGHTMAPS_PER_SURFACE = shared.MAXLIGHTMAPS // 4
)

type gl3state_t struct {
	// TODO: what of this do we need?
	fullscreen bool

	prev_mode int

	// each lightmap consists of 4 sub-lightmaps allowing changing shadows on the same surface
	// used for switching on/off light and stuff like that.
	// most surfaces only have one really and the remaining for are filled with dummy data
	lightmap_textureIDs []uint32

	currenttexture  uint32 // bound to GL_TEXTURE0
	currentlightmap int    // lightmap_textureIDs[currentlightmap] bound to GL_TEXTURE1 // [MAX_LIGHTMAPS][MAX_LIGHTMAPS_PER_SURFACE];
	currenttmu      uint32 // GL_TEXTURE0 or GL_TEXTURE1

	//float camera_separation;
	//enum stereo_modes stereo_mode;

	currentVAO           uint32
	currentVBO           uint32
	currentEBO           uint32
	currentShaderProgram uint32
	currentUBO           uint32

	// NOTE: make sure si2D is always the first shaderInfo (or adapt GL3_ShutdownShaders())
	si2D      gl3ShaderInfo_t // shader for rendering 2D with textures
	si2Dcolor gl3ShaderInfo_t // shader for rendering 2D with flat colors
	si3Dlm    gl3ShaderInfo_t // a regular opaque face (e.g. from brush) with lightmap
	// TODO: lm-only variants for gl_lightmap 1
	si3Dtrans       gl3ShaderInfo_t // transparent is always w/o lightmap
	si3DcolorOnly   gl3ShaderInfo_t // used for beams - no lightmaps
	si3Dturb        gl3ShaderInfo_t // for water etc - always without lightmap
	si3DlmFlow      gl3ShaderInfo_t // for flowing/scrolling things with lightmap (conveyor, ..?)
	si3DtransFlow   gl3ShaderInfo_t // for transparent flowing/scrolling things (=> no lightmap)
	si3Dsky         gl3ShaderInfo_t // guess what..
	si3Dsprite      gl3ShaderInfo_t // for sprites
	si3DspriteAlpha gl3ShaderInfo_t // for sprites with alpha-testing

	si3Dalias      gl3ShaderInfo_t // for models
	si3DaliasColor gl3ShaderInfo_t // for models w/ flat colors

	// NOTE: make sure siParticle is always the last shaderInfo (or adapt GL3_ShutdownShaders())
	siParticle gl3ShaderInfo_t // for particles. surprising, right?

	vao3D, vbo3D uint32 // for brushes etc, using 10 floats and one uint as vertex input (x,y,z, s,t, lms,lmt, normX,normY,normZ ; lightFlags)

	// the next two are for gl3config.useBigVBO == true
	// int vbo3Dsize;
	// int vbo3DcurOffset;

	vaoAlias, vboAlias, eboAlias uint32 // for models, using 9 floats as (x,y,z, s,t, r,g,b,a)
	vaoParticle, vboParticle     uint32 // for particles, using 9 floats (x,y,z, size,distance, r,g,b,a)

	// UBOs and their data
	uniCommonData gl3UniCommon_t
	uni2DData     gl3Uni2D_t
	uni3DData     gl3Uni3D_t
	uniLightsData gl3UniLights_t
	uniCommonUBO  uint32
	uni2DUBO      uint32
	uni3DUBO      uint32
	uniLightsUBO  uint32
}

// attribute locations for vertex shaders
const (
	GL3_ATTRIB_POSITION   = 0
	GL3_ATTRIB_TEXCOORD   = 1 // for normal texture
	GL3_ATTRIB_LMTEXCOORD = 2 // for lightmap
	GL3_ATTRIB_COLOR      = 3 // per-vertex color
	GL3_ATTRIB_NORMAL     = 4 // vertex normal
	GL3_ATTRIB_LIGHTFLAGS = 5 // uint, each set bit means "dyn light i affects this surface"
)

/*
 * skins will be outline flood filled and mip mapped
 * pics and sprites with alpha will be outline flood filled
 * pic won't be mip mapped
 *
 * model skin
 * sprite frame
 * wall texture
 * pic
 */
type imagetype_t int

const (
	it_skin   imagetype_t = 0
	it_sprite imagetype_t = 1
	it_wall   imagetype_t = 2
	it_pic    imagetype_t = 3
	it_sky    imagetype_t = 4
)

type modtype_t int

const (
	mod_bad    modtype_t = 0
	mod_brush  modtype_t = 1
	mod_sprite modtype_t = 2
	mod_alias  modtype_t = 3
)

/* NOTE: struct image_s* is what re.RegisterSkin() etc return so no gl3image_s!
 *       (I think the client only passes the pointer around and doesn't know the
 *        definition of this struct, so this being different from struct image_s
 *        in ref_gl should be ok)
 */
type gl3image_t struct {
	name          string /* game path, including extension */
	itype         imagetype_t
	width, height int /* source image */
	//int upload_width, upload_height;    /* after power of two and picmip */
	registration_sequence int         /* 0 = free */
	texturechain          *msurface_t /* for sort-by-texture world drawing */
	texnum                uint32      /* gl texture binding */
	sl, tl, sh, th        float32     /* 0,0 - 1,1 unless part of the scrap */
	// qboolean scrap; // currently unused
	has_alpha bool
}

const MAX_GL3TEXTURES = 1024

// TODO: do we need the following configurable?
const gl3_solid_format = gl.RGB
const gl3_alpha_format = gl.RGBA
const gl3_tex_solid_format = gl.RGB
const gl3_tex_alpha_format = gl.RGBA

type gl3lightmapstate_t struct {
	internal_format          int
	current_lightmap_texture int // index into gl3state.lightmap_textureIDs[]

	//msurface_t *lightmap_surfaces[MAX_LIGHTMAPS]; - no more lightmap chains, lightmaps are rendered multitextured

	allocated [BLOCK_WIDTH]int

	/* the lightmap texture data needs to be kept in
	   main memory so texsubimage can update properly */
	lightmap_buffers [MAX_LIGHTMAPS_PER_SURFACE][]byte
}

type qGl3 struct {
	ri                       shared.Refimport_t
	gl3config                gl3config_t
	gl3state                 gl3state_t
	vid                      viddef_t
	window                   *sdl.Window
	context                  sdl.GLContext
	gl_msaa_samples          *shared.CvarT
	r_vsync                  *shared.CvarT
	gl_retexturing           *shared.CvarT
	vid_fullscreen           *shared.CvarT
	r_mode                   *shared.CvarT
	r_customwidth            *shared.CvarT
	r_customheight           *shared.CvarT
	vid_gamma                *shared.CvarT
	gl_anisotropic           *shared.CvarT
	gl_texturemode           *shared.CvarT
	gl_drawbuffer            *shared.CvarT
	r_clear                  *shared.CvarT
	gl3_particle_size        *shared.CvarT
	gl3_particle_fade_factor *shared.CvarT
	gl3_particle_square      *shared.CvarT

	gl_lefthand *shared.CvarT
	r_gunfov    *shared.CvarT
	r_farsee    *shared.CvarT

	gl3_intensity      *shared.CvarT
	gl3_intensity_2D   *shared.CvarT
	r_lightlevel       *shared.CvarT
	gl3_overbrightbits *shared.CvarT

	r_norefresh    *shared.CvarT
	r_drawentities *shared.CvarT
	r_drawworld    *shared.CvarT
	gl_nolerp_list *shared.CvarT
	gl_nobind      *shared.CvarT
	r_lockpvs      *shared.CvarT
	r_novis        *shared.CvarT
	r_speeds       *shared.CvarT
	gl_finish      *shared.CvarT

	gl_cull          *shared.CvarT
	gl_zfix          *shared.CvarT
	r_fullbright     *shared.CvarT
	r_modulate       *shared.CvarT
	gl_lightmap      *shared.CvarT
	gl_shadows       *shared.CvarT
	gl3_debugcontext *shared.CvarT
	gl3_usebigvbo    *shared.CvarT
	r_fixsurfsky     *shared.CvarT

	gl3_worldmodel *gl3model_t

	gl3depthmin, gl3depthmax float64

	gl3_newrefdef shared.Refdef_t

	gl_filter_min int32
	gl_filter_max int32

	gl3textures    []gl3image_t
	numgl3textures int

	draw_chars               *gl3image_t
	vbo2D, vao2D, vao2Dcolor uint32

	c_brush_polys, c_alias_polys            int
	c_visible_lightmaps, c_visible_textures int

	/* view origin */
	vup        [3]float32
	vpn        [3]float32
	vright     [3]float32
	gl3_origin [3]float32

	gl3_viewcluster, gl3_viewcluster2, gl3_oldviewcluster, gl3_oldviewcluster2 int

	gl3_visframecount int /* bumped when going to a new PVS */
	gl3_framecount    int /* used for dlight push checking */

	d_8to24table []uint32

	vsyncActive bool

	// static gl3model_t *loadmodel;
	mod_novis [shared.MAX_MAP_LEAFS / 8]byte
	// gl3model_t mod_known[MAX_MOD_KNOWN];
	mod_known  []gl3model_t
	mod_inline []gl3model_t
	// static int mod_numknown;
	registration_sequence int
	// static byte *mod_base;

	frustum [4]shared.Cplane_t

	gl3_particletexture *gl3image_t
	gl3_notexture       *gl3image_t

	_surf_modelorg [3]float32

	currententity *shared.Entity_t
	currentmodel  *gl3model_t

	gl3_lms gl3lightmapstate_t

	gl3_alpha_surfaces *msurface_t

	// sky
	skymins          [2][6]float32
	skymaxs          [2][6]float32
	sky_min, sky_max float32
	skyrotate        float32
	skyaxis          [3]float32
	sky_images       [6]*gl3image_t
	c_sky            int
	vtx_sky          [12]float32
	tex_sky          [8]float32
	index_vtx        int
	index_tex        int

	r_dlightframecount int
	pointcolor         [3]float32
	lightplane         *shared.Cplane_t /* used as shadow plane */
	lightspot          [3]float32
}

// gl3_lightmap.c
const GL_LIGHTMAP_FORMAT = gl.RGBA

func (T *qGl3) useProgram(shaderProgram uint32) {
	if shaderProgram != T.gl3state.currentShaderProgram {
		T.gl3state.currentShaderProgram = shaderProgram
		gl.UseProgram(shaderProgram)
	}
}

func (T *qGl3) bindVAO(vao uint32) {
	if vao != T.gl3state.currentVAO {
		T.gl3state.currentVAO = vao
		gl.BindVertexArray(vao)
	}
}

func (T *qGl3) bindVBO(vbo uint32) {
	if vbo != T.gl3state.currentVBO {
		T.gl3state.currentVBO = vbo
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	}
}

func (T *qGl3) bindEBO(ebo uint32) {
	if ebo != T.gl3state.currentEBO {
		T.gl3state.currentEBO = ebo
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	}
}

func (T *qGl3) selectTMU(tmu uint32) {
	if T.gl3state.currenttmu != tmu {
		gl.ActiveTexture(tmu)
		T.gl3state.currenttmu = tmu
	}
}
