---
id: tasks-ui
title: Terminal UI Engine
sidebar_position: 4
---
# Terminal UI Engine

This section documents the internal mechanics of the lightweight terminal UI rendering engine within Aeroflare.


The terminal UI engine (`internal/ui/`) is a lightweight, dependency-free implementation relying solely on the standard `fmt` and `strings` packages. It does not use heavy TUI frameworks, opting instead for manual string manipulation and ANSI escape codes.


### Box Rendering (`internal/ui/box.go`)
The `PrintSummaryBox` function constructs framed summary boxes for data presentation.
- **Dynamic Sizing:** It calculates the required box width by iterating through the `BoxField` structs (which contain `Label` and `Value`), determining the maximum label width, and adjusting the overall width to accommodate the longest combined line.
- **Unicode Borders:** It uses standard Unicode box-drawing characters (`╭`, `─`, `╮`, `│`, `├`, `┤`, `╰`, `╯`) to construct the frame. Formatting ensures labels are right-aligned to a common colon boundary.

### Table Rendering (`internal/ui/table.go`)
The `PrintTable` function generates Nushell-style data tables.
- **Dynamic Column Widths:** It performs a pass over all headers and rows (`[][]string`) to compute the maximum width (`colWidths`) needed for each column.
- **ANSI Colors:** Hardcoded ANSI escape sequences (`\x1b[90m` for Gray borders, `\x1b[36m` for Cyan headers, and `\x1b[0m` to Reset) are used to provide distinct styling without external dependencies.
- **Intersection Drawing:** A helper function `drawLine` dynamically constructs the horizontal border lines, taking care of placing intersection characters (`┬`, `┼`, `┴`) accurately based on the calculated `colWidths`.
