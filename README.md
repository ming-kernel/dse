###LockService
1. Primay and Backup can make an agreement using two stragegies: operation logic and raw state, which one should choose?
  
    -- I use the state of the lock, not the operation
  
2. To avoid duplicate request from client, Backup can store previous state or previous reply; which one should choose? If we choose previous state, then we may re-run the service logic in response to the client.
    
    -- I use the previous reply

###Primary/Backup Key/Value Service

Situations that **LockService** do not handle:

	a) client thinks primary is dead (due to network failures or packet loss) and switchs to the backup, but the primary is actually alive, the primary will miss some of the client's operations.
	b) recovering servers: if the primary crashes, but is later repaired, it cann't be re-intergrated to the system, which means the system cannot tolearte further failures. 


Questions:

	1. The acknowledgment rule prevents the view service from getting more than one view ahead of the key/value servers. If the view service could get arbitrarily far ahead, then it would need a more complex design in which it kept a history of views, allowed key/value servers to ask about old views, and garbage-collected information about old views when appropriate.

	2. How to ensure new primary has up-to-date replica of state?

    -- only promote previous backup
    
	3. How to avoid promoting a state-less backup?

    -- primary in each view must ack that view to viewserver; viewserver must stay with current view until acked.
    


