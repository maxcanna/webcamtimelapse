# WebCamTimeLapse [![CI Status](https://github.com/maxcanna/webcamtimelapse/workflows/CI/badge.svg)](https://github.com/maxcanna/webcamtimelapse/actions) [![](https://img.shields.io/github/license/maxcanna/webcamtimelapse.svg?maxAge=2592000)](https://github.com/maxcanna/webcamtimelapse/blob/master/LICENSE)

Easily create time-lapse videos from webcam images.

### How do I get set up?

#### Build
`$ ./build.sh`

#### Run

##### CLI

`$ java -jar WebCamTimeLapse-cli.jar`

As shown in help there are several options available:
* `URL` (mandatory) Webcam image URL
* `interval` Interval between each capture
* `frames` Number of frames to capture before stopping
* `filename` Output file name
* `fps` FPS of generated video
* `frameduration` Duration of each frame
* `height` Video height
* `width` Video width
* `quality` Video quality


##### GUI #####

`$ java -XstartOnFirstThread -jar WebCamTimeLapse-gui.jar`

![](https://cloud.githubusercontent.com/assets/1881831/20040332/63bf4d30-a455-11e6-9ed9-e817a451616d.png)

### Source image examples:

* [http://ww3.comune.fe.it/webcam/piazza_trento_e_trieste.jpg](http://ww3.comune.fe.it/webcam/piazza_trento_e_trieste.jpg)
* [http://ww3.comune.fe.it/webcam/piazza_municipale.jpg](http://ww3.comune.fe.it/webcam/piazza_municipale.jpg)

### Credits

Uses work from [Werner Randelshofer](http://www.randelshofer.ch/)
