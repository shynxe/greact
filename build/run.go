package build

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Run is the main function of the run command
func Run(args []string) {
	Build(args)
	run()
}

// run runs the built application
func run() {
	cmd := exec.Command("go", "build", "-o", "app")
	output, err := cmd.Output()
	if err != nil {
		printError(err)
	}
	fmt.Println(string(output))

	cmd = exec.Command("./app")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		printError(err)
	}
	if err := cmd.Start(); err != nil {
		printError(err)
	}

	// Create a channel to receive signals and register it to receive SIGKILL signals
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT)

	// Start a goroutine to listen for SIGKILL signals and terminate the cmd process when one is received
	go func() {
		sig := <-sigchan
		if sig == syscall.SIGINT {
			fmt.Println("[greact] hope you enjoyed the app!")
		}
		cmd.Process.Kill()
	}()

	if _, err := io.Copy(os.Stdout, stdout); err != nil {
		printError(err)
	}
	if err := cmd.Wait(); err != nil {
		printError(err)
	}

	// Remove the app file when the function returns, even if it is terminated by a SIGKILL signal
	defer os.Remove("app")
}

func printError(err error) {
	fmt.Println("[greact] error:", err)
	os.Exit(1)
}
