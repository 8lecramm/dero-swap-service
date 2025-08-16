package main

import (
	"sync"
	"time"
)

type backoff struct {
	userlist map[string]int64
	sync.Mutex
}

var Delay = backoff{userlist: make(map[string]int64)}

func (u *backoff) AddUser(address string) {

	u.Lock()
	defer u.Unlock()

	u.userlist[address] = time.Now().UnixMilli()
}

func (u *backoff) DelUser(address string) {
	delete(u.userlist, address)
}

func (u *backoff) CheckUser(address string) bool {

	u.Lock()
	defer u.Unlock()

	if _, ok := u.userlist[address]; ok {
		return true
	} else {
		return false
	}
}

func (u *backoff) CheckBackoff() {

	u.Lock()
	defer u.Unlock()

	for i, e := range u.userlist {
		added := time.UnixMilli(e)
		if time.Since(added) >= time.Minute*2 {
			Delay.DelUser(i)
		}
	}
}
