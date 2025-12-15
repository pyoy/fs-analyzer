#!/bin/sh
#-----------------------------------------------
# Function: A POSIX-compliant shell script designed to locate the top subdirectories within a specified path, sorted by file size and number of files.
#
# Standard: Complies with POSIX standards
# Runs on: dash (Debian/Ubuntu), bash (RHEL/CentOS/RockyLinux/Almalinux/OpenEuler/AnolisOS), zsh (macOS)
#
# Create:   touch /root/sh/find-heavy-dirs.sh; chmod 700 /root/sh/find-heavy-dirs.sh
# Run ex:   sh /root/sh/find-heavy-dirs.sh
#
#-----------------------------------------------
start_time=$(date +%s)

#Full version of the script
export ver=3.01.20251214.ce

# Function to display usage information
usage() {
  echo "Usage: $0 [--path <path1> path2...] [--maxdepth <N>] [--top <N>] [--verbose] [--display-runtime] [--version]" >&2
  echo "Options:" >&2
  echo "  --path <path...>: One or more paths to search. Default is current directory." >&2
  echo "  --maxdepth <N>:   Limit the search to N levels deep (default: unlimited)." >&2
  echo "  --top <N>:        Display the top N entries (default: 20)." >&2
  echo "  --verbose:        Show detailed progress information." >&2
  echo "  --display-runtime:Show total execution time." >&2
  echo "  --version:        Show program version." >&2
  echo "  -h, --help:       Show this help message." >&2
  exit 1
}

# --- Argument Parsing ---
# We build command strings for 'eval' to correctly handle paths with spaces
# and optional arguments in a POSIX-compliant way (no arrays).

path_list=""
max_depth_arg=""
top_n=20
verbose=0
display_runtime=0

# Check if no arguments are provided
# if [ $# -eq 0 ]; then
#   usage
# fi

while [ $# -gt 0 ]; do
  case "$1" in
    --path)
      shift # Move past --path
      if [ $# -eq 0 ] || [ "$(echo "$1" | cut -c1-2)" = "--" ]; then
        echo "Error: --path requires at least one argument." >&2
        usage
      fi
      # Consume all path arguments until the next option (starting with --)
      while [ $# -gt 0 ] && [ "$(echo "$1" | cut -c1-2)" != "--" ]; do
        # Quote paths to handle spaces
        path_list="$path_list \"$1\""
        shift
      done
      ;;
    --maxdepth)
      shift # Move past --maxdepth
      if [ -z "$1" ]; then
        echo "Error: --maxdepth requires a numeric value." >&2
        usage
      fi
      # Check if $1 is a valid non-negative integer
      case "$1" in
        *[!0-9]*)
          echo "Error: --maxdepth value '$1' is not a valid positive integer." >&2
          usage
          ;;
        *)
          # Add the maxdepth argument for find
          max_depth_arg="-maxdepth $1"
          shift
          ;;
      esac
      ;;
    --top)
      shift
      if [ -z "$1" ]; then
        echo "Error: --top requires a numeric value." >&2
        usage
      fi
      case "$1" in
        *[!0-9]*)
          echo "Error: --top value '$1' is not a valid positive integer." >&2
          usage
          ;;
        *)
          top_n=$1
          shift
          ;;
      esac
      ;;
    --verbose)
      verbose=1
      shift
      ;;
    --display-runtime)
      display_runtime=1
      shift
      ;;
    --version)
      echo "$ver"
      exit 0
      ;;
    -h | --help)
      usage
      ;;
    *)
      echo "Error: Unknown option '$1'" >&2
      usage
      ;;
  esac
done

# Validate that paths were provided
if [ -z "$path_list" ]; then
  # Default to current directory if no path specified
  path_list=" \".\""
fi

if [ "$verbose" -eq 1 ]; then
  echo "Starting scan (Ver: $ver)..."
  echo "Targets: $path_list"
  if [ -n "$max_depth_arg" ]; then
    echo "Max Depth: ${max_depth_arg#-maxdepth }"
  fi
fi

# --- Define Filters ---

# Paths to exclude (system mount points)
# We use -path, so these must be the exact absolute paths.
prune_paths="\( -path /proc -o -path /dev -o -path /sys -o -path /run \)"



# --- Human Readable Conversion Function (POSIX AWK) ---
# This AWK script takes the KB value ($1) and converts it to B, K, M, G, T, P units.
# The calculation uses 1024 for binary units (KiB, MiB, etc.), but the label uses K, M, G, T.
human_readable_awk='
{
    size_kb = $1;
    path = $2;
    units = " KB MB GB TB PB";
    # Start with KB (index 1 in split array)
    s=1; 
    
    # KB to Bytes for precise human-readable starting point (if size < 1MB)
    size_bytes = size_kb * 1024; 
    
    # We must preserve the original KB value for correct sorting later.
    # We print the raw KB value, then the human-readable string, then the path.
    # The raw KB value will be used by the subsequent 'sort' command.
    printf "%s ", size_kb;

    # Convert to higher units (MB, GB, etc.)
    # Loop while size is > 1024 KB and we have units left (max PB here)
    while(size_kb >= 1024 && s < 5) { 
        size_kb /= 1024;
        s++;
    }
    
    # Extract the unit string (v[1]="KB", v[2]="MB", etc.)
    split(units, v, " ");

    # Print the formatted human-readable size
    # We use 1 decimal place for MB, GB, TB, etc.
    if (s == 1) {
        # Print KB as integer if it did not exceed 1024
        printf "%7.0f%s %s\n", size_kb, v[s], path;
    } else {
        # Print higher units with one decimal place
        printf "%7.1f%s %s\n", size_kb, v[s], path;
    }
}
'


# --- 1. Top 20 by Size ---

echo "--- Top $top_n largest subdirectories by size (KB by default) ---"

# We use 'eval' to correctly construct the find command.
# This allows $path_list to expand into multiple, quoted arguments
# and $max_depth_arg to be included only if it was set.
#
# Command logic:
# 1. find $path_list: Search the user-provided paths.
# 2. $prune_paths -prune: If a path matches /proc, /dev, etc., do not descend into it.
# 3. -o: Otherwise (if not pruned)...
# 4. \( ... \): Execute the following group.
# 5. $max_depth_arg: Apply -maxdepth if specified.
# 6. -mindepth 1: Do not include the starting paths (e.g., /var) themselves.
# 7. -type d: Find only directories.
# 8. -exec du -sk {} \;: For each directory found, run 'du -sk' (summary, in kilobytes).
#    This is slow but accurate, as it runs 'du' for every single subdirectory.
# 9. | sort -nr: Sort the output numerically (n) and in reverse (r) order.
# 10. | head -n 20: Get the top 20 results.

eval "find $path_list $max_depth_arg -mindepth 1 $prune_paths -prune -o \( -type d -exec du -sk {} \; \)" | \
  sort -nr | \
  head -n "$top_n" | \
  # Pipe the output to AWK for human-readable conversion
  awk "$human_readable_awk"

# --- 2. Top 20 by File Count ---

echo ""
echo "--- Top $top_n largest subdirectories by file count ---"

# Command logic:
# 1. find ... -type f -print: Find all files (not directories) and print their paths.
# 2. | awk ...: Process each file path.
#    -F/: Set the field separator to '/'.
#    For a path like "/var/log/syslog" (NF=4):
#    It loops from i=2 to NF-1 (i.e., 2 and 3).
#    i=2: path="/var". Prints "/var".
#    i=3: path="/var/log". Prints "/var/log".
#    This credits both /var and /var/log for the file "syslog".
# 3. | sort: Sort all the printed parent paths.
# 4. | uniq -c: Count identical adjacent lines (e.g., "150 /var/log").
# 5. | sort -nr: Sort the counts numerically and in reverse.
# 6. | head -n 20: Get the top 20 results.

eval "find $path_list $max_depth_arg -mindepth 1 $prune_paths -prune -o \( -type f -print \)" | \
  awk -F/ '{
    for (i = 2; i < NF; i++) {
      path = "/"
      for (j = 2; j <= i; j++) {
        path = path $j (j == i ? "" : "/")
      }
      print path
    }
  }' | \
  sort | \
  uniq -c | \
  sort -nr | \
  head -n "$top_n"

if [ "$display_runtime" -eq 1 ]; then
  echo ""
  stop_time=$(date +%s)
  echo "Processed in $((stop_time-start_time)) second(s)"
fi


#-----------------------------------------------
# Change History:
# date       ver   note
# 2025/12/14 v3.01 Added --top, --verbose, --display-runtime, --version, and default path "."
# 2025/11/05 v2.01 Improved, and cross-platform support in unix.
# 2024/09/16 v1.01 Initial creation.
#-----------------------------------------------