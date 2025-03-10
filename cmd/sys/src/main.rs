use std::env;
use std::fs::OpenOptions;
use std::io::Write;
use std::process;
use sysinfo::{System, CpuRefreshKind, RefreshKind, MemoryRefreshKind};

fn print_system_info() {
    let sys = System::new_with_specifics(
        RefreshKind::new()
            .with_cpu(CpuRefreshKind::new().with_cpu_usage())
            .with_memory(MemoryRefreshKind::new())
    );

    println!("System Information:");
    println!("  OS: {}", std::env::consts::OS);
    println!("  Arch: {}", std::env::consts::ARCH);
    println!("  Family: {}", std::env::consts::FAMILY);
    if let Ok(hostname) = hostname::get() {
        if let Some(hostname_str) = hostname.to_str() {
            println!("  Hostname: {}", hostname_str);
        }
    }
    println!("  CPU Count: {}", sys.cpus().len());
}

fn print_system_status() {
    println!("System Status:");
    let mut sys = System::new_with_specifics(
        RefreshKind::new()
            .with_cpu(CpuRefreshKind::new().with_cpu_usage())
            .with_memory(MemoryRefreshKind::new())
    );
    sys.refresh_memory();
    sys.refresh_cpu();

    // Load average (on Unix-like systems)
    #[cfg(target_family = "unix")]
    {
        let load_avg = System::load_average();
        println!("  Load Average: {:.2} {:.2} {:.2}", 
                load_avg.one, load_avg.five, load_avg.fifteen);
    }

    // Memory information (convert from bytes to MB)
    let total_mem = sys.total_memory() / 1024;
    let free_mem = sys.free_memory() / 1024;
    let used_mem = total_mem - free_mem;
    println!("  Memory:");
    println!("    Total: {} MB", total_mem);
    println!("    Free: {} MB", free_mem);
    println!("    Used: {} MB", used_mem);

    // CPU information
    println!("  CPU Cores: {}", sys.cpus().len());
    println!("  CPU Usage:");
    for (i, cpu) in sys.cpus().iter().enumerate() {
        println!("    CPU {}: {:.1}%", i, cpu.cpu_usage());
    }
}

fn print_help() {
    println!("Usage: sys <COMMAND>\n");
    println!("Commands:");
    println!("  info    List system information");
    println!("  status  Check system status");
    println!("  help    Print this message or the help of the given subcommand(s)\n");
    println!("Options:");
    println!("  -h, --help     Print help");
    println!("  -V, --version  Print version");
}

fn main() {
    // 将调试信息写入文件
    let mut debug_args = Vec::new();
    if let Ok(mut file) = OpenOptions::new()
        .create(true)
        .append(true)
        .open("/tmp/sys_debug.log")
    {
        writeln!(file, "\n=== New Execution ===").ok();
        writeln!(file, "Process ID: {}", process::id()).ok();
        writeln!(file, "Current dir: {:?}", env::current_dir().unwrap_or_default()).ok();
        
        // 记录原始参数
        writeln!(file, "Raw Arguments (std::env::args_os):").ok();
        for (i, arg) in env::args_os().enumerate() {
            writeln!(file, "  arg[{}]: {:?}", i, arg).ok();
        }
        
        // 记录 UTF-8 参数
        writeln!(file, "UTF-8 Arguments (std::env::args):").ok();
        for (i, arg) in env::args().enumerate() {
            writeln!(file, "  arg[{}]: {:?}", i, arg).ok();
        }

        // 记录所有环境变量
        writeln!(file, "Environment Variables:").ok();
        for (key, value) in env::vars() {
            writeln!(file, "  {}={}", key, value).ok();
        }

        // 尝试直接从 /proc/self/cmdline 读取命令行
        if let Ok(cmdline) = std::fs::read_to_string("/proc/self/cmdline") {
            writeln!(file, "Content of /proc/self/cmdline:").ok();
            let args: Vec<&str> = cmdline.split('\0').collect();
            for (i, arg) in args.iter().enumerate() {
                writeln!(file, "  arg[{}]: {:?}", i, arg).ok();
                if i > 0 && !arg.is_empty() {
                    debug_args.push(arg.to_string());
                }
            }
        }
    }

    // 使用从 /proc/self/cmdline 读取的参数
    let args = if debug_args.is_empty() {
        // 如果无法从 /proc/self/cmdline 读取，则使用 env::args()
        env::args().skip(1).collect::<Vec<String>>()
    } else {
        debug_args
    };

    // 如果没有参数，显示帮助信息
    if args.is_empty() {
        print_help();
        return;
    }

    // 处理第一个参数，忽略大小写
    match args[0].to_lowercase().as_str() {
        "info" => print_system_info(),
        "status" => print_system_status(),
        "-h" | "--help" | "help" => print_help(),
        "-v" | "--version" => println!("sys 0.1.0"),
        cmd => {
            if let Ok(mut file) = OpenOptions::new()
                .create(true)
                .append(true)
                .open("/tmp/sys_debug.log")
            {
                writeln!(file, "Unknown command: {:?}", cmd).ok();
            }
            print_help();
        }
    }
}
