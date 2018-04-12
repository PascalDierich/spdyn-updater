package main

import (
	"fmt"
	"net/http"
)

// Host represents one instance of server to be updated.
type Host struct {
	UpdateHost string `json:"updateHost"`
	HostName   string `json:"host"`
	User       string `json:"user"`
	Pwd        string `json:"password"`
	IsToken    bool   `json:"isToken"`
}

const urlStrFmt string = "http://%s/nic/update?hostname=%s&myip=%s"

// Update sends the new IP to the updateHost.
func (h *Host) Update(to string, myIP string, ch chan<- string) {
	msg := fmt.Sprintf("[HOST: %s]\n", h.HostName)
	defer func() {
		ch <- msg
	}()

	urlStr := fmt.Sprintf(urlStrFmt, h.UpdateHost, h.HostName, myIP)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		msg += fmt.Sprintf("\t\tERROR: %v\n", err)
		return
	}

	req.Host = to
	req.SetBasicAuth(h.User, h.Pwd)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		msg += fmt.Sprintf("\t\tERROR: %v\n", err)
		return
	}

	msg += fmt.Sprintf("\t\tStatus Code: %d\n", resp.StatusCode)
}
