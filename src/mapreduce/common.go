package mapreduce

import (
	"fmt"
	"strconv"
)

// Debugging enabled?
const debugEnabled = false

// DPrintf will only print if the debugEnabled const has been set to true
func debug(format string, a ...interface{}) (n int, err error) {
	if debugEnabled {
		n, err = fmt.Printf(format, a...)
	}
	return
}

// jobPhase indicates whether a job is scheduled as a map or reduce task.
type jobPhase string

const (
	mapPhase    jobPhase = "Map"
	reducePhase          = "Reduce"
)

// KeyValue is a type used to hold the key/value pairs passed to the map and
// reduce functions.
type KeyValue struct {
	Key   string
	Value string
}

// reduceName constructs the name of the intermediate file which map job
// <MapJob> produces for reduce job <ReduceJob>.
func reduceName(jobName string, MapJob int, ReduceJob int) string {
	return "mrtmp." + jobName + "-" + strconv.Itoa(MapJob) + "-" + strconv.Itoa(ReduceJob)
}

// mergeName constructs the name of the output file of reduce job <ReduceJob>
func mergeName(jobName string, ReduceJob int) string {
	return "mrtmp." + jobName + "-res-" + strconv.Itoa(ReduceJob)
}
