//! Generate `*_proto` crates under `crates/proto/<domain>/` via prost-build.
//! Layout matches Bazel `rust_prost_library` (`crate::proto::<domain>::v1`).

mod domains;

use std::fs;
use std::path::{Path, PathBuf};
use std::process;

use domains::{Domain, DOMAINS};

fn main() {
    if let Err(err) = run() {
        eprintln!("gen-proto failed: {err}");
        process::exit(1);
    }
}

fn run() -> Result<(), Box<dyn std::error::Error>> {
    let crate_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let core_dir = crate_dir
        .parent()
        .and_then(|p| p.parent())
        .ok_or("clients/core")?;
    let repo_root = core_dir
        .parent()
        .and_then(|p| p.parent())
        .ok_or("repo root")?;
    let proto_root = repo_root.join("proto");
    let out_root = core_dir.join("crates/proto");

    if std::env::var_os("PROTOC").is_none() {
        let home_protoc =
            PathBuf::from(std::env::var_os("HOME").unwrap_or_default()).join(".local/bin/protoc");
        if home_protoc.is_file() {
            std::env::set_var("PROTOC", home_protoc);
        }
    }

    for domain in topological_order(DOMAINS)? {
        generate_domain(&proto_root, &out_root, domain)?;
        println!(
            "generated {} {} → crates/proto/{}",
            domain.name, domain.version, domain.name
        );
    }
    Ok(())
}

fn topological_order(domains: &[Domain]) -> Result<Vec<&Domain>, String> {
    let mut remaining: Vec<&Domain> = domains.iter().collect();
    let mut done = std::collections::HashSet::new();
    let mut ordered = Vec::with_capacity(domains.len());
    while !remaining.is_empty() {
        let (ready, blocked): (Vec<_>, Vec<_>) = remaining
            .into_iter()
            .partition(|d| d.deps.iter().all(|dep| done.contains(*dep)));
        if ready.is_empty() {
            return Err(format!(
                "proto domain cycle: {:?}",
                blocked.iter().map(|d| d.name).collect::<Vec<_>>()
            ));
        }
        for d in &ready {
            done.insert(d.name);
        }
        ordered.extend(ready);
        remaining = blocked;
    }
    Ok(ordered)
}

fn rust_mod(name: &str) -> String {
    name.into()
}

fn generate_domain(
    proto_root: &Path,
    out_root: &Path,
    domain: &Domain,
) -> Result<(), Box<dyn std::error::Error>> {
    let domain_dir = out_root.join(domain.name);
    let src_dir = domain_dir.join("src");
    fs::create_dir_all(&src_dir)?;

    let tmp = domain_dir.join(".prost-tmp");
    if tmp.exists() {
        fs::remove_dir_all(&tmp)?;
    }
    fs::create_dir_all(&tmp)?;

    let mut config = prost_build::Config::new();
    config.out_dir(&tmp);
    config.type_attribute(".", "#[derive(serde::Serialize, serde::Deserialize)]");
    config.message_attribute(".", "#[serde(default)]");

    for dep in domain.deps {
        let dependency = DOMAINS
            .iter()
            .find(|candidate| candidate.name == *dep)
            .ok_or_else(|| format!("unknown proto dependency: {dep}"))?;
        let path = format!(
            "::{dep}_proto::proto::{}::{}",
            rust_mod(dep),
            dependency.version
        );
        config.extern_path(format!(".proto.{dep}.{}", dependency.version), path);
    }

    let protos: Vec<PathBuf> = domain
        .srcs
        .iter()
        .map(|s| proto_root.join(domain.name).join(domain.version).join(s))
        .collect();
    for p in &protos {
        if !p.is_file() {
            return Err(format!("missing proto: {}", p.display()).into());
        }
    }

    let mut includes = vec![proto_root.to_path_buf()];
    for candidate in [
        PathBuf::from(std::env::var_os("HOME").unwrap_or_default())
            .join(".local/protoc-29.3/include"),
        PathBuf::from("/opt/homebrew/include"),
        PathBuf::from("/usr/local/include"),
        PathBuf::from("/usr/include"),
    ] {
        if candidate.join("google/protobuf/descriptor.proto").is_file() {
            includes.push(candidate);
            break;
        }
    }
    if let Ok(extra) = std::env::var("PROTO_INCLUDE") {
        includes.push(PathBuf::from(extra));
    }
    config.compile_protos(&protos, &includes)?;
    write_lib_rs(&src_dir, domain, &tmp)?;
    let _ = fs::remove_dir_all(&tmp);
    Ok(())
}

fn write_lib_rs(
    src_dir: &Path,
    domain: &Domain,
    tmp: &Path,
) -> Result<(), Box<dyn std::error::Error>> {
    let candidates = [format!("proto.{}.v1.rs", domain.name)];
    let generated = candidates.iter().map(|n| tmp.join(n)).find(|p| p.is_file());
    let Some(generated) = generated else {
        let entries: Vec<_> = fs::read_dir(tmp)?
            .filter_map(|e| e.ok())
            .map(|e| e.file_name().to_string_lossy().into_owned())
            .collect();
        return Err(format!("expected one of {candidates:?}, found: {entries:?}").into());
    };
    let body = fs::read_to_string(&generated)?;
    let mod_name = rust_mod(domain.name);
    let version = domain.version;
    let lib = format!(
        "//! Auto-generated by do_worker_proto_gen. Do not edit.\n\
         #![allow(clippy::all, dead_code, unused_imports, unused_variables)]\n\
         \n\
         pub mod proto {{\n\
             pub mod {mod_name} {{\n\
                 pub mod {version} {{\n\
         {body}\n\
                 }}\n\
             }}\n\
         }}\n"
    );
    fs::write(src_dir.join("lib.rs"), lib)?;
    Ok(())
}
