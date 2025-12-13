use clap::{Parser, Subcommand};

mod lb;
mod bind;
mod add;
mod search;
mod modify;
mod delete;
mod test;
mod setup;

use lb::run_job;

#[derive(Debug, Parser)]
#[command(version, about = "LDAP Benchmarking Tool")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
    #[arg(short, long, global = true, action = clap::ArgAction::Count)]
    verbose: u8,
}

#[derive(Debug, Subcommand)]
enum Commands {
    #[command(about = "LDAP BIND Benchmarking")]
    Bind(bind::BindArgs),
    #[command(about = "LDAP ADD Benchmarking")]
    Add(add::AddArgs),
    #[command(about = "LDAP SEARCH Benchmarking")]
    Search(search::SearchArgs),
    #[command(about = "LDAP MODIFY Benchmarking")]
    Modify(modify::ModifyArgs),
    #[command(about = "LDAP DELETE Benchmarking")]
    Delete(delete::DeleteArgs),
    #[command(about = "Dummy benchmark for testing")]
    Test(test::TestArgs),
    #[command(about = "Setup base entry")]
    Base(setup::SetupBaseArgs),
    #[command(about = "Setup person entries")]
    Person(setup::SetupPersonArgs),
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    match cli.command {
        Commands::Bind(args) => {
            run_job::<bind::BindJob>(&args).await;
        }
        Commands::Add(args) => {
            run_job::<add::AddJob>(&args).await;
        }
        Commands::Search(args) => {
            run_job::<search::SearchJob>(&args).await;
        }
        Commands::Modify(args) => {
            run_job::<modify::ModifyJob>(&args).await;
        }
        Commands::Delete(args) => {
            run_job::<delete::DeleteJob>(&args).await;
        }
        Commands::Test(args) => {
            run_job::<test::TestJob>(&args).await;
        }
        Commands::Base(args) => {
            setup::run_base(&args).await;
        }
        Commands::Person(args) => {
            setup::run_person(&args).await;
        }
    }
}
