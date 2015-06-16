package rps

import(
	`sync`
	`time`
	`github.com/0xor1/sid`
	`github.com/0xor1/sus`
	`github.com/0xor1/oak`
	`golang.org/x/net/context`
	`google.golang.org/appengine/datastore`
	`github.com/qedus/nds`
)

const(
	kind = `game`
)

var(
	lastGaeClearOut time.Time
	mtx sync.Mutex
)

func NewGaeGameStore() oak.EntityStore {
	pre := func() {
		myLastGaeClearOutInst := lastGaeClearOut
		if lastGaeClearOut.IsZero() || time.Since(lastGaeClearOut).Hours() >= 1 {
			mtx.Lock()
			if lastGaeClearOut != myLastGaeClearOutInst {
				mtx.Unlock()
				return
			}
			lastGaeClearOut = time.Now()
			mtx.Unlock()
			q := datastore.NewQuery(kind).Filter(`DeleteAfter <=`, time.Now()).KeysOnly()
			keys := []*datastore.Key // TODO - make this with a large len and cap and keep increasing in large chunk sizes, appending one every time is slow and inefficient
			for iter := q.Run(context.Background()); ; {
				key, err := iter.Next(nil)
				if err == datastore.Done {
					break
				}
				if err != nil {
					return
				}
				append(keys, key)
			}
			nds.DeleteMulti(context.Background(), keys)
		}
		return
	}
	return &gameStore{preprocess: pre, inner: sus.NewGaeStore(kind, sid.Uuid, func()sus.Version{return NewGame()})}
}

func NewLocalGameStore() oak.EntityStore {
	pre := func(){}
	return &gameStore{preprocess: pre, inner: sus.NewJsonMemoryStore(sid.Uuid, func()sus.Version{return NewGame()})}
}

type gameStore struct {
	preprocess	func()
	inner sus.Store
}

func (gs *gameStore) Create() (string, oak.Entity, error) {
	go gs.preprocess()
	id, v, err := gs.inner.Create()
	var e oak.Entity
	if err == nil && v != nil {
		e = oak.Entity(v)
	}
	return id, e, err
}

func (gs *gameStore) Read(entityId string) (oak.Entity, error) {
	go gs.preprocess()
	v, err := gs.inner.Read(entityId)
	var e oak.Entity
	if err == nil && v != nil {
		e = oak.Entity(v)
	}
	return e, err
}

func (gs *gameStore) Update(entityId string, entity oak.Entity) (error) {
	go gs.preprocess()
	return gs.inner.Update(entityId, entity)
}