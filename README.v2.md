# Spot 🚀

Spot is a next-generation, ultra-fast static code analysis platform that allows engineering teams to define, automate,
and enforce their own architectural and quality standards using a single, clear configuration file. 

Unlike traditional code quality tools that restrict you to a rigid set of pre-baked rules, Spot puts you completely in
control of your codebase's governance.

---

## What is Spot?

Spot is a lightweight, universal gatekeeper for your source code. It reads your codebase, understands its structure, and
verifies that your team’s internal design patterns, safety rules, and architectural guidelines are being followed
perfectly on every single code commit.

Think of Spot as a tailor-made automated code reviewer that knows exactly how *your* company wants to build software,
running continuously inside your development workflow.

---

## What Can It Do?

* **Enforce architectural cleanliness:**: Automatically ensure that structural rules—like making sure configuration blocks or critical declarations are always encapsulated inside the correct namespaces—are followed perfectly across thousands of files.
* **Catch team-specific technical debt:** Ban specific anti-patterns, legacy variable conventions, or dangerous structural layouts unique to your organization before they ever make it to a code review.
* **Scale across languages seamlessly:** Instead of running ten different tools to manage a modern multi-language system, Spot uses a single, unified blueprint language to analyze your entire development footprint.
* **Run universally:** Spot integrates natively into local developer terminals, pre-commit hooks, and any CI/CD automation pipeline (GitHub Actions, GitLab CI, etc.), acting as a seamless quality gate for your pull requests.

---

## Why choose Spot over traditional tools (Like SonarQube)?

If you have ever managed code quality at scale, you know that existing enterprise platforms quickly become a bottleneck.
Spot was built from scratch to fix the three fundamental flaws of traditional static analysis:

### 1. Custom rules take minutes, not weeks
With traditional platforms, writing a custom rule to enforce a company-specific design pattern requires setting up a
massive separate development environment, writing heavy object-oriented plugin code, compiling binaries, and restarting
servers. If you use cloud-managed versions, custom rules are often completely blocked.

* **The Spot difference:**: Spot introduces **Rules-as-code**. You can write a powerful structural compliance rule in 6 lines of plain, readable text, check it straight into your Git repository next to your application code, and run it instantly.

### 2. Immediate feedback (zero startup overhead)
Traditional local code scanners are heavy. They force developers to wait 10 to 15 seconds just for the analysis software
to boot up and initialize before it even begins reading the code, completely breaking a developer's focus.

* **The Spot difference:**: Spot is engineered for hyper-performance. It boots up, maps your custom rules, and executes full-scale analyses in **under 10 milliseconds**. It provides immediate feedback as fast as you can save a file.

### 3. Lightweight and Maintenance-Free
Managing enterprise code quality often means deploying and maintaining heavy, resource-hungry servers that require
constant updates, complex database tuning, and dedicated infrastructure teams.

* **The Spot difference:**: Spot is a single, zero-dependency executable binary that runs entirely locally on your machine or inside your existing pipeline runners. It requires zero server infrastructure, zero databases, and zero ambient background maintenance.

---

## How it works in practice

Spot simplifies quality governance into a single declarative configuration file checked right into your project. 

```text
scope {
    include "**/*.cs"
}

rules {
    rule EnforceUsingDirectivesInsideNamespace {
        match UsingDirective
        where UsingDirective.is_at_file_root == true
        report error "Architectural Violation: Using statements must live inside a namespace."
    }
}

The moment a developer opens a pull request, Spot scans the changes. If a rule is violated, Spot points out the exact
file boundary and prevents the technical debt from bleeding into your production branches.
