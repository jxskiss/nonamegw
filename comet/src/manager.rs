use log::{debug, error, info, warn};
use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::atomic::{AtomicU64, Ordering, AtomicUsize};
use std::sync::Arc;
use std::time::{SystemTime, Duration};
use tokio::sync::mpsc;

use crate::connection::Connection;
use crate::connid::{Generator, ShortConnectionId, ConnectionId};
use crate::linked_list::{LinkedList, NodePtr, Node};
use crate::nats::NatsService;

pub struct Manager {
    clients: dashmap::DashMap<ShortConnectionId, Arc<Connection>>,
    hb_manager: HeartbeatManager,
    uid_gen: Generator,
    nats: NatsService,
}

impl Manager {
    pub fn add_connection(&self, conn: Arc<Connection>) {
        let short_id = conn.uid.to_short();
        self.clients.insert(short_id, conn);
        self.hb_manager.touch(short_id);
    }

    pub fn touch_connection(&self, conn: &Connection) {
        info!("touching connection: uid= {}", conn.uid.to_string());
        self.hb_manager.remove(conn.uid.to_short());
    }

    pub fn remove_connection(&self, conn: &Connection) {
        let short_id = conn.uid.to_short();
        self.clients.remove(&short_id);
        self.hb_manager.remove(short_id);
    }

    pub fn get(&self, uid: ConnectionId) -> Option<Arc<Connection>> {
        let short_id = uid.to_short();
        match self.clients.get(&short_id) {
            None => None,
            Some(kv) => Some(kv.value().clone()),
        }
    }
}

struct HeartbeatManager {

    // TODO: spsc or spmc?
    pub tasks: mpsc::Receiver<Arc<Connection>>,

    touch: mpsc::Sender<ShortConnectionId>,
    remove: mpsc::Sender<ShortConnectionId>,

    pending: AtomicUsize,
}

struct HeartbeatNode {
    idx: u64,
    conn: Arc<Connection>,
}

impl HeartbeatNode {
    fn new(idx: u64, conn: Arc<Connection>) -> Self {
        HeartbeatNode { idx, conn }
    }
}

impl HeartbeatManager {

    fn new() -> Self {
        let (touch_tx, mut touch_rx) = mpsc::channel(10000);
        let (remove_tx, mut remove_rx) = mpsc::channel(10000);
        let (tasks_tx, mut tasks_rx) = mpsc::channel(10000);
        let _self = Self {
            tasks: tasks_rx,
            touch: touch_tx,
            remove: remove_tx,
            pending: Default::default(),
        };
        tokio::spawn(async move {
            _self.run(touch_rx, remove_rx, tasks_tx);
        });
        _self
    }

    pub async fn touch(&self, short_id: ShortConnectionId) {
        if conn.get_ping_interval() > 0 {
            self.touch.send(short_id).await;
        }
    }

    pub async fn remove(&self, short_id: ShortConnectionId) {
        self.remove.send(short_id).await;
    }

    async fn run(
        &self,
        mut touch: mpsc::Receiver<ShortConnectionId>,
        mut remove: mpsc::Receiver<ShortConnectionId>,
        tasks: mpsc::Sender<Arc<Connection>>,
    ) {
        let mut list: LinkedList<RefCell<HeartbeatNode>> = LinkedList::new();
        let mut map: HashMap<ShortConnectionId, NodePtr<RefCell<HeartbeatNode>>> = HashMap::new();
        let counter = AtomicU64::default();
        let pre_index = 0u64;

        let handle_touch = |short_id: ShortConnectionId| {
            let idx = counter.fetch_add(1, Ordering::SeqCst) + 1;
            if let Some(node_ptr) = map.get(&short_id) {
                let node = unsafe { *node_ptr as &mut Node<RefCell<HeartbeatNode>> };
                *node.idx = idx;
                list.move_to_back(*node_ptr);
            } else {
                let hb_node = HeartbeatNode::new(idx, conn);
                let node = RefCell::new(hb_node);
                let list_node = list.push_back(node);
                map.insert(short_id, list_node);
            }
        };

        let handle_remove = |short_id: ShortConnectionId| {
            if let Some(node_ptr) = map.get(&short_id) {
                map.remove(&short_id);
                unsafe { list.remove(*node_ptr) };
            }
        };

        let handle_tick = || {
            const _LIMIT: usize = 1000;
            let front = list.front();
            if front.is_null() {
                return;
            }

            let node_ptr = front;
            let count: usize = 0;
            loop {
                if count >= _LIMIT || node_ptr.is_null() {
                    break;
                }
                let node = unsafe { node_ptr as &mut Node<RefCell<HeartbeatNode>> };
                if *node.idx <= pre_index {
                    break;
                }
                if let Ok(_) = tasks.try_send(node.conn) {
                    self.pending.fetch_add(1, Ordering::SeqCst);
                } else {
                    break;
                }
            }
        };

        let ticker = tokio::time::interval(Duration::SECOND);

        loop {
            tokio::select! {
                Some(conn) = touch.recv() => {
                    handle_touch(conn);
                },
                Some(conn) = remove.recv() => {
                    handle_remove(conn);
                },
                _ = ticker.tick() => {
                    handle_tick();
                }
            }
        }
    }

    // fn ping(&self, tasks: &mut mpsc::Receiver<Arc<Connection>>) {
    //     while let Some(conn) = tasks.recv().await {
    //         if !conn.should_ping(SystemTime::now()) {
    //             continue;
    //         }
    //         tokio::spawn(async move {
    //             // FIXME: error handling
    //             conn.ping().await.unwrap();
    //             self.pending.fetch_add(1, Ordering::SeqCst);
    //         });
    //     }
    // }
}
