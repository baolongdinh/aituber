# Agent Memory Skill — Long-term Context

## TRIGGER
Read this file **FIRST** in every new session.
Perform **2 mandatory steps**: read old context → do task → write new context.

---

## STEP 1: READ CONTEXT AT SESSION START

Before doing any task in a new session, read in this order:

```
1. .antigravity/context/PROJECT_STATE.md   → Where project is, what's been done
2. .antigravity/context/DECISIONS.md       → Technical decisions that were made
3. .antigravity/context/FEATURES.md        → List of features done/in progress/not started
```

> If user asks "where are we?" or "what have we done?" → read context and answer based on that.

---

## STEP 2: WRITE CONTEXT AFTER EACH TASK

After completing a task and before committing, update context:

### Update `PROJECT_STATE.md`
- Write **current state** of the project
- What was completed in the just-finished task
- What's still in progress (if any)

### Update `FEATURES.md`
- Mark feature just done: `[x]` if complete, `[/]` if in progress
- Add new feature if discovered during work

### Update `DECISIONS.md`
- If just made important technical decision (schema, architecture, new pattern) → write here
- No need to record daily small decisions

---

## CONTEXT WRITING FORMAT

### When updating PROJECT_STATE.md
```markdown
## [YYYY-MM-DD] — [Task name just done]

**Completed:**
- ...

**Current State:**
- Backend: [description]
- Frontend: [description]
- Contracts: [description]

**Not Done / In Progress:**
- ...
```

### When updating DECISIONS.md
```markdown
## [YYYY-MM-DD] — [Decision name]

**Problem:** ...
**Chose:** ...
**Reason:** ...
```

---

## DO / DON'T

✅ **DO**
- Read context before each new session
- Update context after each important task
- Write enough detail for next session to understand without asking

❌ **DON'T**
- NEVER write too detailed (don't copy/paste code into context)
- NEVER skip context update after big task
- NEVER record small decisions without long-term reference value
