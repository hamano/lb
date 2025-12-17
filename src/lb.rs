//! Common benchmark framework for LDAP operations

use std::sync::Arc;
use std::time::Instant;
use async_trait::async_trait;
use clap::Args;
use histogram::Histogram;
use ldap3::LdapConnAsync;
use tokio::sync::Barrier;
use tokio::sync::mpsc;

#[derive(Debug, Args, Clone)]
pub struct CommonArgs {
    /// Number of parallel tasks (default: 10)
    #[arg(short, long, global = true, default_value_t = 2)]
    pub concurrency: usize,
    /// Number of requests (default: 10)
    #[arg(short, long, global = true, default_value_t = 10)]
    pub number: usize,
    /// bind DN (default: "cn=Manager,dc=example,dc=com")
    #[arg(short = 'D', long, global = true, default_value_t = String::from("cn=Manager,dc=example,dc=com"))]
    pub bind_dn: String,
    /// bind PW (default: "secret")
    #[arg(short = 'w', long, global = true, default_value_t = String::from("secret"))]
    pub bind_pw: String,
    // URL (default: http://localhost:396)
    #[arg(index = 1, help = "")]
    pub url: String,
}

/// Trait for args that contain CommonArgs
pub trait HasCommonArgs {
    fn common(&self) -> &CommonArgs;
}

// Histogram configuration
pub const HISTOGRAM_GROUPING_POWER: u8 = 7;
pub const HISTOGRAM_MAX_VALUE_POWER: u8 = 32;

/// Result from a single task
#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct TaskResult {
    pub tid: usize,
    pub count: usize,
    pub success: usize,
    pub start_time: Instant,
    pub end_time: Instant,
    pub histogram_data: Vec<u64>,
}

impl TaskResult {
    #[allow(dead_code)]
    pub fn elapsed_time(&self) -> f64 {
        self.end_time.duration_since(self.start_time).as_secs_f64()
    }
}

/// Base structure for benchmark jobs
pub struct BaseJob {
    pub tid: usize,
    pub ldap: Option<ldap3::Ldap>,
    pub count: usize,
    pub success: usize,
    pub histogram: Histogram,
}

impl BaseJob {
    pub fn new(tid: usize) -> Self {
        BaseJob {
            tid,
            ldap: None,
            count: 0,
            success: 0,
            histogram: Histogram::new(HISTOGRAM_GROUPING_POWER, HISTOGRAM_MAX_VALUE_POWER)
                .expect("Failed to create histogram"),
        }
    }

    pub async fn connect(&mut self, url: &str) -> bool {
        match LdapConnAsync::new(url).await {
            Ok((conn, ldap)) => {
                ldap3::drive!(conn);
                self.ldap = Some(ldap);
                true
            }
            Err(e) => {
                eprintln!("task[{}]: failed to connect: {}", self.tid, e);
                false
            }
        }
    }

    pub async fn bind(&mut self, bind_dn: &str, bind_pw: &str) -> bool {
        if let Some(ref mut ldap) = self.ldap {
            match ldap.simple_bind(bind_dn, bind_pw).await {
                Ok(result) => {
                    if result.rc != 0 {
                        eprintln!("task[{}]: bind error: {}", self.tid, result.text);
                        return false;
                    }
                    true
                }
                Err(e) => {
                    eprintln!("task[{}]: bind error: {}", self.tid, e);
                    false
                }
            }
        } else {
            false
        }
    }

    pub async fn unbind(&mut self) {
        if let Some(ref mut ldap) = self.ldap {
            let _ = ldap.unbind().await;
        }
    }

    pub fn record_latency(&mut self, duration_us: u64) {
        let _ = self.histogram.increment(duration_us);
    }

    pub fn to_result(&self, start_time: Instant, end_time: Instant) -> TaskResult {
        TaskResult {
            tid: self.tid,
            count: self.count,
            success: self.success,
            start_time,
            end_time,
            histogram_data: self.histogram.as_slice().to_vec(),
        }
    }
}

/// Trait for benchmark jobs
#[async_trait]
pub trait Job: Send + 'static {
    type Args: HasCommonArgs + Clone + Send + Sync + 'static;

    fn new(tid: usize, args: &Self::Args) -> Self;
    fn args(&self) -> &Self::Args;
    fn base(&mut self) -> &mut BaseJob;
    async fn prepare(&mut self) -> bool;
    async fn request(&mut self) -> bool;

    async fn connect(&mut self) -> bool {
        let url = self.args().common().url.clone();
        self.base().connect(&url).await
    }

    async fn finish(&mut self) {
        self.base().unbind().await;
    }

    async fn run(&mut self) -> TaskResult {
        let common = self.args().common().clone();
        let num_per_task = (common.number + common.concurrency - 1) / common.concurrency;

        let start_time = Instant::now();

        for _ in 0..num_per_task {
            if self.request().await {
                self.base().success += 1;
            }
            self.base().count += 1;
        }

        let end_time = Instant::now();

        self.finish().await;
        self.base().to_result(start_time, end_time)
    }
}

/// Run benchmark with the given job type
pub async fn run_benchmark<J: Job>(args: &J::Args) -> Vec<TaskResult> {
    let common = args.common();
    let c = common.concurrency;
    let (tx, mut rx) = mpsc::channel(c);
    let barrier = Arc::new(Barrier::new(c));
    let args = Arc::new(args.clone());

    for i in 0..c {
        let barrier = barrier.clone();
        let tx = tx.clone();
        let args = args.clone();

        tokio::spawn(async move {
            let mut job = J::new(i, &args);
            
            if !job.connect().await {
                return;
            }
            if !job.prepare().await {
                return;
            }
            
            // Wait for all tasks to be ready
            barrier.wait().await;

            let result = job.run().await;
            let _ = tx.send(result).await;
        });
    }

    drop(tx);

    let mut results = Vec::new();
    while let Some(result) = rx.recv().await {
        results.push(result);
    }
    results
}

/// Run benchmark and print report
pub async fn run_job<J: Job>(args: &J::Args) {
    let common = args.common().clone();
    let results = run_benchmark::<J>(args).await;
    print_report(&results, common.concurrency, false);
}

/// Print benchmark report
pub fn print_report(results: &[TaskResult], concurrency: usize, show_histogram: bool) {
    if results.is_empty() {
        eprintln!("No results received");
        return;
    }

    // Find first start time and last end time
    let first_time = results.iter().map(|r| r.start_time).min().unwrap();
    let last_time = results.iter().map(|r| r.end_time).max().unwrap();
    let taken_time = last_time.duration_since(first_time).as_secs_f64();

    let total_request: usize = results.iter().map(|r| r.count).sum();
    let success_request: usize = results.iter().map(|r| r.success).sum();
    let success_rate = if total_request > 0 {
        (success_request as f64 / total_request as f64 * 100.0) as i32
    } else {
        0
    };

    let rpq = if taken_time > 0.0 {
        total_request as f64 / taken_time
    } else {
        0.0
    };

    let tpr = if total_request > 0 {
        concurrency as f64 * taken_time * 1000.0 / total_request as f64
    } else {
        0.0
    };
    let tpr_all = if total_request > 0 {
        taken_time * 1000.0 / total_request as f64
    } else {
        0.0
    };

    // Print benchmark results
    println!();
    println!("=== Benchmark Results ===");
    println!("Concurrency Level: {}", concurrency);
    println!("Total Requests: {}", total_request);
    println!("Success Requests: {}", success_request);
    println!("Success Rate: {}%", success_rate);
    println!("Time taken for tests: {:.3} seconds", taken_time);
    println!("Requests per second: {:.2} [#/sec] (mean)", rpq);
    println!("Time per request: {:.3} [ms] (mean)", tpr);
    println!(
        "Time per request: {:.3} [ms] (mean, across all concurrent requests)",
        tpr_all
    );

    if show_histogram {
        let merged_histogram = merge_histograms(results);
        println!();
        println!("=== Latency Histogram ===");
        print_histogram_stats(&merged_histogram);
    }
}

fn merge_histograms(results: &[TaskResult]) -> Histogram {
    let mut merged = Histogram::new(HISTOGRAM_GROUPING_POWER, HISTOGRAM_MAX_VALUE_POWER)
        .expect("Failed to create histogram");

    for result in results {
        if let Ok(worker_hist) = Histogram::from_buckets(
            HISTOGRAM_GROUPING_POWER,
            HISTOGRAM_MAX_VALUE_POWER,
            result.histogram_data.clone(),
        ) {
            if let Ok(new_merged) = merged.checked_add(&worker_hist) {
                merged = new_merged;
            }
        }
    }

    merged
}

fn print_histogram_stats(histogram: &Histogram) {
    let total: u64 = histogram.as_slice().iter().sum();
    println!("Total samples: {}", total);

    if total == 0 {
        println!("No latency data collected");
        return;
    }

    // Print percentiles
    let percentiles = [50.0, 75.0, 90.0, 95.0, 99.0, 99.9];
    println!();
    println!("Percentile Latencies:");
    for p in percentiles {
        if let Ok(Some(bucket)) = histogram.percentile(p) {
            let mid = (bucket.start() + bucket.end()) / 2;
            let (val, unit) = format_latency(mid);
            println!(
                "  p{:>5.1}: {:>8} {}",
                p, val, unit
            );
        }
    }

    println!();
    println!("Latency Distribution:");
    print_histogram_bars(histogram, total);
}

fn print_histogram_bars(histogram: &Histogram, total: u64) {
    const MAX_BAR_WIDTH: usize = 40;
    const MAX_ROWS: usize = 10;

    let mut buckets: Vec<(u64, u64, u64)> = Vec::new();
    for bucket in histogram.iter() {
        if bucket.count() > 0 {
            buckets.push((bucket.start(), bucket.end(), bucket.count()));
        }
    }

    if buckets.is_empty() {
        return;
    }

    let display_buckets: Vec<(u64, u64, u64)> = if buckets.len() > MAX_ROWS {
        let buckets_per_group = (buckets.len() + MAX_ROWS - 1) / MAX_ROWS;
        buckets
            .chunks(buckets_per_group)
            .map(|chunk| {
                let start = chunk.first().unwrap().0;
                let end = chunk.last().unwrap().1;
                let count: u64 = chunk.iter().map(|(_, _, c)| c).sum();
                (start, end, count)
            })
            .collect()
    } else {
        buckets
    };

    let max_count = display_buckets.iter().map(|(_, _, c)| *c).max().unwrap_or(1);

    for (start, end, count) in display_buckets {
        let percentage = count as f64 / total as f64 * 100.0;
        let bar_width = (count as f64 / max_count as f64 * MAX_BAR_WIDTH as f64) as usize;
        let bar: String = "#".repeat(bar_width);

        let (start_val, end_val, unit) = format_latency_range(start, end);

        if start_val == end_val {
            println!(
                "  {:>8} {:>2}: {:>6} ({:>5.1}%) {}",
                start_val, unit, count, percentage, bar
            );
        } else {
            println!(
                "  {:>8}-{:<8} {:>2}: {:>6} ({:>5.1}%) {}",
                start_val, end_val, unit, count, percentage, bar
            );
        }
    }
}

fn format_latency(us: u64) -> (String, &'static str) {
    if us >= 1_000_000 {
        (format!("{:.2}", us as f64 / 1_000_000.0), "s")
    } else if us >= 1_000 {
        (format!("{:.2}", us as f64 / 1_000.0), "ms")
    } else {
        (us.to_string(), "us")
    }
}

fn format_latency_range(start_us: u64, end_us: u64) -> (String, String, &'static str) {
    if end_us >= 1_000_000 {
        (
            format!("{:.2}", start_us as f64 / 1_000_000.0),
            format!("{:.2}", end_us as f64 / 1_000_000.0),
            "s",
        )
    } else if end_us >= 1_000 {
        (
            format!("{:.2}", start_us as f64 / 1_000.0),
            format!("{:.2}", end_us as f64 / 1_000.0),
            "ms",
        )
    } else {
        (start_us.to_string(), end_us.to_string(), "us")
    }
}
