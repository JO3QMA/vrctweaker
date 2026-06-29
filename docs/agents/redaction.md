# Redaction for public contribution artifacts

Agents and humans must not put **identifying VRChat user information** into text that is stored in Git or published on GitHub.

Full glossary term: **Redacted reproduction** in root `CONTEXT.md`.

## Scope

Applies when writing or editing:

| Artifact | Examples |
|----------|----------|
| Pull requests | Title, body, review summaries drafted by agents |
| Issues | Title, body, comments |
| Commits | Message, branch name |
| Agent outputs | PR drafts, issue drafts, triage notes under `docs/ai_dlc/`, chat summaries intended for paste into GitHub |

**Out of scope (this policy):** unit/E2E test fixtures and committed test log snippets. Prefer synthetic IDs for **new** tests (`usr_e2e_*`, `usr_a1111111-...`). Existing tests may still contain real IDs from log captures; cleaning them is a separate task.

## Forbidden in scope

Do **not** include:

1. VRChat **display names** (e.g. a friend's in-game name)
2. VRChat **`usr_*` user IDs**
3. **Profile URLs** (`https://vrchat.com/home/user/usr_...`)
4. VRChat **login usernames** (account name, not display name)
5. **Instance strings** that embed `usr_*` (e.g. `~hidden(usr_...)~region(jp)`)

Also avoid quoting private messages, bio text, or location strings that identify a specific person.

## Allowed substitutes

Use **Redacted reproduction** instead:

- **Counts and deltas** — e.g. official friend count 557 vs cached 483
- **Status categories** — offline, active, ask me
- **Role phrases** — "an offline friend", "two missing offline friends", "the reporter's account"
- **Synthetic IDs** in examples only when necessary for code discussion — `usr_e2e_test`, `usr_a1111111-1111-4111-8111-111111111101`
- **Technical symptoms** — "RefreshFriends leaves DB short by N friends on the offline pass"

## Agent checklist

Before `git commit`, `gh pr create`, or `gh issue create`:

- [ ] No real display names in title/body/message
- [ ] No `usr_` UUIDs (except documented synthetic test prefixes in code, not in PR/Issue text)
- [ ] No `vrchat.com/home/user/` links
- [ ] No login usernames
- [ ] No full instance strings with embedded user IDs
- [ ] Verification steps use counts or roles, not named users

If the user pastes identifying data in chat, **do not copy it** into PR/Issue/commit text. Refer to the symptom abstractly.

## Examples

### Bad (PR context)

```markdown
UserA (usr_aaaaaaaa-...) and UserB were missing after refresh.
https://vrchat.com/home/user/usr_aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa
```

### Good

```markdown
Several offline friends were missing from the Tweaker list while visible in the
official VRChat client. After cache clear and refresh, cached friend count matched
the official total (557).
```

### Bad (commit message)

```text
fix: sync UserA friend row
```

### Good

```text
fix(friends): fetch all offline friends when API returns short pages
```

## Related

- `.cursor/rules/redaction-public-artifacts.mdc` — always-on agent reminder
- `.cursor/commands/make-pr.md` — PR workflow includes this check
- `docs/agents/domain.md` — domain glossary usage
