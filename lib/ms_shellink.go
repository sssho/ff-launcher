package lib

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const LINKFLAGS_OFFSET = 20
const HEADERSIZE = 0x4c

type Header struct {
	LinkFlags           uint32
	FileAttributesFlags uint32
}

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

type LinkTargetIDList struct {
	IDListSize uint16
}

type LinkInfo struct {
	LinkInfoSize                    uint32
	LinkInfoFlags                   uint32
	LocalBasePathOffset             uint32
	LocalBasePath                   string
	CommonNetworkRelativeLinkOffset uint32
	CommonPathSuffixOffset          uint32
	CommonPathSuffix                string
}

type CommonNetworkRelativeLink struct {
	CommonNetworkRelativeLinkSize  uint32
	CommonNetworkRelativeLinkFlags uint32
	NetNameOffset                  uint32
	DeviceNameOffset               uint32
	NetName                        string
	DeviceName                     string
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

type StringData struct {
	NAME_STRING            string
	RELATIVE_PATH          string
	WORKING_DIR            string
	COMMAND_LINE_ARGUMENTS string
	ICON_LOCATION          string
}

func ParseHeader(file *os.File) (h Header, err error) {
	_, err = file.Seek(LINKFLAGS_OFFSET, io.SeekStart)
	if err != nil {
		return h, err
	}
	err = binary.Read(file, binary.LittleEndian, &(h.LinkFlags))
	if err != nil {
		return h, err
	}
	err = binary.Read(file, binary.LittleEndian, &(h.FileAttributesFlags))
	if err != nil {
		return h, err
	}
	return h, nil
}

func ParseLinkTargetIDList(file *os.File, offset int32) (l LinkTargetIDList, err error) {
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return l, err
	}
	err = binary.Read(file, binary.LittleEndian, &(l.IDListSize))
	if err != nil {
		return l, err
	}
	return l, nil
}

func ParseLinkInfo(file *os.File, offset int32) (l LinkInfo, err error) {
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return l, err
	}
	err = binary.Read(file, binary.LittleEndian, &(l.LinkInfoSize))
	if err != nil {
		return l, err
	}
	_, err = file.Seek(int64(offset+8), io.SeekStart)
	if err != nil {
		return l, err
	}
	err = binary.Read(file, binary.LittleEndian, &(l.LinkInfoFlags))
	if err != nil {
		return l, err
	}
	VolumeIDAndLocalBasePath := (l.LinkInfoFlags & 1) == 1
	if VolumeIDAndLocalBasePath {
		_, err = file.Seek(int64(offset+16), io.SeekStart)
		if err != nil {
			return l, err
		}
		err = binary.Read(file, binary.LittleEndian, &(l.LocalBasePathOffset))
		if err != nil {
			return l, err
		}
		var localBasePathAddr int64 = int64(offset) + int64(l.LocalBasePathOffset)
		_, err = file.Seek(localBasePathAddr, io.SeekStart)
		if err != nil {
			return l, err
		}
		l.LocalBasePath, err = ReadShiftJISByte(file, localBasePathAddr)
		if err != nil {
			return l, err
		}
	}
	CommonNetworkRelativeLinkAndPathSuffix := ((l.LinkInfoFlags >> 1) & 1) == 1
	if CommonNetworkRelativeLinkAndPathSuffix {
		_, err = file.Seek(int64(offset+20), io.SeekStart)
		if err != nil {
			return l, err
		}
		err = binary.Read(file, binary.LittleEndian, &(l.CommonNetworkRelativeLinkOffset))
		if err != nil {
			return l, err
		}
	}
	_, err = file.Seek(int64(offset+24), io.SeekStart)
	if err != nil {
		return l, err
	}
	err = binary.Read(file, binary.LittleEndian, &(l.CommonPathSuffixOffset))
	if err != nil {
		return l, err
	}
	if l.CommonPathSuffixOffset != 0 {
		var commonPathSuffixAddr int64 = int64(offset) + int64(l.CommonPathSuffixOffset)
		_, err = file.Seek(commonPathSuffixAddr, io.SeekStart)
		if err != nil {
			return l, err
		}
		l.CommonPathSuffix, err = ReadShiftJISByte(file, commonPathSuffixAddr)
		if err != nil {
			return l, err
		}
	}
	return l, nil
}

func ReadShiftJISByte(r io.Reader, addr int64) (s string, err error) {
	// https://zenn.dev/mattn/articles/fd545a14b0ffdf
	jr := transform.NewReader(r, japanese.ShiftJIS.NewDecoder())
	br := bufio.NewReader(jr)
	s, err = br.ReadString(0)
	if err != nil {
		return s, err
	}
	// remove null char
	s = s[:len(s)-1]
	return s, nil
}

func ParseCommonNetworkRelativeLink(file *os.File, offset int64) (c CommonNetworkRelativeLink, err error) {
	_, err = file.Seek(int64(offset+4), io.SeekStart)
	if err != nil {
		return c, err
	}
	err = binary.Read(file, binary.LittleEndian, &(c.CommonNetworkRelativeLinkFlags))
	if err != nil {
		return c, err
	}
	_, err = file.Seek(int64(offset+8), io.SeekStart)
	if err != nil {
		return c, err
	}
	err = binary.Read(file, binary.LittleEndian, &(c.NetNameOffset))
	if err != nil {
		return c, err
	}
	var netNameAddr int64 = int64(offset) + int64(c.NetNameOffset)
	_, err = file.Seek(netNameAddr, io.SeekStart)
	if err != nil {
		return c, err
	}
	jr := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	br := bufio.NewReader(jr)
	c.NetName, err = br.ReadString(0)
	if err != nil {
		return c, err
	}
	// remove null cahr
	c.NetName = c.NetName[:len(c.NetName)-1]
	return c, nil
}

func ParseStringDataOne(file *os.File, offset int64) (s string, n int, err error) {
	var countCharacters uint16
	var o = offset
	_, err = file.Seek(o, io.SeekStart)
	if err != nil {
		return s, 0, err
	}
	err = binary.Read(file, binary.LittleEndian, &countCharacters)
	if err != nil {
		return s, 0, err
	}
	b := make([]byte, countCharacters*2)
	_, err = file.Read(b)
	if err != nil {
		return s, 0, err
	}
	utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	ur := transform.NewReader(bytes.NewReader(b), utf16.NewDecoder())
	bb, err := io.ReadAll(ur)
	if err != nil {
		return s, 0, err
	}
	n = 2 + int(countCharacters)*2
	return string(bb), n, err
}

func ParseStringData(file *os.File, offset int64, flag LinkFlags) (s StringData, err error) {
	var o = offset
	var ss string
	var n int
	if flag.HasName {
		ss, n, err = ParseStringDataOne(file, o)
		if err != nil {
			return s, err
		}
		s.NAME_STRING = ss
		o = o + int64(n)
	}
	if flag.HasRelativePath {
		ss, n, err = ParseStringDataOne(file, o)
		if err != nil {
			return s, err
		}
		s.RELATIVE_PATH = ss
		o = o + int64(n)
	}
	if flag.HasWorkingDir {
		ss, n, err = ParseStringDataOne(file, o)
		if err != nil {
			return s, err
		}
		s.WORKING_DIR = ss
		o = o + int64(n)
	}
	if flag.HasArguments {
		ss, n, err = ParseStringDataOne(file, o)
		if err != nil {
			return s, err
		}
		s.COMMAND_LINE_ARGUMENTS = ss
		o = o + int64(n)
	}
	if flag.HasIconLocation {
		ss, _, err = ParseStringDataOne(file, o)
		if err != nil {
			return s, err
		}
		s.ICON_LOCATION = ss
	}
	return s, nil
}

func ResolveShortcut(file *os.File) (path string, netname string, isdir bool, args string, err error) {
	// Header
	header, err := ParseHeader(file)
	if err != nil {
		return "", "", false, "", err
	}
	lf := NewLinkFlags(header.LinkFlags)
	if !lf.HasLinkInfo {
		return "", "", false, "", fmt.Errorf("linkinfo not found")
	}
	af := NewFileAttributesFlags(header.FileAttributesFlags)
	// LinkTargetIDList
	var linkTargetIDListSize uint32
	var linkTargetIDList LinkTargetIDList
	if lf.HasLinkTargetIDList {
		linkTargetIDList, err = ParseLinkTargetIDList(file, HEADERSIZE)
		if err != nil {
			return "", "", false, "", err
		}
		linkTargetIDListSize = 2 + uint32(linkTargetIDList.IDListSize)
	}
	// LinkInfo
	var linkInfoStartAddr int64 = int64(HEADERSIZE + linkTargetIDListSize)
	linkInfo, err := ParseLinkInfo(file, int32(linkInfoStartAddr))
	if err != nil {
		return "", "", false, "", err
	}
	// CommonNetworkRelativeLink
	var commonNetworkRelativeLink CommonNetworkRelativeLink
	if ((linkInfo.LinkInfoFlags >> 1) & 1) == 1 {
		var commonNetworkRelativeLinkAddr int64 = int64(linkInfoStartAddr) + int64(linkInfo.CommonNetworkRelativeLinkOffset)
		_, err = file.Seek(commonNetworkRelativeLinkAddr, io.SeekStart)
		if err != nil {
			return "", "", false, "", err
		}
		commonNetworkRelativeLink, err = ParseCommonNetworkRelativeLink(file, commonNetworkRelativeLinkAddr)
		if err != nil {
			return "", "", false, "", err
		}
	}
	netpath := filepath.Join(strings.ToLower(commonNetworkRelativeLink.NetName), linkInfo.CommonPathSuffix)
	if linkInfo.LocalBasePath == "" && netpath == "" {
		return "", "", false, "", errors.New("no target path found")
	}
	// StringData
	var stringDataStartAddr int64 = linkInfoStartAddr + int64(linkInfo.LinkInfoSize)
	stringData, err := ParseStringData(file, stringDataStartAddr, lf)
	if err != nil {
		return "", "", false, "", err
	}
	return linkInfo.LocalBasePath, netpath, af.FILE_ATTRIBUTE_DIRECTORY, stringData.COMMAND_LINE_ARGUMENTS, nil
}
