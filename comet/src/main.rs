#![feature(duration_constants)]

use env_logger;
use futures_util::{SinkExt, StreamExt};
use log::*;
use std::net::SocketAddr;
use std::str::FromStr;
use tokio::net::{TcpListener, TcpStream};
use tokio_tungstenite::{accept_async, tungstenite::Error};
use tungstenite::Result;
use warp::Filter;
use axum::prelude::*;

mod connid;
mod connmgr;
mod nats;
mod proto;
mod server;
mod upgoing;
mod manager;
mod connection;
mod linked_list;

async fn handle_connection(peer: SocketAddr, stream: TcpStream) -> Result<()> {
    let mut ws_stream = accept_async(stream).await.expect("failed to accept");
    info!("new websocket connection: {}", peer);

    while let Some(msg) = ws_stream.next().await {
        let msg = msg?;
        if msg.is_text() || msg.is_binary() {
            ws_stream.send(msg).await?;
        }
    }
    Ok(())
}

async fn accept_connection(peer: SocketAddr, stream: TcpStream) {
    if let Err(e) = handle_connection(peer, stream).await {
        match e {
            Error::ConnectionClosed | Error::Protocol(_) | Error::Utf8 => (),
            err => error!("error processing connection: {}", err),
        }
    }
}

#[tokio::main]
async fn main() {
    env_logger::init();

    let app = server::routes2();

    let addr: SocketAddr = "127.0.0.1:9002".parse().unwrap();
    hyper::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
