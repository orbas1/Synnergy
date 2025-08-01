package core

import "sync"

// QuorumTracker tracks votes from validators or token holders and checks
// whether a quorum has been reached. It is a generic helper used by
// consensus and governance modules.

type QuorumTracker struct {
	mu        sync.Mutex
	threshold int // number of votes required to pass
	votes     map[Address]struct{}
	total     int // total eligible voters
}

// NewQuorumTracker returns a tracker requiring `threshold` votes out of
// `total` possible. Threshold must be >0 and <= total.
func NewQuorumTracker(total, threshold int) *QuorumTracker {
	if threshold <= 0 || threshold > total {
		threshold = total
	}
	return &QuorumTracker{
		threshold: threshold,
		votes:     make(map[Address]struct{}),
		total:     total,
	}
}

// AddVote records a vote from the given address. Duplicate votes are ignored.
// It returns the current number of unique votes.
func (qt *QuorumTracker) AddVote(addr Address) int {
	qt.mu.Lock()
	qt.votes[addr] = struct{}{}
	n := len(qt.votes)
	qt.mu.Unlock()
	return n
}

// HasQuorum returns true if the number of unique votes is greater or equal
// to the required threshold.
func (qt *QuorumTracker) HasQuorum() bool {
	qt.mu.Lock()
	n := len(qt.votes)
	qt.mu.Unlock()
	return n >= qt.threshold
}

// Reset clears all recorded votes.
func (qt *QuorumTracker) Reset() {
	qt.mu.Lock()
	qt.votes = make(map[Address]struct{})
	qt.mu.Unlock()
}

// The following standalone helpers are wired into the opcode dispatcher. They
// operate on a global tracker instance returned by CurrentQuorumTracker().

var globalQuorum *QuorumTracker
var qmu sync.Mutex

// InitQuorumTracker initialises the global tracker used by opcodes.
func InitQuorumTracker(total, threshold int) {
	qmu.Lock()
	globalQuorum = NewQuorumTracker(total, threshold)
	qmu.Unlock()
}

// CurrentQuorumTracker returns the active global tracker.
func CurrentQuorumTracker() *QuorumTracker {
	qmu.Lock()
	defer qmu.Unlock()
	if globalQuorum == nil {
		globalQuorum = NewQuorumTracker(1, 1)
	}
	return globalQuorum
}

// Quorum_AddVote exposes AddVote for opcode use.
func Quorum_AddVote(addr Address) int { return CurrentQuorumTracker().AddVote(addr) }

// Quorum_HasQuorum checks if the threshold is met.
func Quorum_HasQuorum() bool { return CurrentQuorumTracker().HasQuorum() }

// Quorum_Reset clears the global vote map.
func Quorum_Reset() { CurrentQuorumTracker().Reset() }
