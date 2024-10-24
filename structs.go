package main

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

type Field_type int

const (
	Show Field_type = iota + 1 // EnumIndex = 1
	Hide
)

// String - Creating common behavior - give the type a String function
func (f Field_type) String() string {
	return [...]string{"Show", "Hide"}[f-1]
}

// EnumIndex - Creating common behavior - give the type a EnumIndex function
func (f Field_type) EnumIndex() int {
	return int(f)
}

type Popup_type int

const (
	User Popup_type = iota + 1 // EnumIndex = 1
	Sudo
	//Password
)

// String - Creating common behavior - give the type a String function
func (p Popup_type) String() string {
	return [...]string{"User", "User", "Password"}[p-1]
}

// EnumIndex - Creating common behavior - give the type a EnumIndex function
func (p Popup_type) EnumIndex() int {
	return int(p)
}

// struct to save some variables that should be accessed from functions
type State struct {
	password   string
	user       string
	commit_msg string
	input      string
	command    string
	mode       Mode
}

// PatchConfig represents the structure of the configuration file
type Config struct {
	GitDir string `json:"GitDir"`
}
