package raft

//
// Raft tests.
//
// we will use the original test_test.go to test your code for grading.
// so, while you can modify this code to help you debug, please
// test with the original before submitting.
//

import "testing"
import "fmt"
import "time"
import "math/rand"
import "sync/atomic"

func TestInitialElection(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: initial election ...\n")

	// is a leader elected?
	cfg.checkOneLeader()

	// does the leader+term stay the same there is no failure?
	term1 := cfg.checkTerms()
	time.Sleep(2 * time.Second)
	term2 := cfg.checkTerms()
	if term1 != term2 {
		t.Fatalf("term changed even though there were no failures")
	}

	fmt.Printf("  ... Passed\n")
}

func TestReElection(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: election after network failure ...\n")

	leader1 := cfg.checkOneLeader()

	// if the leader disconnects, a new one should be elected.
	cfg.disconnect(leader1)
	cfg.checkOneLeader()

	// if the old leader rejoins, that shouldn't
	// disturb the old leader.
	cfg.connect(leader1)
	leader2 := cfg.checkOneLeader()

	// if there's no quorum, no leader should
	// be elected.
	cfg.disconnect(leader2)
	cfg.disconnect((leader2 + 1) % servers)
	time.Sleep(2 * time.Second)
	cfg.checkNoLeader()

	// if a quorum arises, it should elect a leader.
	cfg.connect((leader2 + 1) % servers)
	cfg.checkOneLeader()

	// re-join of last node shouldn't prevent leader from existing.
	cfg.connect(leader2)
	cfg.checkOneLeader()

	fmt.Printf("  ... Passed\n")
}

func TestBasicAgree(t *testing.T) {
	servers := 5
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: basic agreement ...\n")

	iters := 3
	for index := 1; index < iters+1; index++ {
		nd, _ := cfg.nCommitted(index)
		if nd > 0 {
			t.Fatalf("some have committed before Start()")
		}

		cfg.one(index*100, servers)
	}

	fmt.Printf("  ... Passed\n")
}

//
// does Raft discard log entries in response to a call
// to Snapshotted()?
// does Raft still work after discarding some of log?
//
func TestSnapshot1(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: snapshot and log trim ...\n")

	n := 100
	for i := 0; i < n; i++ {
		cfg.one(i, servers)
	}

	before := make([]int, servers)
	for i := 0; i < servers; i++ {
		before[i] = cfg.saved[i].RaftLogSize() // bytes of persisted state
		if before[i] < n*8 || before[i] > n*40 {
			t.Fatalf("server %v uses an unreasonable %v bytes of persist storage",
				i, before[i])
		}
	}

	for i := 0; i < servers; i++ {
		cfg.rafts[i].Snapshotted(n/2 + i)
		cfg.one(1000+i, servers)
	}

	after := make([]int, servers)
	for i := 0; i < servers; i++ {
		after[i] = cfg.saved[i].RaftLogSize() // bytes of persisted state
		if after[i] > 9*(before[i]/10) {
			t.Fatalf("server %v persist state did not shrink enough: %v to %v\n",
				i, before[i], after[i])
		}
	}

	for i := 0; i < 3; i++ {
		cfg.one(20000+i, servers)
	}

	fmt.Printf("  ... Passed\n")
}

func TestFailAgree(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: agreement despite follower failure ...\n")

	cfg.one(101, servers)

	// follower network failure
	leader := cfg.checkOneLeader()
	cfg.disconnect((leader + 1) % servers)

	// agree despite one failed server?
	cfg.one(102, servers-1)
	cfg.one(103, servers-1)
	time.Sleep(1 * time.Second)
	cfg.one(104, servers-1)
	cfg.one(105, servers-1)

	// failed server re-connected
	cfg.connect((leader + 1) % servers)

	// agree with full set of servers?
	cfg.one(106, servers)
	time.Sleep(1 * time.Second)
	cfg.one(107, servers)

	fmt.Printf("  ... Passed\n")
}

func TestFailNoAgree(t *testing.T) {
	servers := 5
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: no agreement if too many followers fail ...\n")

	cfg.one(10, servers)

	// 3 of 5 followers disconnect
	leader := cfg.checkOneLeader()
	cfg.disconnect((leader + 1) % servers)
	cfg.disconnect((leader + 2) % servers)
	cfg.disconnect((leader + 3) % servers)

	index, _, ok := cfg.rafts[leader].Start(20)
	if ok != true {
		t.Fatalf("leader rejected Start()")
	}
	if index != 2 {
		t.Fatalf("expected index 2, got %v", index)
	}

	time.Sleep(2 * time.Second)

	n, _ := cfg.nCommitted(index)
	if n > 0 {
		t.Fatalf("%v committed but no majority", n)
	}

	// repair failures
	cfg.connect((leader + 1) % servers)
	cfg.connect((leader + 2) % servers)
	cfg.connect((leader + 3) % servers)

	// the disconnected majority may have chosen a leader from
	// among their own ranks, forgetting index 2.
	// or perhaps
	leader2 := cfg.checkOneLeader()
	index2, _, ok2 := cfg.rafts[leader2].Start(30)
	if ok2 == false {
		t.Fatalf("leader2 rejected Start()")
	}
	if index2 < 2 || index2 > 3 {
		t.Fatalf("unexpected index %v", index2)
	}

	cfg.one(1000, servers)

	fmt.Printf("  ... Passed\n")
}

func TestConcurrentStarts(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: concurrent Start()s ...\n")

	leader := cfg.checkOneLeader()
	iters := 5
	for ii := 0; ii < iters; ii++ {
		go func(i int) {
			_, _, ok := cfg.rafts[leader].Start(100 + i)
			if ok != true {
				t.Fatalf("Start() failed")
			}
		}(ii)
	}

	cmds := []int{}
	for index := 1; index <= iters; index++ {
		cmd := cfg.wait(index, servers)
		if ix, ok := cmd.(int); ok {
			cmds = append(cmds, ix)
		} else {
			t.Fatalf("value %v is not an int", cmd)
		}
	}
	for ii := 0; ii < iters; ii++ {
		x := 100 + ii
		ok := false
		for j := 0; j < len(cmds); j++ {
			if cmds[j] == x {
				ok = true
			}
		}
		if ok == false {
			t.Fatalf("cmd %v missing", x)
		}
	}

	fmt.Printf("  ... Passed\n")
}

func TestRejoin(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: rejoin of partitioned leader ...\n")

	cfg.one(101, servers)

	// leader network failure
	leader1 := cfg.checkOneLeader()
	cfg.disconnect(leader1)

	// make old leader try to agree on some entries
	cfg.rafts[leader1].Start(102)
	cfg.rafts[leader1].Start(103)
	cfg.rafts[leader1].Start(104)

	// new leader commits, also for index=2
	cfg.one(103, 2)

	// new leader network failure
	leader2 := cfg.checkOneLeader()
	cfg.disconnect(leader2)

	// old leader connected again
	cfg.connect(leader1)

	cfg.one(104, 2)

	// all together now
	cfg.connect(leader2)

	cfg.one(105, servers)

	fmt.Printf("  ... Passed\n")
}

func TestBackup(t *testing.T) {
	servers := 5
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: leader backs up quickly over incorrect follower logs ...\n")

	cfg.one(rand.Int(), servers)

	// put leader and one follower in a partition
	leader1 := cfg.checkOneLeader()
	cfg.disconnect((leader1 + 2) % servers)
	cfg.disconnect((leader1 + 3) % servers)
	cfg.disconnect((leader1 + 4) % servers)

	// submit lots of commands that won't commit
	for i := 0; i < 50; i++ {
		cfg.rafts[leader1].Start(rand.Int())
	}

	time.Sleep(500 * time.Millisecond)

	cfg.disconnect((leader1 + 0) % servers)
	cfg.disconnect((leader1 + 1) % servers)

	// allow other partition to recover
	cfg.connect((leader1 + 2) % servers)
	cfg.connect((leader1 + 3) % servers)
	cfg.connect((leader1 + 4) % servers)

	// lots of successful commands to new group.
	for i := 0; i < 50; i++ {
		cfg.one(rand.Int(), 3)
	}

	// now another partitioned leader and one follower
	leader2 := cfg.checkOneLeader()
	other := (leader1 + 2) % servers
	if leader2 == other {
		other = (leader2 + 1) % servers
	}
	cfg.disconnect(other)

	// lots more commands that won't commit
	for i := 0; i < 50; i++ {
		cfg.rafts[leader2].Start(rand.Int())
	}

	time.Sleep(500 * time.Millisecond)

	// bring original leader back to life,
	for i := 0; i < servers; i++ {
		cfg.disconnect(i)
	}
	cfg.connect((leader1 + 0) % servers)
	cfg.connect((leader1 + 1) % servers)
	cfg.connect(other)

	// lots of successful commands to new group.
	for i := 0; i < 50; i++ {
		cfg.one(rand.Int(), 3)
	}

	// now everyone
	for i := 0; i < servers; i++ {
		cfg.connect(i)
	}
	cfg.one(rand.Int(), servers)

	fmt.Printf("  ... Passed\n")
}

func TestCount(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: RPC counts aren't too high ...\n")

	leader := cfg.checkOneLeader()

	total1 := 0
	for j := 0; j < servers; j++ {
		total1 += cfg.rpcCount(j)
	}

	iters := 10
	cmds := []int{}
	for index := 1; index < iters+2; index++ {
		x := int(rand.Int31())
		cmds = append(cmds, x)
		index1, _, ok := cfg.rafts[leader].Start(x)
		if ok != true || index != index1 {
			t.Fatalf("Start() failed")
		}
	}

	for index := 1; index < iters+1; index++ {
		cmd := cfg.wait(index, servers)
		if ix, ok := cmd.(int); ok == false || ix != cmds[index-1] {
			t.Fatalf("wrong value %v committed for index %v\n", cmd, index)
		}
	}

	total2 := 0
	for j := 0; j < servers; j++ {
		total2 += cfg.rpcCount(j)
	}

	time.Sleep(1 * time.Second)

	total3 := 0
	for j := 0; j < servers; j++ {
		total3 += cfg.rpcCount(j)
	}

	if total1 > 30 || total1 < 1 {
		t.Fatalf("too many or few RPCs (%v) to elect initial leader\n", total1)
	}

	if total2-total1 > (iters+3)*3 {
		t.Fatalf("too many RPCs (%v) for %v entries\n", total2-total1, iters)
	}

	if total3-total2 > 3*20 {
		t.Fatalf("too many RPCs (%v) for 1 second of idleness\n", total3-total2)
	}

	fmt.Printf("  ... Passed\n")
}

func TestPersist1(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: basic persistence ...\n")

	cfg.one(11, servers)

	// crash and re-start all
	for i := 0; i < servers; i++ {
		cfg.start1(i)
	}
	for i := 0; i < servers; i++ {
		cfg.disconnect(i)
		cfg.connect(i)
	}

	cfg.one(12, servers)

	leader1 := cfg.checkOneLeader()
	cfg.disconnect(leader1)
	cfg.start1(leader1)
	cfg.connect(leader1)

	cfg.one(13, servers)

	leader2 := cfg.checkOneLeader()
	cfg.disconnect(leader2)
	cfg.one(14, servers-1)
	cfg.start1(leader2)
	cfg.connect(leader2)

	cfg.wait(4, servers) // wait for leader2 to join before killing i3

	i3 := (cfg.checkOneLeader() + 1) % servers
	cfg.disconnect(i3)
	cfg.one(15, servers-1)
	cfg.start1(i3)
	cfg.connect(i3)

	cfg.one(16, servers)

	fmt.Printf("  ... Passed\n")
}

func TestPersist2(t *testing.T) {
	servers := 5
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: more persistence ...\n")

	index := 1
	for iters := 0; iters < 5; iters++ {
		cfg.one(10+index, servers)
		index++

		leader1 := cfg.checkOneLeader()

		cfg.disconnect((leader1 + 1) % servers)
		cfg.disconnect((leader1 + 2) % servers)

		cfg.one(10+index, servers-2)
		index++

		cfg.disconnect((leader1 + 0) % servers)
		cfg.disconnect((leader1 + 3) % servers)
		cfg.disconnect((leader1 + 4) % servers)

		cfg.start1((leader1 + 1) % servers)
		cfg.start1((leader1 + 2) % servers)
		cfg.connect((leader1 + 1) % servers)
		cfg.connect((leader1 + 2) % servers)

		time.Sleep(1 * time.Second)

		cfg.start1((leader1 + 3) % servers)
		cfg.connect((leader1 + 3) % servers)

		cfg.one(10+index, servers-2)
		index++

		cfg.connect((leader1 + 4) % servers)
		cfg.connect((leader1 + 0) % servers)
	}

	cfg.one(1000, servers)

	fmt.Printf("  ... Passed\n")
}

func TestPersist3(t *testing.T) {
	servers := 3
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: partitioned leader and one follower crash, leader restarts ...\n")

	cfg.one(101, 3)

	leader := cfg.checkOneLeader()
	cfg.disconnect((leader + 2) % servers)

	cfg.one(102, 2)

	cfg.crash1((leader + 0) % servers)
	cfg.crash1((leader + 1) % servers)
	cfg.connect((leader + 2) % servers)
	cfg.start1((leader + 0) % servers)
	cfg.connect((leader + 0) % servers)

	cfg.one(103, 2)

	cfg.start1((leader + 1) % servers)
	cfg.connect((leader + 1) % servers)

	cfg.one(104, servers)

	fmt.Printf("  ... Passed\n")
}

func TestFigure8(t *testing.T) {
	servers := 5
	cfg := make_config(t, servers, false)
	defer cfg.cleanup()

	fmt.Printf("Test: Figure 8 ...\n")

	cfg.one(rand.Int(), 1)

	nup := servers
	for iters := 0; iters < 1000; iters++ {
		leader := -1
		for i := 0; i < servers; i++ {
			if cfg.rafts[i] != nil {
				_, _, ok := cfg.rafts[i].Start(rand.Int())
				if ok {
					leader = i
				}
			}
		}

		if (rand.Int() % 1000) < 100 {
			ms := (rand.Int63() % 500)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		} else {
			ms := (rand.Int63() % 13)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}

		if leader != -1 {
			cfg.crash1(leader)
			nup -= 1
		}

		if nup < 3 {
			s := rand.Int() % servers
			if cfg.rafts[s] == nil {
				cfg.start1(s)
				cfg.connect(s)
				nup += 1
			}
		}
	}

	for i := 0; i < servers; i++ {
		if cfg.rafts[i] == nil {
			cfg.start1(i)
			cfg.connect(i)
		}
	}

	cfg.one(rand.Int(), servers)

	fmt.Printf("  ... Passed\n")
}

func TestUnreliableAgree(t *testing.T) {
	servers := 5
	cfg := make_config(t, servers, true)
	defer cfg.cleanup()

	fmt.Printf("Test: unreliable agreement ...\n")

	for iters := 1; iters < 50; iters++ {
		for j := 0; j < 4; j++ {
			go cfg.one((100*iters)+j, 1)
		}
		cfg.one(iters, 1)
	}

	cfg.setunreliable(false)

	cfg.one(100, servers)

	fmt.Printf("  ... Passed\n")
}

func TestFigure8Unreliable(t *testing.T) {
	servers := 5
	cfg := make_config(t, servers, true)
	defer cfg.cleanup()

	fmt.Printf("Test: Figure 8 (unreliable) ...\n")

	cfg.one(rand.Int()%10000, 1)

	nup := servers
	for iters := 0; iters < 1000; iters++ {
		leader := -1
		for i := 0; i < servers; i++ {
			_, _, ok := cfg.rafts[i].Start(rand.Int() % 10000)
			if ok && cfg.connected[i] {
				leader = i
			}
		}

		if (rand.Int() % 1000) < 100 {
			ms := (rand.Int63() % 500)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		} else {
			ms := (rand.Int63() % 13)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}

		if leader != -1 && (rand.Int()%1000) < 500 {
			cfg.disconnect(leader)
			nup -= 1
		}

		if nup < 3 {
			s := rand.Int() % servers
			if cfg.connected[s] == false {
				cfg.connect(s)
				nup += 1
			}
		}
	}

	for i := 0; i < servers; i++ {
		if cfg.connected[i] == false {
			cfg.connect(i)
		}
	}

	cfg.one(rand.Int()%10000, servers)

	fmt.Printf("  ... Passed\n")
}

func internalChurn(t *testing.T, unreliable bool) {

	if unreliable {
		fmt.Printf("Test: unreliable churn ...\n")
	} else {
		fmt.Printf("Test: churn ...\n")
	}

	servers := 5
	cfg := make_config(t, servers, unreliable)
	defer cfg.cleanup()

	stop := int32(0)

	// create concurrent clients
	cfn := func(me int, ch chan []int) {
		var ret []int
		ret = nil
		defer func() { ch <- ret }()
		values := []int{}
		for atomic.LoadInt32(&stop) == 0 {
			x := rand.Int()
			index := -1
			ok := false
			for i := 0; i < servers; i++ {
				// try them all, maybe one of them is a leader
				cfg.mu.Lock()
				rf := cfg.rafts[i]
				cfg.mu.Unlock()
				if rf != nil {
					index1, _, ok1 := rf.Start(x)
					if ok1 {
						ok = ok1
						index = index1
					}
				}
			}
			if ok {
				// maybe leader will commit our value, maybe not.
				// but don't wait forever.
				for _, to := range []int{10, 20, 50, 100, 200} {
					nd, cmd := cfg.nCommitted(index)
					if nd > 0 {
						if xx, ok := cmd.(int); ok {
							if xx == x {
								values = append(values, x)
							}
						} else {
							cfg.t.Fatalf("wrong command type")
						}
						break
					}
					time.Sleep(time.Duration(to) * time.Millisecond)
				}
			} else {
				time.Sleep(time.Duration(79+me*17) * time.Millisecond)
			}
		}
		ret = values
	}

	ncli := 3
	cha := []chan []int{}
	for i := 0; i < ncli; i++ {
		cha = append(cha, make(chan []int))
		go cfn(i, cha[i])
	}

	for iters := 0; iters < 20; iters++ {
		if (rand.Int() % 1000) < 200 {
			i := rand.Int() % servers
			cfg.disconnect(i)
		}

		if (rand.Int() % 1000) < 500 {
			i := rand.Int() % servers
			if cfg.rafts[i] == nil {
				cfg.start1(i)
			}
			cfg.connect(i)
		}

		if (rand.Int() % 1000) < 200 {
			i := rand.Int() % servers
			if cfg.rafts[i] != nil {
				cfg.crash1(i)
			}
		}

		time.Sleep(777 * time.Millisecond)
	}

	time.Sleep(1500 * time.Millisecond)
	cfg.setunreliable(false)
	for i := 0; i < servers; i++ {
		if cfg.rafts[i] == nil {
			cfg.start1(i)
		}
		cfg.connect(i)
	}

	atomic.StoreInt32(&stop, 1)

	values := []int{}
	for i := 0; i < ncli; i++ {
		vv := <-cha[i]
		if vv == nil {
			t.Fatal("client failed")
		}
		values = append(values, vv...)
	}

	time.Sleep(1500 * time.Millisecond)

	lastIndex := cfg.one(rand.Int(), servers)

	really := make([]int, lastIndex+1)
	for index := 1; index <= lastIndex; index++ {
		v := cfg.wait(index, servers)
		if vi, ok := v.(int); ok {
			really = append(really, vi)
		} else {
			t.Fatalf("not an int")
		}
	}

	for _, v1 := range values {
		ok := false
		for _, v2 := range really {
			if v1 == v2 {
				ok = true
			}
		}
		if ok == false {
			cfg.t.Fatalf("didn't find a value")
		}
	}

	fmt.Printf("  ... Passed\n")
}

func TestReliableChurn(t *testing.T) {
	internalChurn(t, false)
}

func TestUnreliableChurn(t *testing.T) {
	internalChurn(t, true)
}

// implement GC of log / snapshot + recovery from snapshot
// implement raftkv persistence (maybe only makes sense w/ snapshot)
// test that GC actually frees memory
