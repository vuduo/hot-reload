package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const webpackConfigFile = "webpack.config.js"

func main() {

	// define the initial directory that we are going to search through
	// as well as a command to run after linking
	var directory, command string
	flag.StringVar(&directory, "directory", "", "the path of the webpack directory")
	flag.StringVar(&command, "command", "", "command to run after node modules are linked")
	flag.Parse()

	if directory == "" {
		directory = os.Getenv("DIRECTORY")
	}

	if directory == "" {
		directory = findWebpackDirectory("/app")
	} else {
		directory = "/app/" + directory
	}

	if directory == "" {
		fmt.Println("No webpack directory found.\nPlease specify the directory using -e \"DIRECTORY=web\"")
		return
	}

	if command == "" {
		command = os.Getenv("COMMAND")
	}

	if command == "" {
		fmt.Println("No command specified that should be run.\nPlease specify using -e \"COMMAND=npm run dev\"")
	}

	// print the information that was parsed from the flags
	fmt.Printf("DIRECTORY:  %s\n", directory)
	fmt.Printf("COMMAND:    %s\n", command)

	// link the global node modules into the local directory
	err := symlinkGlobalNodeModules(directory)
	if err != nil {
		fmt.Println("could not symlink the global node modules:", err)
		return
	}

	// start webpack with the command supplied
	err = runCommand(directory, command)
	if err != nil {
		fmt.Println("could run the command supplied:", err)
		return
	}

}

// findWebpackDirectory will try to find the webpack directory by looking for
// a webpack.config.js file.
func findWebpackDirectory(searchDirectory string) string {

	var webpackDirectory string

	// find the directory that contains a webpack config file
	err := filepath.Walk(searchDirectory, func(filePath string, f os.FileInfo, err error) error {

		// try to find the webpack directory file
		if strings.Contains(filePath, webpackConfigFile) {
			webpackDirectory = path.Dir(filePath)
			return filepath.SkipDir
		}

		excludeDirs := []string{"/vendor", "/node_modules", ".git", ".svn"}
		if containsAny(filePath, excludeDirs) {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}

	return webpackDirectory
}

// containsAny will check whether any of the matches is part of the given
// string.
func containsAny(source string, matches []string) bool {

	for _, element := range matches {
		if strings.Contains(source, element) == true {
			return true
		}
	}

	return false
}

// symlinkGlobalNodeModules will link the globally installed node modules in the
// container into the node_modules folder of the project.
func symlinkGlobalNodeModules(directory string) error {

	// define the path the node_modules directory
	nodeModules := directory + "/node_modules"

	// check if a node directory exists in the same directory
	// or create if necessary
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		os.Mkdir(nodeModules, 0777)
	}

	// symlink the global node modules into the directory
	symlink := exec.Command("ln", "-s", "-f", "/usr/local/lib/node_modules/*", nodeModules)

	// redirect all output to the standard console
	symlink.Stdout = os.Stdout
	symlink.Stderr = os.Stderr

	err := symlink.Run()
	return err

}

// runCommand will try to start a webpack dev server using the command
// supplied by the user.
func runCommand(directory string, command string) error {

	// split the command into separate entries
	items := strings.Split(command, " ")

	var webpack *exec.Cmd

	switch len(items) {
	case 0:
		return errors.New("No command specified")

	case 1:
		// start the webpack dev server using an npm run command
		webpack = exec.Command(items[0])

	default:
		// start the webpack dev server using an npm run command
		webpack = exec.Command(items[0], items[1:]...)
	}

	// set the current directory to the webpack directory
	os.Chdir(directory)

	// redirect all output to the standard console
	webpack.Stdout = os.Stdout
	webpack.Stderr = os.Stderr

	// try to run the webpack service
	err := webpack.Run()
	return err

}
