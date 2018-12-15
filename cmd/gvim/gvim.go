package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jroimartin/gocui"
)

type VimEditor struct {
	mode bool
}

type CommandEditor struct {
	command string
	exec    bool
}

var fileName string = "default.txt"

//Edit Implement the Editor interface
func (ve *VimEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	mode := ve.mode
	if mode {
		ve.InsertMode(v, key, ch, mod)
	} else {
		ve.NormalMode(v, key, ch, mod)
	}
}

//InsertMode of the Editor
func (ve *VimEditor) InsertMode(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	//fmt.Fprintf(v, "I")
	switch {
	case key == gocui.KeyEsc:
		ve.mode = false
		v.Title = "G-Vim Read"
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyEnter:
		v.EditNewLine()
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}

//NormalMode of the Editor
func (ve *VimEditor) NormalMode(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	//fmt.Fprintf(v, "N")
	//v.Title = "G-Vim Read"
	switch {
	case ch == 'i':
		ve.mode = true
		v.Title = "G-Vim Write"
	case ch == 'j':
		v.MoveCursor(0, 1, false)
	case ch == 'k':
		v.MoveCursor(0, -1, false)
	case ch == 'h':
		v.MoveCursor(-1, 0, false)
	case ch == 'l':
		v.MoveCursor(1, 0, false)
	case key == gocui.KeyEsc:
		ve.mode = false
	}
}

func (ce *CommandEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	//fmt.Fprintf(v, "C")
	switch {
	case key == gocui.KeyEsc:
		ce.exec = false
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyEnter:
		ce.exec = true
	}

}

func main() {

	args := os.Args
	if len(args) == 2 {
		fileName = os.Args[1]
	} else if len(args) > 2 {
		panic("Give a proper file name")
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.InputEsc = true
	g.Mouse = true
	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("command", gocui.KeyEsc, gocui.ModNone, editorMode); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("command", gocui.KeyEnter, gocui.ModNone, executeCommand); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlE, gocui.ModNone, commandMode); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("main", 0, 0, maxX, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			log.Panicln(err)
		}
		v.Title = "G-Vim Read"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.Editor = &VimEditor{}
		isFileExist := isFileExists(fileName)
		if isFileExist {
			data, err := ioutil.ReadFile(fileName)
			check(err)
			content := string(data)
			v.Clear()
			fmt.Fprintf(v, content)
		} else {
			os.Create(fileName)
		}
		if _, err = g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("command", 0, maxY-3, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			log.Panicln(err)
		}

		v.Title = "Command"
		v.Editable = true
		v.Wrap = true
		v.Frame = true

		v.Editor = &CommandEditor{}
		g.SetViewOnBottom("command")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func editorMode(g *gocui.Gui, v *gocui.View) error {
	v, err := g.SetCurrentView("main")
	check(err)
	g.InputEsc = true
	g.Mouse = true
	v.Editor = &VimEditor{}
	v.Title = "G-Vim Read"
	g.SetViewOnTop("main")
	return nil
}

func commandMode(g *gocui.Gui, v *gocui.View) error {
	v, err := g.SetCurrentView("command")
	check(err)
	v.Editor = &CommandEditor{}
	v.Clear()
	g.SetViewOnTop("command")
	return nil
}

func executeCommand(g *gocui.Gui, v *gocui.View) error {
	//fmt.Fprintln(v, "Quit called ")
	command := strings.TrimSpace(v.ViewBuffer())
	command = strings.Replace(command, "\n", "", 1)
	g.Cursor = false
	//fmt.Fprint(v, command)
	switch command {
	case "q":
		return gocui.ErrQuit
	case "w":
		commandForWrite(g, v)
	}
	return nil
}

func commandForWrite(g *gocui.Gui, v *gocui.View) {
	if fileName == "" {
		fmt.Scanf("Please give a file name > %s", &fileName)
	}
	f, err := os.Create(fileName)
	check(err)
	main, err := g.View("main")
	content := main.ViewBuffer()
	defer f.Close()
	bytes, err := f.WriteString(content)
	v.Clear()
	fmt.Fprintf(v, "Wrote %d bytes into file %s", bytes, fileName)
	f.Sync()
}

func isFileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if err == nil {
		return true
	} else if !os.IsNotExist(err) {
		check(err)
	}
	return false
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
