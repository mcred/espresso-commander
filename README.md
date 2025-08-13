# Espresso Commander
Demo application to get system information and ping a host through an HTTP server.

## Requirements
* go: 1.23 +
* GNU make

## Usage
The application supports runs a localhost HTTP server on port 8080 and supports the two following request:
* ping
* sysinfo

### ping
Determine how long it takes for a remote host to respond. `type` is a required string and should be `ping`. `payload` is a required string and should be a valid host.  

Sample Request:
```shell
curl -X POST http://localhost:8080/execute -d '{"type":"ping", "payload":"www.google.com"}'
```
Sample Response:
```json
{
  "success": true,
  "data": 34919000
}
```

### sysinfo
Reports basic information about the host system. `type` is a required string and should be `sysinfo`. `payload` is not required and will be ignored if provided. 

Sample Request:
```shell
 curl -X POST http://localhost:8080/execute -d '{"type":"sysinfo", "payload":"www.google.com"}'
```
```json
{
  "success": true,
  "data": {
    "Hostname": "Mac",
    "IPAddress": "172.31.10.59"
  }
}
```

## Getting Started
There are two main ways to run this application: directly as a compiled binary or installed system executable. The following steps assume you are using MacOS. If you are using windows, only `make run` should work.  

### Running
```shell
make run
```
Builds the application and runs attached to the terminal. The application logs will output to the terminal directly. Use `Ctrl+C` to stop this process.

### Installing
```shell
make install
```
Builds the application and copies the binary to /usr/local/bin, adds the plist to /Library/LaunchDaemons, and starts the service. The output from this command includes API location, log locations, and commands for starting and stopping the service:
```
Service Information:
  - Binary: /usr/local/bin/espresso-commander
  - Service: io.mcred.espresso-commander
  - Logs: /var/log/espresso-commander.log
  - Errors: /var/log/espresso-commander.error.log
  - API: http://localhost:8080/execute

Commands:
  - Check status: sudo launchctl list | grep io.mcred
  - View logs: tail -f /var/log/espresso-commander.log
  - Stop service: sudo launchctl unload /Library/LaunchDaemons/io.mcred.espresso-commander.plist
  - Start service: sudo launchctl load /Library/LaunchDaemons/io.mcred.espresso-commander.plist
  - Uninstall: sudo ./installer/uninstall.sh
```

### Uninstalling
```shell
make uninstall
```
Stops the running service, removes the LaunchDaemon, binary, and logs. 

## Releasing
It is also possible to install the service from a macOS package installer (.pkg) release packing. To create the installer package:
```shell
make package
```
Then go to `./dist` in Finder, run the `EspressoCommander.pkg`, and follow the installer instructions. See [Demo](#demo) for example. 

## Testing
```shell
make test
```
This runs the included unit tests for main.go and commander.go.

## Demo

https://github.com/user-attachments/assets/ce8b0b6d-22a2-4be8-a1f2-e35762a3650b

