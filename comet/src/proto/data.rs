#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ConnectionInfo {
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
pub struct TokenInfo {
    #[prost(string, tag="1")]
    pub id: ::prost::alloc::string::String,
    #[prost(int64, tag="2")]
    pub sign_time_msec: i64,
    #[prost(int64, tag="3")]
    pub app_id: i64,
    #[prost(int64, tag="4")]
    pub user_id: i64,
    #[prost(int64, tag="5")]
    pub device_id: i64,
}
