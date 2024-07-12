package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////////
// global variables
var cmdArray [100]string

type updateMsg struct {
	// array to save items that can be selected in a displayed list
	listitems []string
	// map to contain the selected items
	selected map[int]string
	// position of the cursor
	cursor int
	// header
	header string
	// footer
	footer string
}

//////////////////////////////////////////////////////////////////////////////////
// Tui definition and functions

// define Tui structure
type Tui struct {
	// array to save items that can be selected in a displayed list
	listitems []string
	// map to contain the selected items
	selected map[int]string
	// position of the cursor
	cursor int
	// mode defines if git online to clone, git offline to push or browser to browse files is active
	mode string
	// string for the header
	header string
	//string for the footer
	footer string

	// save the configs in the Tui-structure itself
	configs map[string]string
	// home of the curent user
	userhome string

	// saves the current folder for the browser-mode
	curfolder string
}

// first initialization of tui
func initTui() (t Tui) {
	return Tui{
		listitems: []string{"git online", "git offline", "browser"},
		selected:  make(map[int]string),
		header:    "nomispaz linutil: first select the operation mode.\n\n",
		footer:    "\nAvailable functions\n",
		curfolder: "",
	}
}

// Perform some initial I/O, for now, parse the config file 
func (t *Tui) Init() tea.Cmd {
	t.configs = parseconfigfile()
	return nil
}

func (t *Tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case updateMsg:
		t.selected = msg.selected
		t.listitems = msg.listitems
		t.cursor = msg.cursor
		t.footer = msg.footer

	// Is it a key press?
	case tea.KeyMsg:

		// What was the actual key pressed?
		switch msg.String() {

		// enter selection for operational mode --> remove all prior selections and the operational mode
		case "m":
			// return to mode selection Reset all selections
			t.listitems = []string{"git online", "git offline", "browser"}
			t.selected = make(map[int]string)
			t.mode = ""
			t.cursor = 0
			t.footer  = "\nAvailable functions\n" +
				    "- q     : quit\n" +
				    "- Enter : execute selected item\n"

		case "e":
			// execute file
			if t.mode == "browser" {
				command := ""
				for i := range t.selected {
					selectedfile := t.curfolder + "/" + t.selected[i]
					command += "chmod +x " +
					selectedfile +
					"; " +
					selectedfile +
					"; read -n 1 -s -r -p 'Press key to continue.'; clear; "
				}
			return t,tea.ExecProcess(exec.Command("bash", "-c", command), nil)
			}

		case "b":
			// if in browser-mode, one level up
			if t.mode == "browser" {
				lastInd := strings.LastIndex(t.curfolder, "/")
				t.curfolder = t.curfolder[:lastInd]

				command := "ls " + t.curfolder
			
				footer := "\nAvailable functions\n" +
					"- Enter : open selected folder or file\n" +
					"- b     : one level up\n" +
					"- e     : execute selected file\n" +
					"- m     : return to mode selection\n" +
					"- q     : quit"
				return t, tea.Cmd(func() tea.Msg {
					return runCmd(command, footer)
				})
			}

		case "c":
			// if in mode git online, clone the selected repositories
			if t.mode == "git online" {
				command := "echo cloning repository "
				for i := range t.selected {
					command += t.selected[i] + 
					"; git clone https://github.com/" +
					t.configs["gituser"]+ "/" +
					t.selected[i] +
					" " + t.configs["gitfolder"] + "/" +
					t.selected[i] +
					"; "
				}
				return t,tea.ExecProcess(exec.Command("bash", "-c", command), nil)
			}

		case "p":
			// if in mode git offline, push selected repositories
			if t.mode == "git offline" {
				command := "echo pushing repository "
				for i := range t.selected {
					command += t.selected[i] + 
						"; cd " +
						t.configs["gitfolder"] + "/" +
						t.selected[i] + 
						"; git add .; " +
						"git commit -m " +
						"'pushed by linutil'" + 
						"; git push; "
				}
				return t, tea.ExecProcess(exec.Command("bash", "-c", command),nil)
			}
			
		// These keys should exit the program.
		case "ctrl+c", "q":
			return t, tea.Quit

		// The "up" keys move the cursor up
		case "up":
			if t.cursor > 0 {
				t.cursor--
			}

			// The "down" keys move the cursor down
		case "down":
			if t.cursor < len(t.listitems)-1 {
				t.cursor++
			}

		// The spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case " ":
			_, ok := t.selected[t.cursor]
			if ok {
				delete(t.selected, t.cursor)
			} else {
				t.selected[t.cursor] = t.listitems[t.cursor]
			}

		// enter --> perform action according to selected entries
		case "enter":
			// if operational mode is not set
			if t.mode == "" {
				for i := range t.selected {
					t.mode = t.selected[i]
				}
			}
			if t.mode == "git online" {
				// get online repositories
				command := "curl https://api.github.com/users/nomispaz/repos | grep full_name | cut -d':' -f 2 | cut -d'\"' -f 2 | cut -d'/' -f 2"
				footer := "\nAvailable functions\n" +
					"- c     : clone selected repos\n" +
					"- m     : return to mode selection\n" +
					"- q     : quit"

				return t, tea.Cmd(func() tea.Msg {
					return runCmd(command, footer)
				})
			}
			if t.mode == "git offline" {
				// get repositories already downloaded to gitfolder
				command := "ls " +
					t.configs["gitfolder"]
				footer := "\nAvailable functions\n" +
					"- p     : push selected repos\n" +
					"- m     : return to mode selection\n" +
					"- q     : quit"

				return t, tea.Cmd(func() tea.Msg {
					return runCmd(command, footer)
				})
			}
			if t.mode == "browser" {
				command := "ls "
				if t.curfolder == "" {
					t.curfolder = t.configs["defaultfolder"]
					command += t.curfolder
				} else {
					for i := range t.selected {
						t.curfolder = t.curfolder + "/" +
						t.selected[i]
						command += t.curfolder
					}
				}
				footer := "\nAvailable functions\n" +
					"- Enter : open selected folder or file\n" +
					"- b     : one level up\n" +
					"- e     : execute selected file\n" +
					"- m     : return to mode selection\n" +
					"- q     : quit"
				return t, tea.Cmd(func() tea.Msg {
					return runCmd(command, footer)
				})
			}
		}
	
	}
	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return t, nil
}

func (t *Tui) View() string {
	// The header
	s := t.header

	// Iterate over our choices
	for i, choice := range t.listitems {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if t.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := t.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += t.footer
	// Send the UI for rendering
	return s
}

//////////////////////////////////////////////////////////////////////////////////

// parseconfigfile: read config file from users configdir and parse settings
func parseconfigfile() map[string]string {

	configs := make(map[string]string)

	configdir, err := os.UserConfigDir()

	if err != nil {
		fmt.Println("No configfile found")
	}
	
	fileContent, err := os.ReadFile(configdir + "/linutil/config")
	
	if err != nil {
		panic(err)
	}

	// split file at endline
	fileContent_split := strings.Fields(string(fileContent))
	
	for _, s := range fileContent_split {
		// split row at = sign
		row_split := strings.Split(s, "=")
		configs[row_split[0]] = row_split[1]
	}

	return configs
}
// used to run commands that are parsed and returned to update the view
func runCmd(command string, footer string) tea.Msg {
	cmd, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		panic(err)
	}
	
	// convert result byte to string and split at newline
	result := string(cmd)
	result_split := strings.Fields(result)

	return updateMsg{
		listitems: result_split,
		selected:  make(map[int]string),
		cursor: 0,
		footer: footer,
	}
}

// ////////////////////////////////////////////////////////////////////////////////
//
// main function
func main() {

	userhome, _ := os.UserHomeDir()

	m := initTui()
	m.footer += "- q     : quit\n"
	m.footer += "- Enter : execute selected item\n"

	m.userhome = userhome

	p := tea.NewProgram(&m)
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Program ended unexpectedly due to: %v", err)
		os.Exit(1)
	}
}
