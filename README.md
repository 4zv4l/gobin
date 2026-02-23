# Gobin

A fast, zero-dependency [termbin](https://termbin.com/) written in Go.  
It allows you to quickly create text pastes directly from your command line using netcat (nc) and serves them via a built-in HTTP/HTTPS web server.  

## Installation

Because Gobin uses exactly zero external dependencies, building it is as simple as: `go build`.  

## Usage

You can use `-h` or `--help` to view the configuration flags:
```
Usage of gobin:
  -address string
        bind to this address (default "127.0.0.1")
  -cert-path string
        certificate path for tls
  -directory string
        directory to save/serve the pastes (default "/tmp")
  -gc
        delete old paste if the pool is full
  -loglevel string
        log message up to that level (default "INFO")
  -max-dir-size int
        max directory size allowed in byte (default 104857600)
  -max-file-size int
        max file size allowed in byte (default 10485760)
  -pkey-path string
        private key path for tls
  -randlen int
        IDs length (default 4)
  -tcp-port int
        bind to this port (tcp server) (default 9999)
  -timeout int
        timeout in second to receive a paste (default 1)
  -url string
        uses this url when generating links
  -web-port int
        bind to this port (web server) (default 4433)
```

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
