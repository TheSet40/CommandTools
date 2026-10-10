
# The goal

**A copy of Star Wars: Empire at War with the Thrawn's Revenge 3.5 mod, as close to 1:1 as
possible, plus the improvements in [`Improvements.md`](Improvements.md) over the base mod.** Built
in this order, each step completed before the next where possible:

1. Space battles.
2. The galactic map.
3. Ground battles.

The mod's assets are used as they are, wherever they are accessible, and the mod's data is treated
as the definitive answer to almost every question a 1:1 copy raises. A departure from the mod is
either an item of `Improvements.md`, something the engine cannot yet do, or a mistake; the first
two are recorded in [`Notes/divergences.md`](Notes/divergences.md) with their reason, so they can be told from the third.

# The baseline is Thrawn's Revenge, not vanilla Empire at War

**Every rule this game implements should match the Thrawn's Revenge mod, not the base game.** Where
the two differ, the mod wins — silently assuming vanilla behaviour has already produced real bugs.
When a combat rule is unclear, the answer is in the mod's data, not in memory of how Empire at War
played.

**[`Notes/mod-index.md`](Notes/mod-index.md) is the index to the mod's files.** Read it before
touching mod XML: it maps the layout, the record types, and — most importantly — the three
inheritance mechanisms (`Variant_of_Existing_Type`, `Template_*` records, dummy/container records).
A field's real value is often several records up a variant chain, so reading one record and stopping
gives the wrong answer.

Places the mod diverges from vanilla in ways that are easy to get wrong are recorded in
[`Notes/thrawnsRevengeRules.md`](Notes/thrawnsRevengeRules.md), with the file path and quoted XML
behind each one. Add to it whenever research turns up another.

# Notes

**Documents that are not plans go in [`Notes/`](Notes/)** -- reference material (the mod index, the
mod's rules, surveys), write-ups the user asked for, and anything deferred or too large to start
soon, which goes in [`Notes/deferred.md`](Notes/deferred.md). `.agent/{date}-plans/` holds plans
only: work about to be done, with its checkboxes. A plan that finds something it will not do writes
it up in `Notes/` and links it, rather than leaving it in the plan.

# Architecture

**Read [ARCHITECTURE.md](ARCHITECTURE.md) before doing any structural work.** It is the map of this repository: the engine/game seam, what lives in which package, how a frame flows from `Application` to the renderer, the contracts that are easy to break silently (instancing, assets, the grid), and a register of known caveats and compromises.
If the document doesn't exist, create it and add an overview of the repository's architecture.

Consult it when you need to:
- find where something lives, or which package owns a concern
- understand how the engine and a game talk to each other — through `Scene`, `AppState` and the `Game` interface, and deliberately nothing else
- add a mesh, an instanced layer, a texture, a tile kind or a whole game end-to-end
- check whether a problem you just hit is already a known caveat

### Keeping ARCHITECTURE.md current

Update it **in the same change** that makes it wrong. Specifically:

- An executable, engine package or game is added, removed or renamed
- The engine/game seam changes — anything added to `Game`, `Scene`, `AppState` or `EngineSettings`
- A folder convention changes, or a new top-level folder appears
- A rendering pass, shader or framebuffer is added or reordered
- A contract that fails **silently** changes: instance `revision`, `Grid::setWall`, asset lists in `CMakeLists.txt`, the model/texture loaders
- A file format gains a version or a record (`sceneFile`, `.rrb` boards, `.course` files)
- The steps to add a mesh, layer, tile kind, board or game change
- The demo's regression numbers move (draw calls, triangles, the camera pose `PERFORMANCE.md` was measured from)

### Recording caveats and compromises

When you knowingly ship a shortcut, workaround, partial implementation or disabled feature, **add a row to the caveats register — the last section of [ARCHITECTURE.md](ARCHITECTURE.md)** — do not leave it only in a commit message or a code comment. Refer to it by name rather than by number: sections get renumbered, and a stale cross-reference sends the next reader nowhere.

Each entry needs a file reference, a one-line reason it matters, and a severity:

- 🔴 exploitable or data-losing
- 🟠 broken or silently disabled behaviour
- 🟡 a deliberate design compromise with a real cost
- ⚪ hygiene / dead code

If the compromise was made *because* the correct fix was too large, say so — that is the information a future reader needs. Remove the row when you fix it. The register is not a general TODO list; it is for things that would surprise someone reading the code.



### 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

- State assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them — don't pick silently unless one option scales mutch better than the others.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No abstractions for single-use code.
- No error handling for impossible scenarios.
- Follow YAGNI priciples wherever possible unless told otherwise.

**The test:** Would a senior engineer say this is overcomplicated? If yes, simplify.

### 3. Focus
- when implementing new functionality, consider the performance impact and how the performance scales with data size.
- if the performance impact is large, consider if the changes improve the dynamisity or error rate of the system enough to warrant the change, if not then ask.  



## Security Guidelines
- Never hardcode secrets in configuration files. All sensitive values must be injected via environment variables or a secrets manager.
- Never expose internal exception messages, stack traces, or system details to API consumers.** This prevents information disclosure that attackers could exploit.
- Always validate and sanitize user input at system boundaries.
- Always use parameterized queries. Never concatenate user input into SQL.
- Never log sensitive data. This includes passwords, tokens, credit card numbers, personally identifiable information (PII), etc.


### ✅ Always
- Always prioritize self describing code over comments
- Organize code by domain package, not by technical layer



### 🚫 Never
- Commit secrets, API keys, or credentials without asking first and only ask when nessesary



# Task execution plan
Important: Always plan the task step by step before writing code. Ask for permission to proceed with the plan if there are any complex steps or more than 2 Intermediate steps.
Important: Before proceeding with the plan or making changes, create a new file named `.agent/{date}-plans/name-of-the-task.md` in the git repository root, if not in a git repo ask where to put the file. Based on the approved plan, list all necessary implementation steps as GitHub-style checkboxes (`- [ ] Step Description`). Use sub-bullets for granular details within each main step.

- Plans should be detailed enough to execute without ambiguity
- Each task in the plan must include at least one validation test to verify it works
- Assess complexity and single-pass feasibility - can an agent realistically complete this in one go?
- Include a complexity indicator at the top of each plan:
✅ Simple - Single-pass executable, low risk
⚠️ Intermediate - May need iteration, some complexity
🔴 Complex - Break into sub-plans before executing


**After you successfully complete each step, update the `.agent/{date}-plans/name-of-the-task.md` file by changing the corresponding checkbox from `- [ ]` to `- [x]`. to mark the step as complete**
Announce which step you are starting.

### Code Rules
- always use double quotes when possible and always use semicolons
- always use curly braces for all control flow statements, even if they are single
- always use camelcase when naming things,
- always write code and comments in english