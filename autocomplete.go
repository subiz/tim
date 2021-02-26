package tim

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/armon/go-radix"
)

type PrefixTerm struct {
	*sync.Mutex
	col        string
	accid      string
	tree       *radix.Tree
	isLoad     bool
	latestRead int64
}

type AutocompleteMgr struct {
	*sync.Mutex
	prefixTermMap map[string]*PrefixTerm
	buildTermFunc func(col, accid string) []string
	stopchan      chan struct{} // signal to stop scheduling
	isrunning     bool
	isbuilding    bool
}

func NewPrefixTermMgr(buildTermFunc func(col, accid string) []string) *AutocompleteMgr {
	return &AutocompleteMgr{
		Mutex:         &sync.Mutex{},
		prefixTermMap: make(map[string]*PrefixTerm),
		buildTermFunc: buildTermFunc,
		stopchan:      make(chan struct{}),
	}
}

func (mgr *AutocompleteMgr) CleanPrefixTerm() {
	mgr.Lock()
	defer mgr.Unlock()
	rmarr := make([]string, 0)
	now := time.Now().UnixNano() / 1e6
	for key, prefixTerm := range mgr.prefixTermMap {
		if prefixTerm.latestRead < now-24*60*60*1000 {
			rmarr = append(rmarr, key)
		}
	}
	for _, rmkey := range rmarr {
		delete(mgr.prefixTermMap, rmkey)
	}
}

func (mgr *AutocompleteMgr) BuildPrefixTerm() {
	var prefixTerms []*PrefixTerm
	mgr.Lock()
	if mgr.isbuilding {
		mgr.Unlock()
		return
	}
	mgr.isbuilding = true
	prefixTerms = make([]*PrefixTerm, len(mgr.prefixTermMap))
	index := 0
	for _, prefixTerm := range mgr.prefixTermMap {
		prefixTerms[index] = prefixTerm
		index++
	}
	mgr.Unlock()

	for _, prefixTerm := range prefixTerms {
		start := time.Now()
		terms := mgr.buildTermFunc(prefixTerm.col, prefixTerm.accid)
		fmt.Println("BUILDPREFIXTERMMMM-DB", prefixTerm.col, prefixTerm.accid, time.Since(start))
		prefixTerm.Lock()
		start = time.Now()
		for _, term := range terms {
			prefixTerm.tree.Insert(term, nil)
		}
		fmt.Println("BUILDPREFIXTERMMMM-TREE", prefixTerm.col, prefixTerm.accid, time.Since(start))
		prefixTerm.Unlock()
	}

	mgr.Lock()
	mgr.isbuilding = false
	mgr.Unlock()
}

func (mgr *AutocompleteMgr) AsyncTerm() chan struct{} {
	mgr.Lock()
	defer mgr.Unlock()
	if mgr.isrunning {
		return mgr.stopchan
	}
	mgr.isrunning = true
	ticker := time.NewTicker(1 * time.Second)
	var tickerCount int64
	go func() {
		for {
			select {
			case <-ticker.C:
				atomic.AddInt64(&tickerCount, 1)
				if tickerCount%(5*60) == 0 {
					go mgr.BuildPrefixTerm()
				}
				if tickerCount%(7*60) == 0 {
					go mgr.CleanPrefixTerm()
				}
			case <-mgr.stopchan:
				ticker.Stop()
				mgr.isrunning = false
				return
			}
		}
	}()

	return mgr.stopchan
}

func (mgr *AutocompleteMgr) GetPrefixTerm(col, accid string) *PrefixTerm {
	var prefixTerm *PrefixTerm
	joinkey := col + "-" + accid
	mgr.Lock()
	prefixTerm, _ = mgr.prefixTermMap[joinkey]
	if prefixTerm == nil {
		mgr.prefixTermMap[joinkey] = &PrefixTerm{Mutex: &sync.Mutex{}, tree: radix.New(), col: col, accid: accid}
		prefixTerm = mgr.prefixTermMap[joinkey]
	}
	mgr.Unlock()
	prefixTerm.Lock()
	defer prefixTerm.Unlock()
	if prefixTerm.isLoad {
		return prefixTerm
	}
	start := time.Now()
	terms := mgr.buildTermFunc(col, accid)
	fmt.Println("LOADTERMMMM-DB", col, accid, time.Since(start))
	start = time.Now()
	for _, term := range terms {
		prefixTerm.tree.Insert(term, nil)
	}
	fmt.Println("LOADTERMMMM-TREE", col, accid, time.Since(start))
	prefixTerm.isLoad = true
	prefixTerm.latestRead = time.Now().UnixNano() / 1e6
	return prefixTerm
}

func (mgr *AutocompleteMgr) Insert(col, accid, term string) {
	prefixTerm := mgr.GetPrefixTerm(col, accid)
	prefixTerm.Lock()
	prefixTerm.tree.Insert(term, nil)
	prefixTerm.Unlock()
}

func (mgr *AutocompleteMgr) Inserts(col, accid string, terms []string) {
	prefixTerm := mgr.GetPrefixTerm(col, accid)
	prefixTerm.Lock()
	for _, term := range terms {
		prefixTerm.tree.Insert(term, nil)
	}
	prefixTerm.Unlock()
}

func (mgr *AutocompleteMgr) KeysWithPrefix(col, accid, prefix string) []string {
	prefixTerm := mgr.GetPrefixTerm(col, accid)
	terms := []string{}
	fn := func(s string, v interface{}) bool {
		terms = append(terms, s)
		return false
	}
	prefixTerm.Lock()
	prefixTerm.tree.WalkPrefix(prefix, fn)
	prefixTerm.latestRead = time.Now().UnixNano() / 1e6
	prefixTerm.Unlock()
	limit := 5
	if len(terms) <= limit {
		return terms
	}
	out := make([]string, limit)
	k := len(terms) / limit
	j := 0
	for i := 0; i < limit; i++ {
		out[i] = terms[j]
		j += k
	}
	return out
}
