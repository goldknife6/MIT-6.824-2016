package raftkv

import "labrpc"
import "fmt"
import "log"
import "raft"
import "sync"
import "bytes"
import "sync/atomic"
import "encoding/gob"


const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}


type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
}

type RaftKV struct {
	mu      sync.Mutex
	me      int
	dead    int32 // for testing
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	// Your definitions here.
}


func (kv *RaftKV) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
}

func (kv *RaftKV) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
}

// tell the server to shut itself down.
// please do not change these two functions.
func (kv *RaftKV) Kill() {
	atomic.StoreInt32(&kv.dead, 1)
	kv.rf.Kill()
}

// call this to find out if the server is dead.
func (kv *RaftKV) isdead() bool {
	return atomic.LoadInt32(&kv.dead) != 0
}

//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
//
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, oldpersister *raft.Persister, maxlogsize int) *RaftKV {
	// call gob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	gob.Register(Op{})

	kv := new(RaftKV)
	kv.me = me
	kv.maxlogsize = maxlogsize

	// Your initialization code here.

	if oldpersister != nil {
		fmt.Printf("%v restoring snapshot on startup\n", kv.me)
		kv.restoreSnapshot(oldpersister.ReadSnapshot())
	}

	kv.applyCh = make(chan raft.ApplyMsg)
	kv.rf = raft.Make(servers, me, persister, oldpersister, kv.applyCh)


	return kv
}
