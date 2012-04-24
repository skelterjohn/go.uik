package uik

import (
	"fmt"
	"github.com/skelterjohn/go.wde"
	"time"
)

var WindowGenerator func(parent wde.Window, width, height int) (window wde.Window, err error)

var StartTime = time.Now()

func TimeSinceStart() (dur time.Duration) {
	return time.Duration(time.Now().UnixNano() - StartTime.UnixNano())
}

func Report(args ...interface{}) {
	now := TimeSinceStart() / time.Millisecond
	ts := fmt.Sprintf("%07d", now)
	for _, arg := range args {
		ts += " " + fmt.Sprint(arg)
	}
	fmt.Println(ts)
}
