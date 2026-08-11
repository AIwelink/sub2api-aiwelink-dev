# AIwelink API GitHub Homepage Design

## Goal

Rebrand the repository's GitHub homepage as **AIwelink API**, promote the official service at <https://api.aiwelink.cc>, and remove third-party promotional content without making the technical documentation inaccurate.

## Scope

- Update `README.md`, `README_CN.md`, and `README_JA.md` consistently.
- Replace the current Sub2API header and logo with the exact product name `AIwelink API` and the official AIwelink icon already used by the website project.
- Add a prominent, localized link to <https://api.aiwelink.cc> and a concise description of the multi-model AI API gateway and management service.
- Remove the sponsor section, Trendshift badge, ecosystem section, and Star History section from all three README files.
- Preserve functional documentation, deployment commands, configuration keys, image names, source paths, license information, and necessary technical links.

## Naming Boundary

User-facing homepage branding must use the exact spelling and capitalization `AIwelink API`. Existing technical identifiers such as `sub2api`, repository URLs, executable names, Docker images, directory names, and configuration values remain unchanged wherever they describe real interfaces or commands.

## Environment Example Audit

Audit every environment-like file without exposing its values. Confirm whether each file is tracked or ignored, scan example files for known secret formats and sensitive-looking non-placeholder values, and report findings separately from the README changes. Do not modify real `.env` files. Do not change safe example defaults unless the audit establishes that a value is sensitive.

## Verification

- Confirm all three README files show `AIwelink API` and link to `https://api.aiwelink.cc`.
- Confirm the removed promotional sections and their links no longer appear.
- Confirm technical installation and deployment content remains present.
- Validate local Markdown links and the copied brand image path.
- Repeat the sanitized environment-example audit and verify real environment files remain untracked and ignored.
