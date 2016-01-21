package mapreduce

import "fmt"
import "os"
import "log"
import "net/rpc"
import "net"

// Worker holds the state for a server waiting for DoJob or Shutdown RPCs
type Worker struct {
	name   string
	Map    func(string, string) []KeyValue
	Reduce func(string, []string) string
	nRPC   int
	nJobs  int
	l      net.Listener
}

// DoJob is called by the master when a new job is being scheduled on this
// worker.
func (wk *Worker) DoJob(arg *DoJobArgs, _ *struct{}) error {
	fmt.Printf("%s: given %v job #%d on file %s (nios: %d)\n",
		wk.name, arg.Phase, arg.JobNumber, arg.File, arg.NumOtherPhase)

	switch arg.Phase {
	case mapPhase:
		doMap(arg.JobName, arg.JobNumber, arg.File, arg.NumOtherPhase, wk.Map)
	case reducePhase:
		doReduce(arg.JobName, arg.JobNumber, arg.NumOtherPhase, wk.Reduce)
	}

	fmt.Printf("%s: %v job #%d done\n", wk.name, arg.Phase, arg.JobNumber)
	return nil
}

// Shutdown is called by the master when all work has been completed.
// We should respond with the number of jobs we have processed.
func (wk *Worker) Shutdown(_ *struct{}, res *ShutdownReply) error {
	debug("Shutdown %s\n", wk.name)
	res.Njobs = wk.nJobs
	wk.nRPC = 1 // OK, because the same thread reads nRPC
	wk.nJobs--  // Don't count the shutdown RPC
	return nil
}

// Tell the master we exist and ready to work
func (wk *Worker) register(master string) {
	args := new(RegisterArgs)
	args.Worker = wk.name
	ok := call(master, "Master.Register", args, new(struct{}))
	if ok == false {
		fmt.Printf("Register: RPC %s register error\n", master)
	}
}

// RunWorker sets up a connection with the master, registers its address, and
// waits for jobs to be scheduled.
func RunWorker(MasterAddress string, me string,
	MapFunc func(string, string) []KeyValue,
	ReduceFunc func(string, []string) string,
	nRPC int,
) {
	debug("RunWorker %s\n", me)
	wk := new(Worker)
	wk.name = me
	wk.Map = MapFunc
	wk.Reduce = ReduceFunc
	wk.nRPC = nRPC
	rpcs := rpc.NewServer()
	rpcs.Register(wk)
	os.Remove(me) // only needed for "unix"
	l, e := net.Listen("unix", me)
	if e != nil {
		log.Fatal("RunWorker: worker ", me, " error: ", e)
	}
	wk.l = l
	wk.register(MasterAddress)

	// DON'T MODIFY CODE BELOW
	for wk.nRPC != 0 {
		conn, err := wk.l.Accept()
		if err == nil {
			wk.nRPC--
			go rpcs.ServeConn(conn)
			wk.nJobs++
		} else {
			break
		}
	}
	wk.l.Close()
	debug("RunWorker %s exit\n", me)
}
