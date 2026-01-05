package config

// CortexfileTemplate is the default template for a new Cortexfile.yml
const CortexfileTemplate = `# Cortexfile.yml - Cortex Workflow Configuration
# Documentation: https://github.com/obliviious/cortex

# Optional: Set working directory for all agents
# workdir: /path/to/project

# Define your AI agents
agents:
  # Analyzer agent for code analysis
  analyzer:
    tool: claude-code
    model: sonnet

  # Reviewer agent for code review
  reviewer:
    tool: claude-code
    model: sonnet

  # Coder agent for implementation
  coder:
    tool: claude-code
    model: sonnet

# Define your workflow tasks
tasks:
  # Analysis task - runs first
  analyze:
    agent: analyzer
    prompt: |
      Analyze the codebase structure and identify:
      1. Main components and their responsibilities
      2. Key dependencies and their versions
      3. Potential areas for improvement

      Provide a concise summary.

  # Review task - runs after analysis
  review:
    agent: reviewer
    prompt: |
      Review the codebase for:
      1. Code quality issues
      2. Security vulnerabilities
      3. Performance concerns
      4. Best practices violations

      Provide actionable recommendations.

  # Implementation task - depends on analysis and review
  implement:
    agent: coder
    needs: [analyze, review]
    write: true
    prompt: |
      Based on the analysis:
      {{.analyze}}

      And the review findings:
      {{.review}}

      Implement the suggested improvements.

# Optional: Local settings (override global config)
# settings:
#   parallel: true
#   max_parallel: 2
#   verbose: false
#   stream: true
`

// MasterCortexTemplate is the default template for a new MasterCortex.yml
const MasterCortexTemplate = `# MasterCortex.yml - Multi-Project Workflow Orchestration
# Documentation: https://github.com/obliviious/cortex

# Name of this master workflow
name: multi-project-workflow

# Description of what this workflow does
description: Orchestrates multiple Cortex workflows across projects

# Execution mode: "sequential" or "parallel"
mode: sequential

# Maximum parallel workflows (only used in parallel mode, 0 = unlimited)
max_parallel: 2

# Stop on first error (default: true for sequential, false for parallel)
stop_on_error: true

# Global variables available to all workflows
variables:
  environment: development
  output_dir: ./results

# Define workflows to run
workflows:
  # Run the main project workflow
  - name: main-project
    path: ./Cortexfile.yml
    # workdir: ./main-project  # Optional: override working directory

  # Run workflows in subdirectories
  - name: backend
    path: ./backend/Cortexfile.yml
    workdir: ./backend

  - name: frontend
    path: ./frontend/Cortexfile.yml
    workdir: ./frontend
    needs: [backend]  # Wait for backend to complete first

  # Use glob patterns to run multiple similar projects
  # - name: microservices
  #   path: "./services/*/Cortexfile.yml"

  # Disabled workflow (kept for reference)
  # - name: experimental
  #   path: ./experimental/Cortexfile.yml
  #   enabled: false
`

// MinimalCortexfileTemplate is a minimal template for quick start
const MinimalCortexfileTemplate = `# Cortexfile.yml
agents:
  assistant:
    tool: claude-code
    model: sonnet

tasks:
  main:
    agent: assistant
    write: true
    prompt: |
      # Your prompt here
      Describe what you want the AI to do.
`
