---
model: opus
effort: medium
automerge: false
---

# The auto-merge workflow re-arms auto-merge after pushes from outside contributors

## Problem
Found by the security scan of 2026-11-02 (finding M2). Severity: Medium. It can be triggered by the author of a pull request coming from a fork, once a maintainer has labelled it `automerge`.

GitHub disables auto-merge when someone without write access pushes to a pull request, so that changes made after the review don't get merged unseen. `.github/workflows/auto-merge.yml` undoes that protection: every 30 minutes (`.github/workflows/auto-merge.yml:26-27`) and after each push to main, it lists every open PR labelled `automerge` (`:53`), forks included, and enables auto-merge again with `CI_PAT` when it is off (`:72-78`). The protection of main requires no review, only the `build` and `json-schema-check` checks. So after a maintainer labels a fork PR, its author can push any further commit, and it is merged into main within 30 minutes once CI is green, without anyone looking at it. Main feeds the releases of a project used by third parties.

Same-repository branches (Renovate, maintainers) aren't concerned: only people with write access can push to them.

## Proposal
1. In the loop of `.github/workflows/auto-merge.yml:61-91`, also fetch `isCrossRepository` (and `headRepositoryOwner`) with `gh pr view`, and skip a PR coming from a fork with a `::warning::` saying that fork PRs must be merged by hand. Do it before enabling auto-merge and before updating the branch.
2. Pass `--match-head-commit "${head}"` to `gh pr merge --auto` (`:74`), so auto-merge is only enabled for the commit the workflow just inspected.
3. Update the "Automerge" section of `CLAUDE.md` and the workflow's header comment: the label only applies to same-repository PRs.

## Tests
No test harness for workflows. Check with `actionlint` (or `gh workflow view`) that the file is valid, then:
- On a scratch fork PR labelled `automerge`, trigger the workflow with `workflow_dispatch`: the run logs the fork warning and `gh pr view <n> --json autoMergeRequest` stays `null`. Before the fix, auto-merge gets enabled.
- Positive control: a same-repository PR labelled `automerge` still gets auto-merge enabled and its branch updated when it is behind main.

## Open questions
- Should fork PRs be supported at all, for example by recording the reviewed head SHA when the label is added and refusing any other one? Recommended: no, skip fork PRs; they are rare and can be merged by hand.

## Resolved open questions
- Skip fork PRs; do not support them (autopilot: recommended answer). They are rare and can be merged by hand.
