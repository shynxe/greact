package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/shynxe/greact/config"
)

// Dev is the main function of the dev command
func Dev(args []string) {
	// 1. build pages
	// 2. run client
	Build(args)

	go watchClient()
	watchServer()
}

// Watch the client source directory for changes
// Set the client build to be executed when a change is detected
// Run the event loop to watch for changes
func watchClient() {
	clientWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer clientWatcher.Close()

	err = clientWatcher.Add(config.SourcePath)
	if err != nil {
		log.Fatal(err)
	}

	onChange := func(e fsnotify.Event) {
		log.Println("Detected a client change: ", e.Name)
		build()
	}

	for {
		select {
		case event := <-clientWatcher.Events:
			onChange(event)
		case err := <-clientWatcher.Errors:
			log.Println("error:", err)
		}
	}
}

// Watch the current directory (only the ".go" files)
// Set the "go run ." command to be executed when a change is detected
// Run the event loop to watch for changes
// When a change is detected, the server is restarted (the previous command is killed)
func watchServer() {
	serverWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer serverWatcher.Close()

	err = serverWatcher.Add(".")
	if err != nil {
		log.Fatal(err)
	}

	// Create a channel to receive signals and register it to receive SIGINT signals
	sigchan := make(chan os.Signal, 1)

	var cmd *exec.Cmd
	shouldReturn := buildApp(cmd)
	if shouldReturn {
		return
	}

	signal.Notify(sigchan, syscall.SIGINT)

	cmd, shouldReturn1 := runApp(cmd)
	if shouldReturn1 {
		return
	}

	// Start a goroutine to listen for SIGINT signals and terminate the cmd process when one is received
	go killApp(cmd, sigchan)

	onChange := func(e fsnotify.Event) {
		fileExt := e.Name[len(e.Name)-3:]
		if fileExt == ".go" {
			log.Println("Detected a server change: ", e.Name)

			// Kill the previous command
			if cmd != nil {
				if err := cmd.Process.Kill(); err != nil {
					log.Println("Error killing command: ", err)
				} else {
					log.Println("Killed command: ", cmd)
				}
			}

			// Build "./app"
			cmd = exec.Command("go", "build", "-o", "app")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Println("Error building command: ", err)
				return
			}

			signal.Notify(sigchan, syscall.SIGINT)

			// Run "./app"
			cmd = exec.Command("./app")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				log.Println("Error starting command: ", err)
				return
			}

			// Start a goroutine to listen for SIGINT signals and terminate the cmd process when one is received
			go func() {
				sig := <-sigchan
				fmt.Println("[greact] hope you developed something awesome! (received signal: ", sig, ")")
				cmd.Process.Kill()
				os.Exit(0)
			}()

			defer os.Remove("app")
		}
	}

	for {
		select {
		case event := <-serverWatcher.Events:
			onChange(event)
		case err := <-serverWatcher.Errors:
			log.Println("error:", err)
		}
	}
}

func killApp(cmd *exec.Cmd, sigchan chan os.Signal) {
	sig := <-sigchan
	fmt.Println("[greact] hope you developed something awesome! (received signal: ", sig, ")")
	cmd.Process.Kill()
	os.Remove("app")
	os.Exit(0)
}

func runApp(cmd *exec.Cmd) (*exec.Cmd, bool) {
	cmd = exec.Command("./app")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Println("Error starting command: ", err)
		return nil, true
	}
	return cmd, false
}

func buildApp(cmd *exec.Cmd) bool {
	cmd = exec.Command("go", "build", "-o", "app")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println("Error building command: ", err)
		return true
	}
	return false
}
