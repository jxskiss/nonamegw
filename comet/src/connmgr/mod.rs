use crate::connid::{ConnectionId, ShortConnectionId};
use dashmap::DashMap;

pub struct Connection {}

impl Connection {
    pub fn id(&self) -> &ConnectionId {
        todo!()
    }

    pub fn short_id(&self) -> ShortConnectionId {
        todo!()
    }
}

// pub struct ConnectionHub<'a> {
//     connections: DashMap<ShortConnectionId, Box<Connection>>,
//
//     // pingMgr
//     // tagMgr
//
//     // closingConns chan net.Conn
// }
//
// impl ConnectionHub {
//     pub fn add(&self, conn: &Connection) {
//         self.connections.insert(conn.short_id(), conn);
//
//         todo!()
//     }
//
//     pub fn remove(&self, conn: &Connection) {
//         self.connections.remove(conn.short_id());
//
//         todo!()
//     }
//
//     pub fn touch(&self, conn: &Connection) {
//         todo!()
//     }
//
//     pub fn size(&self) -> usize {
//         self.connections.len()
//     }
//
//     pub fn get(&self, id: impl Into<ShortConnectionId>) -> Box<ConnectionId> {
//         let short_id = id.into();
//         self.connections.get(short_id)
//     }
// }

/*
func NewConnHub(size int) *connHub {
    h := &connHub{
        size:    size,
        shards:  make([]*shard, 0, size),
        tagMgr:  &tagManager{},
        pingMgr: NewPingManager(),
    }
    for i := 0; i < size; i++ {
        s := &shard{
            m: make(map[int64]*Connection),
        }
        h.shards = append(h.shards, s)
    }
    h.closingConns = make(chan net.Conn, 512)
    go h.closer()
    return h
}

type connHub struct {
    size   int
    shards []*shard

    pingMgr *pingManager
    tagMgr  *tagManager

    closingConns chan net.Conn
}

type shard struct {
    sync.RWMutex
    m map[int64]*Connection
}

func (h *connHub) hashshard(uid UID) int {
    return int(uid.Counter()) % h.size
}

func (h *connHub) hashping(interval int) int {
    return interval
}

// Remove removes the passed in connection from the hub.
func (h *connHub) Remove(conn *Connection) {
    s := h.shards[h.hashshard(conn.uid)]
    s.Lock()
    delete(s.m, conn.uid.Short())
    s.Unlock()

    h.pingMgr.Remove(conn)

    connMeta := conn.meta
    productId := connMeta.GetProductId()
    appId := connMeta.GetAppId()
    for k, v := range connMeta.GetTags() {
        if tag := maketag(k, v); tag != "" {
            h.tagMgr.Remove(productId, appId, tag, conn.uid.Short())
        }
    }
}

// Add add the provided connection to the connection hub.
func (h *connHub) Add(conn *Connection) {
    if G.LogEnabled(zap.DebugLevel) {
        G.L.Debug("adding connection", zap.String("uid", conn.meta.ConnUid))
    }
    s := h.shards[h.hashshard(conn.uid)]
    s.Lock()
    s.m[conn.uid.Short()] = conn
    s.Unlock()
    h.pingMgr.Touch(conn)
}

// Touch touch the provided connection to help manage periodic tasks, such as heartbeat.
func (h *connHub) Touch(conn *Connection) {
    if G.LogEnabled(zap.DebugLevel) {
        G.L.Debug("touching connection", zap.String("uid", conn.meta.ConnUid))
    }
    h.pingMgr.Touch(conn)
}

// Get returns connection of the given uid in the hub, when the connection
// is found, the return value ok will be true, else the returned value conn
// is nil and ok will be false.
func (h *connHub) Get(uid UID) (conn *Connection, ok bool) {
    s := h.shards[h.hashshard(uid)]
    s.RLock()
    conn, ok = s.m[uid.Short()]
    s.RUnlock()
    return
}

// List returns all connections in the hub in no particular order.
func (h *connHub) List() []*Connection {
    conns := make([]*Connection, 0, h.Len()+1000)
    h.Iter(func(c *Connection) error {
        conns = append(conns, c)
        return nil
    })
    return conns
}

// Len returns the number of connections in the hub.
func (h *connHub) Len() (n int) {
    for _, s := range h.shards {
        s.RLock()
        n += len(s.m)
        s.RUnlock()
    }
    return
}

// Iter calls function fn on each connection of the hub, in no particular
// order. If fn returns non-nil error, the iteration will be aborted with
// the same error as return value.
func (h *connHub) Iter(fn func(*Connection) error) (err error) {
    for _, s := range h.shards {
        s.RLock()
        for _, c := range s.m {
            if fnerr := fn(c); fnerr != nil {
                err = fnerr
                break
            }
        }
        s.RUnlock()
        if err != nil {
            return
        }
    }
    return
}

func (h *connHub) Close() {
    var wg sync.WaitGroup
    var conns = h.List()
    wg.Add(len(conns))
    for _, conn := range conns {
        conn := conn
        err := G.Pool.Schedule(func() {
            conn.CloseWithFrame(ws.CompiledCloseGoingAway)
            wg.Done()
        })
        if err != nil {
            wg.Done()
        }
    }
    wg.Wait()
}

// UpdateTags removes old tags of the connection specified by the param
// from the hub and adds the connection's present tags to the hub.
//
// NOTE: the connection's metaMu lock should be held by the caller.
func (h *connHub) UpdateTags(conn *Connection, oldTags map[string]string) {
    meta := conn.meta
    productId := meta.GetProductId()
    appId := meta.GetAppId()
    shortUid := conn.uid.Short()
    removed, added := difftags(oldTags, meta.GetTags())

    if G.LogEnabled(zap.DebugLevel) {
        G.L.Debug("update connection tags",
            zap.String("uid", meta.ConnUid),
            zap.Strings("removed", removed),
            zap.Strings("added", added),
        )
    }

    for _, tag := range removed {
        h.tagMgr.Remove(productId, appId, tag, shortUid)
    }
    for _, tag := range added {
        h.tagMgr.Add(productId, appId, tag, shortUid)
    }
}

// FindTags returns connections in the hub which have the given tags.
func (h *connHub) FindTags(productId, appId int32, tags map[string]string) []*Connection {
    strTags := make([]string, 0, len(tags))
    for k, v := range tags {
        if t := maketag(k, v); t != "" {
            strTags = append(strTags, t)
        }
    }
    if G.LogEnabled(zap.DebugLevel) {
        G.L.Debug("find tagged connections", zap.Strings("tags", strTags))
    }

    conns := make([]*Connection, 0, 128)
    shortUids := h.tagMgr.Find(productId, appId, strTags)
    for _, shortId := range shortUids {
        uid, _ := G.UIDMaker.FromShort(shortId)
        conn, ok := h.Get(uid)
        if !ok {
            continue
        }
        conns = append(conns, conn)
    }
    return conns
}

func difftags(oldTags, newTags map[string]string) (removed, added []string) {
    var oTag, nTag string
    for k, nVal := range newTags {
        if nTag = maketag(k, nVal); nTag == "" {
            continue
        }
        if oVal, ok := oldTags[k]; ok {
            if oVal == nVal { // not changed
                continue
            }
            removed = append(removed, maketag(k, oVal))
        }
        added = append(added, nTag)
    }
    for k, oVal := range oldTags {
        if oTag := maketag(k, oVal); oTag == "" {
            continue
        }
        if _, ok := newTags[k]; ok {
            // already processed above
            continue
        }
        // not find in new tags, it's removed
        removed = append(removed, oTag)
    }
    return
}

func (h *connHub) PendingPingTasks() int {
    return h.pingMgr.Pending()
}

func (h *connHub) ScheduleClose(conn net.Conn) {
    h.closingConns <- conn
}

func (h *connHub) closer() {
    type closingConn struct {
        conn   net.Conn
        tsNano int64
    }

    // reuse conns slice memory since we tick somewhat frequently
    pool := newChanPool(5, func() interface{} {
        return make([]*closingConn, 0, 128)
    })

    pending := list.New()
    ticker := wheel.NewTicker(100 * time.Millisecond)
    for {
        select {
        case conn := <-h.closingConns:
            pending.PushBack(&closingConn{
                conn:   conn,
                tsNano: wheel.CheapNowNano(),
            })
        case <-ticker.C:
            if pending.Front() == nil {
                continue
            }
            const timeout = int64(time.Second)
            var now = wheel.CheapNowNano()
            var conns = pool.get().([]*closingConn)[:0]
            for e := pending.Front(); e != nil; e = e.Next() {
                v := e.Value.(*closingConn)
                if now-v.tsNano < timeout {
                    break
                }
                conns = append(conns, v)
                pending.Remove(e)
            }
            if len(conns) == 0 {
                pool.put(conns)
                continue
            }

            // close the underlying connections
            go func(conns []*closingConn) {
                for i, c := range conns {
                    c.conn.Close()
                    conns[i] = nil // avoid resource leak
                }
                // reuse memory buffer that is not larger than 32k
                if len(conns) <= 4096 {
                    pool.put(conns)
                }
            }(conns)
        }
    }
}

func newChanPool(size int, factory func() interface{}) *chanPool {
    return &chanPool{
        ch:      make(chan interface{}, size),
        factory: factory,
    }
}

type chanPool struct {
    ch      chan interface{}
    factory func() interface{}
}

func (p *chanPool) get() interface{} {
    select {
    case v := <-p.ch:
        return v
    default:
        return p.factory()
    }
}

func (p *chanPool) put(v interface{}) {
    select {
    case p.ch <- v:
    default:
    }
}
 */
