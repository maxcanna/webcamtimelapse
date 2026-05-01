# WebCamTimeLapse [![CI Status](https://github.com/maxcanna/webcamtimelapse/workflows/CI/badge.svg)](https://github.com/maxcanna/webcamtimelapse/actions) [![](https://img.shields.io/github/license/maxcanna/webcamtimelapse.svg?maxAge=2592000)](https://github.com/maxcanna/webcamtimelapse/blob/master/LICENSE)

Easily create time-lapse videos from webcam images. This project has been fully rewritten in Go, dropping its old Java dependency and allowing native binaries for Windows, Linux, and macOS.

### How do I get set up?

#### Build

You need Go 1.26.2 installed on your machine.

**Build CLI:**
```bash
$ go build -o webcamtimelapse-cli ./cmd/cli
```

**Build GUI:**
```bash
$ go build -o webcamtimelapse-gui ./cmd/gui
```
Note: Building the GUI version requires CGO and system dependencies (`libgl1-mesa-dev`, `xorg-dev`, etc. on Linux).

#### Run

##### macOS Pre-requisites

If you downloaded the binaries on macOS, Gatekeeper might prevent them from running. You can remove the quarantine attribute and make them executable:

```bash
$ xattr -d com.apple.quarantine webcamtimelapse-cli
$ chmod +x webcamtimelapse-cli

$ xattr -d com.apple.quarantine webcamtimelapse-gui
$ chmod +x webcamtimelapse-gui
```

##### CLI

```bash
$ ./webcamtimelapse-cli
```

As shown in help there are several options available:
* `-url` (mandatory) Webcam image URL
* `-interval` Interval between each capture
* `-frames` Number of frames to capture before stopping
* `-filename` Output file name
* `-fps` FPS of generated video
* `-quality` Video quality (0.0 to 1.0)


##### GUI

```bash
$ ./webcamtimelapse-gui
```

### Source image examples:

* [http://ww3.comune.fe.it/webcam/piazza_trento_e_trieste.jpg](http://ww3.comune.fe.it/webcam/piazza_trento_e_trieste.jpg)
* [http://ww3.comune.fe.it/webcam/piazza_municipale.jpg](http://ww3.comune.fe.it/webcam/piazza_municipale.jpg)
