---
name: NATS-only compose refactor 0.3.1
overview: "Убрать режим `direct` и легаси-документацию: один путь **NATS → ingest-worker → Neo4j**; слить лишние compose-файлы в [docker-compose.yml](docker-compose.yml); код скрейперов без ветвления по `INGEST_MODE`; затем полный прогон, export, пак и релиз **v0.3.1-graph-pack** (новый sha256)."
todos:
  - id: compose-unify
    content: Слить deploy/scrape-nats в docker-compose.yml; удалить neo4j.yml + лишние compose; обновить команды в доках
    status: pending
  - id: code-nats-only
    content: Убрать INGEST_MODE/direct из всех scrapers (cmd, config, usecase); sbom только file/URL CVE; поправить логи proxy «direct»
    status: pending
  - id: docs-diagrams
    content: "README Mermaid + runtime/ingest-contract/coding-style/ingest-worker: только NATS, без legacy"
    status: pending
  - id: e2e-scrape-build
    content: Полный build + scrape up; проверить логи producer/worker
    status: pending
  - id: pack-031-release
    content: export → build v0.3.1; gh release v0.3.1-graph-pack; обновить graph-bootstrap.sh и ссылки на ZIP
    status: pending
isProject: false
---

# NATS-only рефакторинг, один compose, релиз 0.3.1

## Цель

- **Только NATS:** скрейперы не пишут в Bolt; единственный write-path — **ingest-worker**. Удалить ветки **`direct`**, флаги/дефолты **`INGEST_MODE=direct`**, упоминания «legacy / optional direct» из доков и диаграмм.
- **Один основной compose:** свернуть [docker-compose.deploy.yml](docker-compose.deploy.yml), [docker-compose.scrape-nats.yml](docker-compose.scrape-nats.yml), [docker-compose.neo4j.yml](docker-compose.neo4j.yml) в [docker-compose.yml](docker-compose.yml) (профили **`deploy`**, **`scrape`**, **`mcp`** как сейчас + единый **`testpack`** или замена на док — см. ниже).
- **Граф-пак 0.3.1:** после реального наполнения БД — `export` → `build-graph-pack` → новый **sha256** → GitHub release **`v0.3.1-graph-pack`** и обновление URL в [docker/graph-bootstrap.sh](docker/graph-bootstrap.sh) (и ссылок в доках с v0.3.0 на v0.3.1 где это дефолт bootstrap).

## Почему раньше ZIP «одинакового веса»

Уже выяснено: [manifest.v0.2.0.json](data/neo4j_user_export/releases/manifest.v0.2.0.json) и [manifest.v0.3.0.json](data/neo4j_user_export/releases/manifest.v0.3.0.json) содержат **одинаковый `sha256` для `graph.cypher`** — переупаковка без нового [scripts/export-graph-cypher.sh](scripts/export-graph-cypher.sh).

## 1. Compose: один файл + удаление лишних

| Файл | Действие |
|------|----------|
| [docker-compose.scrape-nats.yml](docker-compose.scrape-nats.yml) | **Удалить**; в [docker-compose.yml](docker-compose.yml) для сервисов **`vuln`/`ti`/`lola`/`ds`/`sbom`/`coderules`/`nuclei`**: убрать **`NEO4J_*`**, зафиксировать поведение как NATS-only; **`depends_on`**: только **`nats` healthy** (не ждать Neo4j на продьюсере). **`ingest-worker`** оставить с **`neo4j` + `nats`**. |
| [docker-compose.deploy.yml](docker-compose.deploy.yml) | **Удалить**; перенести сервис **`nginx-lb`** (профиль **`deploy`**) в основной [docker-compose.yml](docker-compose.yml). |
| [docker-compose.neo4j.yml](docker-compose.neo4j.yml) | **Удалить** (дубликат Neo4j без bootstrap/API). |
| [docker-compose.testpack.yml](docker-compose.testpack.yml) | Либо **встроить** профиль **`testpack`** в `docker-compose.yml` через отдельный сервис-«patch» невозможен без дублирования `graph-bootstrap`. Практичный вариант: **один** фрагмент в [docs/threatintel-runtime.md](docs/threatintel-runtime.md) (override `graph-bootstrap.volumes` + `GRAPH_PACK_DEFAULT=0`) и **удалить файл**; либо оставить **один** минимальный `docker-compose.testpack.yml` — выбрать при реализации по удобству CI (сейчас на него есть ссылки только в доках). |

Обновить все команды в [README.md](README.md), [docs/threatintel-runtime.md](docs/threatintel-runtime.md), [docs/deploy.md](docs/deploy.md), [scrapers/README.md](scrapers/README.md) — без `-f docker-compose.scrape-nats.yml` / `-f docker-compose.deploy.yml`.

## 2. Код: убрать `direct`

Затронуть по схеме «один модуль — один путь NATS»:

- [scrapers/sbom](scrapers/sbom): [cmd/main.go](scrapers/sbom/cmd/main.go), [internal/config](scrapers/sbom/internal/config/config.go), [internal/usecase/scrape.go](scrapers/sbom/internal/usecase/scrape.go) — убрать `IngestModeDirect`, Neo4j store из usecase для скрейпа; CVE только **file/URL** ([cvesource](scrapers/sbom/internal/cvesource)).
- Аналогично: [scrapers/ti](scrapers/ti/cmd/main.go), [scrapers/vuln/internal/components/init.go](scrapers/vuln/internal/components/init.go), [scrapers/lola](scrapers/lola/internal/components/init.go), [scrapers/ds](scrapers/ds/cmd/main.go), [scrapers/coderules](scrapers/coderules/cmd/main.go), [scrapers/nuclei](scrapers/nuclei/cmd/main.go) + `internal/config` / usecase — дефолт и единственный режим NATS; удалить инициализацию Neo4j writer в процессе скрейпера.
- [pkg/ingestv1](pkg/ingestv1) / [docs/ingest-contract.md](docs/ingest-contract.md) / [docs/coding-style.md](docs/coding-style.md) / [scrapers/ingest-worker/README.md](scrapers/ingest-worker/README.md): убрать формулировки «как direct» / «direct fallback».
- [scrapers/ingest-worker/cmd/main.go](scrapers/ingest-worker/cmd/main.go): без изменения семантики MERGE (по желанию переименовать комментарии).

**Внимание:** логи вида «running **direct**» в [scrapers/ti/internal/feeds/runner.go](scrapers/ti/internal/feeds/runner.go) / [scrapers/ds](scrapers/ds/internal/usecase/ingest.go) — это про **HTTP proxy**, не ingest; переименовать в логах на «without proxy», чтобы не путать с legacy ingest.

## 3. Документация и диаграммы

- [README.md](README.md) Mermaid: только **NATS → worker → Neo4j** для скрейперов; убрать стрелки `direct: MERGE`.
- [docs/threatintel-runtime.md](docs/threatintel-runtime.md): таблица **`INGEST_MODE`** удалена или заменена одной строкой «режим фиксирован NATS»; убрать секцию про optional scrape-nats override.
- [CONTRIBUTING.md](CONTRIBUTING.md), [AGENTS.md](AGENTS.md): не описывать `direct`.

## 4. Валидация и релиз 0.3.1

1. **`docker compose --profile scrape build`** (сеть/`GOPROXY` при необходимости).
2. **`docker compose --profile scrape up -d`** (при чистом графе — `down -v` по согласованию); поднять лимиты env для «максимума» осознанно.
3. Логи: продьюсеры публикуют; **`ingest-worker`** без постоянных NAK.
4. **`./scripts/export-graph-cypher.sh`** → **`GRAPH_PACK_VERSION=v0.3.1 ./scripts/build-graph-pack.sh`** (или `EXPORT_FIRST=1`); убедиться, что **sha256 ≠** прежний `b4fd360a…`.
5. **`curl`** API `/health` + несколько REST; **`docker compose --profile mcp run --rm -i mcp`**; при **`--profile deploy`** — порт LB.
6. **`gh release create v0.3.1-graph-pack`** с новым ZIP; обновить **`DEFAULT_PACK_URL`** в [docker/graph-bootstrap.sh](docker/graph-bootstrap.sh) и пути в доках/testpack-фрагменте на **`v0.3.1`**.

## Риски

- Удаление **`direct`** ломает локальный «быстрый» запуск без NATS — это **намеренно** по ТЗ.
- Скрейперы без **`depends_on: neo4j`** могут стартовать раньше БД; очередь JetStream это переживёт, worker стартует после Neo4j.
