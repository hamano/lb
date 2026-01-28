use async_trait::async_trait;
use clap::Args;
use std::time::Instant;
use uuid::Uuid;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Args, Clone)]
pub struct AddArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// Use UUID for cn attribute
    #[arg(long)]
    pub uuid: bool,

    /// Base DN for entries
    #[arg(short = 'b', long, default_value = "dc=example,dc=com")]
    pub base_dn: String,
}

impl HasCommonArgs for AddArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct AddJob {
    base: BaseJob,
    args: AddArgs,
}

#[async_trait]
impl Job for AddJob {
    type Args = AddArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        AddJob {
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
        let cn = if self.args.uuid {
            Uuid::new_v4().to_string()
        } else {
            format!("{}", self.base.start_index + self.base.count)
        };

        let dn = format!("cn={},{}", cn, self.args.base_dn);
        let tid_str = self.base.tid.to_string();

        let attrs = vec![
            ("objectClass", vec!["person"].into_iter().collect()),
            ("cn", vec![cn.as_str()].into_iter().collect()),
            ("sn", vec![tid_str.as_str()].into_iter().collect()),
            ("userPassword", vec!["secret"].into_iter().collect()),
        ];

        if let Some(ref mut ldap) = self.base.ldap {
            let start_time = Instant::now();
            let result = ldap.add(&dn, attrs).await;
            let duration = start_time.elapsed();

            self.base.record_latency(duration.as_micros() as u64);

            match result {
                Ok(res) => res.rc == 0,
                Err(e) => {
                    eprintln!("add error: {}", e);
                    false
                }
            }
        } else {
            false
        }
    }
}
