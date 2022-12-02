use std::sync::atomic::{AtomicU64, Ordering};
use warp::ws::{WebSocket, Message};
use std::time::{SystemTime, Duration, UNIX_EPOCH};
use anyhow::Result;

use crate::connid::ConnectionId;
use crate::proto::protocol::Connection as ProtocolConnection;

pub struct Connection {
    pub uid: ConnectionId,
    pub meta: ProtocolConnection,
    pub stream: WebSocket,

    // TODO
    // tags: Vec<String>,

    create_time_sec: u64,
    touch_time_sec: AtomicU64,
    access_time_sec: AtomicU64,
    ping_interval_sec: AtomicU64,
}

impl Connection {
    pub fn new(uid: ConnectionId, meta: ProtocolConnection, stream: WebSocket) -> Self {
        Self {
            uid,
            meta,
            stream,
            create_time_sec: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            touch_time_sec: Default::default(),
            access_time_sec: Default::default(),
            ping_interval_sec: Default::default(),
        }
    }

    pub fn register() {
        todo!()
    }

    pub fn get_create_time(&self) -> SystemTime {
        UNIX_EPOCH + Duration::from_secs(self.create_time_sec)
    }

    pub fn get_touch_time(&self) -> SystemTime {
        let touch_t = self.touch_time_sec.load(Ordering::SeqCst);
        UNIX_EPOCH + Duration::from_secs(touch_t)
    }

    pub fn set_touch_time(&self, t: SystemTime) {
        let touch_t = t.duration_since(UNIX_EPOCH).unwrap().as_secs();
        self.touch_time_sec.store(touch_t, Ordering::SeqCst);
    }

    pub fn get_access_time(&self) -> SystemTime {
        let access_t = self.access_time_sec.load(Ordering::SeqCst);
        UNIX_EPOCH + Duration::from_secs(access_t)
    }

    pub fn set_access_time(&self, t: SystemTime) {
        let access_t = t.duration_since(UNIX_EPOCH).unwrap().as_secs();
        self.access_time_sec.store(access_t, Ordering::SeqCst);
    }

    pub fn get_ping_interval(&self) -> Duration {
        Duration::from_secs(self.ping_interval_sec.load(Ordering::SeqCst))
    }

    pub fn should_ping(&self) -> bool {
        let interval = self.get_ping_interval();
        let access_t = self.get_access_time();
        interval > 0 &&
            SystemTime::now().duration_since(access_t).unwrap() > 0
    }

    pub async fn ping(&self) -> Result<()> {
        match self.stream
            .send(Message::ping(vec![])) // FIXME
            .await {
            Ok(_) => Ok(()),
            Err(err) => Err(anyhow::Error::new(err)),
        }
    }
}
