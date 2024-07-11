package main

import (
	"bufio"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////////
// global variables

// TuiState is used to determine if the program should be stopped completely or just the tui to be interrupted
// -1 : close program
//
//	0 : interrupt tui
//	1 : restart tui
var TuiState int

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
}

// first initialization of tui
func initTui() (t Tui) {
	return Tui{
		listitems: []string{"git online", "git offline", "browser"},
		selected:  make(map[int]string),
		header:    "nomispaz linutil: first select the operation mode.\n\n",
		footer:    "\nAvailable functions\n",
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
			t.listitems = []string{"git online", "git offline", "browser"}
			t.selected = make(map[int]string)
			t.mode = ""
			t.cursor = 0
			t.footer  = "\nAvailable functions\n" +
				    "- q     : quit\n" +
				    "- Enter : execute selected item\n"

		case "c":
			if t.mode == "git online" {
				cmdArrayIdx := 0
				for i := range t.selected {
					cmdArray[cmdArrayIdx] = "git clone https://github.com/" + t.selected[i] + " /home/simonheise/test/" + t.selected[i]
					cmdArrayIdx += 1
				}
				TuiState = 0
				return t, tea.Quit

			}

		case "p":
			if t.mode == "git offline" {
				command := ""
				for i := range t.selected {
					command += "echo pushing " + t.selected[i] + "; cd /home/simonheise/git_repos/" + t.selected[i] + "; git add .; git commit -m " + string(i) + "; git push; "
				}
				return t, tea.ExecProcess(exec.Command("bash", "-c", command),nil)
			}
			
		// These keys should exit the program.
		case "ctrl+c", "q":
			TuiState = -1
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
				return t, tea.Cmd(getGitRepo)
			}
			if t.mode == "git offline" {
				return t, tea.Cmd(func() tea.Msg {
					cmd, _ := exec.Command("bash", "-c", "ls /home/simonheise/git_repos/").Output()
					// convert result byte to string and split at newline
					result := string(cmd)
					result_split := strings.Fields(result)
					return updateMsg {
						listitems: result_split,
						selected:  make(map[int]string),
						cursor: 0,
						footer: "\nAvailable functions\n" +
							"- p     : push selected repos\n" +
							"- m     : return to mode selection\n" +
							"- q     : quit",
					}})
				// return t, tea.ExecProcess(exec.Command("nvim", "/home/simonheise/test.sh"),nil)
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

// run command
func runCommand() {
	command := ""

	for idx := range len(cmdArray) {
		if cmdArray[idx] != "" {
			command += cmdArray[idx] + "; "
		} else {
			break
		}
	}

	cmd := exec.Command("bash", "-c", command)

	// get a pipe to read from standard output
	stdout, _ := cmd.StdoutPipe()

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan struct{})

	// Create a scanner which scans stdout in a line-by-line fashion
	cmd_scanner := bufio.NewScanner(stdout)

	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {
		// Read line by line and process it
		for cmd_scanner.Scan() {
			line := cmd_scanner.Text()
			fmt.Println(line)
		}
		// We're all done, unblock the channel
		done <- struct{}{}
	}()

	cmd.Run()

	// Wait for all output to be processed
	<-done

	// remove previous command array and reinitialize that array
	cmdArray = [100]string{}
}

// parseconfigfile: read config file from users configdir and parse settings
func parseconfigfile() map[string]string {

	configs := make(map[string]string)

	configdir, err := os.UserConfigDir()

	if err != nil {
		fmt.Println("No configfile found\n")
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

// get entries of git user
func getGitRepo() tea.Msg {
	cmd, err := exec.Command("bash", "-c", "curl https://api.github.com/users/nomispaz/repos | grep full_name | cut -d':' -f 2 | cut -d'\"' -f 2 | cut -d'/' -f 2").Output()
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
		footer: "\nAvailable functions\n" +
			"- c     : clone selected repos\n" +
			"- m     : return to mode selection\n" +
			"- q     : quit",
		}
}

// ////////////////////////////////////////////////////////////////////////////////
//
// main function
func main() {

	m := initTui()
	m.footer += "- q     : quit\n"
	m.footer += "- Enter : execute selected item\n"

	TuiState = 1

	for {
		if TuiState == 1 {
			p := tea.NewProgram(&m)
			_, err := p.Run()
			if err != nil {
				fmt.Printf("Program ended unexpectedly due to: %v", err)
				os.Exit(1)
			}
			if TuiState == 0 {
				runCommand()
				TuiState = 1
			}
			if TuiState == -1 {
				break
			}
		}

	}
}
