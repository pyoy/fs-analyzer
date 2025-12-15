package main

/*
Program: find_heavy_dirs
Description:
    This program scans specified directories (or the current directory by default) to identify
    the top N largest subdirectories by total size and the top N subdirectories with the
    most files. It uses an efficient bottom-up aggregation algorithm to calculate sizes
    and file counts, avoiding redundant traversals.

Usage:
    find_heavy_dirs [options]

Options:
    --path <dir1> [dir2...]   Specify directories to scan. Default is current directory.
    --top <N>                 Display the top N entries. Default is 20.
    --maxdepth <N>            Maximum recursion depth. Default is 1000000.
    --verbose                 Show detailed progress information. Default is false.
    --display-runtime         Show total execution time at the end. Default is false.
    --version                 Show program version. Default is false.
    -h, --help                Show help message.
*/

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// --- Configuration & Constants ---

var (
	version        = "find-heavy-dirs version 3.01.20251214.go"
	excludePaths   = []string{"/proc", "/dev", "/sys", "/run"}
	excludeMap     map[string]bool
	targetPaths    []string
	maxDepth       = 1000000 // Default 1000000
	topN           = 20      // Default 20
	verbose        = false   // Default false
	displayRuntime = false   // Default false
	showVersion    = false   // Default false
)

// --- Data Structures ---

type DirStat struct {
	Path      string
	TotalSize int64
	FileCount int64
	Depth     int
}

// Map to store scan results, Key is the absolute path of the directory
var dirStats = make(map[string]*DirStat)

// --- Main Program ---

func main() {
	startTime := time.Now()

	// Parse arguments
	parseArgs()

	// Initialize exclude list
	excludeMap = make(map[string]bool)
	for _, p := range excludePaths {
		excludeMap[p] = true
	}

	// Optimize target paths: Remove subdirectories if their parent is also in the list to avoid double counting
	targetPaths = removeSubdirectories(targetPaths)

	if verbose {
		fmt.Printf("Starting scan (Ver: %s)...\n", version)
		fmt.Printf("Targets: %v\n", targetPaths)
		if maxDepth > -1 {
			fmt.Printf("Max Depth: %d\n", maxDepth)
		}
	}

	// Execute scan
	totalFiles := 0
	for _, root := range targetPaths {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			fmt.Printf("Error resolving path %s: %v\n", root, err)
			continue
		}
		n := scanDirectory(absRoot)
		totalFiles += n
	}

	if verbose {
		fmt.Printf("Scan complete. Found %d files. Aggregating data...\n", totalFiles)
	}

	// Data Aggregation (Bottom-Up calculation)
	aggregateStats()

	// Output results
	// Convert Map to Slice for sorting
	var statsList []*DirStat
	for _, s := range dirStats {
		// Filter out results not under the search root paths (due to bottom-up aggregation, parent of roots might be included, need to exclude)
		if isUnderTargets(s.Path) {
			statsList = append(statsList, s)
		}
	}

	// Sort by size Top N
	sort.Slice(statsList, func(i, j int) bool {
		return statsList[i].TotalSize > statsList[j].TotalSize
	})
	printTable(fmt.Sprintf("Top %d Largest Subdirectories by Size", topN), statsList, true)

	// Sort by file count Top N
	sort.Slice(statsList, func(i, j int) bool {
		return statsList[i].FileCount > statsList[j].FileCount
	})
	printTable(fmt.Sprintf("Top %d Subdirectories by File Count", topN), statsList, false)

	// End statistics
	if displayRuntime {
		duration := time.Since(startTime)
		fmt.Printf("\nProcessed in %.2f second(s)\n", duration.Seconds())
	}
}

// --- Core Logic ---

// scanDirectory traverses the directory tree, recording only file sizes and counts directly belonging to that directory
func scanDirectory(root string) int {
	count := 0
	rootDepth := strings.Count(root, string(os.PathSeparator))

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Ignore permission errors, continue scanning
			if verbose {
				fmt.Printf("Warning: Access denied or error at %s: %v\n", path, err)
			}
			return nil
		}

		// Check exclude paths (Prune)
		if d.IsDir() && excludeMap[path] {
			return filepath.SkipDir
		}

		// Check depth
		currentDepth := strings.Count(path, string(os.PathSeparator)) - rootDepth
		if maxDepth != -1 && currentDepth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Statistics logic
		if !d.IsDir() {
			// It's a file: get size and record to its parent directory
			info, err := d.Info()
			if err == nil {
				dirPath := filepath.Dir(path)
				s := getDirStat(dirPath)
				s.TotalSize += info.Size()
				s.FileCount++ // Record direct file count
				count++
			}
		} else {
			// It's a directory: ensure it exists in Map (even empty directories need to be recorded)
			getDirStat(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path %s: %v\n", root, err)
	}
	return count
}

// aggregateStats bubbles up data from bottom to top
// Original scan only recorded the direct parent directory of files.
// This function accumulates the size and count of subdirectories to their parent directories, up to the search root.
func aggregateStats() {
	// Get all directory paths
	paths := make([]string, 0, len(dirStats))
	for p := range dirStats {
		paths = append(paths, p)
	}

	// Sort by path depth descending (deepest directories first)
	// This ensures when processing a parent, its children are already calculated
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) > len(paths[j]) // Simple approximation of depth by string length
	})

	// Bubble up accumulation
	for _, p := range paths {
		parent := filepath.Dir(p)

		// Prevent self-aggregation (root directory's parent is itself in some OS/cases)
		if parent == p {
			continue
		}

		// If parent is also within our statistics scope (i.e., not above root), accumulate
		// Note: Check if parent is already initialized
		if parentStat, ok := dirStats[parent]; ok {
			childStat := dirStats[p]
			parentStat.TotalSize += childStat.TotalSize
			parentStat.FileCount += childStat.FileCount
		}
	}
}

// getDirStat safely retrieves or initializes Map entry
func getDirStat(path string) *DirStat {
	if _, ok := dirStats[path]; !ok {
		dirStats[path] = &DirStat{Path: path}
	}
	return dirStats[path]
}

// isUnderTargets checks if the path is under the user-specified search paths
func isUnderTargets(path string) bool {
	for _, root := range targetPaths {
		absRoot, _ := filepath.Abs(root)
		if strings.HasPrefix(path, absRoot) {
			return true
		}
	}
	return false
}

// removeSubdirectories cleans up the target list by removing subdirectories that are already covered by parent directories in the list
func removeSubdirectories(paths []string) []string {
	if len(paths) == 0 {
		return paths
	}

	// Convert to absolute paths
	absPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			fmt.Printf("Warning: Could not resolve path %s: %v\n", p, err)
			continue
		}
		absPaths = append(absPaths, abs)
	}

	// Sort paths to ensure parents come before children
	sort.Strings(absPaths)

	var clean []string
	for _, p := range absPaths {
		if len(clean) == 0 {
			clean = append(clean, p)
			continue
		}

		last := clean[len(clean)-1]
		// Check if p is a subdirectory of last
		// Must handle case where p is "/tmp/foo" and last is "/tmp"
		// Also handle "/tmp" and "/tmp" (duplicates)
		// And "/tmp" vs "/tmp2" (not subdirectory)

		if p == last {
			continue // Duplicate
		}

		// Ensure strictly child (starts with parent + separator)
		// or parent is root "/"
		isSub := false
		if strings.HasPrefix(p, last+string(os.PathSeparator)) {
			isSub = true
		} else if last == filepath.VolumeName(last)+string(os.PathSeparator) && strings.HasPrefix(p, last) {
			// specific check for root (e.g. C:\ or /)
			isSub = true
		}

		if !isSub {
			clean = append(clean, p)
		}
	}
	return clean
}

// --- Argument Parsing ---

func parseArgs() {
	args := os.Args[1:]
	// If no arguments provided, defaults will be used (targetPaths handled below)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--path":
			// Read all subsequent non-option arguments as paths
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				targetPaths = append(targetPaths, args[i+1])
				i++
			}
		case "--maxdepth":
			if i+1 < len(args) {
				val, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Println("Error: --maxdepth requires a numeric value")
					os.Exit(1)
				}
				maxDepth = val
				i++
			}
		case "--top":
			if i+1 < len(args) {
				val, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Println("Error: --top requires a numeric value")
					os.Exit(1)
				}
				topN = val
				i++
			}
		case "--verbose":
			verbose = true
		case "--display-runtime":
			displayRuntime = true
		case "--version":
			fmt.Printf("%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
			os.Exit(0)
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		default:
			fmt.Printf("Unknown option: %s\n", arg)
			printUsage()
			os.Exit(1)
		}
	}

	if len(targetPaths) == 0 {
		// Default to current directory if no path specified
		targetPaths = append(targetPaths, ".")
	}
}

func printUsage() {
	fmt.Println("Usage: find_heavy_dirs [--path <path1> path2...] [--maxdepth <N>] [--top <N>] [--verbose] [--display-runtime] [--version]")
	fmt.Println("Options:")
	fmt.Println("  --path <path...>: One or more paths to search. Default is current directory.")
	fmt.Println("  --maxdepth <N>:   Limit the search to N levels deep. Default is 1000000.")
	fmt.Println("  --top <N>:        Display the top N entries. Default is 20.")
	fmt.Println("  --verbose:        Show detailed progress information.")
	fmt.Println("  --display-runtime:Show total execution time.")
	fmt.Println("  --version:        Show program version.")
	fmt.Println("  -h, --help:       Show this help message.")
}

// --- Formatting Tools ---

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func printTable(title string, list []*DirStat, isSize bool) {
	fmt.Println("\n--- " + title + " ---")
	// Simple table header
	fmt.Printf("%-15s | %-50s\n", "Metric", "Path")
	fmt.Println(strings.Repeat("-", 70))

	limit := topN
	if len(list) < limit {
		limit = len(list)
	}

	for i := 0; i < limit; i++ {
		s := list[i]
		var valStr string
		if isSize {
			valStr = formatBytes(s.TotalSize)
		} else {
			valStr = fmt.Sprintf("%d Files", s.FileCount)
		}

		// Simple truncated path display to prevent ugly wrapping
		displayPath := s.Path
		if len(displayPath) > 80 {
			displayPath = "..." + displayPath[len(displayPath)-77:]
		}

		fmt.Printf("%-15s | %s\n", valStr, displayPath)
	}
}

/*
Change History:
2025-12-14:
 - Fixed a bug where the root directory was double-counted (self-aggregation) in statistics.
 - Added verbose warnings for access denied errors during scanning.

2025-12-13:
 - Added --top N parameter to specify the number of top entries (default 20).
 - Added --verbose parameter to control detailed output (default false).
 - Added --display-runtime parameter to control runtime display (default false).
 - Added --path parameter to specify multiple paths (default current directory).
 - Changed default --maxdepth to 1000000.
 - Added header and footer comments.
 - Added --version parameter to show program version.
*/
