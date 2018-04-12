# Spdyn-Updater

`go get github.com/pascaldierich/spdyn-updater`

The spdyn-updater is inspired by the spdns Dynamic DNS Update-Client from <http://my5cent.spdns.de/de/posts/spdns-dynamic-dns-update-client>.

The spdyn-updater is written in Go with no external dependencys.

### Usage

Write a `host.json` file with your server configurations which looks like this:
```javascript
{  
   "updateHost":"update.spdyn.de",
   "host":"your.host.com",
   "user":"user@example.com",
   "password":"password",
   "isToken":false
}
```

Change the spdyn-directory if neccessary with the `-d` flag, default is `${HOME}/.spdyn`.

The spdyn-updater supports IPv4 and IPv6.

### Setup for Raspberry Pi

1. Cross-compile the code with 
`env GOOS=linux GOARCH=arm go build` 

2. Run `scp spdyn-updater user@raspberry-pi:` to copy the runnable to user's home directory.

3. Login to the RasPi.

4. Create a folder `.spdyn` in your home directory and `mv  spdyn-updater .spdyn/`

5. Write the `host.json` file.
Either in `.spdyn/` or redefine the path with the `-h` flag.

6. Add CronJob (_as root_) to `/etc/crontab`.<br>
To check every 10 min, use:<br>
`*/10 * * * * user /home/user/.spdyn/spdyn-updater`

