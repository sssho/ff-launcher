package lib

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const LINKFLAGS_OFFSET = 20
const HEADERSIZE = 0x4c

type LinkFlags struct {
	HasLinkTargetIDList         bool
	HasLinkInfo                 bool
	HasName                     bool
	HasRelativePath             bool
	HasWorkingDir               bool
	HasArguments                bool
	HasIconLocation             bool
	IsUnicode                   bool
	ForceNoLinkInfo             bool
	HasExpString                bool
	RunInSeparateProcess        bool
	Unused1                     bool
	HasDarwinID                 bool
	RunAsUser                   bool
	HasExpIcon                  bool
	NoPidlAlias                 bool
	Unused2                     bool
	RunWithShimLayer            bool
	ForceNoLinkTrack            bool
	EnableTargetMetadata        bool
	DisableLinkPathTracking     bool
	DisableKnownFolderTracking  bool
	DisableKnownFolderAlias     bool
	AllowLinkToLink             bool
	UnaliasOnSave               bool
	PreferEnvironmentPath       bool
	KeepLocalIDListForUNCTarget bool
}

func NewLinkFlags(flag uint32) LinkFlags {
	var l LinkFlags
	l.HasLinkTargetIDList = ((flag >> 0) & 1) == 1
	l.HasLinkInfo = ((flag >> 1) & 1) == 1
	l.HasName = ((flag >> 2) & 1) == 1
	l.HasRelativePath = ((flag >> 3) & 1) == 1
	l.HasWorkingDir = ((flag >> 4) & 1) == 1
	l.HasArguments = ((flag >> 5) & 1) == 1
	l.HasIconLocation = ((flag >> 6) & 1) == 1
	l.IsUnicode = ((flag >> 7) & 1) == 1
	l.ForceNoLinkInfo = ((flag >> 8) & 1) == 1
	l.HasExpString = ((flag >> 9) & 1) == 1
	l.RunInSeparateProcess = ((flag >> 10) & 1) == 1
	l.Unused1 = ((flag >> 11) & 1) == 1
	l.HasDarwinID = ((flag >> 12) & 1) == 1
	l.RunAsUser = ((flag >> 13) & 1) == 1
	l.HasExpIcon = ((flag >> 14) & 1) == 1
	l.NoPidlAlias = ((flag >> 15) & 1) == 1
	l.Unused2 = ((flag >> 16) & 1) == 1
	l.RunWithShimLayer = ((flag >> 17) & 1) == 1
	l.ForceNoLinkTrack = ((flag >> 18) & 1) == 1
	l.EnableTargetMetadata = ((flag >> 19) & 1) == 1
	l.DisableLinkPathTracking = ((flag >> 20) & 1) == 1
	l.DisableKnownFolderTracking = ((flag >> 21) & 1) == 1
	l.DisableKnownFolderAlias = ((flag >> 22) & 1) == 1
	l.AllowLinkToLink = ((flag >> 23) & 1) == 1
	l.UnaliasOnSave = ((flag >> 24) & 1) == 1
	l.PreferEnvironmentPath = ((flag >> 25) & 1) == 1
	l.KeepLocalIDListForUNCTarget = ((flag >> 26) & 1) == 1
	return l
}

type FileAttributesFlags struct {
	FILE_ATTRIBUTE_READONLY            bool
	FILE_ATTRIBUTE_HIDDEN              bool
	FILE_ATTRIBUTE_SYSTEM              bool
	Reserved1                          bool
	FILE_ATTRIBUTE_DIRECTORY           bool
	FILE_ATTRIBUTE_ARCHIVE             bool
	Reserved2                          bool
	FILE_ATTRIBUTE_NORMAL              bool
	FILE_ATTRIBUTE_TEMPORARY           bool
	FILE_ATTRIBUTE_SPARSE_FILE         bool
	FILE_ATTRIBUTE_REPARSE_POINT       bool
	FILE_ATTRIBUTE_COMPRESSED          bool
	FILE_ATTRIBUTE_OFFLINE             bool
	FILE_ATTRIBUTE_NOT_CONTENT_INDEXED bool
	FILE_ATTRIBUTE_ENCRYPTED           bool
}

func NewFileAttributesFlags(flag uint32) FileAttributesFlags {
	var f FileAttributesFlags
	f.FILE_ATTRIBUTE_READONLY = ((flag >> 0) & 1) == 1
	f.FILE_ATTRIBUTE_HIDDEN = ((flag >> 1) & 1) == 1
	f.FILE_ATTRIBUTE_SYSTEM = ((flag >> 2) & 1) == 1
	f.Reserved1 = ((flag >> 3) & 1) == 1
	f.FILE_ATTRIBUTE_DIRECTORY = ((flag >> 4) & 1) == 1
	f.FILE_ATTRIBUTE_ARCHIVE = ((flag >> 5) & 1) == 1
	f.Reserved2 = ((flag >> 6) & 1) == 1
	f.FILE_ATTRIBUTE_NORMAL = ((flag >> 7) & 1) == 1
	f.FILE_ATTRIBUTE_TEMPORARY = ((flag >> 8) & 1) == 1
	f.FILE_ATTRIBUTE_SPARSE_FILE = ((flag >> 9) & 1) == 1
	f.FILE_ATTRIBUTE_REPARSE_POINT = ((flag >> 10) & 1) == 1
	f.FILE_ATTRIBUTE_COMPRESSED = ((flag >> 11) & 1) == 1
	f.FILE_ATTRIBUTE_OFFLINE = ((flag >> 12) & 1) == 1
	f.FILE_ATTRIBUTE_NOT_CONTENT_INDEXED = ((flag >> 13) & 1) == 1
	f.FILE_ATTRIBUTE_ENCRYPTED = ((flag >> 14) & 1) == 1
	return f
}

func ResolveShortcut(file *os.File) (path string, isdir bool, args string, err error) {
	_, err = file.Seek(LINKFLAGS_OFFSET, io.SeekStart)
	if err != nil {
		return "", false, "", err
	}
	var LinkFlags uint32
	err = binary.Read(file, binary.LittleEndian, &(LinkFlags))
	if err != nil {
		return "", false, "", err
	}
	lf := NewLinkFlags(LinkFlags)
	if !lf.HasLinkInfo {
		return "", false, "", fmt.Errorf("linkinfo not found")
	}
	var FileAttributesFlags uint32
	err = binary.Read(file, binary.LittleEndian, &(FileAttributesFlags))
	if err != nil {
		return "", false, "", err
	}
	af := NewFileAttributesFlags(FileAttributesFlags)
	var LinkTargetIDListSize uint16
	if lf.HasLinkTargetIDList {
		_, err = file.Seek(HEADERSIZE, io.SeekStart)
		if err != nil {
			return "", false, "", err
		}
		err = binary.Read(file, binary.LittleEndian, &(LinkTargetIDListSize))
		if err != nil {
			return "", false, "", err
		}
	}
	var LinkInfoStartAddr int64 = int64(HEADERSIZE + 2 + LinkTargetIDListSize)
	_, err = file.Seek(LinkInfoStartAddr+8, io.SeekStart)
	if err != nil {
		return "", false, "", err
	}
	var LinkInfoFlags uint32
	err = binary.Read(file, binary.LittleEndian, &(LinkInfoFlags))
	if err != nil {
		return "", false, "", err
	}
	VolumeIDAndLocalBasePath := (LinkInfoFlags & 1) == 1
	if !VolumeIDAndLocalBasePath {
		return "", false, "", fmt.Errorf("localbasepath not found")
	}
	var LocalBasePathOffset uint32
	_, err = file.Seek(LinkInfoStartAddr+16, io.SeekStart)
	if err != nil {
		return "", false, "", err
	}
	err = binary.Read(file, binary.LittleEndian, &(LocalBasePathOffset))
	if err != nil {
		return "", false, "", err
	}
	var LocalBasePathAddr int64 = int64(LinkInfoStartAddr + int64(LocalBasePathOffset))
	_, err = file.Seek(LocalBasePathAddr, io.SeekStart)
	if err != nil {
		return "", false, "", err
	}
	// https://zenn.dev/mattn/articles/fd545a14b0ffdf
	jr := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	br := bufio.NewReader(jr)
	path, err = br.ReadString(0)
	if err != nil || path == "" {
		return "", false, "", err
	}
	args = "" // TODO
	return path, af.FILE_ATTRIBUTE_DIRECTORY, args, nil
}
