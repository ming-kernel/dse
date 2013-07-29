package lockservice

import "net"
import "net/rpc"
import "log"
import "sync"
import "fmt"
import "os"
import "io"
import "time"

type LockServer struct {
	mu    sync.Mutex
	l     net.Listener
	dead  bool // for test_test.go
	dying bool // for test_test.go

	am_primary bool   // am I the primary?
	backup     string // backup's port

	// for each lock name, is it locked?
	locks map[string]bool

	last_seq   int64
	last_reply bool
}

//
// server Lock RPC handler.
//

// only primary calls this function
func (ls *LockServer) sync(seq int64, lockname string, lockvalue bool, reply bool) {
	args := &SyncArgs{}
	args.Seq = seq
	args.Lockname = lockname
	args.LockValue = lockvalue
	args.Reply = reply
	var sync_reply SyncReply

	// fmt.Printf("Primary sync lockname = %s, value = %v\n", args.Lockname, args.Value)

	call(ls.backup, "LockServer.Sync", args, &sync_reply)
}

func (ls *LockServer) Sync(args *SyncArgs, reply *SyncReply) error {
	// fmt.Printf("Sync lockname = %s, value = %v\n", args.Lockname, args.LockValue)
	ls.locks[args.Lockname] = args.LockValue
	ls.last_seq = args.Seq
	ls.last_reply = args.Reply
	return nil
}

// you will have to modify this function
//
func (ls *LockServer) Lock(args *LockArgs, reply *LockReply) error {

	// Your code here.
	if ls.am_primary {
		// fmt.Printf("Primary: Lock: lockname = %s, seq = %v\n",
		// 	args.Lockname, args.Seq)
	} else {
		// fmt.Printf("Backup: Lock: lockname = %s, seq = %v, last_seq = %v\n",
		// 	args.Lockname, args.Seq, ls.last_seq)
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	// backup checks sequence number
	if !ls.am_primary && args.Seq == ls.last_seq {
		// fmt.Printf("Backup: last_reply = %v\n", ls.last_reply)
		reply.OK = ls.last_reply
		return nil
	}

	locked, _ := ls.locks[args.Lockname]

	if locked {
		reply.OK = false
	} else {
		reply.OK = true
		ls.locks[args.Lockname] = true
	}

	// sync current lock state, and current reply
	if ls.am_primary {
		ls.sync(args.Seq, args.Lockname, ls.locks[args.Lockname], reply.OK)
	}

	return nil
}

//
// server Unlock RPC handler.
//
func (ls *LockServer) Unlock(args *UnlockArgs, reply *UnlockReply) error {

	// Your code here.
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// backup checks sequence number
	if !ls.am_primary && args.Seq == ls.last_seq {
		reply.OK = ls.last_reply
		return nil
	}

	locked, _ := ls.locks[args.Lockname]

	if locked {
		reply.OK = true
		ls.locks[args.Lockname] = false
	} else {
		reply.OK = false
	}

	if ls.am_primary {
		ls.sync(args.Seq, args.Lockname, ls.locks[args.Lockname], reply.OK)
	}
	return nil
}

//
// tell the server to shut itself down.
// for testing.
// please don't change this.
//
func (ls *LockServer) kill() {
	ls.dead = true
	ls.l.Close()
}

//
// hack to allow test_test.go to have primary process
// an RPC but not send a reply. can't use the shutdown()
// trick b/c that causes client to immediately get an
// error and send to backup before primary does.
// please don't change anything to do with DeafConn.
//
type DeafConn struct {
	c io.ReadWriteCloser
}

func (dc DeafConn) Write(p []byte) (n int, err error) {
	return len(p), nil
}
func (dc DeafConn) Close() error {
	return dc.c.Close()
}
func (dc DeafConn) Read(p []byte) (n int, err error) {
	return dc.c.Read(p)
}

func StartServer(primary string, backup string, am_primary bool) *LockServer {
	ls := new(LockServer)
	ls.backup = backup
	ls.am_primary = am_primary
	ls.locks = map[string]bool{}

	// Your initialization code here.

	// me is port number as a string
	me := ""
	if am_primary {
		me = primary
	} else {
		me = backup
	}

	// tell net/rpc about our RPC server and handlers.
	rpcs := rpc.NewServer()
	rpcs.Register(ls)

	// prepare to receive connections from clients.
	// change "unix" to "tcp" to use over a network.
	os.Remove(me) // only needed for "unix"
	l, e := net.Listen("unix", me)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	ls.l = l

	// please don't change any of the following code,
	// or do anything to subvert it.

	// create a thread to accept RPC connections from clients.
	go func() {
		for ls.dead == false {
			conn, err := ls.l.Accept()
			if err == nil && ls.dead == false {
				if ls.dying {
					// process the request but force discard of reply.

					// without this the connection is never closed,
					// b/c ServeConn() is waiting for more requests.
					// test_test.go depends on this two seconds.
					go func() {
						time.Sleep(2 * time.Second)
						// if ls.am_primary {
						// 	fmt.Printf("Primary: to die\n")
						// } else {
						// 	fmt.Printf("Backup: to die\n")
						// }

						conn.Close()
					}()
					ls.l.Close()

					// this object has the type ServeConn expects,
					// but discards writes (i.e. discards the RPC reply).
					deaf_conn := DeafConn{c: conn}

					rpcs.ServeConn(deaf_conn)

					ls.dead = true
				} else {
					go rpcs.ServeConn(conn)
				}
			} else if err == nil {
				conn.Close()
			}
			if err != nil && ls.dead == false {
				fmt.Printf("LockServer(%v) accept: %v\n", me, err.Error())
				ls.kill()
			}
		}
	}()

	return ls
}
