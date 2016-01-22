package raft

//
// support for Raft and kvraft to save persistent
// Raft state (log &c) and k/v server snapshots.
//
// we will use the original persister.go to test your code for grading.
// so, while you can modify this code to help you debug, please
// test with the original before submitting.
//

import "sync"

type Persister struct {
	mu       sync.Mutex
	raftlog  []byte
	snapshot []byte
}

func MakePersister() *Persister {
	return &Persister{}
}

func (ps *Persister) Copy() *Persister {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	np := MakePersister()
	np.raftlog = ps.raftlog
	np.snapshot = ps.snapshot
	return np
}

func (ps *Persister) SaveRaftLog(data []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.raftlog = data
}

func (ps *Persister) ReadRaftLog() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.raftlog
}

func (ps *Persister) RaftLogSize() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return len(ps.raftlog)
}

func (ps *Persister) SaveSnapshot(snapshot []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.snapshot = snapshot
}

func (ps *Persister) ReadSnapshot() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.snapshot
}
