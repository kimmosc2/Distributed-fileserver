package customLog

import "Distributed-fileserver/zaplogger"

var Logger = zaplogger.GetLoggerToFile("/data/logfile/download.log")