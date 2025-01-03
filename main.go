package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// colors for printout
var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

// variable that will contain the provided configuration
var config Config

// struct that contains the definition of the Tui (variables and widgets)
type Tui struct {
	state State

	app  *tview.Application
	flex *tview.Flex
	// flexTop      *tview.Flex
	// flexBottom   *tview.Flex
	flexLeftCol  *tview.Flex
	flexRightCol *tview.Flex

	flexPopup         *tview.Flex
	flexPopupUsername *tview.InputField
	flexPopupPassword *tview.InputField
	flexPopupCommit   *tview.InputField

	pages    *tview.Pages
	menu     *tview.List
	contents *tview.TextView
}

// Initiate the Tui (variables and widgets)
func (t *Tui) Init() {

	t.state.password = "none"
	t.state.command = "none"
	t.state.user = "none"
	t.state.commit_msg = "none"
	t.state.input = "none"
	t.state.mode = None

	t.app = tview.NewApplication()
	t.contents = tview.NewTextView()
	t.menu = tview.NewList()
	t.flex = tview.NewFlex()
	t.flexLeftCol = tview.NewFlex()
	t.flexRightCol = tview.NewFlex()

	t.flexPopup = tview.NewFlex()
	t.flexPopupUsername = tview.NewInputField()
	t.flexPopupPassword = tview.NewInputField()
	t.flexPopupCommit = tview.NewInputField()

	t.pages = tview.NewPages()
}

func (t *Tui) InputPage(wg *sync.WaitGroup, username Field_type, password Field_type, commit Field_type) {
	defer wg.Done()

	number_inputfields := 0
	var inputvalues = make(map[string]string)

	t.flexPopup.Clear()
	t.state.user = "none"
	t.state.password = "none"
	t.state.commit_msg = "none"

	if username == Show {
		t.flexPopupUsername.SetLabel("Enter username: ")
		t.flexPopupUsername.SetDoneFunc(func(key tcell.Key) {
			t.state.user = t.flexPopupUsername.GetText()
			inputvalues["user"] = t.state.user
			if commit == Show {
				t.app.SetFocus(t.flexPopupCommit)
			}
			if password == Show {
				t.app.SetFocus(t.flexPopupPassword)
			}
		})

		t.flexPopup.AddItem(t.flexPopupUsername, 0, 1, true)
		number_inputfields += 1
		inputvalues["user"] = ""
	}
	if password == Show {
		t.flexPopupPassword.SetLabel("Enter password: ")
		t.flexPopupPassword.SetMaskCharacter('*')
		t.flexPopupPassword.SetDoneFunc(func(key tcell.Key) {
			t.state.password = t.flexPopupPassword.GetText()
			inputvalues["password"] = t.state.password
			if username == Show {
				t.app.SetFocus(t.flexPopupUsername)
			}
			if commit == Show {
				t.app.SetFocus(t.flexPopupCommit)
			}
		})

		t.flexPopup.AddItem(t.flexPopupPassword, 0, 1, false)
		if number_inputfields == 0 {
			t.app.SetFocus(t.flexPopupPassword)
		}
		number_inputfields += 1
		inputvalues["password"] = ""
	}
	
	if commit == Show {
		t.flexPopupCommit.SetLabel("Enter commit message: ")
		t.flexPopupCommit.SetDoneFunc(func(key tcell.Key) {
			t.state.commit_msg = t.flexPopupCommit.GetText()
			inputvalues["commit"] = t.state.commit_msg
			if password == Show {
				t.app.SetFocus(t.flexPopupPassword)
			}
			if username == Show {
				t.app.SetFocus(t.flexPopupUsername)
			}
		})

		t.flexPopup.AddItem(t.flexPopupCommit, 0, 1, false)
		if number_inputfields == 0 {
			t.app.SetFocus(t.flexPopupCommit)
		}
		number_inputfields += 1
		inputvalues["commit"] = ""

	}

	t.pages.SendToFront("popup")

	// if all inputfields are filled, close the popup
	input_done := false

	for {
		// sleep of 2 ms to prevent 100% cpu usage
		time.Sleep(2 * time.Millisecond)

		input_done = true
		for _, input := range inputvalues {
			if input == "" {
				input_done = false
			}
		}

		if input_done {
			t.pages.SendToFront("flex")
			break
		}
	}
}

// get mode from list item
func (t *Tui) GetTextFromListItem() {
	current_item_text1, _ := t.menu.GetItemText(t.menu.GetCurrentItem())

	switch t.state.mode {
	case None:
		// no mode selected --> still the starting menu is shown
		switch current_item_text1 {
		case "Clone repos":
			t.state.mode = Clone
			c := make(chan string)

			command := "curl https://api.github.com/users/nomispaz/repos | grep full_name | cut -d':' -f 2 | cut -d'\"' -f 2 | cut -d'/' -f 2"

			go execCmd(c, command, "out")

			// remove all items from the list
			t.menu.Clear()

			// loop through channel of cmd ouput
			for msg := range c {
				t.menu.AddItem(msg, "", '-', t.GetTextFromListItem)
				t.app.ForceDraw()

			}
		case "Push repos":
			t.state.mode = Push
			c := make(chan string)

			command := fmt.Sprintf("ls %s", config.GitDir)

			go execCmd(c, command, "both")

			// remove all items from the list
			t.menu.Clear()

			// loop through channel of cmd ouput
			for msg := range c {
				t.menu.AddItem(msg, "", '-', t.GetTextFromListItem)
				t.app.ForceDraw()
			}
		}
	case Clone:
		// mode clone --> if item is selected, it is cloned
		c := make(chan string)
		//clone the repo
		command := fmt.Sprintf(
			"git clone --progress https://github.com/nomispaz/%s %s/%s",
			current_item_text1,
			config.GitDir,
			current_item_text1,
		)
		go execCmd(c, command, "both")
		var output = ""
		// loop through channel of cmd ouput
		for msg := range c {
			output += output + "\n" + msg
			t.contents.SetText(output)
			t.app.ForceDraw()
		}
	case Push:
		//push the repo
		// wait groups wgp: wait for password, wgu: wait for username
		c := make(chan string)

		var wg sync.WaitGroup
		wg.Add(1)

		go t.InputPage(&wg, Show, Show, Show)

		go func() {

			wg.Wait()

			go execCmd(c, fmt.Sprintf(
				"pushd %s/%s; git add .; git commit -m \"%s\"; git push --progress https://%s:%s@github.com/nomispaz/%s; popd",
				config.GitDir,
				current_item_text1,
				t.state.commit_msg,
				t.state.user,
				t.state.password,
				current_item_text1,
			), "both")

			var output = ""

			for msg := range c {
				output = output + "\n" + msg
				t.contents.SetText(output)
				t.app.ForceDraw()
			}
			t.state.password = "none"
			t.state.user = "none"
			t.state.password = "none"
		}()
	}
}

// setup the interface and functions belonging to the widgets like SetFinishedFunc etc.
func (t *Tui) SetupTUI() {

	// ScrollToEnd is used so that the list is always scrolled together with the text
	t.contents.SetTextAlign(tview.AlignLeft).SetText("").SetDynamicColors(false).SetTextColor(tcell.ColorSlateGrey).ScrollToEnd()

	// hide second row for the menu items
	t.menu.ShowSecondaryText(false)
	t.menu.AddItem("Clone repos", "", '-', t.GetTextFromListItem)
	t.menu.AddItem("Push repos", "", '-', t.GetTextFromListItem)

	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}

	// define the layout of the main window
	t.flexLeftCol.SetDirection(tview.FlexRow)
	t.flexLeftCol.AddItem(t.menu, 0, 100, true).SetBorder(true)
	t.flexRightCol.SetDirection(tview.FlexRow)
	t.flexRightCol.AddItem(t.contents, 0, 100, false).SetBorder(true)

	t.flex.SetDirection(tview.FlexColumn).
		AddItem(t.flexLeftCol, 0, 1, true).
		AddItem(t.flexRightCol, 0, 3, false)

	t.flexPopup.SetDirection(tview.FlexRow)

	// define pages so that we are able to switch between main layout and popups
	t.pages.
		AddPage("popup", modal(t.flexPopup, 40, 10), true, true).
		AddPage("flex", t.flex, true, true)

	// start the app with the main layout shown
	t.pages.SendToFront("flex")
}

// define key bindings for the tui
func (t *Tui) Keybindings() {

	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		// if key is ESC, switch back to root page
		case tcell.KeyEsc:
			t.pages.SendToFront("flex")
		case tcell.KeyCtrlC:
			t.app.Stop()
		case tcell.KeyRune:
			switch event.Rune() {
			// execute command
			// case 'c':

			// i := 10
			// fmt.Println(i)
			}
		}
		return event
	})
}

// Create the main application with a new tui
func CreateApplication() *Tui {
	return new(Tui)
}

func main() {

	// Read patch configuration
	configFilePath := "~/.config/linutil/configs/config.json"
	config = read_config(configFilePath)
	
	tui := CreateApplication()
	tui.Init()
	tui.SetupTUI()
	tui.Keybindings()

	if err := tui.app.SetRoot(tui.pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
