package raft

import "log"
import "sync/atomic"

// Debugging
const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

//
// tell the peer to shut itself down.
// for testing.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
}

//
// has this peer been asked to shut down?
//
func (rf *Raft) isdead() bool {
	return atomic.LoadInt32(&rf.dead) != 0
}
