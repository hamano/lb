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

    #[command(flatten)]
    pub base: BaseArgs,
}

#[derive(Debug, Args)]
pub struct SetupPersonArgs {
    #[command(flatten)]
    pub common: CommonArgs,

    /// Quiet mode
    #[arg(short = 'q', long)]
    pub quiet: bool,

    #[command(flatten)]
    pub person: PersonArgs,
}

#[derive(Debug, Args)]
pub struct BaseArgs {
    /// Base DN
    #[arg(short = 'b', long, default_value = "dc=example,dc=com")]
    pub base_dn: String,
}

#[derive(Debug, Args)]
pub struct PersonArgs {
    /// Base DN for entries
    #[arg(short = 'b', long, default_value = "dc=example,dc=com")]
    pub base_dn: String,

    /// cn attribute prefix
    #[arg(long, default_value = "user")]
    pub cn: String,

    /// sn attribute (defaults to cn if not specified)
    #[arg(long)]
    pub sn: Option<String>,

    /// userPassword attribute
    #[arg(long, default_value = "secret")]
    pub password: String,

    /// First ID number
    #[arg(long, default_value_t = 1)]
    pub first: u32,

    /// Last ID number (0 means single entry without number suffix)
    #[arg(long, default_value_t = 0)]
    pub last: u32,
}

pub async fn run_base(args: &SetupBaseArgs) {
    setup_base(&args.common, args.quiet, &args.base).await;
}

pub async fn run_person(args: &SetupPersonArgs) {
    setup_person(&args.common, args.quiet, &args.person).await;
}

async fn setup_base(common: &CommonArgs, quiet: bool, base_args: &BaseArgs) {
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
        println!("Adding base entry: {}", base_args.base_dn);
    }

    // Extract dc value from base_dn (e.g., "dc=example,dc=com" -> "example")
    let dc_value = base_args
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

    match ldap.add(&base_args.base_dn, attrs).await {
        Ok(result) => {
            if result.rc == 0 {
                if !quiet {
                    println!("Added base entry: {}", base_args.base_dn);
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

async fn setup_person(common: &CommonArgs, quiet: bool, person_args: &PersonArgs) {
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

    if person_args.last > 0 {
        for i in person_args.first..=person_args.last {
            let cn = if person_args.cn.contains('%') {
                // Format string support (e.g., "user%04d" -> "user0001")
                person_args
                    .cn
                    .replace("%d", &i.to_string())
                    .replace("%04d", &format!("{:04}", i))
                    .replace("%03d", &format!("{:03}", i))
                    .replace("%02d", &format!("{:02}", i))
            } else {
                format!("{}{}", person_args.cn, i)
            };
            setup_one_person(quiet, &mut ldap, &cn, person_args).await;
        }
    } else {
        setup_one_person(quiet, &mut ldap, &person_args.cn, person_args).await;
    }

    let _ = ldap.unbind().await;
}

async fn setup_one_person(quiet: bool, ldap: &mut ldap3::Ldap, cn: &str, person_args: &PersonArgs) {
    let sn = person_args.sn.as_deref().unwrap_or(cn);
    let dn = format!("cn={},{}", cn, person_args.base_dn);

    if !quiet {
        println!("Adding person entry: {}", dn);
    }

    let attrs = vec![
        ("objectClass", vec!["person"].into_iter().collect()),
        ("cn", vec![cn].into_iter().collect()),
        ("sn", vec![sn].into_iter().collect()),
        (
            "userPassword",
            vec![person_args.password.as_str()].into_iter().collect(),
        ),
    ];

    match ldap.add(&dn, attrs).await {
        Ok(result) => {
            if result.rc == 0 {
                if !quiet {
                    println!("Added person entry: {}", dn);
                }
            } else {
                eprintln!("Add error for {}: {} (rc={})", dn, result.text, result.rc);
            }
        }
        Err(e) => {
            eprintln!("Add error for {}: {}", dn, e);
        }
    }
}
