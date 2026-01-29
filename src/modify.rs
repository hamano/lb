use async_trait::async_trait;
use clap::Args;
use ldap3::Mod;
use std::time::Instant;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Args, Clone)]
pub struct ModifyArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// Attribute to modify
    #[arg(long, default_value = "sn")]
    pub attr: String,

    /// Value to set
    #[arg(long, default_value = "modified")]
    pub value: String,
}

impl HasCommonArgs for ModifyArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct ModifyJob {
    base: BaseJob,
    args: ModifyArgs,
}

#[async_trait]
impl Job for ModifyJob {
    type Args = ModifyArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        ModifyJob {
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
        let dn = format!("cn={},{}", cn, self.args.common.base_dn);
        let mods = vec![Mod::Replace(
            &self.args.attr,
            [&self.args.value].into_iter().collect(),
        )];

        if let Some(ref mut ldap) = self.base.ldap {
            let start_time = Instant::now();
            let result = ldap.modify(&dn, mods).await;
            let duration = start_time.elapsed();

            self.base.record_latency(duration.as_micros() as u64);

            match result {
                Ok(res) => res.rc == 0,
                Err(e) => {
                    eprintln!("modify error: {}", e);
                    false
                }
            }
        } else {
            false
        }
    }
}
