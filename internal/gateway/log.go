package gateway

import (
	"io/ioutil"
	"log"
	"os"
)

var SimpleLog = log.New(os.Stdout, "", 0)
var NoLog = log.New(ioutil.Discard, "", 0)
