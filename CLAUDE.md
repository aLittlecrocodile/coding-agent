# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is an autonomous AI agent framework that implements a self-maintaining development loop using Claude as the reasoning engine. The system orchestrates four specialized roles (Planner, Executor, Reviewer, Loop Controller) to iteratively work toward a goal through plan generation, execution suggestion, review, and control decisions.

## Architecture

### Core Loop Driver

**loop_driver.py** - Main orchestration engine that runs the multi-agent loop:
- Manages state across iterations (goal, plan, actions, notes, results, stop_reason)
- Calls each agent role in sequence: Planner → Executor → Reviewer → Loop Controller
- Handles JSON parsing with tolerance for extra text/markdown code blocks
- Implements retry logic (2 attempts) for non-JSON outputs
- Logs all interactions to `runs/run-{timestamp}.jsonl` for audit trail
- Enforces maximum iterations (default 10, configurable via MAX_ITER)

### Agent Roles (prompts/ directory)

Each role is defined as a prompt file that instructs Claude on its specific function:

**00_system_constraints.md** - Global constraints applied to all roles:
- JSON-only output (no explanations, no markdown fences)
- snake_case key naming
- Safety-first: never output destructive commands without marking `destructive=true`
- Minimal viable output when input is ambiguous
- Fail-safe: prefer stop/pause signals on format errors

**01_planner.prompt.md** - Planner role:
- Generates step-by-step plan and next actions
- Identifies risks and assumptions
- Output: `{plan, next_actions, risks, assumptions}`

**02_executor.prompt.md** - Executor role:
- Suggests specific commands/shell actions
- Marks destructive operations
- Provides expected outcomes and fallbacks
- Output: `{commands: [{cmd, why, destructive}], expected_outcomes, fallbacks}`

**03_reviewer.prompt.md** - Reviewer role:
- Evaluates execution suggestions against plan
- Identifies findings (issues/risks) and deltas (deviations)
- Suggests adjustments
- Output: `{findings, deltas, suggested_changes}`

**04_loop_controller.prompt.md** - Loop Controller role:
- Decides whether to continue, adjust plan, pause, or stop
- Provides reasoning for decision
- Optionally patches plan when adjusting
- Output: `{decision, reason, plan_patch?}`

### State Management

The loop maintains a shared state dictionary:
```python
{
    "iteration": int,
    "goal": str,           # From LOOP_GOAL env var
    "plan": list,          # Updated each iteration
    "actions": list,       # Commands suggested by Executor
    "notes": list,         # Findings from Reviewer
    "results": str,        # Deltas from Reviewer
    "stop_reason": str     # Reason for stopping
}
```

## Build and Run Commands

### Dependencies
```bash
pip install anthropic
```

### Environment Variables (Required)
```bash
export ANTHROPIC_API_KEY="your-api-key"
```

### Environment Variables (Optional)
```bash
export LOOP_GOAL="Your development goal here"
export CLAUDE_MODEL="claude-3-opus-20240229"  # Or other Claude models
export MAX_ITER="10"  # Maximum loop iterations
```

### Running the Loop
```bash
python loop_driver.py
```

**Important**: The current implementation does NOT execute commands automatically. It only logs suggested commands from the Executor. To enable actual execution:
1. Modify loop_driver.py:152 to implement command execution with confirmation
2. Capture actual output and pass it to the Reviewer
3. Add proper error handling and rollback mechanisms

## Output and Logging

All loop iterations are logged to `runs/run-{timestamp}.jsonl` in JSONL format:
```json
{"event": "planner", "payload": {...}}
{"event": "executor", "payload": {...}}
{"event": "reviewer", "payload": {...}}
{"event": "controller", "payload": {...}}
```

Each file contains a complete audit trail of one run.

## Design Decisions

- **Safety-first**: No automatic command execution; requires manual integration
- **JSON-only protocol**: All roles output pure JSON for programmatic consumption
- **Retry tolerance**: Handles non-JSON responses with up to 2 correction attempts
- **Structured prompts**: Each role has clearly defined schema and constraints
- **Auditable**: Every interaction logged to timestamped files
- **Graceful degradation**: Missing fields filled with empty strings/lists
- **Schema enforcement**: Required fields guaranteed via `ensure_schema()`

## Extending the Framework

### Adding New Agent Roles
1. Create new prompt file in `prompts/` (e.g., `05_new_role.prompt.md`)
2. Define JSON output schema in the prompt
3. Add call step in `LoopDriver.run()` method
4. Update state handling if role produces new data

### Modifying the Loop
- Edit `LoopDriver.run()` in loop_driver.py (line 135)
- Current sequence: Planner → Executor → Reviewer → Controller
- Can inject custom logic between steps (e.g., actual command execution)

### Custom State Fields
Add fields to `self.state` initialization (line 75) and update state sync logic (line 173)

## Language and Style

- Chinese comments and prompt text throughout
- English variable names and code structure
- JSON schema in prompts uses English keys
- Log messages in Chinese for human readability

## Important Constraints

- The system is designed for supervised, not autonomous, operation
- Commands are suggested, not executed
- No external tool usage beyond Claude API calls
- State persists only in memory during run (not persisted to disk except logs)
- Maximum iterations enforced to prevent infinite loops
- No git integration or file modification (read-only operation)
