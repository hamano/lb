use std::time::Instant;
use async_trait::async_trait;
use clap::Args;
use rand::Rng;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Args, Clone)]
pub struct BindArgs {
    #[command(flatten)]
    pub common: CommonArgs,
}

impl HasCommonArgs for BindArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct BindJob {
    base: BaseJob,
    args: BindArgs,
}

impl BindJob {
    fn generate_dn(&self) -> String {
        let mut rng = rand::rng();
        let id = rng.random_range(0..=self.base.count);
        let base_dn = self
            .args
            .common
            .bind_dn
            .split_once(',')
            .map(|x| x.1)
            .unwrap_or(self.args.common.bind_dn.as_str());

        format!("cn={}-{},{}", self.base.tid, id, base_dn)
    }
}

#[async_trait]
impl Job for BindJob {
    type Args = BindArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        BindJob {
            base: BaseJob::new(tid),
            args: args.clone(),
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
