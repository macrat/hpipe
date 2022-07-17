hpipe
=====

A simple tool for piping TCP connection over HTTP.

This program uses [Upgrade header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Upgrade) of HTTP to make a connection, so there is no protocol overhead.

## Usage
### HTTP -> TCP

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

### TCP -> HTTP

``` bash
$ hpipe -l :1234 http://example.com:8080
```

```
+----------+       +----------------+        +--------------------+
|          |  TCP  |                |  HTTP  |                    |
|  client ---------->    hpipe     ----------->       hpipe       |
|          |       | (0.0.0.0:1234) |        | (example.com:8080) |
+----------+       +----------------+        +--------------------+
```

### stdio -> HTTP

``` bash
$ hpipe http://example.com:8080
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

$ echo 'export http_proxy=http://your-proxy.example.com' >> ~/.bashrc
```

Connect SSH.

``` shell
$ ssh your-server.example.com
```
