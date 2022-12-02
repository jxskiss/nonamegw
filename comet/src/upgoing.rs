use anyhow::{anyhow, Result};

pub fn connect_upgoing() -> Result<()> {
    let nc = nats::connect("127.0.0.1:4222")?;

    nc.publish("upgoing", "hello world")?;

    Ok(())
}
