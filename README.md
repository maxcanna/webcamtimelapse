# WebCamTimeLapse ![](https://img.shields.io/codeship/ebf02f60-54f2-0134-5aad-4e423e130982/master.svg?maxAge=2592000) [![](https://img.shields.io/github/license/maxcanna/webcamtimelapse.svg?maxAge=2592000)](https://github.com/maxcanna/webcamtimelapse/blob/master/LICENSE)

Easily create time-lapse videos from webcam images.

### How do I get set up?

#### Build
`$ ./build.sh`

#### Run

`$ java -jar WebCamTimeLapse.jar`

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

### Source image examples:

* [http://www.hotel-galles-milano-webcam.it/hotel-milano-corso-buenos-aires.jpg](http://www.hotel-galles-milano-webcam.it/hotel-milano-corso-buenos-aires.jpg)
* [http://www.fashionschool.com/fashion_school.jpg](http://www.fashionschool.com/fashion_school.jpg)
* [http://wirednewyork.com/webcam/new-york-live.jpg](http://wirednewyork.com/webcam/new-york-live.jpg)

### Credits

Uses work from [Werner Randelshofer](http://www.randelshofer.ch/)
