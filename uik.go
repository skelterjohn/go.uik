/*
   Copyright 2012 the go.wde authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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

const ReportIDs = false
