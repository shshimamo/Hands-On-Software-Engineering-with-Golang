package memory

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/google/uuid"
	"github.com/shshimamo/Hands-On-Software-Engineering-with-Golang/textindexer/index"
	"golang.org/x/xerrors"
	"sync"
	"time"
)

const batchSize = 10

var _ index.Indexer = (*InMemoryBleveIndexer)(nil)

type bleveDoc struct {
	Title    string
	Content  string
	PageRank float64
}

type InMemoryBleveIndexer struct {
	mu   sync.RWMutex
	docs map[string]*index.Document
	idx  bleve.Index
}

func NewInMemoryBleveIndexer() (*InMemoryBleveIndexer, error) {
	mapping := bleve.NewIndexMapping()
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	return &InMemoryBleveIndexer{
		idx:  idx,
		docs: make(map[string]*index.Document),
	}, nil
}

func (i *InMemoryBleveIndexer) Close() error {
	return i.idx.Close()
}

func (i *InMemoryBleveIndexer) Index(doc *index.Document) error {
	if doc.LinkID == uuid.Nil {
		return xerrors.Errorf("index: %w", index.ErrMissingLinkID)
	}

	doc.IndexedAt = time.Now()
	dcopy := copyDoc(doc)
	key := dcopy.LinkID.String()

	i.mu.Lock()
	defer i.mu.Unlock()

	// 更新する場合、既存のPageRankスコアを保持する。
	if orig, exists := i.docs[key]; exists {
		dcopy.PageRank = orig.PageRank
	}

	if err := i.idx.Index(key, makeBleveDoc(dcopy)); err != nil {
		return xerrors.Errorf("index: %w", err)
	}

	i.docs[key] = dcopy
	i.mu.Unlock()
	return nil
}

func (i *InMemoryBleveIndexer) FindByID(linkID uuid.UUID) (*index.Document, error) {
	return i.findByID(linkID.String())
}

func (i *InMemoryBleveIndexer) findByID(linkID string) (*index.Document, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if d, found := i.docs[linkID]; found {
		return copyDoc(d), nil
	}

	return nil, xerrors.Errorf("find by ID: %w", index.ErrNotFound)
}

// 特定のクエリのインデックスを検索し、結果のイテレータを返します
func (i *InMemoryBleveIndexer) Search(q index.Query) (index.Iterator, error) {
	var bq query.Query
	switch q.Type {
	case index.QueryTypePhrase:
		bq = bleve.NewMatchPhraseQuery(q.Expression)
	default:
		bq = bleve.NewMatchQuery(q.Expression)
	}

	searchReq := bleve.NewSearchRequest(bq)
	searchReq.SortBy([]string{"-PageRank", "-_score"})
	searchReq.Size = batchSize
	searchReq.From = int(q.Offset)
	rs, err := i.idx.Search(searchReq)
	if err != nil {
		return nil, xerrors.Errorf("index: %w", err)
	}

	return &bleveIterator{
		idx: i, searchReq: searchReq, rs: rs, cumIdx: q.Offset}, nil
	}
}

// UpdateScore は、指定されたリンク ID を持つ文書の PageRank スコアを更新する。
// そのような文書が存在しない場合、指定されたスコアを持つプレースホルダ文書が作成される。
func (i *InMemoryBleveIndexer) UpdateScore(linkID uuid.UUID, score float64) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	key := linkID.String()
	doc, found := i.docs[key]
	if !found {
		doc = &index.Document{LinkID: linkID}
		i.docs[key] = doc
	}

	doc.PageRank = score
	if err := i.idx.Index(key, makeBleveDoc(doc)); err != nil {
		return xerrors.Errorf("update score: %w", err)
	}

	return nil
}

// copyDoc ヘルパーは、内部ドキュメントマップに安全に格納できるオリジナルドキュメントのコピーを作成する
func copyDoc(d *index.Document) *index.Document {
	dcopy := new(index.Document)
	*dcopy = *d
	return dcopy
}

// makeBleveDocヘルパーは、検索クエリの一部として使用したいフィールドのみを含む、オリジナルドキュメントの部分的で軽量なビューを返す
func makeBleveDoc(d *index.Document) bleveDoc {
	return bleveDoc{
		Title:    d.Title,
		Content:  d.Content,
		PageRank: d.PageRank,
	}
}
