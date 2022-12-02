use base32;
use byteorder::{ByteOrder, BE};
use std::cmp::Ordering;
use std::fmt;
use std::fmt::Formatter;
use std::net::{Ipv4Addr, Ipv6Addr, SocketAddr, SocketAddrV4, SocketAddrV6};
use std::str::FromStr;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

const B32_ALPHABET: base32::Alphabet = base32::Alphabet::Crockford;

#[non_exhaustive]
pub struct MachineIdType;

impl MachineIdType {
    pub const RANDOM: u8 = 0;
    pub const ADDRESS_V4: u8 = 1;
    pub const ADDRESS_V6: u8 = 2;
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub enum MachineId {
    Random([u8; 18]),
    AddressV4(SocketAddrV4),
    AddressV6(SocketAddrV6),
}

impl Into<MachineId> for SocketAddr {
    fn into(self) -> MachineId {
        match self {
            SocketAddr::V4(v4) => MachineId::AddressV4(v4),
            SocketAddr::V6(v6) => MachineId::AddressV6(v6),
        }
    }
}

#[derive(Clone, Copy, Debug, PartialEq)]
pub struct ConnectionId {
    pub(crate) machine_id: MachineId,
    pub(crate) msec: u64,
    pub(crate) incr: u16,
    pub(crate) rand: u16,
    pub(crate) version: u8,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub struct ShortConnectionId {
    repr: [u8; 10],
}

impl Into<ShortConnectionId> for ConnectionId {
    fn into(self) -> ShortConnectionId {
        self.to_short()
    }
}

impl ToString for ConnectionId {
    fn to_string(&self) -> String {
        let mut machine_id_type: u8;
        let mut machine_id_buf = [0u8; 18];
        match self.machine_id {
            MachineId::Random(rand) => {
                machine_id_type = MachineIdType::RANDOM;
                machine_id_buf[..].copy_from_slice(&rand);
            }
            MachineId::AddressV4(addr_v4) => {
                machine_id_type = MachineIdType::ADDRESS_V4;
                BE::write_u32(&mut machine_id_buf[..4], (*addr_v4.ip()).into());
                BE::write_u16(&mut machine_id_buf[4..6], addr_v4.port());
            }
            MachineId::AddressV6(addr_v6) => {
                machine_id_type = MachineIdType::ADDRESS_V6;
                BE::write_u128(&mut machine_id_buf[..16], (*addr_v6.ip()).into());
                BE::write_u16(&mut machine_id_buf[16..18], addr_v6.port());
            }
        }

        let mut buf = [0u8; 28];
        BE::write_u48(&mut buf[..6], (self.msec << 2) | machine_id_type as u64);
        buf[6..24].copy_from_slice(&mut machine_id_buf);
        BE::write_u16(&mut buf[24..26], self.incr);
        BE::write_u16(&mut buf[26..28], (self.rand << 2) | self.version as u16);

        base32::encode(B32_ALPHABET, &buf)
    }
}

impl FromStr for ConnectionId {
    type Err = super::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        if s.len() != 45 {
            return Err(super::Error::new("invalid connection id"));
        }
        if let Some(buf) = base32::decode(B32_ALPHABET, s) {
            let (msec, machine_id_type) = {
                let tmp = BE::read_u48(&buf[..6]);
                (tmp >> 2, (tmp & 0x3) as u8)
            };
            let machine_id = match machine_id_type {
                MachineIdType::RANDOM => {
                    let mut rand = [0u8; 18];
                    rand[..].copy_from_slice(&buf[6..24]);
                    MachineId::Random(rand)
                }
                MachineIdType::ADDRESS_V4 => {
                    let ip = BE::read_u32(&buf[6..10]);
                    let port = BE::read_u16(&buf[10..12]);
                    MachineId::AddressV4(SocketAddrV4::new(Ipv4Addr::from(ip), port))
                }
                MachineIdType::ADDRESS_V6 => {
                    let ip = BE::read_u128(&buf[6..22]);
                    let port = BE::read_u16(&buf[22..24]);
                    MachineId::AddressV6(SocketAddrV6::new(Ipv6Addr::from(ip), port, 0, 0))
                }
                _ => {
                    return Err(super::Error::new("invalid connection id"));
                }
            };
            let incr = BE::read_u16(&buf[24..26]);
            let (rand, version) = {
                let tmp = BE::read_u16(&buf[26..28]);
                (tmp >> 2, (tmp & 0x3) as u8)
            };

            return Ok(ConnectionId {
                machine_id,
                msec,
                incr,
                rand,
                version,
            });
        }

        Err(super::Error::new("invalid connection id"))
    }
}

impl PartialOrd for ConnectionId {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        PartialOrd::<String>::partial_cmp(&self.to_string(), &other.to_string())
    }
}

impl ConnectionId {
    pub fn get_time(self) -> SystemTime {
        UNIX_EPOCH + Duration::from_millis(self.msec)
    }

    pub fn get_socket_addr(self) -> Option<SocketAddr> {
        match self.machine_id {
            MachineId::AddressV4(addr) => Some(SocketAddr::V4(addr)),
            MachineId::AddressV6(addr) => Some(SocketAddr::V6(addr)),
            _ => None,
        }
    }

    pub fn get_version(self) -> u8 {
        self.version
    }

    pub fn to_short(self) -> ShortConnectionId {
        let mut repr = [0u8; 10];
        BE::write_u48(&mut repr[..6], self.msec);
        BE::write_u16(&mut repr[6..8], self.incr);
        BE::write_u16(&mut repr[8..10], (self.rand << 2) | self.version as u16);
        ShortConnectionId { repr }
    }

    pub fn from_short(short: ShortConnectionId, machine_id: MachineId) -> Self {
        let msec = BE::read_u48(&short.repr[..6]);
        let incr = BE::read_u16(&short.repr[6..8]);
        let (rand, version) = {
            let tmp = BE::read_u16(&short.repr[8..10]);
            (tmp >> 2, (tmp & 0x3) as u8)
        };
        ConnectionId {
            machine_id,
            msec,
            incr,
            rand,
            version,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::DateTime;

    fn mock_connection_id() -> ConnectionId {
        let datetime = "2021-07-20T00:40:39.789456+08:00";
        let addr_v6 = "[fdbd:dc02:ff:1:2:225:137:157]:12345";
        ConnectionId {
            msec: DateTime::parse_from_rfc3339(datetime)
                .unwrap()
                .timestamp_millis() as u64,
            machine_id: MachineId::AddressV6(addr_v6.parse().unwrap()),
            incr: 12345,
            rand: 2345,
            version: 0,
        }
    }

    #[test]
    fn test_connection_id_to_string() {
        let id = mock_connection_id();
        println!("{:?}", id);
        assert_eq!(
            id.to_string().as_str(),
            "0QNFX42SPVYVVQ0203ZG00800812A09Q05BK0E9G74JA8"
        )
    }

    #[test]
    fn test_connection_id_to_short() {
        let id = mock_connection_id();
        let short_id = id.to_short();
        let want: [u8; 10] = [1, 122, 191, 164, 22, 109, 48, 57, 36, 164];
        println!("{:?}", id);
        assert_eq!(short_id, ShortConnectionId { repr: want })
    }

    #[test]
    fn test_connection_id_parse() {
        let id = mock_connection_id();
        let id_str = id.to_string().as_str();
        let got: ConnectionId = "0QNFX42SPVYVVQ0203ZG00800812A09Q05BK0E9G74JA8"
            .parse()
            .unwrap();
        assert_eq!(got, id)
    }
}
