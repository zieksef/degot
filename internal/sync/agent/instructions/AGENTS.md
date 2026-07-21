# AGENTS.md

1. Ask, don't assume. If something is unclear, ask before writing a single line. Never make silent assumptions about intent, architecture, or requirements. When running unattended, pick the most reasonable interpretation, proceed, and record the assumption rather than blocking.

2. Don't over-engineer. No speculative abstraction, indirection, or flexibility that isn't needed yet; prefer small composable pieces over tightly coupled ones. Keep changes scoped to the task: leave unrelated code alone, but when the clean fix requires touching adjacent code, do it properly rather than working around it. Surface other bad code or design smells as separate issues.

3. If you see a clearly better approach, say so before implementing. Explain the tradeoff in 2-4 bullets. If the current request is still reasonable, proceed unless the alternative avoids serious risk or wasted work.

4. If what we are trying to do is similar to settled science or industry practice, let me know. We don’t have to reinvent the wheel.

5. Follow the team development standards. Before starting any task that writes, modifies, refactors, or reviews code — or that touches DB schema, APIs, logging, caching, or commits — you MUST invoke the `devguide` skill and follow it; don't substitute memory or guesswork. Before delivering, re-check your work against devguide's completion checklist.