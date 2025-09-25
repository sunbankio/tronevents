
# Implementation Plan: TRON Events Redis Stream Daemon

**Branch**: `001-a-daemon-to` | **Date**: 2025-09-25 | **Spec**: [TRON Events Redis Stream Daemon Spec](spec.md)
**Input**: Feature specification from `/specs/001-a-daemon-to/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
A daemon to continuously monitor the TRON blockchain for new events and publish them to a Redis stream with 7-day retention. The system handles interruptions gracefully by processing missed blocks upon restart while maintaining synchronization with ongoing blockchain activity. Redis will be used for both events publishing and as a worker queue system using the github.com/hibiken/asynq package, implementing priority, backlog, retry, and dead letter queues. The system will use up to 15 workers: one for new blocks (generated every 3 seconds), one for priority queue, and up to 13 for backlog queue with recent blocks (larger block numbers) prioritized over older blocks.

## Technical Context
**Language/Version**: Go (latest stable)
**Primary Dependencies**: github.com/hibiken/asynq (for queue management), Redis client library, existing TRON blockchain scanner (in /scanner/ package)
**Storage**: Redis (for event streaming and queue management)
**Testing**: Go testing package
**Target Platform**: Linux server
**Project Type**: Single service daemon
**Performance Goals**: Demonstrate high-throughput processing of large backlogs (up to 201600 blocks) and robustly handle interruptions and resumption of the process without data loss
**Constraints**: 7-day event retention in Redis stream, graceful handling of daemon interruptions, robust handling of server shutdown during backlog processing
**Scale/Scope**: Continuous monitoring of TRON blockchain, priority processing of recent events over backlog with up to 15 concurrent workers

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Since this project is written in Go and follows a library-first approach:
- Each feature should start as a standalone library that is self-contained and independently testable
- All libraries must have clear purpose with no organizational-only libraries
- TDD is mandatory: Tests written → User approved → Tests fail → Then implement
- Integration tests are required for new library contract tests, contract changes, inter-service communication, and shared schemas
- Structured logging is required for observability
- All PRs/reviews must verify compliance with these constitutional principles

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure]
```

**Structure Decision**: DEFAULT to Option 1 (Single project) since this is a daemon service

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

4. **Add implementation details from user requirements**:
   - Redis will be used for events publishing and worker queue
   - Use github.com/hibiken/asynq package for the queue system
   - Implement priority, backlog, retry, dead letter queues
   - Read last_synced_block using existing scanner.Scan(0) function (scanner already implemented in /scanner/ package)
   - Implement logic for handling synchronization states:
     * Normal operation: when last_synced_block matches returned block, wait 1 second and continue
     * First run or in-sync: when last_synced_block doesn't exist or is zero, or when returned_block_num=last_synced_block+1, update last_synced_block and wait 3 seconds
     * Slight backlog: when lagging at most 20 blocks, push gapped blocks' block numbers to priority queue, update last_synced_block and wait 3 seconds
     * Large backlog: when lagging more than 20 blocks, push block number range (max(returned_block-201600, last_synced+1) to returned_block-1) to backlog queue, update last_synced_block
     * Backlog processing order: Last In, First Out (recent backlogs processed first)
   - Robust handling of server shutdown during backlog processing
   - Worker configuration: max 15 workers (1 for new blocks, 1 for priority queue, up to 13 for backlog queue)

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh claude`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

6. **Additional design considerations**:
   - Design Redis stream structure for TRON events with 7-day retention
   - Design asynq queue configuration for priority, backlog, retry, and dead letter queues
   - Design block processing logic with different states (normal, backlog, etc.)
   - Design persistence mechanism for last_synced_block tracking
   - Design error handling and retry mechanisms with backoff times of 5, 10, 30, 60, 180, 300, 600, 1800, and 3600 seconds
   - Design worker configuration with max 15 workers (1 for new blocks, 1 for priority queue, up to 13 for backlog queue)
   - Design robust handling of server shutdown during backlog processing to prevent data loss
   - Design LIFO (Last In, First Out) processing for backlog queue to prioritize recent blocks
   - Integrate with existing TRON blockchain scanner in /scanner/ package rather than implementing new scanner

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each entity → model creation task [P]
- Each user story → integration test task
- Implementation tasks to make tests pass
- Tasks for Redis integration with asynq queue system
- Tasks for block processing logic with different states
- Tasks for graceful interruption handling and restart logic
- Tasks for worker configuration with 15 concurrent workers
- Tasks for server shutdown handling during backlog processing
- Tasks for implementing LIFO processing for backlog queue
- Tasks for integrating with existing TRON blockchain scanner (instead of implementing new scanner)

**Ordering Strategy**:
- TDD order: Tests before implementation
- Dependency order: Models before services before UI
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 25-30 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [ ] Phase 0: Research complete (/plan command)
- [ ] Phase 1: Design complete (/plan command)
- [ ] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [ ] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
