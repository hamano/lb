use async_trait::async_trait;
use clap::Args;
use ldap3::exop::PasswordModify;
use std::time::Instant;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Args, Clone)]
pub struct PassmodArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// New password to set
    #[arg(long, default_value = "newsecret")]
    pub new_password: String,

    /// Old password (if not provided, uses bind password)
    #[arg(long)]
    pub old_password: Option<String>,
}

impl HasCommonArgs for PassmodArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct PassmodJob {
    base: BaseJob,
    args: PassmodArgs,
}

#[async_trait]
impl Job for PassmodJob {
    type Args = PassmodArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        PassmodJob {
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
        let user_dn = format!("cn={},{}", cn, self.args.common.base_dn);
        
        let old_pw = self.args.old_password.as_deref()
            .unwrap_or(&self.args.common.bind_pw);

        if let Some(ref mut ldap) = self.base.ldap {
            let start_time = Instant::now();
            
            let exop = PasswordModify {
                user_id: Some(&user_dn),
                old_pass: Some(old_pw),
                new_pass: Some(&self.args.new_password),
            };
            
            let result = ldap.extended(exop).await;
            let duration = start_time.elapsed();

            self.base.record_latency(duration.as_micros() as u64);

            match result {
                Ok(res) => res.1.rc == 0,
                Err(e) => {
                    eprintln!("passmod error: {}", e);
                    false
                }
            }
        } else {
            false
        }
    }
}
