use warp;
use warp::Filter;
use axum;
use std::sync::{Mutex, Arc};
use std::collections::HashSet;
use tokio::sync::broadcast;
use axum::ws::{WebSocket, Message};
use axum::prelude::*;
use axum::routing::BoxRoute;
use futures::{sink::SinkExt, stream::StreamExt};
use hyper::service::{make_service_fn, service_fn};

// pub fn routes() -> impl warp::Filter<Extract=impl warp::Reply> + Clone {
//     let ping = warp::path!("ping").map(|| "pong");
//     let ws = warp::path!("ws1").map(|| "ws1"); // FIXME
//     let another = warp::path!("another").map(|| "another"); // FIXME
//     warp::get().and(ws.or(ping).or(another))
// }

pub fn routes2() -> BoxRoute<Body> {
    let app =
        route("/ping", get(|| async { "pong" }))
        .route("/ws1", get(|| async { "ws1" }))
        .route("/another", get(|| async { "another" }));
    app.boxed()
}

// Our shared state
struct AppState {
    user_set: Mutex<HashSet<String>>,
    tx: broadcast::Sender<String>,
}

async fn websocket(stream: WebSocket, state: extract::Extension<Arc<AppState>>) {
    let state = state.0;

    // By splitting we can send and receive at the same time.
    let (mut sender, mut receiver) = stream.split();

    // Username gets set in the receive loop, if its valid.
    let mut username = String::new();

    // Loop until a text message is found.
    while let Some(Ok(msg)) = receiver.next().await {
        if let Some(name) = msg.to_str() {
            // If username that is sent by cleint is not taken, fill username string.
            check_username(&state, &mut username, name);

            // If not empty we want to quit the loop else we want to quit function.
            if !username.is_empty() {
                break;
            } else {
                // Only send our client that username is taken.
                let _ = sender.send(Message::text("Username already taken.")).await;
                return;
            }
        }
    }

    // Subscribe before sending joined message.
    let mut rx = state.tx.subscribe();

    // Send joined message to all subscribers.
    let msg = format!("{} joined.", username);
    let _ = state.tx.send(msg);

    // This task will receive broadcast messages and send text message to our client.
    let mut send_task = tokio::spawn(async move {
        while let Ok(msg) = rx.recv().await {
            // In any websocket error, break loop.
            if sender.send(Message::text(msg)).await.is_err() {
                break;
            }
        }
    });

    // Clone things we want to pass to the receiving task.
    let tx = state.tx.clone();
    let name = username.clone();

    // This task will receive messages from client and send them to broadcast subscribers.
    let mut recv_task = tokio::spawn(async move {
        while let Some(Ok(msg)) = receiver.next().await {
            if let Some(text) = msg.to_str() {
                // Add username before message.
                let _ = tx.send(format!("{}: {}", name, text));
            }
        }
    });

    // If any one of the tasks exit, abort the other.
    tokio::select! {
        _ = (&mut send_task) => recv_task.abort(),
        _ = (&mut recv_task) => send_task.abort(),
    }
    ;

    // Send user left message.
    let msg = format!("{} left.", username);
    let _ = state.tx.send(msg);

    // Remove usename from map so new clients can take it.
    state.user_set.lock().unwrap().remove(&username);
}

fn check_username(state: &AppState, string: &mut String, name: &str) {
    let mut user_set = state.user_set.lock().unwrap();

    if !user_set.contains(name) {
        user_set.insert(name.to_owned());

        string.push_str(name);
    }
}
