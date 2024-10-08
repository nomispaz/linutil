package main

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Mode - Custom type to hold value for mode of operation
type Mode int

// Declare related constants for each weekday starting with index 1
const (
	None Mode = iota + 1 // EnumIndex = 1
	Clone
	Push
)

// String - Creating common behavior - give the type a String function
func (m Mode) String() string {
	return [...]string{"None", "Clone", "Push"}[m-1]
}

// EnumIndex - Creating common behavior - give the type a EnumIndex function
func (m Mode) EnumIndex() int {
	return int(m)
}

// struct to save some variables that should be accessed from functions
type State struct {
	password string
	command  string
	mode     Mode
}

// struct that contains the definition of the Tui (variables and widgets)
type Tui struct {
	state State

	app             *tview.Application
	flex            *tview.Flex
	flexTop         *tview.Flex
	flexBottom      *tview.Flex
	flexLeftCol     *tview.Flex
	flexRightCol    *tview.Flex
	modal           *tview.Modal
	pages           *tview.Pages
	menu            *tview.List
	contents        *tview.TextView
	password_prompt *tview.InputField
}

// Initiate the Tui (variables and widgets)
func (t *Tui) Init() {

	t.state.password = "none"
	t.state.command = "none"
	t.state.mode = None

	t.app = tview.NewApplication()
	t.contents = tview.NewTextView()
	t.menu = tview.NewList()
	t.flex = tview.NewFlex()
	t.flexLeftCol = tview.NewFlex()
	t.flexRightCol = tview.NewFlex()
	t.pages = tview.NewPages()
	t.password_prompt = tview.NewInputField()
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

			command := "ls /home/simonheise/git_repos/"

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
		t.contents.Clear()
		c := make(chan string)
		//clone the repo
		command := fmt.Sprintf("git clone --progress https://github.com/nomispaz/%s /home/simonheise/git_repos/%s", current_item_text1, current_item_text1)
		go execCmd(c, command, "both")
		var output = ""
		// loop through channel of cmd ouput
		for msg := range c {
			output += output + "\n" + msg
			t.contents.SetText(output)
			t.app.ForceDraw()
		}
	case Push:
		t.contents.Clear()
		//push the repo
		t.pages.SendToFront("password")
		t.app.ForceDraw()

		// wait for password to be entered
		go func() {
			for {
				// if the password was entered, continue
				if t.state.password != "none" {

					var c = make(chan string)
					// execute the command and feed password to command. The command returns each outut line via channel
					//go execCmd(c, , t.state.password), "both")
					go execCmd(c, fmt.Sprintf("pushd /home/simonheise/git_repos/%s; git add .; git commit -m \"%s\"; git push https://nomispaz:%s@github.com/nomispaz/%s; popd", current_item_text1, "Test", t.state.password, current_item_text1), "both")

					var output = ""

					// loop through channel of cmd ouput
					for msg := range c {
						output = output + "\n" + msg
						t.contents.SetText(output)
						t.app.ForceDraw()
					}
					break
				}

			}

			t.state.command = "none"
		}()
	}
}

// setup the interface and functions belonging to the widgets like SetFinishedFunc etc.
func (t *Tui) SetupTUI() {

	// ScrollToEnd is used so that the list is always scrolled together with the text
	t.contents.SetTextAlign(tview.AlignLeft).SetText("").SetDynamicColors(false).SetTextColor(tcell.ColorSlateGrey).ScrollToEnd()

	// define password prompt for the sudo password
	t.password_prompt.SetLabel("Enter root password: ")
	t.password_prompt.SetFinishedFunc(func(key tcell.Key) {
		t.state.password = t.password_prompt.GetText()
		t.pages.SendToFront("flex")
	})

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

	// define pages so that we are able to switch between main layout and popups
	t.pages.
		AddPage("password", modal(t.password_prompt, 40, 10), true, true).
		AddPage("flex", t.flex, true, true)

	// start the app with the main layout shown
	t.pages.SendToFront("flex")
}

// wrapper to execute comands that require a sudo password
func (t *Tui) handle_sudo_cmd(command string) {
	// show the password prompt
	t.pages.SendToFront("password")
	t.app.ForceDraw()

	// wait for password to be entered
	go func() {
		for {
			// if the password was entered, continue
			if t.state.password != "none" {

				var c = make(chan string)
				// execute the command and feed password to command. The command returns each outut line via channel
				//go execCmd(c, , t.state.password), "both")
				go execCmd(c, fmt.Sprintf(command, t.state.password), "both")

				var output = ""

				// loop through channel of cmd ouput
				for msg := range c {
					output = output + "\n" + msg
					t.contents.SetText(output)
					t.app.ForceDraw()
				}
				break
			}

		}

		t.state.command = "none"
	}()
}

// execute command and return the combined stderr and stdout via pipe to a channel. Channel is closed at the end to prevend a deadlock
func execCmd(c chan string, command string, mode string) {
	cmd := exec.Command("bash", "-c", command)

	// create a pipe for stdout
	stdout, _ := cmd.StdoutPipe()
	if mode == "both" {
		// combine outputs of stderr and stdout
		cmd.Stderr = cmd.Stdout
	}
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		c <- scanner.Text()
	}

	cmd.Wait()
	// to prevent deadlock panic after the function finished
	close(c)
}

// Create the main application with a new tui
func CreateApplication() *Tui {
	return new(Tui)
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
			case 'c':
				go t.handle_sudo_cmd("echo %s | sudo -S ls -l")
			}
		}
		return event
	})
}

func main() {

	tui := CreateApplication()
	tui.Init()
	tui.SetupTUI()
	tui.Keybindings()

	if err := tui.app.SetRoot(tui.pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
