package tim

import (
	"fmt"
	"sync"
	"time"

	"github.com/armon/go-radix"
)

type PrefixTerm struct {
	*sync.Mutex
	tree   *radix.Tree
	isLoad bool
}

type AutocompleteMgr struct {
	*sync.Mutex
	prefixTermMap map[string]*PrefixTerm
	buildTermFunc func(col, accid string) []string
}

func NewPrefixTermMgr(buildTermFunc func(col, accid string) []string) *AutocompleteMgr {
	return &AutocompleteMgr{
		Mutex:         &sync.Mutex{},
		prefixTermMap: make(map[string]*PrefixTerm),
		buildTermFunc: buildTermFunc,
	}
}

func (mgr *AutocompleteMgr) GetPrefixTerm(col, accid string) *PrefixTerm {
	var prefixTerm *PrefixTerm
	joinkey := col + "-" + accid
	mgr.Lock()
	prefixTerm, _ = mgr.prefixTermMap[joinkey]
	if prefixTerm == nil {
		mgr.prefixTermMap[joinkey] = &PrefixTerm{Mutex: &sync.Mutex{}, tree: radix.New()}
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
	prefixTerm.isLoad = true
	fmt.Println("LOADTERMMMM-TREE", col, accid, time.Since(start))
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
	out := []string{}
	fn := func(s string, v interface{}) bool {
		out = append(out, s)
		return false
	}
	prefixTerm.Lock()
	prefixTerm.tree.WalkPrefix(prefix, fn)
	prefixTerm.Unlock()
	return out
}
