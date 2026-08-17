---
title: "Contributing Guide"
weight: 70
---

# Contributing to Enterprise Task Engine

We want to make contributing to this project as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

---

## Contributor License Agreement (CLA)

Contributions to this project must be accompanied by a **Contributor License Agreement (CLA)**. You (or your employer) retain the copyright to your contribution; this simply gives us permission to use and redistribute your contributions as part of the project.

- Head over to **[cla.developers.google.com](https://cla.developers.google.com/)** to see your current agreements on file or to sign a new one.
- You generally only need to submit a CLA once, so if you've already submitted one (even if it was for a different Google open source project), you probably don't need to do it again.

---

## Code of Conduct

This project has adopted the **[Google Open Source Community Guidelines](https://opensource.google/conduct/)**. By participating in this project and its communications, you agree to abide by its terms.

If you observe unacceptable behavior, please report it to the project owners listed in **[owners.txt](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/owners.txt)** or via official Google Open Source channels.

---

## Contribution Workflow

### 1. Source Code Licensing Header
Every new or modified source file must include the standard Apache License Version 2.0 header at the top of the file:

```text
Copyright 2026 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
*(Use `//` for Go/TS/JS/Kotlin/Java, `#` for Python/Shell/Bazel/Terraform, `--` for SQL, and `<!-- -->` for HTML).*

### 2. Code Style & Engineering Standards
All code must adhere to **[Google Style Guides](https://google.github.io/styleguide/)**:
- **Go (`/cmd`, `/pkg`)**: Follow Go idiomatic practices, static linting via `golangci-lint`, and format with `gofmt` / `gofumpt`.
- **Python (`/cmd/task_agent`)**: Managed via `uv`, linted and formatted with `Ruff`, requiring strict PEP 484 type annotations on all signatures.
- **TypeScript & React (`/web`)**: Modern React functional components with arrow syntax, Material 3 baseline, and strict TypeScript compilation.
- **Kotlin & Android (`/apps/gtasks`)**: M3 Jetpack Compose UI patterns, repository/domain layer isolation, and Gradle Kotlin DSL formatting.

### 3. Testing Requirements
- Ensure all automated unit and integration tests pass before submitting a pull request:
  ```bash
  bazel test //...
  ```
- Any new business logic, agent tools, or A2UI component rendering must be covered by comprehensive unit tests or end-to-end verification scripts (`scripts/e2e_test.py`).

### 4. Pull Request Review & Approval
1. Fork or branch from `main` using descriptive branch names (`feature/...`, `fix/...`, `docs/...`).
2. Write concise, imperative-mood commit messages explaining the problem and solution.
3. Submit your pull request for review by project maintainers defined in **[owners.txt](file:///Users/rmcguinness/Projects/internal/gemini_task_engine/owners.txt)**.
4. Once CLA checks, Bazel CI builds, and owner code reviews pass, your changes will be merged.
