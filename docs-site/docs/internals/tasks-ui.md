---
id: tasks-ui
title: Background Tasks & UI
sidebar_position: 4
---
# Background Tasks & UI

This section documents the internal mechanics of the background tasks (such as garbage collection) and the terminal UI rendering engine.

## Garbage Collection Algorithm

The garbage collection implementation in Aeroflare (`src/gc.go`) operates on the `PushCacheIndex` to manage storage and prune unreferenced paths.

### Reference Extraction and Graph Building
1. **Narinfo Parsing:** For each entry in the index, the algorithm extracts references by parsing the `Narinfo` string. It specifically looks for lines starting with `References:` and splits the subsequent fields (e.g., `hash-name`) by `-` to capture the base hashes.
2. **Graph Construction:** A directed dependency graph is built in memory, represented as a `map[string][]string` mapping each hash to its direct dependencies.

### Live Path Traversal
The GC algorithm employs a breadth-first search (BFS) to discover all reachable, "live" paths:
1. **GCRoots:** The traversal starts from the roots defined in `index.GCRoots`.
2. **Queue-based BFS:** A simple slice (`[]string`) acts as a queue, paired with a `map[string]bool` (`live`) to track visited nodes. The traversal visits each referenced node recursively until the queue is exhausted.

### Dead Path Identification and Sweeping
1. **Identification:** Any hash present in `index.Entries` but not in the `live` map is marked as a dead path.
2. **Sweeping:** The algorithm iterates through the sorted dead paths, removing them from `index.Entries` and accumulating the freed bytes based on each entry's `NarSize`.
3. **Early Exit Constraint:** If `maxFreed > 0`, the sweep phase terminates early as soon as the total `FreedBytes` meets or exceeds `maxFreed`.

The `RunGC` function returns a `GCResult` struct containing sorted slices of `LivePaths` and `DeadPaths`, along with the total `FreedBytes`.

## Terminal UI Rendering Engine

The terminal UI engine (`src/ui/`) is a lightweight, dependency-free implementation relying solely on the standard `fmt` and `strings` packages. It does not use heavy TUI frameworks, opting instead for manual string manipulation and ANSI escape codes.

### Box Rendering (`src/ui/box.go`)
The `PrintSummaryBox` function constructs framed summary boxes for data presentation.
- **Dynamic Sizing:** It calculates the required box width by iterating through the `BoxField` structs (which contain `Label` and `Value`), determining the maximum label width, and adjusting the overall width to accommodate the longest combined line.
- **Unicode Borders:** It uses standard Unicode box-drawing characters (`╭`, `─`, `╮`, `│`, `├`, `┤`, `╰`, `╯`) to construct the frame. Formatting ensures labels are right-aligned to a common colon boundary.

### Table Rendering (`src/ui/table.go`)
The `PrintTable` function generates Nushell-style data tables.
- **Dynamic Column Widths:** It performs a pass over all headers and rows (`[][]string`) to compute the maximum width (`colWidths`) needed for each column.
- **ANSI Colors:** Hardcoded ANSI escape sequences (`\x1b[90m` for Gray borders, `\x1b[36m` for Cyan headers, and `\x1b[0m` to Reset) are used to provide distinct styling without external dependencies.
- **Intersection Drawing:** A helper function `drawLine` dynamically constructs the horizontal border lines, taking care of placing intersection characters (`┬`, `┼`, `┴`) accurately based on the calculated `colWidths`.
