package mapreduce

import "fmt"

// schedule starts and waits for all jobs in the given phase (Map or Reduce).
func (mr *Master) schedule(phase jobPhase) {
	var njobs int
	var nios int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		njobs = len(mr.files)
		nios = mr.nReduce
	case reducePhase:
		njobs = mr.nReduce
		nios = len(mr.files)
	}

	fmt.Printf("Schedule: %v %v jobs (%d I/Os)\n", njobs, phase, nios)

	// All njobs jobs have to be scheduled on workers, and only once all of
	// them have been completed successfully should the function return.
	// Remember that workers may fail, and that any given worker may finish
	// multiple jobs.
	//
	// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
	//
	fmt.Printf("Schedule: %v phase done\n", phase)
}
