use async_trait::async_trait;
use clap::Args;
use std::time::Instant;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Args, Clone)]
pub struct DeleteArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// Base DN for entries
    #[arg(short = 'b', long, default_value = "dc=example,dc=com")]
    pub base_dn: String,
}

impl HasCommonArgs for DeleteArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct DeleteJob {
    base: BaseJob,
    args: DeleteArgs,
}

#[async_trait]
impl Job for DeleteJob {
    type Args = DeleteArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        DeleteJob {
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
        let common = self.args.common();
        self.base.bind(&common.bind_dn, &common.bind_pw).await
    }

    async fn request(&mut self) -> bool {
        let cn = self.base.start_index + self.base.count;
        let dn = format!("cn={},{}", cn, self.args.base_dn);

        if let Some(ref mut ldap) = self.base.ldap {
            let start_time = Instant::now();
            let result = ldap.delete(&dn).await;
            let duration = start_time.elapsed();

            self.base.record_latency(duration.as_micros() as u64);

            match result {
                Ok(res) => res.rc == 0,
                Err(e) => {
                    eprintln!("delete error: {}", e);
                    false
                }
            }
        } else {
            false
        }
    }
}
