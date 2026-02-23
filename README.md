# Gobin

A fast, zero-dependency [termbin](https://termbin.com/) written in Go.  
It allows you to quickly create text pastes directly from your command line using netcat (nc) and serves them via a built-in HTTP/HTTPS web server.  

## Installation

Because Gobin uses exactly zero external dependencies, building it is as simple as: `go build`.  

## Usage

You can use `-h` or `--help` to view the configuration flags.  
Note that `-directory` is the only mandatory flag.  

### Log Levels: DEBUG, INFO, WARN, ERROR.

> Tip: Gobin logs straight to stdout.  
> To save logs to a file while watching them live, you can use tee: ./gobin -directory ./tmp | tee server.log  

## Examples

Start the Gobin server, specifying a directory for storage: `./gobin -directory ./tmp`  

Create a paste from another terminal using nc: `echo just testing! | nc 127.0.0.1 9999`  

Access the paste via curl or any web browser: `curl http://127.0.0.1:4433/test`  

You can stop the server at any time by pressing Ctrl-C.  
The server will gracefully shut down its active listeners before exiting.  

### Custom IDs
You can use custom IDs for pastes by simply pre-creating files in the `-directory`.  
During startup, Gobin scans the directory, registers existing filenames, and avoids generating collisions.  
If `-gc` is enabled, the garbage collector will eventually recycle these filenames when space runs out.  
