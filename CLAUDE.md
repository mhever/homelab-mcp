## Workflow Orchestration

### 1. Mandatory Handoffs (Gemini Integration)
- **Code Generation & Modification:** You MUST delegate all code writing, refactoring, and file edits to the `gemini-coder` subagent. Do not apply code yourself.
- **Code Review:** You MUST pass all modified files to the `gemini-reviewer` subagent.
- **Iteration:** If `gemini-reviewer` flags issues, feed the exact output back to `gemini-coder` for remediation. Repeat until the reviewer passes the code.

### 2. Execution Rules
- **Plan First:** For tasks requiring 3+ steps, write a checklist to `tasks/todo.md` before execution.
- **Halt on Failure:** If a step fails twice, STOP. Do not loop blindly. Re-evaluate the plan.
- **Verification:** Run tests and check logs to prove the implementation works before marking a task complete.
- **Autonomous Resolution:** If given a bug report, use `gemini-coder` to resolve it directly based on logs/errors. 

### 3. Task Management & Memory
- **Tracking:** Update `tasks/todo.md` as items are completed.
- **Lessons Learned:** Upon user correction, immediately log the required behavior change in `tasks/lessons.md`.
- **Review History:** Read `tasks/lessons.md` at the start of every session to prevent repeating mistakes.

### 4. Code Standards
- Keep changes localized. Only touch necessary files.
- Resolve root causes; implement zero temporary workarounds.

Current plan is in @plan.md