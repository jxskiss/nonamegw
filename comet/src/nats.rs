use anyhow::{anyhow, Result};
use async_nats as nats;

use crate::proto::messag::{UpgoingMessage, DowngoingMessage};
use crate::proto::protocol::{Event};

const UPGOING_MESSAGE_TOPIC: &str = "broker.upgoingMessage";
const EVENT_TOPIC: &str = "broker.event";
const DOWNGOING_MESSAGE_TOPIC: &str = "broker.{}.downgoingMessage";

pub struct NatsConfig {
    pub server_url: String,
    pub machine_id: String,
    pub send_message: Box<dyn async Fn(DowngoingMessage) -> Result<()>>,
}

pub struct NatsService {
    config: NatsConfig,
    conn: nats::Connection,
}

impl NatsService {
    pub async fn new(cfg: NatsConfig) -> Self {

        // FIXME: "127.0.0.1:4222"
        let nc = nats::connect(cfg.server_url.as_str()).unwrap();

        Self {
            config: cfg,
            conn: nc,
        }
    }

    pub async fn setup(&self) -> Result<()> {
        self._subscribe_downgoing_messages().await?;
        Ok(())
    }

    async fn _subscribe_downgoing_messages(&self) -> Result<()> {
        let subject = format!(DOWNGOING_MESSAGE_TOPIC, self.config.machine_id);
        let sub = self.conn.subscribe(&subject).await?;
        tokio::spawn(async move || {
            while let Some(msg) = sub.next().await {
                let mut pb_msg: DowngoingMessage = DowngoingMessage::fixme(); // FIXME
                pb_msg.decode(msg.data).unwrap();
                self.config.send_message(pb_msg).await.unwrap();
            };
            sub.drain().await;
        });
        Ok(())
    }

    pub async fn call_rpc<REQ, RSP>(&self, subject: &str, req: REQ, resp: RSP) -> Result<()>
        where REQ: prost::Message,
              RSP: prost::Message,
    {
        let buf = req.encode_to_vec();
        let rsp_msg = self.conn.request(subject, buf).await?;
        resp.decode(rsp_msg.data)?;
        Ok(())
    }

    async fn _send<T: prost::Message>(&self, subject: &str, data: T) -> Result<()> {
        let buf = data.encode_to_vec();
        self.conn.publish(subject, buf).await?;
        Ok(())
    }

    pub async fn send_upgoing_message(&self, message: UpgoingMessage) -> Result<()> {
        self._send(UPGOING_MESSAGE_TOPIC, message).await
    }

    pub async fn send_upgoing_event(&self, event: Event) -> Result<()> {
        self._send(EVENT_TOPIC, event).await
    }
}
