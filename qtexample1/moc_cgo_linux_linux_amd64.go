// +build !ubports

package main

/*
#cgo CFLAGS: -pipe -O2 -Wall -W -D_REENTRANT -fPIC -DQT_NO_DEBUG -DQT_WIDGETS_LIB -DQT_GUI_LIB -DQT_QML_LIB -DQT_NETWORK_LIB -DQT_CORE_LIB
#cgo CXXFLAGS: -pipe -O2 -Wall -W -D_REENTRANT -fPIC -DQT_NO_DEBUG -DQT_WIDGETS_LIB -DQT_GUI_LIB -DQT_QML_LIB -DQT_NETWORK_LIB -DQT_CORE_LIB
#cgo CXXFLAGS: -I../../widgets -I. -I/home/rob/Qt/5.13.1/gcc_64/include -I/home/rob/Qt/5.13.1/gcc_64/include/QtWidgets -I/home/rob/Qt/5.13.1/gcc_64/include/QtGui -I/home/rob/Qt/5.13.1/gcc_64/include/QtQml -I/home/rob/Qt/5.13.1/gcc_64/include/QtNetwork -I/home/rob/Qt/5.13.1/gcc_64/include/QtCore -I. -isystem /usr/include/libdrm -I/home/rob/Qt/5.13.1/gcc_64/mkspecs/linux-g++
#cgo LDFLAGS: -O1 -Wl,-rpath,/home/rob/Qt/5.13.1/gcc_64/lib
#cgo LDFLAGS:  /home/rob/Qt/5.13.1/gcc_64/lib/libQt5Widgets.so /home/rob/Qt/5.13.1/gcc_64/lib/libQt5Gui.so /home/rob/Qt/5.13.1/gcc_64/lib/libQt5Qml.so /home/rob/Qt/5.13.1/gcc_64/lib/libQt5Network.so /home/rob/Qt/5.13.1/gcc_64/lib/libQt5Core.so -lGL -lpthread
#cgo CFLAGS: -Wno-unused-parameter -Wno-unused-variable -Wno-return-type
#cgo CXXFLAGS: -Wno-unused-parameter -Wno-unused-variable -Wno-return-type
*/
import "C"
