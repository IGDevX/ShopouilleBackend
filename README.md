<|start_of_focus|># Shopouille Backend

This monorepo hosts the Shopouille backend services, shared libraries, and infrastructure tooling.

## Release Tagging Strategy

All release deployments are orchestrated through Git tags that follow [Semantic Versioning](https://semver.org/) with a `v` prefix.

### Tag patterns

- `vMAJOR.MINOR.PATCH` → immutable production releases (for example `v1.4.0`)
- `vMAJOR.MINOR.PATCH-rc.NUMBER` → immutable staging release candidates (for example `v1.4.0-rc.3`)
- `vMAJOR.MINOR.PATCH-{alpha,beta}.NUMBER` → optional internal validation tags (for example `v1.4.0-alpha.1`)
- Every release-related tag must be created from the `main` branch and is protected from deletion or force-push through tag protection rules.

### Release flow (alpha/beta → rc → final)

1. (Optional) Cut `alpha`/`beta` tags from commits on `main` to share early builds with reviewers or testers.
2. Once the change set is feature-complete, cut an `rc` tag from the current `main` commit. Each `rc` is a staging promotion and **must never be moved**.
3. After the latest RC has been validated in staging, create the final `vMAJOR.MINOR.PATCH` tag **on the exact same commit**. No rebuild occurs between RC and final.

CI reacts to these tags as follows:

- Tags that match `v*-rc.*` trigger staging deployment workflows.
- Tags that match `v[0-9]+.[0-9]+.[0-9]+` trigger production deployment workflows.

### Promotion workflow (RC → final)

The reusable workflows under `.github/workflows` consume the tags described above:

- RC tags deploy to staging and run release-candidate validations.
- Once an RC is approved, the final tag reuses the same commit SHA, so the production release is a straight promotion with no rebuild.

### Valid tag examples

- Production: `v2.3.5`
- Release Candidate: `v2.3.5-rc.2`
- Alpha/Beta (optional): `v2.3.5-alpha.1`, `v2.3.5-beta.4`

### Creating tags

Cut RC and final tags locally, then push them to origin:

```bash
# Ensure the commit you want is the tip of main
git checkout main
git pull origin main

# Create and push an RC tag for staging
git tag v2.3.5-rc.1
git push origin v2.3.5-rc.1

# After staging validation, reuse the same commit for the final tag
git tag v2.3.5
git push origin v2.3.5
```

Never retag or delete RC/final tags. If a new build is required, bump the patch version and restart the RC cycle.
