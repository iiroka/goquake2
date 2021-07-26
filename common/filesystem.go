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
 * The Quake II file system, implements generic file system operations
 * as well as the .pak file format and support for .pk3 files.
 *
 * =======================================================================
 */
package common

import (
	"fmt"
	"goquake2/shared"
	"os"
	"strings"
)

/* The .pak files are just a linear collapse of a directory tree */

const IDPAKHEADER = (('K' << 24) + ('C' << 16) + ('A' << 8) + 'P')

type dpackfile_t struct {
	Name    string
	Filepos int32
	Filelen int32
}

func dpackFile(data []byte) dpackfile_t {
	r := dpackfile_t{}
	r.Name = shared.ReadString(data, 56)
	r.Filepos = shared.ReadInt32(data[56:])
	r.Filelen = shared.ReadInt32(data[60:])
	return r
}

const dpackfile_size = 56 + 2*4

type dpackheader_t struct {
	Ident  int32 /* == IDPAKHEADER */
	Dirofs int32
	Dirlen int32
}

func dpackHeader(data []byte) dpackheader_t {
	r := dpackheader_t{}
	r.Ident = shared.ReadInt32(data)
	r.Dirofs = shared.ReadInt32(data[4:])
	r.Dirlen = shared.ReadInt32(data[8:])
	return r
}

const dpackheader_size = 3 * 4

const MAX_FILES_IN_PACK = 4096
const MAX_HANDLES = 512

type fsPackFile_t struct {
	name   string
	size   int
	offset int64 /* Ignored in PK3 files. */
}

type fsPack_t struct {
	name string
	pak  *os.File
	// int numFiles;
	// FILE *pak;
	// unzFile *pk3;
	// qboolean isProtectedPak;
	files []fsPackFile_t
}

type fsSearchPath_t struct {
	path string    /* Only one used. */
	pack *fsPack_t /* (path or pack) */
}

const (
	maxHANDLES = 512
	maxMODS    = 32
	maxPAKS    = 100
)

type qFileHandle struct {
	handle *os.File
	end    uint64
	offset uint64
	owns   bool
}

/*
 * Finds a free fileHandle_t.
 */
func (T *qCommon) fsHandleForFile(path string) (*qFileHandle, error) {

	for i, handle := range T.filehandles {
		if handle.handle == nil {
			return &T.filehandles[i], nil
		}
	}
	/* Failed. */
	return nil, T.Com_Error(shared.ERR_DROP, "FS_HandleForFile: none free")
}

func (f *qFileHandle) Close() {
	if f.owns {
		f.handle.Close()
	}
	f.handle = nil
	f.offset = 0
	f.end = 0
}

func (f *qFileHandle) Read(size int) []byte {
	if f.offset >= f.end {
		return nil
	}
	len := size
	if len > int(f.end-f.offset) {
		len = int(f.end - f.offset)
	}
	data := make([]byte, len)
	sz, _ := f.handle.ReadAt(data, int64(f.offset))
	f.offset += uint64(sz)
	return data[0:sz]
}

func (T *qCommon) FS_FOpenFile(name string, gamedir_only bool) (shared.QFileHandle, error) {
	// file_from_protected_pak = false;
	handle, err := T.fsHandleForFile(name)
	if err != nil {
		return nil, err
	}
	path := strings.ToLower(name)

	/* Search through the path, one element at a time. */
	for _, search := range T.fs_searchPaths {

		// if (gamedir_only) {
		// 	if (strstr(search->path, FS_Gamedir()) == NULL) {
		// 		continue;
		// 	}
		// }

		// Evil hack for maps.lst and players/
		// TODO: A flag to ignore paks would be better
		// if ((strcmp(fs_gamedirvar->string, "") == 0) && search->pack) {
		// 	if ((strcmp(name, "maps.lst") == 0)|| (strncmp(name, "players/", 8) == 0)) {
		// 		continue;
		// 	}
		// }

		/* Search inside a pack file. */
		if search.pack != nil {
			pack := search.pack

			for _, f := range pack.files {
				if f.name == path {
					/* Found it! */
					if T.fs_debug.Bool() {
						T.Com_Printf("FS_FOpenFile: '%s' (found in '%s').\n", path, pack.name)
					}

					handle.handle = pack.pak
					handle.offset = uint64(f.offset)
					handle.end = uint64(f.offset) + uint64(f.size)
					handle.owns = false

					return handle, nil
				}
			}
		} else {
			/* Search in a directory tree. */
			path := fmt.Sprintf("%v/%v", search.path, path)
			fhandle, err := os.Open(path)
			if err == nil {

				if T.fs_debug.Bool() {
					T.Com_Printf("FS_FOpenFile: '%s' (found in '%s').\n", path, search.path)
				}

				handle.handle = fhandle
				handle.offset = 0

				st, _ := fhandle.Stat()
				size := st.Size()
				handle.end = uint64(size)
				handle.owns = true

				return handle, nil
			}

		}
	}
	if T.fs_debug.Bool() {
		T.Com_Printf("FS_FOpenFile: couldn't find '%s'.\n", path)
	}
	return nil, nil
}

/*
 * Filename are reletive to the quake search path. A null buffer will just
 * return the file length without loading.
 */
func (T *qCommon) LoadFile(path string) ([]byte, error) {
	// file_from_protected_pak = false;
	// handle = FS_HandleForFile(name, f);
	// Q_strlcpy(handle->name, name, sizeof(handle->name));
	// handle->mode = FS_READ;
	path = strings.ToLower(path)

	/* Search through the path, one element at a time. */
	for _, search := range T.fs_searchPaths {

		// Evil hack for maps.lst and players/
		// TODO: A flag to ignore paks would be better
		// if ((strcmp(fs_gamedirvar->string, "") == 0) && search->pack) {
		// 	if ((strcmp(name, "maps.lst") == 0)|| (strncmp(name, "players/", 8) == 0)) {
		// 		continue;
		// 	}
		// }

		/* Search inside a pack file. */
		if search.pack != nil {
			pack := search.pack

			for _, f := range pack.files {
				if f.name == path {
					/* Found it! */
					if T.fs_debug.Bool() {
						T.Com_Printf("FS_LoadFile: '%s' (found in '%s').\n", path, pack.name)
					}

					bfr := make([]byte, f.size)
					_, err := pack.pak.ReadAt(bfr, f.offset)
					if err != nil {
						return nil, err
					}

					return bfr, nil
				}
			}
		} else {
			/* Search in a directory tree. */
			path := fmt.Sprintf("%v/%v", search.path, path)
			handle, err := os.Open(path)
			if err == nil {

				if T.fs_debug.Bool() {
					T.Com_Printf("FS_LoadFile: '%s' (found in '%s').\n", path, search.path)
				}

				st, _ := handle.Stat()
				size := st.Size()

				bfr := make([]byte, int(size))
				_, err := handle.Read(bfr)
				if err != nil {
					return nil, err
				}

				handle.Close()
				return bfr, nil
			}

		}
	}
	if T.fs_debug.Bool() {
		T.Com_Printf("FS_LoadFile: couldn't find '%s'.\n", path)
	}
	return nil, nil
}

/*
 * Takes an explicit (not game tree related) path to a pak file.
 *
 * Loads the header and directory, adding the files at the beginning of the
 * list so they override previous pack files.
 */
func (T *qCommon) loadPAK(packPath string) (*fsPack_t, error) {
	//  int i; /* Loop counter. */
	//  int numFiles; /* Number of files in PAK. */
	//  FILE *handle; /* File handle. */
	//  fsPackFile_t *files; /* List of files in PAK. */
	//  fsPack_t *pack; /* PAK file. */
	//  dpackheader_t header; /* PAK file header. */
	//  dpackfile_t *info = NULL; /* PAK info. */

	handle, err := os.Open(packPath)
	if err != nil {
		return nil, nil
	}

	bfr := make([]byte, dpackheader_size)
	handle.Read(bfr)

	header := dpackHeader(bfr)
	if header.Ident != IDPAKHEADER {
		handle.Close()
		return nil, T.Com_Error(shared.ERR_FATAL, "loadPAK: '%v' is not a pack file\n", packPath)
	}

	numFiles := header.Dirlen / dpackfile_size

	if (numFiles == 0) || (header.Dirlen < 0) || (header.Dirofs < 0) {
		handle.Close()
		return nil, T.Com_Error(shared.ERR_FATAL, "loadPAK: '%v' is too short.", packPath)
	}

	if numFiles > MAX_FILES_IN_PACK {
		T.Com_Printf("loadPAK: '%s' has %v > %v files\n",
			packPath, numFiles, MAX_FILES_IN_PACK)
	}

	bfr = make([]byte, header.Dirlen)

	files := make([]fsPackFile_t, numFiles)

	handle.ReadAt(bfr, int64(header.Dirofs))

	/* Parse the directory. */
	for i := 0; i < int(numFiles); i++ {
		info := dpackFile(bfr[i*dpackfile_size:])
		files[i].name = info.Name
		files[i].size = int(info.Filelen)
		files[i].offset = int64(info.Filepos)
	}

	pack := fsPack_t{}
	pack.name = packPath
	pack.pak = handle
	pack.files = files

	T.Com_Printf("Added packfile '%v' (%v files).\n", packPath, numFiles)

	return &pack, nil
}

func (T *qCommon) addDirToSearchPath(dir string, create bool) error {

	// Set the current directory as game directory. This
	// is somewhat fragile since the game directory MUST
	// be the last directory added to the search path.
	T.fs_gamedir = dir

	// 	if (create) {
	// 		FS_CreatePath(fs_gamedir);
	// 	}

	// Add the directory itself.
	search := fsSearchPath_t{}
	search.path = dir
	T.fs_searchPaths = append(T.fs_searchPaths, search)

	// We need to add numbered paks in the directory in
	// sequence and all other paks after them. Otherwise
	// the gamedata may break.
	// 	for (i = 0; i < sizeof(fs_packtypes) / sizeof(fs_packtypes[0]); i++) {
	for j := 0; j < maxPAKS; j++ {
		path := fmt.Sprintf("%v/pak%v.pak", dir, j)

		// 			switch (fs_packtypes[i].format)
		// 			{
		// 				case PAK:
		pack, err := T.loadPAK(path)
		if err != nil {
			return nil
		}

		// 					if (pack)
		// 					{
		// 						pack->isProtectedPak = true;
		// 					}

		// 					break;
		// 				case PK3:
		// 					pack = FS_LoadPK3(path);

		// 					if (pack)
		// 					{
		// 						pack->isProtectedPak = false;
		// 					}

		// 					break;
		// 			}

		// 			if (pack == NULL)
		// 			{
		// 				continue;
		// 			}

		if pack != nil {
			search = fsSearchPath_t{}
			search.pack = pack
			T.fs_searchPaths = append(T.fs_searchPaths, search)
		}
	}
	// 	}

	// 	// And as said above all other pak files.
	// 	for (i = 0; i < sizeof(fs_packtypes) / sizeof(fs_packtypes[0]); i++) {
	// 		Com_sprintf(path, sizeof(path), "%s/*.%s", dir, fs_packtypes[i].suffix);

	// 		// Nothing here, next pak type please.
	// 		if ((list = FS_ListFiles(path, &nfiles, 0, 0)) == NULL)
	// 		{
	// 			continue;
	// 		}

	// 		Com_sprintf(path, sizeof(path), "%s/pak*.%s", dir, fs_packtypes[i].suffix);

	// 		for (j = 0; j < nfiles - 1; j++)
	// 		{
	// 			// If the pak starts with the string 'pak' it's ignored.
	// 			// This is somewhat stupid, it would be better to ignore
	// 			// just pak%d...
	// 			if (glob_match(path, list[j]))
	// 			{
	// 				continue;
	// 			}

	// 			switch (fs_packtypes[i].format)
	// 			{
	// 				case PAK:
	// 					pack = FS_LoadPAK(list[j]);
	// 					break;
	// 				case PK3:
	// 					pack = FS_LoadPK3(list[j]);
	// 					break;
	// 			}

	// 			if (pack == NULL)
	// 			{
	// 				continue;
	// 			}

	// 			pack->isProtectedPak = false;

	// 			search = Z_Malloc(sizeof(fsSearchPath_t));
	// 			search->pack = pack;
	// 			search->next = fs_searchPaths;
	// 			fs_searchPaths = search;
	// 		}

	// 		FS_FreeList(list, nfiles);
	// 	}
	return nil
}

func (T *qCommon) buildGenericSearchPath(paths map[string]bool) error {
	// We may not use the va() function from shared.c
	// since it's buffersize is 1024 while most OS have
	// a maximum path size of 4096...
	// char path[MAX_OSPATH];

	// fsRawPath_t *search = fs_rawPath;

	for search, create := range paths {
		path := search + "/" + shared.BASEDIRNAME
		err := T.addDirToSearchPath(path, create)
		if err != nil {
			return err
		}
	}

	// // Until here we've added the generic directories to the
	// // search path. Save the current head node so we can
	// // distinguish generic and specialized directories.
	// fs_baseSearchPaths = fs_searchPaths;

	// // We need to create the game directory.
	// Sys_Mkdir(fs_gamedir);

	// // We need to create the screenshot directory since the
	// // render dll doesn't link the filesystem stuff.
	// Com_sprintf(path, sizeof(path), "%s/scrnshot", fs_gamedir);
	// Sys_Mkdir(path);
	return nil
}

func (T *qCommon) buildRawPath() map[string]bool {
	set := make(map[string]bool)

	// Add $HOME/.yq2 (MUST be the last dir!)
	// 	if (!is_portable) {
	homedir, err := os.UserHomeDir()
	if err == nil {
		set[homedir] = true
	}

	// 	// Add $binarydir
	// 	const char *binarydir = Sys_GetBinaryDir();

	// if(binarydir[0] != '\0')
	// 	{
	// 		FS_AddDirToRawPath(binarydir, false);
	// 	}

	// Add $basedir/
	set[T.datadir] = false

	// 	// Add SYSTEMDIR
	// #ifdef SYSTEMWIDE
	// 	FS_AddDirToRawPath(SYSTEMDIR, false);
	// #endif

	// The CD must be the last directory of the path,
	// otherwise we cannot be sure that the game won't
	// stream the videos from the CD.
	if len(T.fs_cddir.String) > 0 {
		set[T.fs_cddir.String] = false
	}
	return set
}

// --------

func (T *qCommon) initFilesystem() error {
	// Register FS commands.
	// 	Cmd_AddCommand("path", FS_Path_f);
	// 	Cmd_AddCommand("link", FS_Link_f);
	// 	Cmd_AddCommand("dir", FS_Dir_f);

	// Register cvars
	T.fs_basedir = T.Cvar_Get("basedir", ".", shared.CVAR_NOSET)
	T.fs_cddir = T.Cvar_Get("cddir", "", shared.CVAR_NOSET)
	T.fs_gamedirvar = T.Cvar_Get("game", "", shared.CVAR_LATCH|shared.CVAR_SERVERINFO)
	T.fs_debug = T.Cvar_Get("fs_debug", "1", 0)

	// Deprecation warning, can be removed at a later time.
	if T.fs_basedir.String != "." {
		T.Com_Printf("+set basedir is deprecated, use -datadir instead\n")
		T.datadir = T.fs_basedir.String
	} else if len(T.datadir) == 0 {
		T.datadir = "."
	}

	// #ifdef _WIN32
	// 	// setup minizip for Unicode compatibility
	// 	fill_fopen_filefunc(&zlib_file_api);
	// 	zlib_file_api.zopen_file = fopen_file_func_utf;
	// #endif

	// Build search path
	paths := T.buildRawPath()
	err := T.buildGenericSearchPath(paths)
	if err != nil {
		return err
	}

	// 	if (fs_gamedirvar->string[0] != '\0')
	// 	{
	// 		FS_BuildGameSpecificSearchPath(fs_gamedirvar->string);
	// 	}
	// #ifndef DEDICATED_ONLY
	// 	else
	// 	{
	// 		// no mod, but we still need to get the list of OGG tracks for background music
	// 		OGG_InitTrackList();
	// 	}
	// #endif

	// Debug output
	T.Com_Printf("Using '%v' for writing.\n", T.fs_gamedir)
	return nil
}
