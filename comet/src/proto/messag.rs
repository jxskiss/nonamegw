#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpgoingMessage {
    #[prost(message, optional, tag="1")]
    pub packet: ::core::option::Option<super::protocol::Packet>,
    #[prost(message, optional, tag="2")]
    pub conn: ::core::option::Option<super::protocol::Connection>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DowngoingMessage {
    #[prost(string, repeated, tag="3")]
    pub conn_ids: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(oneof="downgoing_message::Data", tags="1, 2")]
    pub data: ::core::option::Option<downgoing_message::Data>,
}
/// Nested message and enum types in `DowngoingMessage`.
pub mod downgoing_message {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Data {
        #[prost(message, tag="1")]
        Packet(super::super::protocol::Packet),
        #[prost(bytes, tag="2")]
        BinPacket(::prost::alloc::vec::Vec<u8>),
    }
}
/// TODO
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BroadcastMessage {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TokenKey {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(int64, tag="2")]
    pub enable_time_sec: i64,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CometConfiguration {
    #[prost(string, tag="1")]
    pub token_key: ::prost::alloc::string::String,
    #[prost(message, repeated, tag="2")]
    pub old_token_keys: ::prost::alloc::vec::Vec<TokenKey>,
}
