package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

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

func read_config(configfilepath string) Config {

	// expand $HOME or ~ in filepath
	current_user, err := user.Current()
	if err != nil {
		fmt.Printf(Red + "Failed to read user." + Reset)
		panic(err)
	}
	homedir := current_user.HomeDir

	if strings.HasPrefix(configfilepath, "~/") {
		configfilepath = filepath.Join(homedir, configfilepath[2:])
	}
	if strings.HasPrefix(configfilepath, "$HOME") {
		configfilepath = filepath.Join(homedir, configfilepath[5:])
	}

	// read json file to string
	contents, err := os.ReadFile(configfilepath)

	if err != nil {
		// config file doesn't exist --> create default config
		create_path, _ := filepath.Split(configfilepath)
		err = os.MkdirAll(create_path, 0766)
		err = os.WriteFile(configfilepath, []byte("{\n\"GitDir\": \"~/git_repos/\"\n}"), 0644)

		contents, err = os.ReadFile(configfilepath)

		if err != nil {
			fmt.Printf("Couldn't create config file:")
			panic(err)
		}
	}

	// initialize the config map
	var config Config

	// Unmarshal the JSON into the Config struct
	err = json.Unmarshal(contents, &config)

	if err != nil {
		fmt.Println(Red + "Error unmarshalling JSON." + Reset)
		panic(err)
	}

	return config

}
