// Spdyn-Updater is an update client for the spDYN Dynamic-DNS-Service.
//
// Create a folder and specify the location with the ``-d`` flag,
// default is ``${HOME}/.spdyn`` and specify a ``host.json`` file
// with following fields.
//
//		{ updateHost, host, user, password, isToken }
//
// Setup for Raspberry Pi
//
// Cross compile the code with
//		env GOOS=linux GOENV=arch go build
//
// Move the binary into the ``.spdyn/`` folder in the RasPi's
// home directory and add the ``host.json``.
//
// Add a CronJob, as root to ``/etc/crontab``,
// to check every 10 min, use:
//		*/10 * * * * <user> /home/<user>/.spdyn/spdyn-updater
//
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
	"unicode"
)

const (
	dnsHost    string = "http://checkip.spdyn.de"
	logFmt     string = "\t%s\n"
	logTimeRFC string = time.RFC3339
	spdynDir   string = "/.spdyn/"
	hostName   string = "host.json"
	ipName     string = "spdynuIP.cnf"
	logName    string = "spdynu.log"
)

var lastKnownIP net.IP
var removeCtrlChars = func(b []byte) []byte {
	for i, c := range b {
		if unicode.IsControl(rune(c)) {
			return b[:i]
		}
	}
	return b
}

var u = flag.String("u", "update.spdyn.de", "updateHost")
var d = flag.String("d", "", "spdynDirectory") // default is ${HOME}/.spdyn/

func main() {
	flag.Parse()

	// Get Flags
	updateTo := *u
	dir := *d
	if dir == "" {
		d, err := getHomeDir()
		if err != nil {
			panic(err)
		} else {
			dir = d + spdynDir
		}
	}
	hostPath := dir + hostName
	logPath := dir + logName
	ipPath := dir + ipName

	// Open logfile.
	logf, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer logf.Close()

	var logMsg string
	var log = func(msg string) {
		logMsg = fmt.Sprintf(logFmt, msg)
		logf.WriteString(logMsg)
	}
	var logStart = func() {
		t := time.Now().Format(logTimeRFC)
		logf.WriteString(t + "\n")
	}

	// Get old and current IP.
	if err = getStoredIP(ipPath); err != nil {
		logStart()
		log(err.Error())
		os.Exit(-1)
	}
	oldIP := lastKnownIP
	if err = updateIP(); err != nil {
		logStart()
		log(err.Error())
		os.Exit(-1)
	}

	// Check if update is neccessary.
	if lastKnownIP.Equal(oldIP) {
		return
	}
	logStart()

	// Open host file (.json).
	hostf, err := os.Open(hostPath)
	if err != nil {
		log(err.Error())
		os.Exit(-1)
	}
	defer hostf.Close()

	// Update each decoded host in own goroutine.
	ch := make(chan string)
	numHosts := 0
	decoder := json.NewDecoder(hostf)
	for {
		var host Host
		err := decoder.Decode(&host)
		if err == io.EOF {
			break
		}
		if err != nil {
			log("Check your host.json file: " + err.Error())
			break
		}

		numHosts++
		go host.Update(updateTo, lastKnownIP.String(), ch)
	}

	// Receive hosts' messages.
	for i := 0; i < numHosts; {
		select {
		case msg := <-ch:
			log(msg)
			i++
		default:
			continue
		}
	}

	// Store current IP.
	err = storeIP(ipPath)
	if err != nil {
		log(err.Error())
		os.Exit(-1)
	}
}

// storeIP writes field `lastKnownIP` to path `p`.
func storeIP(p string) error {
	err := os.Remove(p)
	if err != nil {
		// we need to return here, otherwise os.Create truncates to existing file.
		return err
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(lastKnownIP.String())
	if err != nil {
		return err
	}
	return nil
}

// getStoredIP sets field `lastKnownIP` to stored IP, read from path `p`.
func getStoredIP(p string) error {
	ipFile, err := os.Open(p)
	if err != nil {
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		f.Close()
		return nil
	}
	defer ipFile.Close()

	b := make([]byte, 45)
	_, err = ipFile.Read(b)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	lastKnownIP = net.ParseIP(string(removeCtrlChars(b))) // if nil, update will take place.
	return nil
}

// updateIP resets field `lastKnownIP` to current IP, checked through dnsHost.
func updateIP() error {
	resp, err := http.Get(dnsHost)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	newIP := net.ParseIP(string(removeCtrlChars(body)))
	if newIP == nil {
		msg := fmt.Sprintf("Failed to parse IP: %v", body)
		return errors.New(msg)
	}
	lastKnownIP = newIP
	return nil
}

// getHomeDir return the path to the user's home directory.
func getHomeDir() (dir string, err error) {
	dir = os.Getenv("HOME")
	if dir == "" {
		err = errors.New("No HomeDir")
	}
	return
}
