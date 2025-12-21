use std::time::Instant;
use async_trait::async_trait;
use clap::{Args, ValueEnum};
use ldap3::Scope;
use rand::Rng;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Clone, Copy, ValueEnum)]
pub enum SearchScope {
    Base,
    One,
    Sub,
    Children,
}

impl From<SearchScope> for Scope {
    fn from(s: SearchScope) -> Self {
        match s {
            SearchScope::Base => Scope::Base,
            SearchScope::One => Scope::OneLevel,
            SearchScope::Sub => Scope::Subtree,
            SearchScope::Children => Scope::Subtree, // ldap3 doesn't have Children
        }
    }
}

#[derive(Debug, Args, Clone)]
pub struct SearchArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// Base DN for search
    #[arg(short = 'b', long, default_value = "dc=example,dc=com")]
    pub base_dn: String,

    /// Search scope
    #[arg(short = 's', long, value_enum, default_value = "sub")]
    pub scope: SearchScope,

    /// Search filter (use %d for random id substitution)
    #[arg(short = 'f', long, default_value = "(objectClass=*)")]
    pub filter: String,

    /// Attributes to retrieve (comma-separated)
    #[arg(short = 'a', long, default_value = "dn")]
    pub attributes: String,

    /// First ID for random filter generation
    #[arg(long, default_value_t = 1)]
    pub first: u32,

    /// Last ID for random filter generation (0 means no range)
    #[arg(long, default_value_t = 0)]
    pub last: u32,
}

impl HasCommonArgs for SearchArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct SearchJob {
    base: BaseJob,
    args: SearchArgs,
    id_range: u32,
    attrs: Vec<String>,
}

impl SearchJob {
    fn generate_filter(&self) -> String {
        if self.id_range > 0 {
            let id = {
                let mut rng = rand::rng();
                rng.random_range(self.args.first..=self.args.last)
            };
            self.args.filter
                .replace("%d", &id.to_string())
                .replace("%04d", &format!("{:04}", id))
                .replace("%03d", &format!("{:03}", id))
                .replace("%02d", &format!("{:02}", id))
        } else {
            self.args.filter.clone()
        }
    }
}

#[async_trait]
impl Job for SearchJob {
    type Args = SearchArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        let id_range = if args.filter.contains('%') && args.last > 0 {
            args.last - args.first + 1
        } else {
            0
        };
        let attrs: Vec<String> = args.attributes
            .split(',')
            .map(|s| s.trim().to_string())
            .collect();

        SearchJob {
            base: BaseJob::new(tid),
            args: args.clone(),
            id_range,
            attrs,
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
        let filter = self.generate_filter();
        let scope: Scope = self.args.scope.into();
        let attrs: Vec<&str> = self.attrs.iter().map(|s| s.as_str()).collect();

        if let Some(ref mut ldap) = self.base.ldap {
            let start_time = Instant::now();
            let result = ldap.search(&self.args.base_dn, scope, &filter, attrs).await;
            let duration = start_time.elapsed();

            self.base.record_latency(duration.as_micros() as u64);

            match result {
                Ok(result) => result.0.len() > 0,
                Err(_) => false,
            }
        } else {
            false
        }
    }
}
