#[derive(Clone, PartialEq, ::prost::Message)]
pub struct KvEntry {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub value: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Packet {
    #[prost(bytes="vec", tag="1")]
    pub control: ::prost::alloc::vec::Vec<u8>,
    #[prost(int64, tag="2")]
    pub seq_id: i64,
    #[prost(int32, tag="3")]
    pub flag: i32,
    #[prost(int32, tag="4")]
    pub command: i32,
    #[prost(int64, tag="5")]
    pub biz_flag: i64,
    #[prost(message, repeated, tag="6")]
    pub headers: ::prost::alloc::vec::Vec<KvEntry>,
    #[prost(bytes="vec", tag="7")]
    pub payload: ::prost::alloc::vec::Vec<u8>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Connection {
    #[prost(string, tag="1")]
    pub id: ::prost::alloc::string::String,
    #[prost(int64, tag="2")]
    pub app_id: i64,
    #[prost(int64, tag="3")]
    pub user_id: i64,
    #[prost(int64, tag="4")]
    pub device_id: i64,
    #[prost(string, tag="11")]
    pub client_ip: ::prost::alloc::string::String,
    #[prost(string, tag="12")]
    pub client_version: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConnectionList {
    #[prost(message, repeated, tag="1")]
    pub connections: ::prost::alloc::vec::Vec<Connection>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Content {
    #[prost(int64, tag="1")]
    pub biz_flag: i64,
    #[prost(map="string, string", tag="2")]
    pub headers: ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
    #[prost(bytes="vec", tag="3")]
    pub payload: ::prost::alloc::vec::Vec<u8>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Event {
    #[prost(message, optional, tag="1")]
    pub conn: ::core::option::Option<Connection>,
    #[prost(enumeration="event::Type", tag="2")]
    pub r#type: i32,
    #[prost(message, optional, tag="6")]
    pub reconnect_data: ::core::option::Option<event::ReconnectData>,
}
/// Nested message and enum types in `Event`.
pub mod event {
    #[derive(Clone, PartialEq, ::prost::Message)]
    pub struct ReconnectData {
        #[prost(string, tag="1")]
        pub old_id: ::prost::alloc::string::String,
    }
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Type {
        Touch = 0,
        Connect = 1,
        Reconnect = 2,
        Disconnect = 3,
        Kickoff = 4,
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Message {
    #[prost(message, optional, tag="1")]
    pub conn: ::core::option::Option<Connection>,
    #[prost(message, optional, tag="2")]
    pub content: ::core::option::Option<Content>,
}
