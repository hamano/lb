use async_trait::async_trait;
use clap::Args;
use rand::Rng;
use std::time::Instant;

use crate::lb::{BaseJob, CommonArgs, HasCommonArgs, Job};

#[derive(Debug, Args, Clone)]
pub struct TestArgs {
    #[command(flatten)]
    pub common: CommonArgs,
}

impl HasCommonArgs for TestArgs {
    fn common(&self) -> &CommonArgs {
        &self.common
    }
}

pub struct TestJob {
    base: BaseJob,
    args: TestArgs,
}

#[async_trait]
impl Job for TestJob {
    type Args = TestArgs;

    fn new(tid: usize, args: &Self::Args) -> Self {
        TestJob {
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

    async fn connect(&mut self) -> bool {
        true
    }

    async fn prepare(&mut self) -> bool {
        true
    }

    async fn request(&mut self) -> bool {
        let wait_ms = {
            let mut rng = rand::rng();
            rng.random_range(0..=100)
        };

        let start_time = Instant::now();
        tokio::time::sleep(std::time::Duration::from_millis(wait_ms)).await;
        let duration = start_time.elapsed();

        self.base.record_latency(duration.as_micros() as u64);
        true
    }
}
