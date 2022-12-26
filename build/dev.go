package build

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/shynxe/greact/config"
)

var DevWebSocket *websocket.Conn
var timer *time.Timer

const refreshDebounce = time.Millisecond * 200

// Dev is the main function of the dev command
func Dev(args []string) {
	args = append(args, "-dev")
	Build(args)

	go initDevSocket()
	go watchClient(config.StaticPath, handleRefreshClient)
	go watchClient(config.SourcePath, handleBuildClient)
	watchServer()
}

func handleRefreshClient(e fsnotify.Event) {
	if timer == nil {
		timer = time.AfterFunc(refreshDebounce, func() {
			if err := DevWebSocket.WriteMessage(websocket.TextMessage, []byte("refresh")); err != nil {
				fmt.Println("Error sending refresh to websocket:", err)
			}
			timer = nil
		})
	} else {
		timer.Reset(refreshDebounce)
	}
}

func handleBuildClient(e fsnotify.Event) {
	build()
}

func initDevSocket() {
	var upgrader = websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		var err error
		DevWebSocket, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading websocket:", err)
			http.Error(w, "Could not upgrade websocket connection", http.StatusInternalServerError)
			return
		}

		for {
			_, _, err := DevWebSocket.ReadMessage()
			if err != nil {
				break
			}
		}

		defer DevWebSocket.Close()
	})

	http.ListenAndServe(":1501", nil)
}

// Watch the client source directory for changes
func watchClient(path string, onChange func(e fsnotify.Event)) {
	clientWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer clientWatcher.Close()

	err = clientWatcher.Add(path)
	if err != nil {
		log.Fatal(err)
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
			// Kill the previous command
			if cmd != nil {
				if err := cmd.Process.Kill(); err != nil {
					log.Println("Error killing command: ", err)
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

const refreshScript = `<script type="text/javascript">
const socket = new WebSocket('ws://localhost:1501/ws');
  socket.onmessage = function (event) {
    console.log('Received data from server:', event.data);
    if (event.data === 'refresh') {
      window.location.reload();
    }
  }</script>`
