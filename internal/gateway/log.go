package gateway

import (
	"io/ioutil"
	"log"
	"os"
)

var TimestampedLog = log.New(os.Stdout, "", log.LstdFlags)
var SimpleLog = log.New(os.Stdout, "", 0)
var NoLog = log.New(ioutil.Discard, "", 0)
