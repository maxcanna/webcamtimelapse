#!/bin/bash

mkdir -p bin

javac -verbose -cp lib/swt.jar -d bin src/net/luxteam/webcamtimelapse/MainWindow.java src/net/luxteam/webcamtimelapse/WebCamTimeLapse.java src/ch/randelshofer/media/io/FilterImageOutputStream.java src/ch/randelshofer/media/quicktime/QuickTimeOutputStream.java src/ch/randelshofer/media/quicktime/DataAtomOutputStream.java

cd bin

jar cvfe ../WebCamTimeLapse-cli.jar net.luxteam.webcamtimelapse.WebCamTimeLapse .

jar cvfm ../WebCamTimeLapse-gui.jar ../MANIFEST.MF .

cd ../

rm -rf bin
