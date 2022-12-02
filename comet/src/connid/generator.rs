use byteorder::BigEndian;
use rand::prelude::*;
use std::fmt;
use std::fmt::Formatter;
use std::sync::atomic::{AtomicU16, Ordering};
use std::time::{SystemTime, UNIX_EPOCH};

use super::{ConnectionId, MachineId, ShortConnectionId};

pub struct Generator {
    machine_id: MachineId,
    counter: AtomicU16,
    version: u8,
}

impl Generator {
    pub fn new() -> Self {
        let mut machine_id = [0u8; 18];
        thread_rng().fill_bytes(&mut machine_id);
        Self::new_with_machine_id(MachineId::Random(machine_id))
    }

    pub fn new_with_machine_id(machine_id: impl Into<MachineId>) -> Self {
        let mut buf = [0u8; 2];
        thread_rng().fill_bytes(&mut buf);
        let counter = (buf[0] as u16) << 8 | (buf[1] as u16);

        Generator {
            machine_id: machine_id.into(),
            counter: AtomicU16::new(counter),
            version: 0,
        }
    }

    pub fn generate(&self) -> ConnectionId {
        self.generate_with_time(SystemTime::now())
    }

    fn generate_with_time(&self, t: SystemTime) -> ConnectionId {
        let msec = t.duration_since(UNIX_EPOCH).unwrap().as_millis() as u64;
        let counter = self.counter.fetch_add(1, Ordering::SeqCst);
        let rand = thread_rng().next_u32() as u16 & ((1 << 14) - 1);

        ConnectionId {
            msec,
            machine_id: self.machine_id,
            incr: counter,
            rand,
            version: 0,
        }
    }

    pub fn restore_from_short(&self, short: ShortConnectionId) -> ConnectionId {
        ConnectionId::from_short(short, self.machine_id)
    }
}

impl fmt::Debug for Generator {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "Generator{{machine_id: {:?}, counter: {:?}, version: {:?}}}",
            self.machine_id, self.counter, self.version
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generator_basic() {
        let total = 1e6 as u32;
        let g = Generator::new();

        let mut prev_id = g.generate();
        let mut prev_incr = 0u32;

        for i in 0..total {
            let id = g.generate();
            if i > 0 {
                assert_eq!(id.incr, (prev_incr + 1) as u16);
            }
            if id.incr == 0 {
                assert_eq!(prev_incr, 0xFFFF);
            } else {
                assert!(
                    prev_id < id,
                    "{} ({:?}) != {} ({:?}) {}",
                    prev_id.to_string(),
                    prev_id,
                    id.to_string(),
                    id,
                    i
                );
            }

            prev_id = id;
            prev_incr = (id.incr as u32) & 0xFFFF;

            assert_eq!(id.to_string().len(), 45);
            assert_eq!(id.machine_id, g.machine_id);
        }
    }

    #[test]
    fn test_generator_restore_from_short_id() {
        let g = Generator::new();
        let id = g.generate();

        let short = id.to_short();
        let got = g.restore_from_short(short);

        assert_eq!(id, got);
    }
}
