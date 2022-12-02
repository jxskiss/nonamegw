// ---- broker RPC through Nats Server ---- //

#[derive(Clone, PartialEq, ::prost::Message)]
pub struct AuthToken {
    #[prost(string, tag="1")]
    pub token: ::prost::alloc::string::String,
    #[prost(int64, tag="2")]
    pub sign_time_msec: i64,
    #[prost(int64, tag="3")]
    pub app_id: i64,
    #[prost(int64, tag="4")]
    pub user_id: i64,
    #[prost(int64, tag="5")]
    pub device_id: i64,
}
/// Nested message and enum types in `AuthToken`.
pub mod auth_token {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum VerifyCode {
        Success = 0,
        Invalid = 1,
        Expired = 2,
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct VerifyAuthTokenRequest {
    #[prost(string, tag="1")]
    pub token: ::prost::alloc::string::String,
    #[prost(string, tag="2")]
    pub client_ip: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct VerifyAuthTokenResponse {
    #[prost(enumeration="auth_token::VerifyCode", tag="1")]
    pub code: i32,
    #[prost(message, optional, tag="2")]
    pub token: ::core::option::Option<AuthToken>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetCometConfigurationRequest {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetCometConfigurationResponse {
    #[prost(message, optional, tag="1")]
    pub configuration: ::core::option::Option<super::messag::CometConfiguration>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BrokerRequest {
    #[prost(oneof="broker_request::Request", tags="1, 2")]
    pub request: ::core::option::Option<broker_request::Request>,
}
/// Nested message and enum types in `BrokerRequest`.
pub mod broker_request {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Request {
        #[prost(message, tag="1")]
        VerifyAuthTokenRequest(super::VerifyAuthTokenRequest),
        #[prost(message, tag="2")]
        GetCometConfigurationRequest(super::GetCometConfigurationRequest),
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BrokerResponse {
    #[prost(oneof="broker_response::Response", tags="1, 2")]
    pub response: ::core::option::Option<broker_response::Response>,
}
/// Nested message and enum types in `BrokerResponse`.
pub mod broker_response {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Response {
        #[prost(message, tag="1")]
        VerifyAuthTokenResponse(super::VerifyAuthTokenResponse),
        #[prost(message, tag="2")]
        GetCometConfigurationResponse(super::GetCometConfigurationResponse),
    }
}
// ---- broker RPC through Nats Server ---- //

// ---- comet RPC through Nats Server ---- //

#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetConnectionInfoRequest {
    #[prost(string, tag="1")]
    pub conn_id: ::prost::alloc::string::String,
}
/// TODO
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetConnectionInfoResponse {
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CometRequest {
    #[prost(oneof="comet_request::Request", tags="1")]
    pub request: ::core::option::Option<comet_request::Request>,
}
/// Nested message and enum types in `CometRequest`.
pub mod comet_request {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Request {
        #[prost(message, tag="1")]
        GetConnectionInfoRequest(super::GetConnectionInfoRequest),
    }
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CometResponse {
    #[prost(oneof="comet_response::Response", tags="1")]
    pub response: ::core::option::Option<comet_response::Response>,
}
/// Nested message and enum types in `CometResponse`.
pub mod comet_response {
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Response {
        #[prost(message, tag="1")]
        GetConnectionInfoResponse(super::GetConnectionInfoResponse),
    }
}
