//! Connection ID
//! - 6 字节 毫秒时间戳 + 机器ID类型标记
//!   - 高 46位 毫秒时间戳
//!   - 最低 2位 机器ID类型标记
//! - 机器ID
//!   - 18 字节随机数，或
//!   - 16 字节 IPv4 / IPv6 地址 + 2 字节 port
//! - 2 字节 counter，从一个随机数开始
//! - 2 字节 随机数 + 版本号
//!   - 高 14位 随机数
//!   - 最低 2位 版本号
//!   - (后续可扩展版本，从随机数腾出位置携带其他数据)
//! - 共计 28字节，base32 编码长度为 45字节
//! - 不带机器ID 10字节，base32 编码长度为 16字节

use std::fmt;
use std::fmt::Formatter;

pub use connection_id::{ConnectionId, MachineId, MachineIdType, ShortConnectionId};
pub use generator::Generator;

pub mod connection_id;
pub mod generator;

#[derive(Clone, Debug)]
pub struct Error {
    detail: String,
}

impl Error {
    fn new(detail: &str) -> Self {
        Error {
            detail: detail.to_string(),
        }
    }
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.detail)
    }
}
