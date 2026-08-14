# Contributing

`ytunes` is primarily a personal Go learning project — the pace of
review and how much scope creep gets accepted reflects that. Bug
reports and small, focused fixes are very welcome; large feature PRs
are more likely to turn into a discussion first than a quick merge.

## Before you start

- **Bug fixes / small improvements:** open a PR directly.
- **New features:** open an issue first to talk through the idea.
  `PLAN.md`'s "Stop Here" section lists what's deliberately out of
  scope for now (playlists, queues, search, media keys, config files)
  — worth checking there before proposing something in that space.

## Building and testing locally

Requires [mpv](https://mpv.io) and [yt-dlp](https://github.com/yt-dlp/yt-dlp)
on your `PATH` (see the README's Installation section).

```sh
go build ./...
go vet ./...
gofmt -l .   # should print nothing
```

There's no automated test suite yet — changes are verified by running
the app and exercising the affected screens/states directly.

## Code style

- Run `gofmt` before committing; there's no separate linter config.
- Keep changes scoped to what they're fixing — avoid drive-by
  refactors mixed into an unrelated fix, they make the diff harder to
  review.
- This repo's `PLAN.md`/`LEARNING.md`/`PROGRESS.md` document the
  project's own design decisions and reasoning as it was built — worth
  skimming the relevant phase before touching an area you're unfamiliar
  with, since a lot of "why is it built this way" is already answered
  there.

## Submitting a PR

- Describe what changed and why, not just what.
- Reference the issue it addresses, if any.
- Keep it small where possible — easier to review, easier to revert if
  something's wrong.
- Add a one-liner to `CHANGELOG.md` under the current version, tagged
  with your PR number (e.g. `- Fix theme picker crash on empty list
  (#42)`) — this is how the changelog stays a trustworthy record of
  what shipped and where to find the discussion behind it.
