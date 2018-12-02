package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	ipsecConf = "/etc/ipsec.conf"
)

var statsURL string
var retryTimeout uint64
var saveToConf bool
var reload bool
var restart bool
var conn string
var v bool

type stats map[string]percent
type percent struct {
	Percent uint8
}

func checkFlags() {
	flag.StringVar(&statsURL, "statsURL", "https://nordvpn.com/api/server/stats", "Please, specify API url with percentage statistics. [string] (default: https://nordvpn.com/api/server/stats)")
	flag.Uint64Var(&retryTimeout, "retryTimeout", 5, "Please, specify timeout to retry to up connection in seconds. [uint64] (default: 3)")
	flag.BoolVar(&saveToConf, "saveToConf", true, "Please, specify saveToConf flag to save the fastest server to ipsec configuration file. [bool] (default: true)")
	flag.BoolVar(&reload, "reload", true, "Please, specify reload flag to reload ipsec settings. [bool] (default: true)")
	flag.BoolVar(&restart, "restart", false, "Please, specify restart flag to restart ipsec settings. [bool] (default: false)")
	flag.StringVar(&conn, "conn", "up", "Please, specify connection command (\"up\", \"down\" or \"nothing\"). [string] (default: up")
	flag.BoolVar(&v, "v", false, "Please, specify v flag to enable verbose mode. [bool] (default: false)")
	flag.Parse()

	if conn != "up" && conn != "down" && conn != "nothing" {
		fmt.Println("Please? specify \"up\", \"down\" or \"nothing\" for conn parameter")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {

	checkFlags()

	cfg, err := getFile(ipsecConf)
	check(err)
	lines := strings.Split(*cfg, "\n")
	connName := getStringAfterValue(lines, "conn", " ")
	fmt.Println("Your connection name is:", connName)

	switch conn {
	case "down":
		err := connect(connName, "down")
		check(err)
		os.Exit(1)
	case "up":
		err := connect(connName, "down")
		check(err)
	}

	var a stats
	err = getStats(statsURL, &a)
	check(err)

	fastestServer, load := getFastestServer(a)
	if fastestServer == nil {
		log.Fatalln("No result in the server's response")
	}
	fmt.Printf("The fastest server is: %v, load: %d\n", *fastestServer, load)

	oldServer := getStringAfterValue(lines, "nordvpn.com", "=")
	fmt.Println("Your old server is:", oldServer)

	replaceServer(lines, *fastestServer)

	if saveToConf {
		err = writeFile(lines, ipsecConf)
		check(err)
		fmt.Printf("Now your server is: %v\n", *fastestServer)
	}

	if reload {
		err = ipsec("reload")
		check(err)
	}

	if restart {
		err = ipsec("restart")
		check(err)
		time.Sleep(time.Duration(retryTimeout) * time.Second)
	}

	if conn == "up" {
		for err = connect(connName, "up"); err != nil; {
			fmt.Printf("Connection error: %v\nRetrying...\n", err)
			time.Sleep(time.Duration(retryTimeout) * time.Second)
			err = connect(connName, "up")
		}
		check(err)
	}

}

func getStats(url string, res *stats) error {

	client := &http.Client{}
	r, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to get data from server: %q", err)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(res)
}

func getFastestServer(s stats) (res *string, i uint8) {

	i = 100
	for server, v := range s {
		if v.Percent < i {
			i = v.Percent
			res = &server
		}
	}

	return
}

func getFile(fileName string) (*string, error) {

	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to get file content: %q", err)
	}

	content := string(fileContent)

	return &content, nil
}

func writeFile(lines []string, fileName string) (err error) {

	output := strings.Join(lines, "\n")

	err = ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write to file %s: %q", fileName, err)
	}

	return
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func getStringAfter(value string, a string) string {

	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}

	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}

	return value[adjustedPos:len(value)]
}

func ipsec(action string) error {

	cmd := exec.Command("ipsec", action)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ipsec reload failed with %s", err)
	}
	fmt.Printf("%s\n", string(out))

	return nil
}

func connect(connName string, action string) error {

	cmd := exec.Command("ipsec", action, connName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ipsec up %s failed with %q", connName, err)
	}
	if strings.Contains(string(out), "no config named") {
		return fmt.Errorf("config %s is not loaded yet", connName)
	}

	if v {
		fmt.Println(string(out))
	}

	fmt.Printf("Connect function successfully finihed action %q\n", action)

	return nil
}

func getStringAfterValue(lines []string, text string, after string) (result string) {

	for _, line := range lines {
		if strings.Contains(line, text) {
			result = getStringAfter(line, after)
		}
	}

	return
}

func replaceServer(lines []string, newServer string) {
	for i, line := range lines {
		if strings.Contains(line, ".nordvpn.com") {
			lines[i] = "  right=" + newServer
		}
	}
}
