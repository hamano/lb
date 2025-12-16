use std::time::Instant;
use async_trait::async_trait;
use clap::Args;
use rand::Rng;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Args, Clone)]
pub struct BindArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// First ID for random DN generation
    #[arg(long, default_value_t = 1)]
    pub first: u32,

    /// Last ID for random DN generation (0 means no range)
    #[arg(long, default_value_t = 0)]
    pub last: u32,
}

impl HasCommonArgs for BindArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct BindJob {
    base: BaseJob,
    args: BindArgs,
    id_range: u32,
}

impl BindJob {
    fn generate_dn(&self) -> String {
        if self.id_range > 0 {
            let mut rng = rand::rng();
            let id = rng.random_range(self.args.first..=self.args.last);
            self.args.common.bind_dn
                .replace("%d", &id.to_string())
                .replace("%04d", &format!("{:04}", id))
                .replace("%03d", &format!("{:03}", id))
                .replace("%02d", &format!("{:02}", id))
        } else {
            self.args.common.bind_dn.clone()
        }
    }
}

#[async_trait]
impl Job for BindJob {
    type Args = BindArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        let id_range = if args.common.bind_dn.contains('%') && args.last > 0 {
            args.last - args.first + 1
        } else {
            0
        };

        BindJob {
            base: BaseJob::new(tid),
            args: args.clone(),
            id_range,
        }
    }

    fn args(&self) -> &Self::Args {
        &self.args
    }

    fn base(&mut self) -> &mut BaseJob {
        &mut self.base
    }

    async fn prepare(&mut self) -> bool {
        true
    }

    async fn request(&mut self) -> bool {
        let dn = self.generate_dn();

        if let Some(ref mut ldap) = self.base.ldap {
            let start_time = Instant::now();
            let result = ldap.simple_bind(&dn, &self.args.common.bind_pw).await;
            let duration = start_time.elapsed();

            self.base.record_latency(duration.as_micros() as u64);

            match result {
                Ok(res) => res.rc == 0,
                Err(_) => false,
            }
        } else {
            false
        }
    }
}
