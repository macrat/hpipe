hpipe
=====

A simple tool for piping TCP connection over HTTP.

This program can use WebSocket or pure TCP connection using [Upgrade header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Upgrade).
The pure TCP has smaller overhead than WebSocket, but it does not work well in some environments.
The WebSocket is more compatible to most environments such as behind of a fire-wall or proxies.

## Usage
### server: HTTP -> TCP

``` bash
$ hpipe -l :8080 example.com:22
```

```
+----------+        +----------------+       +------------------+
|          |  HTTP  |                |  TCP  |                  |
|  client ----------->    hpipe     ---------->     server      |
|          |        | (0.0.0.0:8080) |       | (example.com:22) |
+----------+        +----------------+       +------------------+
```

### client: TCP -> HTTP

``` bash
$ hpipe -l :1234 http://example.com:8080  # pure TCP mode
$ hpipe -l :1234 ws://example.com:8080    # WebSocket mode
```

```
+----------+       +----------------+        +--------------------+
|          |  TCP  |                |  HTTP  |                    |
|  client ---------->    hpipe     ----------->       hpipe       |
|          |       | (0.0.0.0:1234) |        | (example.com:8080) |
+----------+       +----------------+        +--------------------+
```

### client: stdio -> HTTP

``` bash
$ hpipe http://example.com:8080  # pure TCP mode
$ hpipe ws://example.com:8080    # WebSocket mode
```

```
          +-----------------+        +--------------------+
stdin ----->                |  HTTP  |                    |
          |      hpipe     ----------->       hpipe       |
stdout <----                |        | (example.com:8080) |
          +-----------------+        +--------------------+
```


## Connect SSH over HTTP proxy

__NOTE__: This function is still work in progress.

```
                            HTTP Proxy
client                          +-+      server
+------------------------+      | |      +-------------------------+
|                        |      | |      |                         |
| +----------+   +-----+ |      | |      |  +-----+   +----------+ |
| |SSH client|-->|hpipe|--------| |-------->|hpipe|-->|SSH server| |
| +----------+   +-----+ |      | |      |  +-----+   +----------+ |
|                        |      | |      |                         |
+------------------------+      | |      +-------------------------+
                                +-+
```

### Server

Start Hpipe on your server.

``` shell
$ hpipe -l :8022 localhost:22
```

### Client

Setup ssh-config and HTTP proxy address.

``` shell
$ cat ~/.ssh/config
Host your-server.example.com
	Port 8022
	ProxyCommand hpipe http://%h:%p
	#ProxyCommand hpipe ws://%h:%p

$ echo 'export http_proxy=http://your-proxy.example.com' >> ~/.bashrc
```

Connect SSH.

``` shell
$ ssh your-server.example.com
```
