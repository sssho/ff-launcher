package lib

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Shortcut struct {
	Path    string
	TPath   string
	Args    string
	IsDir   bool
	Parent  string
	Org     Origin
	ModTime time.Time
}

func NewShortcut(dir string, finfo fs.FileInfo, org Origin) (s Shortcut, err error) {
	s.Path = filepath.Join(dir, finfo.Name())
	s.ModTime = finfo.ModTime()
	s.Org = org

	f, err := os.Open(s.Path)
	if err != nil {
		return s, err
	}
	defer f.Close()
	tpath, err := ResolveShortcut(f)
	if err != nil {
		return s, err
	}
	if tpath == "" {
		return s, fmt.Errorf("tpath is nil")
	}
	_, err = os.Stat(tpath)
	if err != nil {
		return s, err
	}
	isdir, err := isDir(tpath)
	if err != nil {
		isdir = false
	}
	var parent string
	if !isdir {
		parent = filepath.Dir(tpath)
	}
	s.TPath = tpath
	s.Args = "" // TODO
	s.IsDir = isdir
	s.Parent = parent
	return s, nil
}

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

func ResolveShortcut(file *os.File) (path string, err error) {
	_, err = file.Seek(LINKFLAGS_OFFSET, io.SeekStart)
	if err != nil {
		return "", err
	}
	var LinkFlags uint32
	err = binary.Read(file, binary.LittleEndian, &(LinkFlags))
	if err != nil {
		return "", err
	}
	lf := NewLinkFlags(LinkFlags)
	if !lf.HasLinkInfo {
		return "", fmt.Errorf("linkinfo not found")
	}
	var LinkTargetIDListSize uint16
	if lf.HasLinkTargetIDList {
		_, err = file.Seek(HEADERSIZE, io.SeekStart)
		if err != nil {
			return "", err
		}
		err = binary.Read(file, binary.LittleEndian, &(LinkTargetIDListSize))
		if err != nil {
			return "", err
		}
	}
	var LinkInfoStartAddr int64 = int64(HEADERSIZE + 2 + LinkTargetIDListSize)
	_, err = file.Seek(LinkInfoStartAddr+8, io.SeekStart)
	if err != nil {
		return "", err
	}
	var LinkInfoFlags uint32
	err = binary.Read(file, binary.LittleEndian, &(LinkInfoFlags))
	if err != nil {
		return "", err
	}
	VolumeIDAndLocalBasePath := (LinkInfoFlags & 1) == 1
	if !VolumeIDAndLocalBasePath {
		return "", fmt.Errorf("localbasepath not found")
	}
	var LocalBasePathOffset uint32
	_, err = file.Seek(LinkInfoStartAddr+16, io.SeekStart)
	if err != nil {
		return "", err
	}
	err = binary.Read(file, binary.LittleEndian, &(LocalBasePathOffset))
	if err != nil {
		return "", err
	}
	var LocalBasePathAddr int64 = int64(LinkInfoStartAddr + int64(LocalBasePathOffset))
	_, err = file.Seek(LocalBasePathAddr, io.SeekStart)
	if err != nil {
		return "", err
	}
	// https://zenn.dev/mattn/articles/fd545a14b0ffdf
	jr := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	br := bufio.NewReader(jr)
	path, err = br.ReadString(0)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s Shortcut) Text() (text string) {
	if s.IsDir {
		text = fmt.Sprintf("%s %s", folderPrefix, s.TPath)
	} else {
		text = fmt.Sprintf("%s %s %s", filePrefix, s.TPath, s.Args)
	}
	text = strings.TrimSpace(text)
	return
}

func NewShortcuts(dir string, origin Origin) ([]Shortcut, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	shortcuts := make([]Shortcut, 0, len(files))
	for _, file := range files {
		shortcut, err := NewShortcut(dir, file, origin)
		if err != nil {
			continue
		}
		shortcuts = append(shortcuts, shortcut)
	}

	return shortcuts, nil
}

func isDir(tpath string) (bool, error) {
	info, err := os.Stat(tpath)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return true, nil
	} else {
		return false, nil
	}
}

func GetShortcutTexts(shortcuts []Shortcut) []string {
	texts := make([]string, 0, len(shortcuts))
	checkDuplicate := make(map[string]bool)
	for _, s := range shortcuts {
		key := strings.TrimSpace(s.TPath + s.Args)
		if checkDuplicate[key] {
			continue
		}
		// Add parent
		if !s.IsDir {
			if !checkDuplicate[s.Parent] {
				texts = append(texts, fmt.Sprintf("%s %s", folderPrefix, s.Parent))
				checkDuplicate[s.Parent] = true
			}
		}
		texts = append(texts, s.Text())
		checkDuplicate[key] = true
	}

	return texts
}
