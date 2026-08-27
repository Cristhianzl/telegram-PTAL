# Heuristics & oracles

Purpose: the idea-generators and problem-detectors that make exploration systematic instead of random. Heuristics tell you *what to try*; oracles tell you *whether what happened is wrong*.

The canonical source for most of these is James Bach's [Heuristic Test Strategy Model (HTSM)](https://www.satisfice.com/download/heuristic-test-strategy-model) at Satisfice, and Michael Bolton's oracle work at [DevelopSense](https://developsense.com).

---

## SFDIPOT — product elements (what to cover)

Sweep the product through seven lenses so coverage isn't accidental. From the HTSM "Product Elements".

| Element     | Probe |
|-------------|-------|
| **S**tructure | The things the product is made of — files, code, hardware, dependencies. What if one is missing or corrupt? |
| **F**unction  | What the product does — features, calculations, error handling, interactions between functions. |
| **D**ata      | What it processes — inputs, outputs, stored state, defaults, large/small/empty data, lifecycle (CRUD). |
| **I**nterfaces | How you reach it — UI, API, CLI, import/export, integrations with other systems. |
| **P**latform  | What it depends on — OS, browser, device, locale, fonts, screen size, network, external services. |
| **O**perations | How it's actually used — typical vs extreme vs malicious usage, who the users are, their environment. |
| **T**ime      | Timing and sequence — concurrency, race conditions, timeouts, time zones, ordering, fast/slow input, day boundaries. |

(Some versions drop the I and call it **SFDPOT**; same idea, Interfaces folded into Function/Platform.)

---

## CRUSSPIC STMPL — quality criteria (what "good" means)

The dimensions a product can be good or bad along. From the HTSM "Quality Criteria". Use it to ask "is this *good enough* at X?" not just "does X work?"

- **C**apability — does it do what it's supposed to?
- **R**eliability — does it keep working under stress, over time, on failure?
- **U**sability — can users actually accomplish their goals?
- **S**ecurity — does it protect data and resist abuse?
- **S**calability — does it hold up as load/data/users grow?
- **P**erformance — is it responsive and resource-efficient?
- **I**nstallability — does it install, configure, upgrade, uninstall cleanly?
- **C**ompatibility — does it play well with other systems, formats, versions?

Plus the development/internal-facing criteria — **STMPL**:

- **S**upportability — can support staff diagnose and help?
- **T**estability — can it be observed and controlled enough to test?
- **M**aintainability — can it be changed safely?
- **P**ortability — can it move to other platforms?
- **L**ocalizability — can it adapt to other languages, regions, formats?

---

## RCRCRC — regression heuristic (where regressions hide)

Where to focus when time is short after a change. From Karen N. Johnson ([RCRCRC](https://karennjohnson.com/regression-testing-cheat-sheet-rcrcrc/)).

- **R**ecent — new features and recently changed code.
- **C**ore — functions essential to the product's purpose; if these break, nothing else matters.
- **R**isky — areas known to be fragile, complex, or historically defect-prone.
- **C**onfiguration — settings, environment, and data that vary between users.
- **R**epaired — code recently fixed (fixes introduce new bugs).
- **C**hronic — areas that break again and again.

---

## FEW HICCUPPS — consistency oracles (how you recognize a bug)

An oracle is a principle by which you decide something is a problem. All of them are **fallible** — none is authoritative alone, so apply several. A bug is usually an *inconsistency* between the product and one of these. From Michael Bolton, [DevelopSense](https://developsense.com/blog/2012/07/few-hiccupps).

A product should be consistent with:

- **F**amiliarity — known patterns of failure (it shouldn't behave like a classic bug).
- **E**xplainability — you should be able to explain its behavior coherently.
- **W**orld — facts about the world (a date of Feb 30 is wrong).
- **H**istory — its own past behavior (it worked yesterday).
- **I**mage — the brand/reputation the org wants to project.
- **C**omparable products — similar tools and competitors.
- **C**laims — what docs, marketing, specs, and help text say it does.
- **U**ser expectations — what a reasonable user would expect.
- **P**roduct — itself; internal consistency across screens/features.
- **P**urpose — the explicit and implicit reasons it exists.
- **S**tandards/Statutes — relevant standards, conventions, laws, regulations.

When you report a bug, name the oracle: "inconsistent with **Claims** — the help text promises X, the product does Y."

---

## Data-variation patterns (how to attack any input)

For every input field, parameter, or payload, run through these deliberately.

- **CRUD lifecycle** — Create, Read, Update, Delete the same entity; check each transition and what happens to related data.
- **Goldilocks** — too big, too small, just right (for sizes, lengths, quantities, durations).
- **Boundaries** — at the limit, just below, just above; watch for off-by-one. (Max length 50 → try 49, 50, 51.)
- **Position** — beginning, middle, end (of a list, string, range, sort order).
- **Count** — zero, one, many (no items, exactly one, a large collection).
- **Data attacks** — empty string, null/absent, very long, special characters, Unicode, emoji, negative numbers, leading/trailing spaces, mixed case, format-breaking values, and injection strings (SQL/HTML/script/path) where relevant.

---

## Whittaker's tours — charters in a box

Each tour is a lens for moving through the product, and each makes a ready-made charter. From James Whittaker's *Exploratory Software Testing*.

- **Guidebook / Landmark tour** — follow the documentation/manual step by step; do the features match the claims? Hit the "landmark" features in sequence.
- **Money tour** — exercise the features that make the money or that sales demos show; these must never break.
- **Intellectual tour** — ask the product hard questions: feed it the most complex, demanding inputs it claims to handle.
- **FedEx tour (follow the data)** — pick a piece of data and follow it everywhere it travels: input, storage, display, export, downstream systems. Does it stay consistent?
- **Supermodel tour** — ignore function; look only at the surface — layout, styling, responsiveness, fonts, alignment, loading states.
- **Garbage Collector tour** — visit every screen/dialog/menu methodically, like cleaning a house room by room; pure coverage.
- **Saboteur / Bad-Neighborhood tour** — actively try to break it: kill the network mid-action, deny permissions, corrupt files, revisit screens already known to be buggy.

Use a tour as the *resource* clause of a charter: "Explore the report export **with a follow-the-data (FedEx) tour** to discover where numbers diverge between the dashboard, the CSV, and the PDF."
