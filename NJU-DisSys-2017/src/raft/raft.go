package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"fmt"
	"labrpc"
	"math/rand"
	"sync"
	"time"
	"sync/atomic"
	"bytes"
	"encoding/gob"
)

// import "bytes"
// import "encoding/gob"

const (
	FOLLOWER = iota
	CANDIDATE
	LEADER

	HEARTBEAT_TIMEOUT = 50//心跳间隔
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type LogEntry struct{
	Index int
	Term int
	Command interface{}
}


//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	// Your data here.
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	votedFor int   //给谁投票
	voteAcquired int  //获得的票数
	state int32		//当前状态
	currentTerm int32	//当前任期
	electionTimer *time.Timer	//选举时间间隔
	voteCh chan struct{}	//成功投票的信号
	appendCh chan struct{}	//成功更新log的信号
	commitCh chan struct{}
	log []LogEntry
	//log[1:commitIndex]已经提交的log，大部分server都达成一致
	commitIndex int	//index of highest log entry konwn to be commited
	lastApplied int	//index of highest log entry applied to state machine
	nextIndex []int	//为每个server维护一个nextIndex，一致性检查，nextIndex之前的日志都和leader一致
	matchIndex []int //每个server已经和leader匹配的最高的log index
	applyCh  chan ApplyMsg
}

// atomic operations
func (rf *Raft) getTerm() int32 {
	return atomic.LoadInt32(&rf.currentTerm)
}

func (rf *Raft) incrementTerm() {
	atomic.AddInt32(&rf.currentTerm, 1)
}

func (rf *Raft) getLastLogIndex() int {
	return len(rf.log) - 1
}
func (rf *Raft) isStateEqual(state int32) bool {
	return atomic.LoadInt32(&rf.state) == state
}
func (rf *Raft) getLastLogTerm() int{
	return rf.log[rf.getLastLogIndex()].Term
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here.
	term = int(rf.getTerm())
	isleader = rf.isStateEqual(LEADER)
	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here.
	// Example:
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here.
	// Example:
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	d.Decode(&rf.currentTerm)
	d.Decode(&rf.votedFor)
	d.Decode(&rf.log)
}


//
// example RequestVote RPC arguments structure.
//
type RequestVoteArgs struct {
	// Your data here.
	Term int32	//请求者当前所处时期
	CandidateId int	//编号

	LastLogIndex int
	LastLogTerm int
}

//
// example RequestVote RPC reply structure.
//
type RequestVoteReply struct {
	// Your data here.
	Term int32	//回复者所处时期
	IsVote bool	//是否投票
}


func randElectionTimeout() time.Duration{
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//return time.Millisecond*time.Duration(r.Int63n(MAX_ELECTION_TIMEOUT-MIN_ELECTION_TIMEOUT)+MIN_ELECTION_TIMEOUT)
	return time.Millisecond*time.Duration(rand.Int63() % 333 + 550)
}

func (rf *Raft) startElection(){
	rf.incrementTerm()	//term+1
	rf.votedFor = rf.me	//给自己投一票
	rf.voteAcquired = 1
	rf.persist()
	rf.electionTimer.Reset(randElectionTimeout())
	go rf.broadcastRequestVote()	//广播投票
}

func (rf *Raft) updateState(state int32){
	if rf.isStateEqual(state){
		return
	}
	//states := []string{"FOLLOWER", "CANDIDATE", "LEADER"}
	//preState := rf.state
	switch state {
	case FOLLOWER:
		rf.state = FOLLOWER
		rf.votedFor = -1
	case CANDIDATE:
		rf.state = CANDIDATE
		rf.startElection()
	case LEADER:
		//leader为每个server维护两个状态数组
		for i, _ := range rf.peers {
			rf.nextIndex[i] = rf.getLastLogIndex() +1
			rf.matchIndex[i] = 0
		}
		rf.state = LEADER
	default:
		fmt.Printf("invalid state %d",state)
	}
	//fmt.Printf("In term %d: Server %d transfer from %s to %s\n", rf.currentTerm, rf.me, states[preState], states[rf.state])

}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here.
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	reply.IsVote = false
	defer rf.persist()
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		return
	}else if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.updateState(FOLLOWER)
		rf.votedFor = args.CandidateId
		reply.IsVote = true
	}else {
		if rf.votedFor == -1 {
			rf.votedFor = args.CandidateId
			reply.IsVote = true
		}else {
			reply.IsVote = false
		}
	}


	//选举限制
	//如果两个日志的任期号不同，任期号大的更新；如果任期号相同，更长的日志更新，更长表示最后一个日志的索引更大
	thisLogIndex := rf.getLastLogIndex()
	thisLogTerm :=rf.getLastLogTerm()

	if thisLogTerm > args.LastLogTerm{
		reply.IsVote = false
	}else if thisLogTerm == args.LastLogTerm {
		if thisLogIndex > args.LastLogIndex {
			reply.IsVote = false
		}
	}


	if reply.IsVote == true {
		go func() {rf.voteCh <- struct{}{}}()
	}
}

type AppendEntriesArgs struct {
	Term int32 	//leader`s term
	LeaderId int	//leader`s Id


	//leader认为的该follower最后一个匹配的位置
	PrevLogIndex int
	//该位置上对应entry的Term
	PreLogTerm int
	Entries []LogEntry
	LeaderCommit int //leader`s commitIndex

}

type AppendEntriesResults struct {
	Term int32
	Success bool

	NextIndex int
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesResults) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()

	reply.Success = false
	if args.Term <rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.NextIndex = rf.getLastLogIndex() + 1
		return
	}else if args.Term > rf.currentTerm{
		rf.currentTerm = args.Term
		rf.updateState(FOLLOWER)
		reply.Success = true
	}else {
		reply.Success = true
	}


	//log一致性检查
	//当前
	if args.PrevLogIndex > rf.getLastLogIndex() {
		reply.Success = false
		reply.NextIndex = rf.getLastLogIndex()+1
		return
	}

	//任期不同的日志删除
	//回退当前不一样任期的日志重试
	if args.PreLogTerm != rf.log[args.PrevLogIndex].Term {

		reply.Success =false


		errorTerm := rf.log[args.PrevLogIndex].Term
		i := args.PrevLogIndex
		for  ; rf.log[i].Term == errorTerm ; i--{

		}
		reply.NextIndex = i+1
		return
	}


	//进行更新log
	conflictIndex := -1
	if rf.getLastLogIndex() < args.PrevLogIndex+len(args.Entries) {
		//检测是否已经添加,如果长度不相等显然没添加
		conflictIndex = args.PrevLogIndex + 1

	}else {
		for i := 0; i <len(args.Entries); i++{
			if rf.log[i+args.PrevLogIndex+1].Term != args.Entries[i].Term{
				conflictIndex = i+args.PrevLogIndex + 1
				break
			}
		}
	}
	if conflictIndex != -1 {
		rf.log = append(rf.log[:args.PrevLogIndex+1], args.Entries...)//rf.log=append(rf.log[:conflictIndex],args.Entries[conflictIndex:]...)
	}

	if args.LeaderCommit > rf.commitIndex {
		if args.LeaderCommit > rf.getLastLogIndex(){
			rf.commitIndex = rf.getLastLogIndex()
		}else{
			rf.commitIndex = args.LeaderCommit
		}
		//rf.commitCh <- struct{}{}
	}
	go func() {rf.appendCh <- struct{}{}}()
}

func (rf *Raft) broadcastRequestVote(){	//广播请求投票
	rf.mu.Lock()
	args := RequestVoteArgs{Term:atomic.LoadInt32(&rf.currentTerm), CandidateId:rf.me,LastLogIndex:rf.getLastLogIndex(),LastLogTerm:rf.log[rf.getLastLogIndex()].Term}
	rf.mu.Unlock()
	for i := range rf.peers{
		if i== rf.me{
			continue
		}
		go func(server int) {
			var reply RequestVoteReply
			if rf.isStateEqual(CANDIDATE) && rf.sendRequestVote(server,&args,&reply){
				rf.mu.Lock()
				defer rf.mu.Unlock()
				if reply.IsVote == true {
					rf.voteAcquired += 1
				}else {
					if reply.Term >rf.currentTerm {
						rf.currentTerm = reply.Term
						rf.updateState(FOLLOWER)
						rf.persist()
					}

				}
			}else {
				//fmt.Println("CANDIDATE: ",rf.isStateEqual(CANDIDATE))
				//fmt.Printf("Server %d send vote req failed.\n", rf.me)
			}
		}(i)
	}
}

func (rf *Raft) broadcastAppendEntries() {	//广播心跳

	//添加日志，返回true代表日志添加不成功需要重新尝试添加
	appendEntriesToServer := func(server int) bool {

		//fmt.Println("server:",server)
		rf.mu.Lock()
		args := AppendEntriesArgs{Term:atomic.LoadInt32(&rf.currentTerm), LeaderId:rf.me,PrevLogIndex:rf.nextIndex[server]-1,
			PreLogTerm:rf.log[rf.nextIndex[server]-1].Term,LeaderCommit:rf.commitIndex}
		if rf.getLastLogIndex() >= rf.nextIndex[server] {
			args.Entries = rf.log[rf.nextIndex[server]:]
		}
		rf.mu.Unlock()
		var reply AppendEntriesResults

		if rf.isStateEqual(LEADER) && rf.sendAppendEntries(server,&args,&reply){
			rf.mu.Lock()
			defer rf.mu.Unlock()
			if reply.Success == true {
				rf.nextIndex[server] += len(args.Entries)
				rf.matchIndex[server] = rf.nextIndex[server] - 1
			}else{

				if rf.state != LEADER {
					return false
				}

				if reply.Term >rf.currentTerm {
					rf.currentTerm = reply.Term
					rf.updateState(FOLLOWER)
				}else{
					//日志没添加成功需要重试
					rf.nextIndex[server] = reply.NextIndex
					return true
				}
			}

		}
		return false
	}



	for i := range rf.peers {
		if i == rf.me {
			continue
		}

		go func(server int){

			for {
				if appendEntriesToServer(server) == false{
					break
				}
			}
		}(i)
	}
}


//leader根据多数投票法来commit log
func (rf *Raft) updateCommitIndex() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	//从leader的最后一个日志开始遍历，测试是否大多数server`log已经和leader匹配
	for i := rf.getLastLogIndex(); i > rf.commitIndex; i--{
		matchedVote := 1
		for i,matched := range rf.matchIndex {
			if i== rf.me {
				continue
			}
			if matched > rf.commitIndex {
				matchedVote += 1
			}
		}

		if matchedVote > len(rf.peers)/2 {
			rf.commitIndex = i
			break
		}
	}
}


//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesResults) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//


func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := -1
	isLeader := true

	term, isLeader = rf.GetState()
	if isLeader == true {

		index = len(rf.log)
		rf.log = append(rf.log,LogEntry{index,term,command})
		rf.persist()

	}

	return index, term, isLeader
}


//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here.
	rf.state = FOLLOWER
	rf.votedFor = -1
	rf.voteCh = make(chan struct{})
	rf.appendCh = make(chan struct{})
	rf.currentTerm = 0
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	rf.applyCh = applyCh
	rf.log = make([]LogEntry,1)
	rf.commitIndex = 0

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go func(){
		rf.electionTimer = time.NewTimer(randElectionTimeout())
		for{
			switch atomic.LoadInt32(&rf.state) {
			case FOLLOWER:
				select {
				case <-rf.voteCh :
					rf.electionTimer.Reset(randElectionTimeout())
				case <-rf.appendCh:
					rf.electionTimer.Reset(randElectionTimeout())
				case <-rf.electionTimer.C:
					rf.mu.Lock()
					rf.updateState(CANDIDATE)
					rf.mu.Unlock()
				}
			case CANDIDATE:
				rf.mu.Lock()
				//rf.startElection()//candidate开始新一轮选举
				//rf.mu.Unlock()

				select {
				case <- rf.appendCh:
					rf.updateState(FOLLOWER)
				case <-rf.electionTimer.C:
					rf.electionTimer.Reset(randElectionTimeout())
					rf.startElection()
					//rf.persist()
				default:
					if rf.voteAcquired> len(rf.peers)/2{
						rf.updateState(LEADER)
					}
				}
				rf.mu.Unlock()

			case LEADER:
				rf.broadcastAppendEntries()
				rf.updateCommitIndex()
				time.Sleep(HEARTBEAT_TIMEOUT*time.Millisecond)
			}
			go rf.applyLog()

		}


	}()

	return rf
}

//apply log to state machine
func (rf *Raft) applyLog() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.commitIndex > rf.lastApplied {
		for i := rf.lastApplied + 1; i <= rf.commitIndex; i++ {
			var msg ApplyMsg
			msg.Index = i
			msg.Command = rf.log[i].Command
			rf.applyCh <- msg
			rf.lastApplied = i
		}
		//fmt.Println("commitIndex",rf.commitIndex)
	}
}
