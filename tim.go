// search in vietnamese is "tim kiem"
// it is a simple library so just call it "tim"
package tim

import (
	"github.com/gocql/gocql"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"fmt"
	"hash/crc32"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

//
// Can be used to:
// + search conversation (conversation service)
// + search lead or user (user service)
// + search account (account service)
// + search for a product (product service)

// Thuật ngữ:
// + Document: thứ mà người dùng cần tìm (conversation, user, lead, account, ...)
// + Query: đoạn text mà user nhập để tìm các document
// + Owner: Người sở hữu document (quan hệ owner và document là N-N). Owner là loại filter duy nhất chúng ta hỗ trợ
//         Tức là: khi tìm hội thoại, ko cho filter theo ngày tạo, số lượng tin nhắn hay trạng thái nhưng cho phép
//         filter theo ai đang có trong cuộc hội thoại này
// + Tokenize: Hàm dùng để biến một đoạn text thành các term
//
//
// Cách sử dụng:
// indexer.AppendText("convo", "acc1", "convo1", "xin chào")
// indexer.AppendText("convo", "acc1", "convo1", "cộng hòa xã hội chủ nghĩa việt nam")
// indexer.UpdateOwner("convo", "acc1", "convo1", []string{"ag1", "ag2"})
//
// indexer.AppendText("convo", "acc1", "convo2", "xin tạm biệt")
// indexer.AppendText("convo", "acc1", "convo2", "độc lập tự do hạnh phúc")
// indexer.UpdateOwner("convo", "acc1", "convo1", []string{"ag1"})
//
// indexer.Search("convo", "acc1", "ag1", "xin")
//   => 2 hits: [convo1, convo2]
// indexer.Search("convo", "acc1", "ag2", "xin")
//   => 1 hit: [convo1]
//
//
// Thiết kế:
// + Trong database chỉ lữu trữ docId, nội dung doc không quan tâm và không lưu
// + Sử dụng 4 bảng:
//   CREATE KEYSPACE tim WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};
//
//   CREATE TABLE tim.doc_term(col ASCII, acc ASCII, term ASCII, doc ASCII, PRIMARY KEY ((col, acc, doc), term));
//     xem docs chứa những term nào, phục vụ mục đích clear term, đánh lại index cho doc
//
//   CREATE TABLE tim.term_doc(col ASCII, acc ASCII, term ASCII, doc ASCII, PRIMARY KEY ((col, acc, term), doc));
//     dùng để tìm kiếm docs theo term
//
//   CREATE TABLE tim.term (col ASCII, acc ASCII, par INT, term ASCII, PRIMARY KEY ((col, acc, par), term));
//     dùng để xây prefix trie, cần thiết cho tính năng suggestion
//
//   CREATE TABLE tim.owner (col ASCII, acc ASCII , par INT, doc ASCII, owners SET<ASCII>, PRIMARY KEY((col, acc, par), doc));
//     dùng để tra cứu quan hệ doc - owner
//

// text => term

// see http://www.clc.hcmus.edu.vn/?page_id=1507
var STOP_WORDS = map[string]bool{"va": true, "cua": true, "co": true, "cac": true, "la": true}

// Search all docs that match the query
func Search(collection, accid, ownerid, query string) ([]string, error) {
	waitforstartup(collection, accid)

	terms := tokenize(query)
	if len(terms) == 0 {
		return []string{}, nil
	}

	// order by length desc
	sort.Slice(terms, func(i, j int) bool { return len(terms[i]) > len(terms[j]) })

	owneridhash := hash(ownerid)

	// match docs must statisfy all terms
	// find term by by term and intersect the results
	// start by the most lengthy term so we can soon drop the unneeded docs

	// contain all matched doc
	match := map[string]bool{}
	for termindex, term := range terms {
		iter := db.Query("SELECT doc FROM tim.term_doc WHERE col=? AND acc=? AND term=?", collection, accid, term).Iter()
		var docid string
		localMatch := map[string]bool{}
		for iter.Scan(&docid) {
			// quick escape since we know this docid will never match
			if termindex != 0 && !match[docid] {
				continue
			}

			ownerLock.Lock()
			owners := ownerM[hash(collection+"-"+accid)][hash(docid)]
			for _, owner := range owners {
				if owner == owneridhash {
					localMatch[docid] = true
					break
				}
			}
			ownerLock.Unlock()
		}

		if err := iter.Close(); err != nil {
			return nil, err
		}

		if termindex == 0 {
			match = localMatch
		} else {
			// AND match with localMatch
			for doc := range match {
				if !localMatch[doc] {
					delete(match, doc)
				}
			}
		}

		// early escape when nothing matched
		if len(match) == 0 {
			break
		}
	}

	hits := []string{}
	for doc := range match {
		if len(hits) > 200 {
			break
		}
		hits = append(hits, doc)
	}
	return hits, nil
}

func ClearText(collection, accid, docId string) error {
	waitforstartup(collection, accid)

	var term string
	iter := db.Query("SELECT term FROM tim.doc_term WHERE col=? AND acc=? AND doc=?", collection, accid, docId).Iter()
	for iter.Scan(&term) {
		if err := db.Query("DELETE FROM tim.term_doc WHERE col=? AND acc=? AND term=? AND doc=?", collection, accid, term, docId).Exec(); err != nil {
			return err
		}
	}
	if err := iter.Close(); err != nil {
		return err
	}

	if err := db.Query("DELETE FROM tim.doc_term WHERE col=? AND acc=? AND AND doc=?", collection, accid, docId).Exec(); err != nil {
		return err
	}
	return nil
}

func AppendText(collection, accid, docId, text string) error {
	waitforstartup(collection, accid)
	terms := tokenize(text)
	for _, term := range terms {
		if err := db.Query("INSERT INTO tim.term_doc(col, acc, term, doc) VALUES(?,?,?,?)", collection, accid, term, docId).Exec(); err != nil {
			return err
		}
		if err := db.Query("INSERT INTO tim.doc_term(col, acc, term, doc) VALUES(?,?,?,?)", collection, accid, term, docId).Exec(); err != nil {
			return err
		}
		par := hash(term) % TERM_PAR
		if err := db.Query("INSERT INTO tim.term(col, acc, par, term) VALUES(?,?,?,?)", collection, accid, par, term).Exec(); err != nil {
			return err
		}
	}
	return nil
}

func UpdateOwner(collection, accid, docId string, owners []string) error {
	waitforstartup(collection, accid)

	par := hash(docId) % PAR
	err := db.Query("INSERT INTO tim.owner(col,acc,par,doc,owners) VALUES(?,?,?,?,?)", collection, accid, par, docId, owners).Exec()
	if err != nil {
		return err
	}
	colhash := hash(collection + "-" + accid)

	ownerLock.Lock()
	if ownerM[colhash] == nil {
		ownerM[colhash] = make(map[uint32][]uint32)
	}
	dochash := hash(docId)
	ownerhashs := make([]uint32, len(owners), len(owners))
	for i, owner := range owners {
		ownerhashs[i] = hash(owner)
	}
	ownerM[colhash][dochash] = ownerhashs
	ownerLock.Unlock()
	return nil
}

func Suggest(collection, accid, query string) []string {
	waitforstartup(collection, accid)
	return nil
}

var startupLock sync.Mutex
var readyM map[string]bool

var db *gocql.Session

const PAR = 100
const TERM_PAR = 100

var ownerLock sync.Mutex

// hash(collection+accid) => hash(doc_id) => hash(owner_ids)
var ownerM map[uint32]map[uint32][]uint32

func waitforstartup(collection, accid string) {
	startupLock.Lock()
	defer startupLock.Unlock()

	if ownerM == nil {
		ownerM = make(map[uint32]map[uint32][]uint32)
	}

	if readyM == nil {
		readyM = make(map[string]bool)
	}

	if readyM[collection+"-"+accid] {
		return
	}

	// connect db
	if db == nil {
		cluster := gocql.NewCluster("db-0")
		cluster.Timeout = 10 * time.Second
		cluster.Keyspace = "tim"
		var err error
		for {
			if db, err = cluster.CreateSession(); err == nil {
				break
			}
			fmt.Println("cassandra", err, ". Retring after 5sec...")
			time.Sleep(5 * time.Second)
		}
	}

	// build owner map for collection acc
	colhash := hash(collection + "-" + accid)

	// should parallel ?
	for par := 0; par < PAR; par++ {
		iter := db.Query("SELECT doc, owners FROM tim.owner WHERE col=? AND acc=? AND par=?", collection, accid, par).Iter()
		var docid string
		var owners []string
		for iter.Scan(&docid, &owners) {
			ownerLock.Lock()
			if ownerM[colhash] == nil {
				ownerM[colhash] = make(map[uint32][]uint32)
			}
			dochash := hash(docid)
			ownerhashs := make([]uint32, len(owners), len(owners))
			for i, owner := range owners {
				ownerhashs[i] = hash(owner)
			}
			ownerM[colhash][dochash] = ownerhashs
			ownerLock.Unlock()
		}
		if err := iter.Close(); err != nil {
			panic(err)
		}
	}
	readyM[collection+"-"+accid] = true
}

var crc32q = crc32.MakeTable(crc32.IEEE)

func hash(text string) uint32 {
	return crc32.Checksum([]byte(text), crc32q)
}

func toASCII(text string) string {
	for k, v := range VNMAP {
		text = strings.Replace(text, k, v, -1)
	}

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	if s, _, err := transform.String(t, text); err == nil {
		text = s
	}

	// remove all non-ascii
	text = strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1
		}
		return r
	}, text)
	return text
}

func n_gram(str string, n int) []string {
	if len(str) == 0 {
		return nil
	}

	var result []string
	for i := 0; i < len(str)-n+1; i++ {
		for j := n; i+j <= len(str); j++ {
			result = append(result, str[i:i+j])
		}
	}
	return result
}

var VNMAP = map[string]string{
	"ạ": "a", "ả": "a", "ã": "a", "à": "a", "á": "a", "â": "a", "ậ": "a", "ầ": "a", "ấ": "a",
	"ẩ": "a", "ẫ": "a", "ă": "a", "ắ": "a", "ằ": "a", "ặ": "a", "ẳ": "a", "ẵ": "a",
	"ó": "o", "ò": "o", "ọ": "o", "õ": "o", "ỏ": "o", "ô": "o", "ộ": "o", "ổ": "o", "ỗ": "o",
	"ồ": "o", "ố": "o", "ơ": "o", "ờ": "o", "ớ": "o", "ợ": "o", "ở": "o", "ỡ": "o",
	"é": "e", "è": "e", "ẻ": "e", "ẹ": "e", "ẽ": "e", "ê": "e", "ế": "e", "ề": "e", "ệ": "e", "ể": "e", "ễ": "e",
	"ú": "u", "ù": "u", "ụ": "u", "ủ": "u", "ũ": "u", "ư": "u", "ự": "u", "ữ": "u", "ử": "u", "ừ": "u", "ứ": "u",
	"í": "i", "ì": "i", "ị": "i", "ỉ": "i", "ĩ": "i",
	"ý": "y", "ỳ": "y", "ỷ": "y", "ỵ": "y", "ỹ": "y",
	"đ": "d",
	"Ạ": "A", "Ả": "A", "Ã": "A", "À": "A", "Á": "A", "Â": "A", "Ậ": "A", "Ầ": "A", "Ấ": "A",
	"Ẩ": "A", "Ẫ": "A", "Ă": "A", "Ắ": "A", "Ằ": "A", "Ặ": "A", "Ẳ": "A", "Ẵ": "A",
	"Ó": "O", "Ò": "O", "Ọ": "O", "Õ": "O", "Ỏ": "O", "Ô": "O", "Ộ": "O", "Ổ": "O", "Ỗ": "O",
	"Ồ": "O", "Ố": "O", "Ơ": "O", "Ờ": "O", "Ớ": "O", "Ợ": "O", "Ở": "O", "Ỡ": "O",
	"É": "E", "È": "E", "Ẻ": "E", "Ẹ": "E", "Ẽ": "E", "Ê": "E", "Ế": "E", "Ề": "E", "Ệ": "E", "Ể": "E", "Ễ": "E",
	"Ú": "U", "Ù": "U", "Ụ": "U", "Ủ": "U", "Ũ": "U", "Ư": "U", "Ự": "U", "Ữ": "U", "Ử": "U", "Ừ": "U", "Ứ": "U",
	"Í": "I", "Ì": "I", "Ị": "I", "Ỉ": "I", "Ĩ": "I",
	"Ý": "Y", "Ỳ": "Y", "Ỷ": "Y", "Ỵ": "Y", "Ỹ": "Y",
	"Đ": "D",
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func Report(collection, accid string) {
	// connect db
	cluster := gocql.NewCluster("db-0")
	cluster.Timeout = 10 * time.Second
	cluster.Keyspace = "tim"
	var err error
	for {
		if db, err = cluster.CreateSession(); err == nil {
			break
		}
		fmt.Println("cassandra", err, ". Retring after 5sec...")
		time.Sleep(5 * time.Second)
	}

	full := map[string]int{}

	shards := makeShards(20) // run 20 threads
	wg := sync.WaitGroup{}

	mutex := &sync.Mutex{}
	wg.Add(len(shards))
	for _, tokens := range shards {
		go func(fromtoken, totoken int64) {
			iter := db.Query("SELECT col,acc,term FROM tim.term_doc WHERE token(col,acc,term)>=? AND token(col,acc,term)<=?", fromtoken, totoken).Iter()
			acc, col, term := "", "", ""

			for iter.Scan(&col, &acc, &term) {
				if accid != "" && acc != accid {
					continue
				}
				if collection != "" && col != collection {
					continue
				}
				mutex.Lock()
				full[term]++
				mutex.Unlock()
			}
			if err := iter.Close(); err != nil {
				panic(err)
			}
			wg.Done()
		}(tokens[0], tokens[1])
	}
	wg.Wait()
	fmt.Println("REPORT of total", len(full), "terms")
	fmt.Println(drawGraph("FOR COLLECTION "+collection+"AND ACC"+accid, full))
}

func makeShards(n int) [][2]int64 {
	d := (9223372036854775807 / n) * 2
	out := make([][2]int64, n, n)

	p := int64(-9223372036854775808)
	for i := 0; i < n; i++ {
		out[i] = [2]int64{p, p + int64(d) - 1}
		p += int64(d)
	}
	out[n-1][1] = 9223372036854775807
	return out
}
