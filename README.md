# Introduction  
This program scans specified directories (or the current directory by default) to identify the top N largest subdirectories by total size and the top N subdirectories with the most files. It uses an efficient bottom-up aggregation algorithm to calculate sizes and file counts, avoiding redundant traversals.  
  
**Key Features:**  
*   **High-Performance Scanning**: Utilizes Go's `filepath.WalkDir` (significantly faster than `filepath.Walk`) for efficient directory traversal.  
*   **Smart Aggregation**: Implements a **Bottom-Up** aggregation algorithm. Unlike traditional shell scripts that often recursively recalculate or use redundant loops, this tool scans the file system only once and aggregates data from the deepest subdirectories up to the root instantly.  
*   **User-Friendly Display**:  
    *   Auto-aligned tabular output.  
    *   Human-readable units (automatically scales to KB, MB, GB, TB).  
    *   Progress indicators (in verbose mode).  
  
# Advantages Over Shell Scripts  
**Algorithmic Efficiency (O(N) vs O(N*Depth))**  
  
*   **The Shell Script Bottleneck**: Traditional shell scripts often rely on `find` combined with `awk` loops to calculate paths, or `du` commands that may perform redundant I/O operations for overlapping directories. In scenarios with deep directory structures or massive file counts, this approach leads to exponential performance degradation and high memory usage (e.g., `sort | uniq` on millions of lines).  
*   **The Go Solution**: This tool performs a **single-pass** traversal using `filepath.WalkDir`. During the scan, it records statistics only for the direct parent directory of each file. Once the scan is complete, it aggregates data using a **Bottom-Up (Deepest First)** strategy, propagating subdirectory totals to their parents in a single efficient step. This design significantly reduces CPU utilization and I/O latency.   
  
# Performance Comparison of Go Executable and Shell Script  
Under the same directory structure and parameters, the Go executable typically runs faster than the shell script. This is because the Go program is statically compiled, avoiding interpreter overhead, while shell scripts rely on the system's shell interpreter.  

In a directory structure with 1,827,814 files totaling 166GB:  

| Metric                 | Go Executable | Shell Script |
| :--------------------- | :------------ | :----------- |
| Runtime                | 91.39 seconds | 775 seconds  |
| CPU Single-Thread Usage| ~57%          | ~18%         |
| Disk %Util             | ~98%          | ~85%         |

**Summary**: In this test environment, the Go program's execution efficiency is 8.48 times that of the shell script. It uses more CPU and disk I/O than the shell script, but because it runs in a single thread, its impact on other programs in a multi-core CPU system is minimal. The disk I/O usage of the Go executable is higher, while the shell script's CPU and disk I/O usage fluctuates significantly, generally imposing less pressure.  

## Statistical Accuracy  
The statistical results of the Go executable and the shell script are largely consistent, but the Go executable can count more files, making its statistics more accurate. The reported size and file count may be slightly higher than those from the shell script.  

# Shell Script Usage Help  
A POSIX-compliant shell script designed to locate the top subdirectories within a specified path, sorted by file size and number of files.  
Runs on: dash (Debian/Ubuntu), bash (RHEL/CentOS/RockyLinux/Almalinux/OpenEuler/AnolisOS), zsh (macOS)  

```  
Usage: ./find-heavy-dirs.sh [--path <path1> path2...] [--maxdepth <N>] [--top <N>] [--verbose] [--display-runtime] [--version]  
Options:  
  --path <path...>: One or more paths to search. Default is current directory.  
  --maxdepth <N>:   Limit the search to N levels deep (default: unlimited).  
  --top <N>:        Display the top N entries (default: 20).  
  --verbose:        Show detailed progress information.  
  --display-runtime:Show total execution time.  
  --version:        Show program version.  
  -h, --help:       Show this help message.  
```  
  
# Install Go Environment  
On Linux systems, you need to install a Go environment (e.g., `yum install go`, or `dnf install go`, or `apt install golang`).  
On macOS systems, you can install Go using Homebrew (e.g., `brew install go`).  
On Windows systems, you can download and install Go from the official Go website (e.g., `https://golang.org/dl/`).  
  
# Go Source Code Execution Example  
```bash  
go run find_heavy_dirs.go --path . .. --top 3 --verbose  
```  
  
# Compilation  
#Initialize module (optional, can compile directly if it's a single file)  
```bash  
test -d fs-analyzer/cmd || mkdir -p fs-analyzer/cmd; cd fs-analyzer/cmd  
#upload find_heavy_dirs.go  
go mod init find_heavy_dirs  
```  
or  
```bash
git clone https://github.com/pyoy/fs-analyzer.git
cd fs-analyzer/cmd
go mod tidy
```
  
To ensure 100% smooth operation on all versions like EL6, EL7, EL8, EL9, EL10, Debian/Ubuntu series systems, it is recommended to use the standard Go community method: disable CGO and static compilation.  
  
Linux (compile to an executable named `find-heavy-dirs`):  
```bash  
CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/find-heavy-dirs ./find_heavy_dirs.go  
```  
  
Windows (cross-compilation):  
```bash  
set GOOS=windows  
set GOARCH=amd64  
go build -ldflags="-s -w" -o ../bin/find-heavy-dirs.exe ./find_heavy_dirs.go  
```  
  
macOS (cross-compilation):  
```bash  
set GOOS=darwin  
set GOARCH=amd64  
go build -ldflags="-s -w" -o ../bin/find-heavy-dirs-darwin-amd64 ./find_heavy_dirs.go  
```  
  
# Go Executable Usage Help  
A standalone binary, no Go environment installation required, just run directly.  
```  
Usage: find_heavy_dirs [--path <path1> path2...] [--maxdepth <N>] [--top <N>] [--verbose] [--display-runtime] [--version]  
Options:  
  --path <path...>: One or more paths to search. Default is current directory.  
  --maxdepth <N>:   Limit the search to N levels deep. Default is 1000000.  
  --top <N>:        Display the top N entries. Default is 20.  
  --verbose:        Show detailed progress information.  
  --display-runtime:Show total execution time.  
  --version:        Show program version.  
  -h, --help:       Show this help message.  
```  
  
## Usage example on Linux  
#Grant execution permissions and Run  
```bash  
chmod +x find-heavy-dirs   
./find-heavy-dirs --path /usr/lib /var --maxdepth 1500 --top 15 --display-runtime  
```  
Compatible with Debian/Ubuntu series systems, and RHEL/CentOS/RockyLinux/Almalinux/OpenEuler/AnolisOS series systems. Supports at least el6-el9 distributions or derivative distributions, and possibly more. Please test it yourself.  
  
Output is as follows:  
```
--- Top 15 Largest Subdirectories by Size ---
Metric          | Path                                              
----------------------------------------------------------------------
1.4 GB          | /var
1.1 GB          | /usr/lib
769.7 MB        | /var/log
660.4 MB        | /usr/lib/firmware
526.9 MB        | /var/cache
524.6 MB        | /var/cache/yum/x86_64
524.6 MB        | /var/cache/yum/x86_64/7
524.6 MB        | /var/cache/yum
368.0 MB        | /var/cache/yum/x86_64/7/updates
324.1 MB        | /var/cache/yum/x86_64/7/updates/gen
206.3 MB        | /usr/lib/golang
139.7 MB        | /usr/lib/firmware/netronome
132.4 MB        | /var/lib
124.8 MB        | /var/lib/rpm
112.2 MB        | /usr/lib/golang/pkg

--- Top 15 Subdirectories by File Count ---
Metric          | Path                                              
----------------------------------------------------------------------
15626 Files     | /usr/lib
7504 Files      | /var
7238 Files      | /var/lib
7202 Files      | /var/lib/yum
7126 Files      | /var/lib/yum/yumdb
5060 Files      | /usr/lib/modules
4519 Files      | /usr/lib/golang
4467 Files      | /usr/lib/golang/src
2557 Files      | /usr/lib/firmware
2531 Files      | /usr/lib/modules/3.10.0-1160.el7.x86_64
2529 Files      | /usr/lib/modules/3.10.0-1160.119.1.el7.x86_64
2510 Files      | /usr/lib/modules/3.10.0-1160.el7.x86_64/kernel
2508 Files      | /usr/lib/modules/3.10.0-1160.119.1.el7.x86_64/kernel
1918 Files      | /var/lib/yum/yumdb/l
1718 Files      | /usr/lib/modules/3.10.0-1160.el7.x86_64/kernel/drivers

Processed in 0.17 second(s)
```
Both Windows and macOS systems can be used as a reference for Linux.  
  
## Usage example on Windows  
```ps1  
cmd /c find-heavy-dirs.exe --path c:\Windows --top 15 --display-runtime  
```  
Output is as follows:  
```  
--- Top 15 Largest Subdirectories by Size ---  
Metric          | Path  
----------------------------------------------------------------------  
31.7 GB         | c:\Windows  
19.8 GB         | c:\Windows\WinSxS  
4.2 GB          | c:\Windows\System32  
2.4 GB          | c:\Windows\SoftwareDistribution  
2.4 GB          | c:\Windows\SoftwareDistribution\Download  
2.3 GB          | c:\Windows\SoftwareDistribution\Download\2f7d46b7f2bbea65e38359aca32fefdd  
2.0 GB          | ...ndows\SoftwareDistribution\Download\2f7d46b7f2bbea65e38359aca32fefdd\Metadata  
1.3 GB          | c:\Windows\SystemApps  
1.0 GB          | c:\Windows\SysWOW64  
1013.4 MB       | ...\Download\2f7d46b7f2bbea65e38359aca32fefdd\Metadata\Windows11.0-KB5068861-x64  
703.1 MB        | ...\Download\2f7d46b7f2bbea65e38359aca32fefdd\Metadata\Windows11.0-KB5043080-x64  
601.3 MB        | c:\Windows\Microsoft.NET  
595.7 MB        | c:\Windows\System32\Microsoft-Edge-WebView  
595.7 MB        | ...microsoft-edge-webview_31bf3856ad364e35_10.0.26100.7171_none_2ed7609d3aa5a301  
579.3 MB        | ...microsoft-edge-webview_31bf3856ad364e35_10.0.26100.6899_none_2e8cf9973add6927  
  
--- Top 15 Subdirectories by File Count ---  
Metric          | Path  
----------------------------------------------------------------------  
461226 Files    | c:\Windows  
270675 Files    | c:\Windows\SoftwareDistribution  
270659 Files    | c:\Windows\SoftwareDistribution\Download  
270650 Files    | c:\Windows\SoftwareDistribution\Download\2f7d46b7f2bbea65e38359aca32fefdd  
269598 Files    | ...ndows\SoftwareDistribution\Download\2f7d46b7f2bbea65e38359aca32fefdd\Metadata  
170013 Files    | ...\Download\2f7d46b7f2bbea65e38359aca32fefdd\Metadata\Windows11.0-KB5068861-x64  
134541 Files    | c:\Windows\WinSxS  
98932 Files     | ...\Download\2f7d46b7f2bbea65e38359aca32fefdd\Metadata\Windows11.0-KB5043080-x64  
38926 Files     | c:\Windows\WinSxS\Manifests  
22444 Files     | c:\Windows\System32  
13383 Files     | c:\Windows\servicing  
13236 Files     | c:\Windows\servicing\Packages  
6647 Files      | c:\Windows\System32\CatRoot  
6647 Files      | c:\Windows\System32\CatRoot\{F750E6C3-38EE-11D1-85E5-00C04FC295EE}  
6629 Files      | c:\Windows\SystemApps  
  
Processed in 14.56 second(s)  
```
  
## Usage example on macOS  
```bash  
chmod +x find-heavy-dirs-darwin-amd64  
./find-heavy-dirs-darwin-amd64 --path /usr/lib --top 15 --display-runtime  
```  
Output is as follows:    
```  
--- Top 15 Largest Subdirectories by Size ---
Metric          | Path                                              
----------------------------------------------------------------------
32.8 MB         | /usr/lib
7.7 MB          | /usr/lib/usd
6.2 MB          | /usr/lib/usd/usd
5.5 MB          | /usr/lib/zsh/5.9
5.5 MB          | /usr/lib/zsh/5.9/zsh
5.5 MB          | /usr/lib/zsh
5.0 MB          | /usr/lib/usd/usd/hdx
5.0 MB          | /usr/lib/usd/usd/hdx/resources
4.9 MB          | /usr/lib/usd/usd/hdx/resources/textures
3.7 MB          | /usr/lib/system
3.6 MB          | /usr/lib/rpcsvc
2.8 MB          | /usr/lib/sasl2
2.4 MB          | /usr/lib/swift
2.2 MB          | /usr/lib/pam
1.5 MB          | /usr/lib/system/introspection

--- Top 15 Subdirectories by File Count ---
Metric          | Path                                              
----------------------------------------------------------------------
493 Files       | /usr/lib
353 Files       | /usr/lib/usd
210 Files       | /usr/lib/usd/libraries
137 Files       | /usr/lib/usd/usd
132 Files       | /usr/lib/usd/libraries/stdlib
109 Files       | /usr/lib/usd/libraries/stdlib/genglsl
37 Files        | /usr/lib/zsh/5.9
37 Files        | /usr/lib/zsh/5.9/zsh
37 Files        | /usr/lib/zsh
32 Files        | /usr/lib/usd/libraries/pbrlib
29 Files        | /usr/lib/usd/libraries/pbrlib/genglsl
28 Files        | /usr/lib/usd/usd/hdSt
28 Files        | /usr/lib/usd/usd/hdSt/resources
26 Files        | /usr/lib/usd/usd/hdSt/resources/shaders
24 Files        | /usr/lib/usd/usd/hdx/resources

Processed in 0.09 second(s)
```