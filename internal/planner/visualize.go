package planner

import (
	"fmt"
	"sort"
	"strings"
)

// GraphFormat specifies the output format for graph rendering
type GraphFormat string

const (
	FormatASCII GraphFormat = "ascii"
	FormatDOT   GraphFormat = "dot"
)

// RenderGraph renders the DAG in the specified format
func RenderGraph(dag *DAG, tasks []ExecutionTask, format GraphFormat) string {
	switch format {
	case FormatDOT:
		return RenderDOT(dag, tasks)
	default:
		return RenderASCII(dag, tasks)
	}
}

// RenderASCII renders the DAG as ASCII art with box-drawing characters
func RenderASCII(dag *DAG, tasks []ExecutionTask) string {
	if dag.Size() == 0 {
		return "No tasks to display.\n"
	}

	levels := BuildExecutionLevels(dag)
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("\n◆ Execution Graph (%d tasks, %d levels)\n", dag.Size(), len(levels)))
	sb.WriteString("═══════════════════════════════════════════════════════\n\n")

	// Build task info map for quick lookup
	taskInfo := make(map[string]ExecutionTask)
	for _, t := range tasks {
		taskInfo[t.Name] = t
	}

	// Render each level
	for levelIdx, level := range levels {
		sb.WriteString(renderLevel(levelIdx, level, dag, taskInfo))

		// Draw connections to next level if not last
		if levelIdx < len(levels)-1 {
			sb.WriteString(renderConnections(level, levels[levelIdx+1], dag))
		}
	}

	// Legend
	sb.WriteString("\n─────────────────────────────────────────────────────────\n")
	sb.WriteString("Legend: ┌─┐ task box │ → dependency │ ▼ flow direction\n")

	return sb.String()
}

// renderLevel renders a single execution level with task boxes
func renderLevel(levelIdx int, level ExecutionLevel, dag *DAG, taskInfo map[string]ExecutionTask) string {
	var sb strings.Builder

	// Sort tasks for consistent output
	sortedTasks := make([]string, len(level.Tasks))
	copy(sortedTasks, level.Tasks)
	sort.Strings(sortedTasks)

	// Calculate box widths
	boxWidth := 14 // minimum width
	for _, name := range sortedTasks {
		if len(name)+4 > boxWidth {
			boxWidth = len(name) + 4
		}
	}
	if boxWidth > 20 {
		boxWidth = 20 // max width
	}

	// Level header
	parallelNote := ""
	if len(level.Tasks) > 1 {
		parallelNote = " (parallel)"
	}
	sb.WriteString(fmt.Sprintf("Level %d%s:\n", levelIdx, parallelNote))

	// Draw boxes - top border
	sb.WriteString("  ")
	for i := range sortedTasks {
		if i > 0 {
			sb.WriteString("   ")
		}
		sb.WriteString("┌")
		sb.WriteString(strings.Repeat("─", boxWidth))
		sb.WriteString("┐")
	}
	sb.WriteString("\n")

	// Draw boxes - content (task name)
	sb.WriteString("  ")
	for i, name := range sortedTasks {
		if i > 0 {
			sb.WriteString("   ")
		}
		displayName := name
		if len(displayName) > boxWidth-2 {
			displayName = displayName[:boxWidth-5] + "..."
		}
		padding := boxWidth - len(displayName)
		leftPad := padding / 2
		rightPad := padding - leftPad
		sb.WriteString("│")
		sb.WriteString(strings.Repeat(" ", leftPad))
		sb.WriteString(displayName)
		sb.WriteString(strings.Repeat(" ", rightPad))
		sb.WriteString("│")
	}
	sb.WriteString("\n")

	// Draw boxes - agent/tool info
	sb.WriteString("  ")
	for i, name := range sortedTasks {
		if i > 0 {
			sb.WriteString("   ")
		}
		info := ""
		if t, ok := taskInfo[name]; ok {
			info = t.Tool
			if t.Model != "" {
				info += "/" + t.Model
			}
		}
		if len(info) > boxWidth-2 {
			info = info[:boxWidth-5] + "..."
		}
		padding := boxWidth - len(info)
		leftPad := padding / 2
		rightPad := padding - leftPad
		sb.WriteString("│")
		sb.WriteString(strings.Repeat(" ", leftPad))
		sb.WriteString(info)
		sb.WriteString(strings.Repeat(" ", rightPad))
		sb.WriteString("│")
	}
	sb.WriteString("\n")

	// Draw boxes - bottom border
	sb.WriteString("  ")
	for i := range sortedTasks {
		if i > 0 {
			sb.WriteString("   ")
		}
		sb.WriteString("└")
		sb.WriteString(strings.Repeat("─", boxWidth))
		sb.WriteString("┘")
	}
	sb.WriteString("\n")

	return sb.String()
}

// renderConnections renders the arrows connecting levels
func renderConnections(currentLevel, nextLevel ExecutionLevel, dag *DAG) string {
	var sb strings.Builder

	// Find which tasks in next level depend on tasks in current level
	hasConnection := false
	for _, nextTask := range nextLevel.Tasks {
		deps := dag.GetDependencies(nextTask)
		for _, dep := range deps {
			for _, currTask := range currentLevel.Tasks {
				if dep == currTask {
					hasConnection = true
					break
				}
			}
		}
	}

	if hasConnection {
		// Simple arrow down
		sb.WriteString("        │\n")
		sb.WriteString("        ▼\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

// RenderDOT renders the DAG in Graphviz DOT format
func RenderDOT(dag *DAG, tasks []ExecutionTask) string {
	var sb strings.Builder

	sb.WriteString("digraph ExecutionGraph {\n")
	sb.WriteString("    rankdir=TB;\n")
	sb.WriteString("    node [shape=box, style=rounded, fontname=\"Arial\"];\n")
	sb.WriteString("    edge [arrowhead=vee];\n\n")

	// Build task info map
	taskInfo := make(map[string]ExecutionTask)
	for _, t := range tasks {
		taskInfo[t.Name] = t
	}

	// Build levels for subgraph grouping
	levels := BuildExecutionLevels(dag)

	// Create subgraphs for each level
	for levelIdx, level := range levels {
		sb.WriteString(fmt.Sprintf("    subgraph cluster_level%d {\n", levelIdx))
		sb.WriteString(fmt.Sprintf("        label=\"Level %d\";\n", levelIdx))
		sb.WriteString("        style=dashed;\n")
		sb.WriteString("        color=gray;\n")

		for _, taskName := range level.Tasks {
			label := taskName
			if t, ok := taskInfo[taskName]; ok {
				label = fmt.Sprintf("%s\\n(%s)", taskName, t.Tool)
				if t.Model != "" {
					label = fmt.Sprintf("%s\\n(%s/%s)", taskName, t.Tool, t.Model)
				}
			}
			sb.WriteString(fmt.Sprintf("        \"%s\" [label=\"%s\"];\n", taskName, label))
		}
		sb.WriteString("    }\n\n")
	}

	// Add edges (dependencies)
	sb.WriteString("    // Dependencies\n")
	for taskName, deps := range dag.Edges {
		for _, dep := range deps {
			sb.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\";\n", dep, taskName))
		}
	}

	sb.WriteString("}\n")

	return sb.String()
}

// RenderCompact renders a compact single-line representation of the DAG
func RenderCompact(dag *DAG) string {
	levels := BuildExecutionLevels(dag)
	if len(levels) == 0 {
		return "No tasks"
	}

	var parts []string
	for _, level := range levels {
		sort.Strings(level.Tasks)
		if len(level.Tasks) == 1 {
			parts = append(parts, level.Tasks[0])
		} else {
			parts = append(parts, "["+strings.Join(level.Tasks, ", ")+"]")
		}
	}

	return strings.Join(parts, " → ")
}
