# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Go Runtime deep-dive research project** that explores the complete lifecycle of a Go program from compilation to execution, scheduling, and system calls. The project uses a simple `main.go` as a specimen to analyze Go's runtime internals at assembly level.

## Key Constraint

**IMPORTANT**: `proc.md` is the user's personal documentation file. Do NOT write to or modify `proc.md`. Instead:
- Create separate analysis files (e.g., `*-analysis.md`, `notes-*.md`)
- Provide research findings that the user can selectively integrate into `proc.md`

## Build and Analysis Commands

### Compilation

```bash
# Standard build
go build -o main_bin main.go

# Verbose build with detailed output
go build -x -work -o main_bin main.go 2>&1 | tee build_log.txt

# Build with escape analysis
go build -gcflags="-m -m" -o main_bin main.go

# Disable optimizations and inlining (for easier analysis)
go build -gcflags="-N -l" -o main_bin main.go
```

### Binary Analysis

```bash
# Check file type
file main_bin

# Read ELF header
readelf -h main_bin

# List sections
readelf -S main_bin

# List program headers
readelf -l main_bin

# View symbols
nm main_bin | grep -E "main|runtime|rt0"

# Disassemble specific function
objdump -d main_bin --start-address=0x<addr> --stop-address=0x<addr> -M intel

# Disassemble entry point (adjust address as needed)
objdump -d main_bin --start-address=0x46ce40 --stop-address=0x46cec0 -M intel
```

### Runtime Tracing and Debugging

```bash
# Scheduler trace (prints every 1000ms)
GODEBUG=schedtrace=1000,scheddetail=1 ./main_bin

# GC trace
GODEBUG=gctrace=1 ./main_bin

# Combined tracing
GODEBUG=schedtrace=1000,scheddetail=1,gctrace=1 ./main_bin

# System call trace (if available)
strace -c ./main_bin

# Check thread creation
timeout 2 ./main_bin &
PID=$!
ps -T -p $PID
```

### Go Runtime Source Analysis

```bash
# Go version
go version

# Find runtime source files
find /usr/local/go/src/runtime -name "*.go" | grep -E "proc|runtime2|asm"

# Key runtime files:
# - /usr/local/go/src/runtime/runtime2.go (g, m, p structs)
# - /usr/local/go/src/runtime/proc.go (scheduler)
# - /usr/local/go/src/runtime/asm_amd64.s (assembly entry points)
```

## Project Structure

### Core Files

- **main.go**: Simple specimen program with goroutines
- **proc.md**: User's primary documentation (DO NOT MODIFY)
- **README.md**: Complete index and knowledge graph

### Analysis Documents

Created as auxiliary research files:

1. **startup-analysis.md**: Compilation → ELF structure → program startup
   - Covers: `_rt0_amd64_linux` → `runtime.rt0_go` → `schedinit`

2. **gmp-scheduler-analysis.md**: GMP scheduling model deep dive
   - Core data structures: `g`, `m`, `p`
   - Scheduler initialization and main loop

3. **goroutine-lifecycle-analysis.md**: Goroutine creation and scheduling
   - `newproc` → `runqput` → `schedule` → `execute` → `goexit`

4. **syscall-and-os-interaction.md**: System calls and OS interaction
   - `entersyscall` / `exitsyscall` mechanisms
   - `sysmon` monitoring thread
   - Hand-off and preemption

### Build Artifacts

- **main_bin**: Compiled executable
- **build_log.txt**: Verbose build output
- **.claude/**: Claude Code session data (auto-generated)

## Research Methodology

This project follows a layered analysis approach:

```
1. Compilation Layer
   ├─ Source code → AST → SSA → Assembly
   └─ Link with runtime → ELF executable

2. Startup Layer
   ├─ OS loader → Entry point (_rt0_amd64_linux)
   ├─ Bootstrap (g0, m0, TLS setup)
   └─ Runtime initialization (schedinit)

3. Scheduling Layer
   ├─ GMP model (Goroutines, Machines, Processors)
   ├─ Work stealing and load balancing
   └─ Preemption and fairness

4. System Call Layer
   ├─ entersyscall / exitsyscall
   ├─ Hand-off mechanism
   └─ Netpoller (for async I/O)
```

## Common Workflows

### Adding New Analysis

When researching a new aspect of Go runtime:

1. Create a new `<topic>-analysis.md` file (NOT in proc.md)
2. Include relevant code snippets, assembly, or runtime source references
3. Use actual program traces (GODEBUG output) to validate findings
4. Cross-reference with Go runtime source code locations

Example structure:
```markdown
# <Topic> Analysis

## Source Code Reference
Location: `/usr/local/go/src/runtime/<file>.go:line`

## Actual Behavior
[Include GODEBUG traces or disassembly]

## Key Findings
[Your analysis]
```

### Disassembling Runtime Functions

To find and disassemble specific runtime functions:

```bash
# Find function address
nm main_bin | grep "runtime.<function_name>"

# Disassemble
objdump -d main_bin --start-address=0x<start> --stop-address=0x<end> -M intel
```

### Analyzing Runtime Source

When referencing Go runtime source:

```bash
# Read specific function
grep -n "^func <name>" /usr/local/go/src/runtime/proc.go

# Read struct definition
grep -n "^type g struct" /usr/local/go/src/runtime/runtime2.go
```

Always include the source file path and line number in documentation.

## Key Technical Details

### Entry Points and Addresses

- Entry point: `0x46ce40` (_rt0_amd64_linux) - varies by build
- Key runtime symbols:
  - `_rt0_amd64`: Platform-agnostic entry
  - `runtime.rt0_go`: Main bootstrap function
  - `runtime.schedinit`: Scheduler initialization
  - `runtime.main`: Main goroutine
  - `runtime.newproc`: Goroutine creation
  - `runtime.mstart`: M (thread) startup

### GMP Model Constants

- Initial goroutine stack: 2KB (`_StackMin`)
- P local queue size: 256 goroutines
- Global runqueue: unbounded (with lock)
- Max M count: 10000 (configurable via `sched.maxmcount`)
- GOMAXPROCS: defaults to CPU count

### Scheduler Trace Fields

When using `GODEBUG=schedtrace=1000,scheddetail=1`:

- **gomaxprocs**: Number of P's
- **idleprocs**: Idle P count
- **threads**: Total OS thread count
- **P status**: 0=idle, 1=running, 2=syscall, 3=gcstop, 4=dead
- **M.curg**: Current goroutine on M
- **G status**: 0=idle, 1=runnable, 2=running, 3=syscall, 4=waiting, 6=dead

## Documentation Style

- Use Chinese for main content (user preference)
- Include both high-level concepts and low-level details (assembly, addresses)
- Cross-reference between documents using relative links
- Provide actual command outputs and traces, not theoretical examples
- Mark source code locations: `runtime/proc.go:782`

## Integration with proc.md

The user maintains `proc.md` as their primary document. Your role:
1. Create comprehensive analysis files
2. Present findings for the user to review
3. Let the user decide what to integrate into `proc.md`
4. Never write directly to `proc.md`
