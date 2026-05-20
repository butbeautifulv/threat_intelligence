---
name: engage tool source map
overview: "Срочный план: построить проверенную карту источников установки для недостающих security тулов (Kali/Debian/upstream), внедрить быстрый fallback installer и закрыть field-phase проверкой preflight + red-vs-blue harness малыми PR через субагентов."
todos:
  - id: source-registry
    content: Собрать и зафиксировать registry источников для missing tools (Kali/pkg tracker/Salsa/upstream).
    status: completed
  - id: fallback-installer
    content: Добавить fallback install режим для недостающих тулов (go/cargo + provenance output).
    status: completed
  - id: preflight-hints
    content: Расширить preflight JSON actionable hints + export missing tool list.
    status: completed
  - id: field-validate
    content: Прогнать install+preflight+harness и задокументировать дефекты/ограничения.
    status: completed
isProject: false
---

# Engage Fast Tool Sourcing Plan

## Цель

Сделать установку недостающих тулов (`httpx`, `nuclei`, `subfinder`, `amass`, `feroxbuster`) практичной и воспроизводимой: не только отметить `MISS`, а дать автоматизированный путь установки из проверенных источников (Kali package pages, Debian/Salsa packaging, upstream source/releases), с проверкой через preflight и lab harness.

## Подтверждённые источники (исследовано)

- Kali tool pages:
  - `httpx-toolkit`: [https://kali.org/tools/httpx-toolkit](https://kali.org/tools/httpx-toolkit)
  - `nuclei`: [https://kali.org/tools/nuclei](https://kali.org/tools/nuclei)
  - `subfinder`: [https://kali.org/tools/subfinder](https://kali.org/tools/subfinder)
  - `amass`: [https://kali.org/tools/amass](https://kali.org/tools/amass)
  - `feroxbuster`: [https://kali.org/tools/feroxbuster](https://kali.org/tools/feroxbuster)
- Kali package tracking / packaging repos (пример паттерна):
  - `httpx-toolkit`: [https://pkg.kali.org/pkg/httpx-toolkit/](https://pkg.kali.org/pkg/httpx-toolkit/), [https://gitlab.com/kalilinux/packages/httpx-toolkit](https://gitlab.com/kalilinux/packages/httpx-toolkit)
  - Аналогично для `nuclei/subfinder/amass/feroxbuster` через `pkg.kali.org/pkg/<name>` + `gitlab.com/kalilinux/packages/<name>`
- Debian packaging provenance pattern (как ты просил на примере):
  - `hashcat`: [https://salsa.debian.org/pkg-security-team/hashcat](https://salsa.debian.org/pkg-security-team/hashcat)
- Upstream source repos (для fallback install):
  - `httpx`: [https://github.com/projectdiscovery/httpx](https://github.com/projectdiscovery/httpx)
  - `nuclei`: [https://github.com/projectdiscovery/nuclei](https://github.com/projectdiscovery/nuclei)
  - `subfinder`: [https://github.com/projectdiscovery/subfinder](https://github.com/projectdiscovery/subfinder)
  - `amass`: [https://github.com/owasp-amass/amass](https://github.com/owasp-amass/amass)
  - `feroxbuster`: [https://github.com/epi052/feroxbuster](https://github.com/epi052/feroxbuster)

## Что уже есть в репо (база для доработки)

- Installer + package map:
  - [`/home/bbv/Desktop/threat_intelligence/scripts/ops/install-engage-host-tools.sh`](/home/bbv/Desktop/threat_intelligence/scripts/ops/install-engage-host-tools.sh)
  - [`/home/bbv/Desktop/threat_intelligence/scripts/ops/engage-tools-packages.yaml`](/home/bbv/Desktop/threat_intelligence/scripts/ops/engage-tools-packages.yaml)
- Проверка наличия тулов:
  - [`/home/bbv/Desktop/threat_intelligence/scripts/engage/preflight-client-tools.sh`](/home/bbv/Desktop/threat_intelligence/scripts/engage/preflight-client-tools.sh)
- Lab harness:
  - [`/home/bbv/Desktop/threat_intelligence/scripts/test/smoke-engage-red-vs-blue.sh`](/home/bbv/Desktop/threat_intelligence/scripts/test/smoke-engage-red-vs-blue.sh)
- Плановый каркас:
  - [`/home/bbv/Desktop/threat_intelligence/.cursor/plans/engage_userfriendly_install_73f6d9c0.plan.md`](/home/bbv/Desktop/threat_intelligence/.cursor/plans/engage_userfriendly_install_73f6d9c0.plan.md)

## Исполнимый merge order (мелкие PR)

1. **PR-A: Source Registry (данные + docs only)**
   - Добавить machine-readable registry источников для missing tools:
     - новый файл [`/home/bbv/Desktop/threat_intelligence/scripts/ops/engage-tools-sources.yaml`](/home/bbv/Desktop/threat_intelligence/scripts/ops/engage-tools-sources.yaml)
   - Обновить runbook:
     - [`/home/bbv/Desktop/threat_intelligence/docs/engage/engage-install-linux.md`](/home/bbv/Desktop/threat_intelligence/docs/engage/engage-install-linux.md)
   - Содержание registry по каждому tool: `kali_tool_page`, `kali_pkg_tracker`, `kali_packaging_repo`, `debian_or_salsa_if_exists`, `upstream_repo`, `preferred_install_methods`.

2. **PR-B: Fallback Installer (код, один скрипт + make target)**
   - Добавить `--fallback` режим в [`/home/bbv/Desktop/threat_intelligence/scripts/ops/install-engage-host-tools.sh`](/home/bbv/Desktop/threat_intelligence/scripts/ops/install-engage-host-tools.sh) или отдельный скрипт `install-engage-host-tools-fallback.sh`.
   - Логика:
     - сначала distro package manager;
     - для недостающих: fallback по registry (`go install` для Go-based, `cargo install` для Rust-based);
     - печатать provenance (откуда поставлено).
   - Добавить `make engage-install-fallback` в [`/home/bbv/Desktop/threat_intelligence/Makefile`](/home/bbv/Desktop/threat_intelligence/Makefile).

3. **PR-C: Preflight -> actionable remediation**
   - Расширить [`/home/bbv/Desktop/threat_intelligence/scripts/engage/preflight-client-tools.sh`](/home/bbv/Desktop/threat_intelligence/scripts/engage/preflight-client-tools.sh):
     - `--json` включает `install_hint`/`source_hint` из registry;
     - новый режим `--emit-missing` (newline list для пайпа в installer).

4. **PR-D: Field validation + bug capture**
   - Прогон на хосте:
     - `make engage-install-plan`
     - `make engage-install-fallback`
     - `./scripts/engage/preflight-client-tools.sh --profile recommended --json`
     - `make test-engage-red-blue`
   - Результаты и дефекты:
     - [`/home/bbv/Desktop/threat_intelligence/docs/engage/engage-red-blue-bugs.md`](/home/bbv/Desktop/threat_intelligence/docs/engage/engage-red-blue-bugs.md)

5. **PR-E+: micro-fixes per defect**
   - Один баг = один PR (`engage/fix-pXX-*`) до достижения DoD.

## Субагентное исполнение (параллельные потоки)

- **Agent-SourceMap** (`engage/install-p09-source-registry`)
  - Заполняет `engage-tools-sources.yaml` по подтверждённым URL.
- **Agent-Installer** (`engage/install-p10-fallback-installer`)
  - Реализует fallback install path и `make engage-install-fallback`.
- **Agent-Preflight** (`engage/install-p11-preflight-remediation`)
  - Добавляет actionable hints + missing-tool export.
- **Agent-Field** (`engage/lab-p12-field-validation`)
  - Гоняет preflight + harness, фиксирует баги в markdown лог.
- **Critic (основной чат)**
  - Проверяет diff size, provenance прозрачность, merge order, и принимает только green slices.

## DoD (не «на бумаге»)

- Для каждого из 5 missing тулов есть запись в `engage-tools-sources.yaml` с проверяемыми URL.
- `make engage-install-fallback` не падает целиком на частично недоступных пакетах и выводит, что установлено, что осталось manual.
- Preflight JSON показывает осмысленные install hints для каждого missing tool.
- `make test-engage-red-blue` проходит после установки fallback-путём.
- Баги/ограничения задокументированы в [`/home/bbv/Desktop/threat_intelligence/docs/engage/engage-red-blue-bugs.md`](/home/bbv/Desktop/threat_intelligence/docs/engage/engage-red-blue-bugs.md).

## Ограничения и риски

- Стандартные репозитории не гарантируют наличие всех security тулов на конкретной Ubuntu/Debian ветке.
- Kali repo как глобальный fallback может конфликтовать с base Debian/Ubuntu pinning; использовать только как явный opt-in путь в документации, не как дефолтный авто-режим.
- `go install`/`cargo install` требуют toolchain и могут отличаться по версии от distro packages; это нужно явно логировать в install output.