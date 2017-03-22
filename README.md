# cryptserver: SHA512, base64 encryption server

Server listening for POST requests containing a `password` parameter, returns SHA512 hash of password in base64 to `/` path. Server gracefully handles shutdown requests both from SIGINT signals and POST requests to `/shutdown` with authorized passwords.


## Installation

- Install [Go v1.8](https://golang.org/doc/install)
- `go get -u github.com/asgaines/cryptserver/...`

## Usage

### Starting Server

- `cryptserver`
- `cryptserver 2> /path/to/logfile`

#### Optional Parameters

- `-port`, or `--port` defines on which port server should listen (default: `8080`)
- `-delay`, or `--delay` defines how much time to wait before responding to requests (default: `5s`)
 - Note: must include unit (e.g. `3s`, `500ms`)

### Issuing Requests

#### Retrieving hashes of passwords
- `curl --data "password=<your-password>" http://localhost:<your-port>`

##### Example
Start server:

`cryptserver`

Issue POST request:

`curl --data "password=superSecretPassword1234" http://localhost:8080`

After default response time wait, receive:

`+PNX5nhK38mR3VJEWOMBweFrgocThl86dbM423y0DOtIsa3xBrUObVwsBnArhv5gFy69uvImsWK74QQogGHeUQ==`

#### Graceful shutdown of server
- `curl --data "password=<valid-password>" http://localhost:<your-port>/shutdown`
- Passwords are validated by including a SHA512 hash of them in base 64 in [shadow file](etc/shadow)
 - For quick testing, `angryMonkey` is configured to work, out of the box. For security purposes, remove this line from the [shadow file](etc/shadow) and replace with new hashes when done testing
 - Create new valid password: `curl --data "password=<new-valid-password>" http://localhost:<your-port>/ >> ./etc/shadow`
- Invalid passwords return error

## Testing

`go test ./...`
