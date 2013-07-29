package lockservice

//
// RPC definitions for a simple lock service.
//
// You will need to modify this file.
//

//
// Lock(lockname) returns OK=true if the lock is not held.
// If it is held, it returns OK=false immediately.
//
type LockArgs struct {
	// Go's net/rpc requires that these field
	// names start with upper case letters!
	Seq      int64
	Lockname string // lock name
}

type LockReply struct {
	Seq int64
	OK  bool
}

//
// Unlock(lockname) returns OK=true if the lock was held.
// It returns OK=false if the lock was not held.
//
type UnlockArgs struct {
	Seq      int64
	Lockname string
}

type UnlockReply struct {
	Seq int64
	OK  bool
}

type SyncArgs struct {
	Seq      int64
	Lockname string
	Value    bool
}

type SyncReply struct {
	Seq int64
	OK  bool
}
