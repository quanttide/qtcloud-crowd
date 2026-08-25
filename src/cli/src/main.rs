use clap::{Parser, Subcommand};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

// ─── CLI ───

/// 量潮众包管理云 CLI —— 管理方工具（任务审核 / 执行方认证 / 结算记录）
///
/// 读写本地数据文件（默认 data/crowd.json，QTCLOUD_CROWD_DATA 可覆盖；
/// 文件不存在时自动初始化为空结构 {tasks,partners,settlements}）。
#[derive(Parser)]
#[command(
    name = "qtcloud-crowd",
    version,
    about = "量潮众包管理云 CLI：管理方工具（任务审核/执行方认证/结算记录）"
)]
struct Cli {
    /// 数据文件路径覆盖（默认 data/crowd.json；未设时读环境变量 QTCLOUD_CROWD_DATA）
    #[arg(long, global = true)]
    data: Option<String>,

    #[command(subcommand)]
    command: Command,
}

#[derive(Subcommand)]
enum Command {
    /// 任务管理（事：任务审核）
    Tasks {
        #[command(subcommand)]
        command: TasksCommand,
    },
    /// 执行方管理（人：名单 + 认证）
    Partners {
        #[command(subcommand)]
        command: PartnersCommand,
    },
    /// 结算管理（财：记录一笔）
    Settlements {
        #[command(subcommand)]
        command: SettlementsCommand,
    },
}

#[derive(Subcommand)]
enum TasksCommand {
    /// 列出任务（id/title/status；可 --status 过滤）
    List {
        /// 按状态过滤（pending/reviewing/done）
        #[arg(long)]
        status: Option<String>,
    },
    /// 审核任务（approve=通过→done；reject=打回→pending；验收准则为空不能通过）
    Review {
        /// 任务 ID
        id: String,
        /// 审核动作（approve/reject）
        #[arg(long)]
        action: String,
    },
}

#[derive(Subcommand)]
enum PartnersCommand {
    /// 列出执行方（id/name/type/certified）
    List,
    /// 认证执行方（待认证 → 已认证）
    Certify {
        /// 执行方 ID
        id: String,
    },
}

#[derive(Subcommand)]
enum SettlementsCommand {
    /// 列出结算记录（任务+执行方+金额+时间）
    List,
    /// 记一笔结算（任务必须已通过验收；同一任务只能结算一次）
    Add {
        /// 任务 ID（必须已通过验收，即 status=done）
        task_id: String,
        /// 执行方 ID
        partner_id: String,
        /// 结算金额（元，必须为正数）
        #[arg(allow_negative_numbers = true)]
        amount: f64,
    },
}

// ─── 数据模型（对齐 src/studio/lib/models/） ───

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Task {
    id: String,
    title: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    content: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    acceptance_criteria: Option<String>,
    status: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Partner {
    id: String,
    name: String,
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    r#type: Option<String>,
    certified: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Settlement {
    id: String,
    task_id: String,
    partner_id: String,
    amount: f64,
    settled_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct CrowdData {
    tasks: Vec<Task>,
    partners: Vec<Partner>,
    settlements: Vec<Settlement>,
}

impl CrowdData {
    fn empty() -> Self {
        Self {
            tasks: Vec::new(),
            partners: Vec::new(),
            settlements: Vec::new(),
        }
    }
}

// 领域枚举取值（对齐 src/studio/lib/models/）
const TASK_STATUSES: [&str; 3] = ["pending", "reviewing", "done"];
const REVIEW_ACTIONS: [&str; 2] = ["approve", "reject"];

fn exit_err(e: &str) -> ! {
    eprintln!("错误: {e}");
    std::process::exit(1);
}

// ─── 数据文件访问 ───

/// 默认数据文件路径（相对当前工作目录）
const DEFAULT_DATA: &str = "data/crowd.json";

/// 解析数据文件路径：--data 参数 > 环境变量 QTCLOUD_CROWD_DATA > data/crowd.json
fn resolve_data_path(cli_data: &Option<String>) -> PathBuf {
    if let Some(d) = cli_data
        && !d.trim().is_empty()
    {
        return PathBuf::from(d);
    }
    match std::env::var("QTCLOUD_CROWD_DATA") {
        Ok(d) if !d.trim().is_empty() => PathBuf::from(d),
        _ => PathBuf::from(DEFAULT_DATA),
    }
}

/// 读取数据文件；文件不存在时初始化为空结构并落盘
fn load_data(path: &Path) -> CrowdData {
    let raw = match fs::read_to_string(path) {
        Ok(raw) => raw,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            let data = CrowdData::empty();
            save_data(path, &data);
            eprintln!("已初始化数据文件: {}", path.display());
            return data;
        }
        Err(e) => exit_err(&format!("读取数据文件 {} 失败: {e}", path.display())),
    };
    if raw.trim().is_empty() {
        return CrowdData::empty();
    }
    serde_json::from_str(&raw)
        .unwrap_or_else(|e| exit_err(&format!("解析数据文件 {} 失败: {e}", path.display())))
}

/// 原子写入数据文件（先写临时文件再 rename），2 空格缩进、空值省略
fn save_data(path: &Path, data: &CrowdData) {
    let mut json = serde_json::to_string_pretty(data)
        .unwrap_or_else(|e| exit_err(&format!("序列化数据失败: {e}")));
    json.push('\n');
    if let Some(parent) = path.parent()
        && !parent.as_os_str().is_empty()
    {
        fs::create_dir_all(parent)
            .unwrap_or_else(|e| exit_err(&format!("创建目录 {} 失败: {e}", parent.display())));
    }
    let tmp = path.with_extension("json.tmp");
    fs::write(&tmp, json).unwrap_or_else(|e| exit_err(&format!("写入临时文件失败: {e}")));
    fs::rename(&tmp, path)
        .unwrap_or_else(|e| exit_err(&format!("写入数据文件 {} 失败: {e}", path.display())));
}

// ─── 工具函数 ───

/// 生成 ID：`{prefix}-{unix秒}`，冲突则自增（对齐 qtcloud-execute 风格）
fn gen_id(prefix: &str, existing: &[String]) -> String {
    let base = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs();
    for i in 0..10000u64 {
        let id = format!("{prefix}-{}", base + i);
        if !existing.contains(&id) {
            return id;
        }
    }
    format!("{prefix}-{base}")
}

/// 当前时间 ISO 8601 UTC（无 chrono 依赖，Howard Hinnant 算法）
fn iso8601_now() -> String {
    let secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs();
    let days = secs / 86400;
    let rem = secs % 86400;
    let (h, mi, s) = (rem / 3600, (rem % 3600) / 60, rem % 60);
    let z = days as i64 + 719_468;
    let era = if z >= 0 { z } else { z - 146_096 } / 146_097;
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let year = if m <= 2 { y + 1 } else { y };
    format!("{year:04}-{m:02}-{d:02}T{h:02}:{mi:02}:{s:02}Z")
}

fn check_status(status: &str) -> Result<(), String> {
    if TASK_STATUSES.contains(&status) {
        Ok(())
    } else {
        Err(format!(
            "非法状态 `{status}`，可选：{}",
            TASK_STATUSES.join("/")
        ))
    }
}

fn check_review_action(action: &str) -> Result<(), String> {
    if REVIEW_ACTIONS.contains(&action) {
        Ok(())
    } else {
        Err(format!(
            "非法审核动作 `{action}`，可选：{}",
            REVIEW_ACTIONS.join("/")
        ))
    }
}

// ─── 子命令实现 ───

/// tasks list：列出任务（id/title/status；可 --status 过滤）
fn tasks_list(data: &CrowdData, status: Option<&str>, path: &Path) {
    if let Some(s) = status
        && let Err(e) = check_status(s)
    {
        exit_err(&e);
    }
    println!("── 任务清单 ──  {}", path.display());
    let filtered: Vec<&Task> = data
        .tasks
        .iter()
        .filter(|t| status.map(|s| t.status == s).unwrap_or(true))
        .collect();
    if filtered.is_empty() {
        println!("  （暂无任务）");
        return;
    }
    for t in &filtered {
        println!("  {:<18} {:<10} {}", t.id, t.status, t.title);
        if let Some(c) = &t.acceptance_criteria {
            println!("      验收准则: {c}");
        }
    }
}

/// tasks review：审核任务（approve=通过→done；reject=打回→pending）
fn tasks_review(data: &mut CrowdData, id: &str, action: &str, path: &Path) {
    if let Err(e) = check_review_action(action) {
        exit_err(&e);
    }
    let task = data
        .tasks
        .iter_mut()
        .find(|t| t.id == id)
        .unwrap_or_else(|| exit_err(&format!("任务不存在: {id}")));
    match action {
        "approve" => {
            if task.status == "done" {
                exit_err(&format!("任务 {id} 已完成，无需重复审核"));
            }
            let criteria = task.acceptance_criteria.as_deref().unwrap_or("").trim();
            if criteria.is_empty() {
                exit_err(&format!(
                    "任务 {id} 验收准则为空，不能通过——说不清验收 = 不发单（模型层校验）"
                ));
            }
            task.status = "done".to_string();
            println!("✓ 已通过验收：{id} → done（{}）", task.title);
        }
        "reject" => {
            if task.status == "done" {
                exit_err(&format!("任务 {id} 已完成，不能打回"));
            }
            task.status = "pending".to_string();
            println!("✓ 已打回：{id} → pending（待补充/修改）");
        }
        _ => unreachable!(),
    }
    save_data(path, data);
}

/// partners list：列出执行方（id/name/type/certified）
fn partners_list(data: &CrowdData, path: &Path) {
    println!("── 执行方清单 ──  {}", path.display());
    if data.partners.is_empty() {
        println!("  （暂无执行方）");
        return;
    }
    for p in &data.partners {
        let t = p.r#type.as_deref().unwrap_or("-");
        let cert = if p.certified {
            "已认证"
        } else {
            "未认证"
        };
        println!("  {:<18} {:<9} {}  {}", p.id, t, cert, p.name);
    }
}

/// partners certify：认证执行方（待认证 → 已认证）
fn partners_certify(data: &mut CrowdData, id: &str, path: &Path) {
    let partner = data
        .partners
        .iter_mut()
        .find(|p| p.id == id)
        .unwrap_or_else(|| exit_err(&format!("执行方不存在: {id}")));
    if partner.certified {
        println!("（执行方 {id} 已是已认证状态）");
    } else {
        partner.certified = true;
        println!("✓ 已认证：{id} → {}（已认证）", partner.name);
    }
    save_data(path, data);
}

/// settlements list：列出结算记录（id/task_id/partner_id/amount/settled_at）
fn settlements_list(data: &CrowdData, path: &Path) {
    println!("── 结算清单 ──  {}", path.display());
    if data.settlements.is_empty() {
        println!("  （暂无结算记录）");
        return;
    }
    for s in &data.settlements {
        println!(
            "  {:<16} {:<18} {:<18} {:>10.2}  {}",
            s.id, s.task_id, s.partner_id, s.amount, s.settled_at
        );
    }
}

/// settlements add：记一笔结算（任务必须已通过验收；同一任务只结算一次）
fn settlements_add(
    data: &mut CrowdData,
    task_id: &str,
    partner_id: &str,
    amount: f64,
    path: &Path,
) {
    let task = data
        .tasks
        .iter()
        .find(|t| t.id == task_id)
        .unwrap_or_else(|| exit_err(&format!("任务不存在: {task_id}")));
    if task.status != "done" {
        exit_err(&format!(
            "任务 {task_id} 未通过验收（status={}），不能结算——验收通过后才能记一笔",
            task.status
        ));
    }
    if data.settlements.iter().any(|s| s.task_id == task_id) {
        exit_err(&format!("任务 {task_id} 已结算过，不能重复结算"));
    }
    if data.partners.iter().find(|p| p.id == partner_id).is_none() {
        exit_err(&format!("执行方不存在: {partner_id}"));
    }
    if amount <= 0.0 {
        exit_err(&format!("金额必须为正数（收到 {amount}）"));
    }
    let existing_ids: Vec<String> = data.settlements.iter().map(|s| s.id.clone()).collect();
    let id = gen_id("stl", &existing_ids);
    let settled_at = iso8601_now();
    data.settlements.push(Settlement {
        id: id.clone(),
        task_id: task_id.to_string(),
        partner_id: partner_id.to_string(),
        amount,
        settled_at: settled_at.clone(),
    });
    println!(
        "✓ 已记录结算：{id} → 任务 {task_id} / 执行方 {partner_id} / {amount:.2} 元 @ {settled_at}"
    );
    save_data(path, data);
}

fn main() {
    let cli = Cli::parse();
    let path = resolve_data_path(&cli.data);

    match &cli.command {
        Command::Tasks { command } => match command {
            TasksCommand::List { status } => {
                let data = load_data(&path);
                tasks_list(&data, status.as_deref(), &path);
            }
            TasksCommand::Review { id, action } => {
                let mut data = load_data(&path);
                tasks_review(&mut data, id, action, &path);
            }
        },
        Command::Partners { command } => match command {
            PartnersCommand::List => {
                let data = load_data(&path);
                partners_list(&data, &path);
            }
            PartnersCommand::Certify { id } => {
                let mut data = load_data(&path);
                partners_certify(&mut data, id, &path);
            }
        },
        Command::Settlements { command } => match command {
            SettlementsCommand::List => {
                let data = load_data(&path);
                settlements_list(&data, &path);
            }
            SettlementsCommand::Add {
                task_id,
                partner_id,
                amount,
            } => {
                let mut data = load_data(&path);
                settlements_add(&mut data, task_id, partner_id, *amount, &path);
            }
        },
    }
}
