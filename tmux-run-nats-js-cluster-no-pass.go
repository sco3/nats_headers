///usr/bin/true; exec /usr/bin/env go run "$0" "$@"

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Function to write a NATS configuration file
func writeConfigFile(
	serverName string, filename string, store string, logFile string,
	clientPort int, clusterPort int, httpPort int, routes []string,
) error {

	configContent := fmt.Sprintf(`
	
	
jetstream: true
port: %d
server_name: %s

log_file: "%s"
debug: false
trace: false
logfile_size_limit: 1000MB
http: localhost:%d
http_port: %d


max_payload: 1100000

jetstream {
  max_memory_store: 20MB
  store_dir: "%s/jetstream"
}


`, clientPort, serverName, logFile,
		httpPort, httpPort, store,
	)
	return os.WriteFile(filename, []byte(configContent), 0644)

}

// Function to start a NATS server in a tmux session
func startNatsServer(sessionName, configFile string, logFile string) error {
	fmt.Printf("config: %s log: %s\n", configFile, logFile)
	cmd := exec.Command(
		"tmux", "new-session",
		"-d",
		"-e", "GOMEMLIMIT=2GiB",
		"-s", sessionName,
		"nats-server", "-js", "-c", configFile,
	)
	return cmd.Run()
}

func main() {
	var err error

	err = exec.Command("pgrep", "-x", "nats-server").Run()
	if err == nil {
		fmt.Println("NATS server is already running")
		os.Exit(0)
	}

	numServers := 1
	if len(os.Args) > 1 {
		numServers, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("Invalid number of servers")
			return
		}
	}
	dir := os.Getenv("HOME") + "/nats-logs"
	err = os.Mkdir(dir, 0755) // 0755 is the permission mode for the directory
	if err != nil {
		fmt.Printf("Directory: %v\n", err)
	}

	// Base ports for the servers
	baseClientPort := 4222
	baseClusterPort := 6222
	baseHttpPort := 8222

	for i := 0; i < numServers; i++ {
		clientPort := baseClientPort + i
		clusterPort := baseClusterPort + i
		httpPort := baseHttpPort + i
		serverName := fmt.Sprintf("nats%d", i)
		configFilename := fmt.Sprintf("%s/%s.conf", dir, serverName)
		storeFilename := fmt.Sprintf("%s/%s_data", dir, serverName)
		logFile := fmt.Sprintf("%s/%s.log", dir, serverName)

		// Generate routes for the cluster configuration
		var routes []string
		for j := 0; j < numServers; j++ {
			if j != i {
				routes = append(routes, fmt.Sprintf("nats-route://localhost:%d", baseClusterPort+j))
			}
		}

		if err := writeConfigFile(serverName, configFilename, storeFilename, logFile, clientPort, clusterPort, httpPort, routes); err != nil {
			fmt.Printf("Failed to write configuration file: %s\n", configFilename)
			continue
		}

		sessionName := fmt.Sprintf("nats%d", i)
		if err := startNatsServer(sessionName, configFilename, logFile); err != nil {
			fmt.Printf("Failed to start NATS server in tmux session: %s\n", sessionName)
			continue
		}
	}

	fmt.Printf("Started %d NATS servers in tmux sessions.\n", numServers)
}
