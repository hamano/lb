use clap::Args;
use ldap3::LdapConnAsync;

use crate::lb::CommonArgs;

#[derive(Debug, Args)]
pub struct SetupBaseArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// Quiet mode
    #[arg(short = 'q', long)]
    pub quiet: bool,
}

pub async fn run_base(args: &SetupBaseArgs) {
    setup_base(&args.common, args.quiet).await;
}

async fn setup_base(common: &CommonArgs, quiet: bool) {
    let (conn, mut ldap) = match LdapConnAsync::new(&common.url).await {
        Ok(result) => result,
        Err(e) => {
            eprintln!("Failed to connect: {}", e);
            return;
        }
    };
    ldap3::drive!(conn);

    if let Err(e) = ldap.simple_bind(&common.bind_dn, &common.bind_pw).await {
        eprintln!("Bind error: {}", e);
        return;
    }

    if !quiet {
        println!("Adding base entry: {}", common.base_dn);
    }

    // Extract dc value from base_dn (e.g., "dc=example,dc=com" -> "example")
    let dc_value = common
        .base_dn
        .split(',')
        .next()
        .and_then(|s| s.strip_prefix("dc="))
        .unwrap_or("example");

    let attrs = vec![
        (
            "objectClass",
            vec!["dcObject", "organization"].into_iter().collect(),
        ),
        ("o", vec!["lb"].into_iter().collect()),
        ("dc", vec![dc_value].into_iter().collect()),
    ];

    match ldap.add(&common.base_dn, attrs).await {
        Ok(result) => {
            if result.rc == 0 {
                if !quiet {
                    println!("Added base entry: {}", common.base_dn);
                }
            } else {
                eprintln!("Add error: {} (rc={})", result.text, result.rc);
            }
        }
        Err(e) => {
            eprintln!("Add error: {}", e);
        }
    }

    let _ = ldap.unbind().await;
}
