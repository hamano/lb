use async_trait::async_trait;
use clap::{Args, ValueEnum};
use ldap3::Scope;
use rand::Rng;
use std::time::Instant;

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
}

impl HasCommonArgs for SearchArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct SearchJob {
    base: BaseJob,
    args: SearchArgs,
    attrs: Vec<String>,
}

impl SearchJob {
    fn generate_filter(&self) -> String {
        let mut rng = rand::rng();
        let id = rng.random_range(0..=self.base.count);
        format!("(cn={}-{})", self.base.tid, id)
    }
}

#[async_trait]
impl Job for SearchJob {
    type Args = SearchArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        let attrs: Vec<String> = args
            .attributes
            .split(',')
            .map(|s| s.trim().to_string())
            .collect();

        SearchJob {
            base: BaseJob::new(tid),
            args: args.clone(),
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
                Ok(result) => !result.0.is_empty(),
                Err(_) => false,
            }
        } else {
            false
        }
    }
}
