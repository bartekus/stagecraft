// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::{collections::BTreeMap, env, fs, path::Path, path::PathBuf};
use walkdir::WalkDir;

#[derive(Serialize, Deserialize)]
struct FileEntry {
    path: String,
    size: u64,
    lines: usize,
    ext: String,
}
#[derive(Serialize, Deserialize)]
struct ModuleHit {
    kind: String,
    file: String,
    name: Option<String>,
    module: Option<String>,
}
#[derive(Serialize, Deserialize)]
struct Index {
    root: String,
    indexedAt: String,
    files: Vec<FileEntry>,
    languages: BTreeMap<String, u64>,
    topDirs: BTreeMap<String, u64>,
    moduleFiles: Vec<ModuleHit>,
    digest: String,
}

fn load_ignores(root: &Path) -> Vec<String> {
    let mut v = vec![
        ".git".into(),"node_modules".into(),"dist".into(),"build".into(),"out".into(),
        "target".into(),"vendor".into(),".cache".into(),".tmp".into(),"coverage".into()
    ];
    let p = root.join("tools/context-compiler/xray/ignore.rules");
    if let Ok(t) = fs::read_to_string(p) {
        v.extend(t.lines().map(|s| s.trim().to_string()).filter(|s| !s.is_empty()));
    }
    v
}
fn should_ignore(rel: &str, ignores: &Vec<String>) -> bool {
    rel.split(std::path::MAIN_SEPARATOR)
        .any(|part| ignores.iter().any(|r| r == part || (r.ends_with('*') && part.starts_with(&r[..r.len()-1]))))
}

fn count_lines(p: &Path) -> usize {
    fs::read_to_string(p).map(|t| t.lines().count()).unwrap_or(0)
}

fn main() {
    let root = env::args().nth(1).unwrap_or_else(|| ".".into());
    let rootp = fs::canonicalize(&root).expect("root");
    let ignores = load_ignores(&rootp);

    let mut files: Vec<FileEntry> = Vec::new();
    for e in WalkDir::new(&rootp).into_iter().filter_map(Result::ok) {
        let p = e.path();
        if p.is_dir() {
            // handled via should_ignore on rel path
            continue;
        }
        if p.is_file() {
            let rel = p.strip_prefix(&rootp).unwrap().to_string_lossy().replace(std::path::MAIN_SEPARATOR, "/");
            if should_ignore(&rel, &ignores) { continue; }
            if let Ok(md) = fs::metadata(p) {
                let ext = p.extension().and_then(|s| s.to_str()).unwrap_or("").to_lowercase();
                files.push(FileEntry{
                    path: rel,
                    size: md.len(),
                    lines: count_lines(p),
                    ext: format!(".{}", ext),
                });
            }
        }
    }

    let mut languages = BTreeMap::<String,u64>::new();
    let mut top_dirs = BTreeMap::<String,u64>::new();
    for f in &files {
        let dir = f.path.split('/').next().unwrap_or(".");
        *top_dirs.entry(dir.to_string()).or_default() += f.size;
        *languages.entry(f.ext.clone()).or_default() += f.size;
    }

    let mut mods = Vec::<ModuleHit>::new();
    if rootp.join("package.json").exists() {
        if let Ok(t) = fs::read_to_string(rootp.join("package.json")) {
            let name = t.contains("\"name\"").then(|| "package".to_string());
            mods.push(ModuleHit{ kind: "npm".into(), file: "package.json".into(), name, module: None });
        }
    }
    if rootp.join("go.mod").exists() {
        if let Ok(t) = fs::read_to_string(rootp.join("go.mod")) {
            let module = t.lines().find(|l| l.starts_with("module ")).map(|l| l.trim_start_matches("module ").trim().to_string());
            mods.push(ModuleHit{ kind: "go".into(), file: "go.mod".into(), name: None, module });
        }
    }
    if rootp.join("Cargo.toml").exists() {
        if let Ok(t) = fs::read_to_string(rootp.join("Cargo.toml")) {
            let name = t.lines().find(|l| l.trim().starts_with("name")).and_then(|l| l.split('=').nth(1)).map(|s| s.replace('"',"").trim().to_string());
            mods.push(ModuleHit{ kind: "cargo".into(), file: "Cargo.toml".into(), name, module: None });
        }
    }
    if rootp.join(".git").exists() {
        mods.push(ModuleHit{ kind: "git".into(), file: ".git".into(), name: None, module: None });
    }

    let mut idx = Index{
        root: rootp.file_name().unwrap_or_default().to_string_lossy().to_string(),
        indexedAt: chrono::Utc::now().to_rfc3339(),
        files, languages, topDirs: top_dirs, moduleFiles: mods,
        digest: String::new(),
    };

    let raw = serde_json::to_vec(&idx).unwrap();
    let sum = &hex::encode(Sha256::digest(&raw))[..16];
    idx.digest = sum.to_string();

    let data_dir = rootp.join("data");
    fs::create_dir_all(&data_dir).ok();
    let out = data_dir.join("index.json");
    fs::write(&out, serde_json::to_vec_pretty(&idx).unwrap()).unwrap();
    println!("Wrote {}", out.display());
}
