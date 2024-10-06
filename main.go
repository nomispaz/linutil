package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type State struct {
	password string
	command  string
}
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

func (t *Tui) Init() {

	t.state.password = "none"
	t.state.command = "none"

	t.app = tview.NewApplication()
	t.contents = tview.NewTextView()
	t.menu = tview.NewList()
	t.flex = tview.NewFlex()
	t.flexLeftCol = tview.NewFlex()
	t.flexRightCol = tview.NewFlex()
	t.pages = tview.NewPages()
	t.password_prompt = tview.NewInputField()
}

func (t *Tui) SetupTUI() {

	t.contents.SetTextAlign(tview.AlignLeft).SetText("").SetDynamicColors(false).SetTextColor(tcell.ColorSlateGrey)

	t.password_prompt.SetLabel("Enter root password: ")
	t.password_prompt.SetFinishedFunc(func(key tcell.Key) {
		t.state.password = t.password_prompt.GetText()
		t.pages.SendToFront("flex")

	})

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

	t.flexLeftCol.SetDirection(tview.FlexRow)
	t.flexLeftCol.AddItem(t.menu, 0, 100, true).SetBorder(true)
	t.flexRightCol.SetDirection(tview.FlexRow)
	t.flexRightCol.AddItem(t.contents, 0, 100, false).SetBorder(true)

	t.flex.SetDirection(tview.FlexColumn).
		AddItem(t.flexLeftCol, 0, 1, false).
		AddItem(t.flexRightCol, 0, 3, false)

	t.pages.
		AddPage("password", modal(t.password_prompt, 40, 10), true, true).
		AddPage("flex", t.flex, true, true)

	t.pages.SendToFront("flex")
}

func (t *Tui) handle_sudo_cmd() {
	t.pages.SendToFront("password")
	t.app.ForceDraw()

	// wait for password to be entered
	go func() {
		for {
			if t.state.password != "none" {

				var c = make(chan string)
				go execCmd(c, fmt.Sprintf("echo %s | sudo -S ls -l", t.state.password))
				//go execCmd(c, "pushd /home/simonheise/git_repos/nompac/overlays/tuxedo-drivers-dkms/; makepkg -cCsrf --skippgpcheck; popd")
				var output = ""

				for msg := range c {

					if msg == "[sudo] password for root: Sorry, try again." {
						t.contents.SetText("Password was not correct, retry.")
						t.state.password = "none"
						t.app.ForceDraw()
						break

					} else {
						if !strings.HasPrefix(strings.Trim(msg, ""), "[sudo]") {
							output = output + "\n" + msg
							t.contents.SetText(output)
							t.app.ForceDraw()
						}
					}

				}
				break
			}
		}
		t.state.command = "none"
	}()
}

func execCmd(c chan string, command string) {
	cmd := exec.Command("bash", "-c", command)

	// create a pipe for stdout
	stdout, _ := cmd.StdoutPipe()
	// combine outputs of stderr and stdout
	cmd.Stderr = cmd.Stdout
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		c <- scanner.Text()
	}

	cmd.Wait()
	// to prevent deadlock panic after the function finished
	close(c)
}

// test documentation
func CreateApplication() *Tui {
	return new(Tui)
}

func (t *Tui) Keybindings() {

	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		// if key is ESC, switch back to root page
		case tcell.KeyEsc:
			t.pages.SendToFront("flex")
		case tcell.KeyRune:
			switch event.Rune() {
			// execute selected script
			// only if file was selected
			case 'q':
				t.app.Stop()

			case 'c':
				// execute commend
				go t.handle_sudo_cmd()
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
