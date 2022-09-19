package lib

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ole "github.com/go-ole/go-ole"
	"github.com/mattn/go-colorable"
)

const logo = `
FF is ff
   version 1.0`

const (
	CMD_ALL          = "a"
	CMD_XLSX         = "e"
	CMD_WORD         = "w"
	CMD_PPT          = "pp"
	CMD_PDF          = "p"
	CMD_VISIO        = "v"
	CMD_TXT          = "t"
	CMD_SORT         = "s"
	CMD_FOLDER       = "f"
	CMD_FOCUS_XLSX   = "fx"
	CMD_FOCUS_FOLDER = "ff"
	CMD_RELOAD       = "R"
	CMD_QUIT         = "q"
)

const (
	BLANK_LINE = "  "
	PROMPT     = " ff > "
	STATUS     = " status:"
	CMDLINES   = 15
)

var out = colorable.NewColorableStdout()

func cursorRight(n int) {
	fmt.Fprintf(out, "\033[%dC", n)
}

func cursorUp(n int) {
	fmt.Fprintf(out, "\033[%dA", n)
}

func killLineAfter() {
	fmt.Fprintf(out, "\033[0K")
}

func killLineAll() {
	fmt.Fprintf(out, "\033[2K")
}

func flushAfter() {
	fmt.Fprintf(out, "\033[0J")
}

func PrintCommands(s SortType) {
	fmt.Printf(" [%2v] %v\n", CMD_ALL, "ALL Files/Folders")
	fmt.Printf(" [%2v] %v\n", CMD_FOLDER, "Folders")
	fmt.Printf(" [%2v] %v\n", CMD_XLSX, "Excel")
	fmt.Printf(" [%2v] %v\n", CMD_WORD, "Word")
	fmt.Printf(" [%2v] %v\n", CMD_PPT, "PowerPoint")
	fmt.Printf(" [%2v] %v\n", CMD_PDF, "PDF")
	fmt.Printf(" [%2v] %v\n", CMD_VISIO, "Visio")
	fmt.Printf(" [%2v] %v\n", CMD_TXT, "Txt")
	fmt.Println(BLANK_LINE)
	fmt.Printf(" [%2v] %v\n", CMD_FOCUS_XLSX, "Focus opened xlsx")
	fmt.Printf(" [%2v] %v\n", CMD_FOCUS_FOLDER, "Focus opened folder")
	fmt.Println(BLANK_LINE)
	fmt.Printf(" [%2v] %v %v\n", CMD_SORT, "Sort Type:", s)
	fmt.Printf(" [%2v] %v\n", CMD_RELOAD, "Reload(TBD)")
	fmt.Printf(" [%2v] %v\n", CMD_QUIT, "Quit")
}

type Action func(app *App) (status string, err error)

type SortType int

func (s SortType) String() string {
	if s == SORT_BY_PATH {
		return "by path name"
	} else {
		return "by access time"
	}
}

const (
	SORT_BY_PATH = iota
	SORT_BY_TIME
)

type App struct {
	config Config
	hist   History
	sort   SortType
	action map[string]Action
}

func NewApp() (a *App, err error) {
	var app App
	err = app.config.Load()
	if err != nil {
		return nil, err
	}
	app.hist = History{}
	app.sort = SORT_BY_TIME
	app.action = make(map[string]Action)
	app.action[CMD_ALL] = actionAll
	app.action[CMD_XLSX] = actionXlsx
	app.action[CMD_WORD] = actionWord
	app.action[CMD_PPT] = actionPpt
	app.action[CMD_PDF] = actionPdf
	app.action[CMD_VISIO] = actionVisio
	app.action[CMD_TXT] = actionTxt
	app.action[CMD_SORT] = actionSort
	app.action[CMD_FOLDER] = actionFolder
	app.action[CMD_FOCUS_XLSX] = actionFocusXlsx
	app.action[CMD_FOCUS_FOLDER] = actionFocusFolder
	app.action[CMD_RELOAD] = actionReload
	app.action[CMD_QUIT] = actionQuit
	return &app, nil
}

func actionSelectbyFF(r io.Reader, prompt string) (status string, err error) {
	selected, err := SelectByFF(r, "", prompt)
	if err != nil {
		return "ff error! try again..", err
	}
	if selected == "" {
		return "selected is empty", errors.New("hoge")
	}
	err = RunApp(selected)
	if err != nil {
		return "runApp err!", fmt.Errorf("RunAPP: selected [%s], %w", selected, err)
	}
	return "run!", nil
}

func actionAll(app *App) (status string, err error) {
	var b bytes.Buffer
	for _, h := range (*app).hist {
		_, err = fmt.Fprintf(&b, "%s\n", h.path)
		if err != nil {
			return "actionAll error!", err
		}
	}
	status, err = actionSelectbyFF(&b, "All files/folders >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionFolder(app *App) (status string, err error) {
	var b bytes.Buffer
	for _, h := range (*app).hist {
		if !h.isDir {
			continue
		}
		_, err = fmt.Fprintf(&b, "%s\n", h.path)
		if err != nil {
			return "actionDir error!", err
		}
	}
	status, err = actionSelectbyFF(&b, "folder >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func filterAndRun(hist History, filter []string, prompt string) (status string, err error) {
	var b bytes.Buffer
	for _, h := range hist {
		for _, f := range filter {
			if strings.HasSuffix(h.path, f) {
				_, err := fmt.Fprintf(&b, "%s\n", h.path)
				if err != nil {
					return "filter error!", err
				}
				break
			}
		}
	}
	status, err = actionSelectbyFF(&b, prompt)
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionXlsx(app *App) (status string, err error) {
	var filter = []string{".xlsx", ".xlsm"}
	status, err = filterAndRun((*app).hist, filter, "excel >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionWord(app *App) (status string, err error) {
	var filter = []string{".docx"}
	status, err = filterAndRun((*app).hist, filter, "word >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionPpt(app *App) (status string, err error) {
	var filter = []string{".pptx"}
	status, err = filterAndRun((*app).hist, filter, "ppt >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionPdf(app *App) (status string, err error) {
	var filter = []string{".pdf"}
	status, err = filterAndRun((*app).hist, filter, "pdf >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionVisio(app *App) (status string, err error) {
	var filter = []string{".vsdx"}
	status, err = filterAndRun((*app).hist, filter, "visio >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionTxt(app *App) (status string, err error) {
	var filter = []string{".txt"}
	status, err = filterAndRun((*app).hist, filter, "txt >")
	if err != nil {
		return status, err
	}
	return "run!", nil
}

func actionSort(app *App) (status string, err error) {
	if app.sort == SORT_BY_PATH {
		app.sort = SORT_BY_TIME
		app.hist.SortByTime()
	} else {
		app.sort = SORT_BY_PATH
		app.hist.SortByPath()
	}
	return "sort type is set", nil
}

func actionFocusXlsx(app *App) (status string, err error) {
	ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED|ole.COINIT_DISABLE_OLE1DDE)
	defer ole.CoUninitialize()
	excel, err := NewExcel()
	if err != nil {
		return "No excel found!", err
	}
	names, err := excel.OpenedWookBookNames()
	if err != nil {
		return "excel open error!", err
	}
	var b bytes.Buffer
	for _, name := range names {
		_, err = fmt.Fprintf(&b, "%s\n", filepath.ToSlash(name))
		if err != nil {
			return "internal error", err
		}
	}
	name, err := SelectByFF(&b, "", "opened excel >")
	if err != nil {
		return "ff error! try again..", err
	}
	if name == "" {
		return "selected is empty", errors.New("hoge")
	}
	err = excel.ActivateWorkBook(strings.TrimSpace(filepath.FromSlash(name)))
	if err != nil {
		return "excel activate error!", err
	}
	return "excel focus done", nil
}

func actionFocusFolder(app *App) (status string, err error) {
	paths, err := GetOpenedFolderPaths()
	if err != nil {
		return "actionFocusFolder error!", err
	}
	if len(paths) == 0 {
		return "No opened folder found!", nil
	}
	var b bytes.Buffer
	for _, path := range paths {
		_, err = fmt.Fprintf(&b, "%s\n", path)
		if err != nil {
			return "actionFocusFolder error!", err
		}
	}
	status, err = actionSelectbyFF(&b, "opened folders >")
	if err != nil {
		return status, err
	}
	return "focus!", nil
}

func actionReload(app *App) (status string, err error) {
	return "TBD!", nil
}

func actionQuit(app *App) (status string, err error) {
	// err = app.hist.Save()
	// if err != nil {
	// 	return "hist save error!", err
	// }
	return "Q", nil
}

func (app *App) Exec(cmd string) (status string, err error) {
	if action, ok := app.action[cmd]; ok {
		status, err = action(app)
		if err != nil {
			return status, err
		}
	} else {
		return fmt.Sprintf("no such command (%v)", cmd), nil
	}
	return status, nil
}

func Run() error {
	app, err := NewApp()
	if err != nil {
		return err
	}
	hist, err := FindHistory(app.config)
	if err != nil {
		return err
	}
	app.hist = hist
	plen := len(PROMPT)
	scanner := bufio.NewScanner(os.Stdin)
	status := ""
	fmt.Fprintln(out, logo)
	fmt.Println("")
	for {
		fmt.Println(PROMPT)
		fmt.Println(BLANK_LINE)
		PrintCommands(app.sort)
		fmt.Println(BLANK_LINE)
		fmt.Println(STATUS, status)
		cursorUp(4 + CMDLINES)
		cursorRight(plen)
		if !scanner.Scan() {
			return errors.New("canceled")
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}
		status, _ = app.Exec(scanner.Text())
		if status == "Q" {
			cursorUp(1)
			killLineAll()
			flushAfter()
			break
		}
		flushAfter()
		cursorUp(1)
		killLineAfter()
	}
	return nil
}
