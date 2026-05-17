---
name: graph layer structure
overview: Привести слой knowledge/ к кодстайлу (cmd → internal → connector), разделив write-path (NATS ingest) и read-path (api+mcp) на два Go-модуля внутри knowledge/, с общим connector для Neo4j. Каталог knowledge/ не переименовываем; repo pkg/* — только wire-типы.
todos:
  - id: doc-graph-semantics
    content: "docs/coding-style.md: ingest pack + serve pack + connector, dependency diagram"
    status: completed
  - id: doc-graph-readme
    content: "graph/README.md: новое дерево ingest/, serve/, connector/"
    status: completed
  - id: doc-runtime-paths
    content: "threatintel-runtime.md: пути бинарей без смены имён compose-сервисов"
    status: completed
  - id: connector-init
    content: Создать knowledge/connector/go.mod, перенести neo4jclient/neo4j и query
    status: completed
  - id: connector-imports-sources
    content: Переключить knowledge/sources/*/storage на knowledge/connector
    status: completed
  - id: connector-imports-appsec
    content: Переключить knowledge/storage/* на knowledge/connector
    status: completed
  - id: connector-imports-api-mcp
    content: Переключить api storage и mcp neo4jconn на connector
    status: completed
  - id: connector-rm-neo4jclient
    content: Удалить knowledge/neo4jclient/
    status: completed
  - id: ingest-mod-init
    content: knowledge/ingest/go.mod + cmd/ingest_worker скелет
    status: completed
  - id: ingest-move-nats-ingestkit
    content: Перенести natsensure и ingestkit в knowledge/ingest/internal/
    status: completed
  - id: ingest-move-source-ti
    content: knowledge/sources/ti → knowledge/ingest/internal/sources/ti
    status: completed
  - id: ingest-move-source-vuln
    content: Перенести vuln source
    status: completed
  - id: ingest-move-source-lola
    content: Перенести lola source
    status: completed
  - id: ingest-move-source-ds
    content: Перенести ds source
    status: completed
  - id: ingest-move-appsec
    content: knowledge/storage/* → knowledge/ingest/internal/appsec/*
    status: completed
  - id: ingest-extract-components
    content: DI из main → internal/components
    status: completed
  - id: ingest-extract-loop
    content: Pull loop → internal/ingest/consumer.go
    status: completed
  - id: ingest-cmd-thin
    content: Тонкий cmd/ingest_worker/main.go
    status: completed
  - id: ingest-rm-legacy-dirs
    content: Удалить ingest_worker/, sources/, storage/, internal/ на верхнем уровне
    status: completed
  - id: serve-mod-init
    content: knowledge/serve/go.mod
    status: completed
  - id: serve-move-api
    content: knowledge/api → knowledge/serve (cmd + internal)
    status: completed
  - id: serve-move-mcp
    content: knowledge/mcp → knowledge/serve/cmd/mcp + transport
    status: completed
  - id: serve-dry-neo4j
    content: Убрать дубли Bolt-обёрток, единый connector
    status: completed
  - id: serve-rm-legacy
    content: Удалить knowledge/api и knowledge/mcp
    status: completed
  - id: graph-work-deploy
    content: knowledge/go.work (3 модуля), Makefile, Dockerfiles
    status: completed
  - id: graph-verify
    content: make test-graph + checklist coding-style
    status: completed
isProject: false
---

# Graph layer — структура, именование, два модуля

## Как называть слой (ответ на «graph или interaction?»)

В Veil три **runtime-слоя** по потоку данных:

```mermaid
flowchart LR
  scrape[scrape acquisition] -->|NATS scrape.>| pipeline[pipeline transform]
  pipeline -->|NATS ingest.>| graphWrite[graph ingest pack]
  graphWrite --> neo4j[(Neo4j)]
  graphRead[graph serve pack] --> neo4j
  clients[HTTP / MCP clients] --> graphRead
```

| Название | Что это | Не путать с |
|----------|---------|-------------|
| **graph/** (оставляем) | Зона Neo4j: запись из `ingest.>` + чтение через API/MCP | «Graph DB» как продукт |
| **ingest pack** | Inbound-адаптер: JetStream consumer → MERGE | Слоем pipeline (тот только публикует `ingestv1`) |
| **serve pack** | Inbound-адаптер: HTTP/MCP → read-only Cypher | Ingest worker |

**Межслойное взаимодействие** — только NATS + [`pkg/ingestv1`](pkg/ingestv1). Код `knowledge/` не импортирует `scrape/` или `pipeline/`.

`ingest_worker` — это **порт/адаптер** (hexagonal: driving adapter с bus), не отдельный «бизнес-слой». Оркестрация MERGE — `internal/usecase`; Cypher — `internal/.../storage`.

---

## Go layering (best practice + ваш кодстайл)

Официальная модель ([go.dev/doc/modules/layout](https://go.dev/doc/modules/layout)): **`cmd/`** — `main`, **`internal/`** — приватно модулю, **`pkg/`** (repo root) — общее между слоями.

**Направление зависимостей (стрелки = «импортирует»):**

```mermaid
flowchart TB
  cmd[cmd]
  components[internal/components]
  transport[internal/transport]
  ingestLoop[internal/ingest]
  usecase[internal/usecase]
  domain[internal/domain]
  storage[internal/storage or sources storage]
  connector[internal/connector]
  repopkg[repo pkg ingestv1 natsjet tidomain]

  cmd --> components
  components --> transport
  components --> ingestLoop
  components --> usecase
  ingestLoop --> usecase
  transport --> usecase
  usecase --> domain
  usecase --> storage
  storage --> connector
  ingestLoop --> repopkg
  usecase --> repopkg
```

Правила (как в [`docs/coding-style.md`](docs/coding-style.md)):

- `domain` — без Bolt, NATS, HTTP
- `usecase` — без subject strings / Cypher
- `cmd` — только wiring: env, `components.Init`, signal, `errgroup`
- `connector` — тонкая обёртка Bolt / JetStream ([`pkg/natsjet`](pkg/natsjet) для stream ensure)

**Не** `internal → cmd`. **Не** `domain → storage`.

---

## Целевая структура (2 модуля + shared connector)

Сейчас ~15 модулей в [`knowledge/go.work`](knowledge/go.work). Цель — **3 модуля**:

```
graph/
  go.work                          # use: ./ingest ./serve ./connector
  connector/                       # shared Neo4j client + query (ex neo4jclient)
    go.mod
    neo4j/
    query/
  ingest/                          # write path (NATS → MERGE)
    go.mod
    cmd/ingest_worker/main.go
    internal/
      config/
      components/                  # DI: stores + domain appliers
      connector/nats/              # ex internal/natsensure
      ingest/                    # pull loop, route by source (ex cmd/main bulk)
      ingestkit/                 # ex internal/ingestkit
      sources/
        ti/   {domain,usecase,storage,envelope}
        vuln/ lola/ ds/
      appsec/
        sbom/ coderules/ nuclei/   # ex knowledge/storage/*
  serve/                           # read path (api + mcp)
    go.mod
    cmd/api/main.go
    cmd/mcp/main.go
    internal/
      config/
      components/
      connector/neo4j/             # uses knowledge/connector
      usecase/
      domain/
      transport/httpserver/        # ex api/internal/transport
      transport/mcpserver/         # ex mcp/internal/transport
```

[`knowledge/api`](knowledge/api/) уже близок к эталону (`cmd` + `internal/components` + `usecase` + `transport`) — переносим в `serve/` с сохранением пакетов.

[`knowledge/ingest_worker/cmd/main.go`](knowledge/ingest_worker/cmd/main.go) (~345 строк) — разбить: `cmd` тонкий, логика в `internal/ingest` + `internal/components`.

---

## Что не делаем в этом рефакторинге

- Переименование `knowledge/` → другое (вы выбрали **keep graph**)
- AppSec symmetry (`sources/` vs `storage/`) — только переезд в `ingest/internal/appsec/`
- Объединение ingest + serve в **один** `go.mod` (вы разделили: serve = api+mcp, ingest = worker)
- Изменения `deploy/scrape`, `pipeline/`, repo `pkg/*` (кроме import path при необходимости)

---

## Стратегия малого diff

- Один todo = один коммит: перенос каталога + `go mod` + правка import path + `go build` одного бинарника
- Сначала создать новые пути с **re-export / type alias**, потом переключить импорты, потом удалить старые каталоги
- После каждого шага: `go build` затронутого `cmd`

---

## Фаза 0 — документация и семантика

| ID | Действие |
|----|----------|
| `doc-graph-semantics` | В [`docs/coding-style.md`](docs/coding-style.md): секция Graph = ingest pack + serve pack + connector; диаграмма зависимостей |
| `doc-graph-readme` | Обновить [`knowledge/README.md`](knowledge/README.md): дерево `ingest/`, `serve/`, `connector/` |
| `doc-runtime-paths` | Проверить [`docs/threatintel-runtime.md`](docs/threatintel-runtime.md) — пути бинарей без изменения имён сервисов compose |

---

## Фаза 1 — `knowledge/connector` (ex neo4jclient)

| ID | Действие |
|----|----------|
| `connector-init` | Создать [`knowledge/connector/go.mod`](knowledge/connector/go.mod); перенести [`knowledge/neo4jclient/neo4j`](knowledge/neo4jclient/neo4j), [`knowledge/neo4jclient/query`](knowledge/neo4jclient/query) |
| `connector-imports-ingest` | Временные re-export в `knowledge/neo4jclient` → forward import (опционально, 1 коммит) |
| `connector-imports-sources` | Переключить `knowledge/sources/*/storage` на `knowledge/connector/neo4j` |
| `connector-imports-appsec` | Переключить `knowledge/storage/*` |
| `connector-imports-api` | Переключить `knowledge/api/internal/storage` |
| `connector-imports-mcp` | Переключить `knowledge/mcp/internal/connector/neo4jconn` → общий connector или thin wrapper |
| `connector-rm-neo4jclient` | Удалить [`knowledge/neo4jclient/`](knowledge/neo4jclient/) |
| `connector-test` | `go test ./...` в connector |

---

## Фаза 2 — `knowledge/ingest` модуль (write path)

| ID | Действие |
|----|----------|
| `ingest-mod-init` | `knowledge/ingest/go.mod`; `cmd/ingest_worker/` скелет |
| `ingest-move-natsensure` | `knowledge/internal/natsensure` → `knowledge/ingest/internal/connector/nats` |
| `ingest-move-ingestkit` | `knowledge/internal/ingestkit` → `knowledge/ingest/internal/ingestkit` |
| `ingest-move-source-ti` | `knowledge/sources/ti` → `knowledge/ingest/internal/sources/ti` (domain, usecase, storage, ingest→`envelope.go`) |
| `ingest-move-source-vuln` | vuln |
| `ingest-move-source-lola` | lola |
| `ingest-move-source-ds` | ds |
| `ingest-move-appsec-sbom` | `knowledge/storage/sbom` → `knowledge/ingest/internal/appsec/sbom` |
| `ingest-move-appsec-coderules` | coderules |
| `ingest-move-appsec-nuclei` | nuclei |
| `ingest-extract-components` | Вынести DI из main в `internal/components` (stores, appliers, close) |
| `ingest-extract-loop` | Pull loop + routing → `internal/ingest/consumer.go` |
| `ingest-cmd-thin` | `cmd/ingest_worker/main.go` только signal + `components.Run` |
| `ingest-rm-old-worker` | Удалить `knowledge/ingest_worker/` |
| `ingest-rm-old-sources` | Удалить `knowledge/sources/` |
| `ingest-rm-old-storage` | Удалить `knowledge/storage/` |
| `ingest-build` | `go build ./knowledge/ingest/cmd/ingest_worker` |

---

## Фаза 3 — `knowledge/serve` модуль (read path)

| ID | Действие |
|----|----------|
| `serve-mod-init` | `knowledge/serve/go.mod` |
| `serve-move-api-cmd` | `knowledge/api/cmd` → `knowledge/serve/cmd/api` |
| `serve-move-api-internal` | `knowledge/api/internal/*` → `knowledge/serve/internal/` (config, components, usecase, domain, transport/httpserver) |
| `serve-move-mcp-cmd` | `knowledge/mcp/cmd` → `knowledge/serve/cmd/mcp` |
| `serve-move-mcp-internal` | mcp transport → `internal/transport/mcpserver`; убрать дубль neo4j conn → `knowledge/connector` |
| `serve-dry-query` | `usecase` использует `connector/query` напрямую; убрать лишний слой в api storage если дублирует connector |
| `serve-rm-old-api` | Удалить `knowledge/api/` |
| `serve-rm-old-mcp` | Удалить `knowledge/mcp/` |
| `serve-build-api` | `go build ./knowledge/serve/cmd/api` |
| `serve-build-mcp` | `go build ./knowledge/serve/cmd/mcp` |

---

## Фаза 4 — go.work, Makefile, deploy

| ID | Действие |
|----|----------|
| `graph-work-rewrite` | [`knowledge/go.work`](knowledge/go.work): только `connector`, `ingest`, `serve` |
| `makefile-test-graph` | [`Makefile`](Makefile) `test-graph`: build ingest_worker + api + mcp из новых путей |
| `docker-ingest-dockerfile` | [`deploy/knowledge/docker/ingest_worker.Dockerfile`](deploy/knowledge/docker/ingest_worker.Dockerfile): `WORKDIR knowledge/ingest` |
| `docker-api-dockerfile` | [`deploy/knowledge/docker/api.Dockerfile`](deploy/knowledge/docker/api.Dockerfile): `WORKDIR knowledge/serve`, `cmd/api` |
| `docker-mcp-if-any` | MCP Dockerfile/compose service если есть |
| `compose-smoke` | `make test-graph` зелёный |

---

## Фаза 5 — выравнивание domain/usecase (кодстайл)

| ID | Действие |
|----|----------|
| `ingest-ti-envelope-package` | `ingest/apply.go` → `envelope.go` или `internal/sources/ti/envelope`; `setup.go` рядом |
| `ingest-domain-pkg-tidomain` | Убрать лишние alias-обёртки в graph TI domain где можно импортировать `pkg/tidomain` из usecase только |
| `ingest-pr-checklist` | Пройти PR checklist из coding-style: cmd без Cypher, usecase без NATS subjects |
| `serve-pr-checklist` | serve: cmd без Cypher; transport без Bolt driver |

---

## Критерии готовности

- `knowledge/go.work` — 3 модуля (`connector`, `ingest`, `serve`)
- Нет `knowledge/sources/`, `knowledge/storage/`, `knowledge/ingest_worker/`, `knowledge/api/`, `knowledge/mcp/`, `knowledge/neo4jclient/` на верхнем уровне
- `ingest_worker` / `api` / `mcp` собираются из `knowledge/ingest` и `knowledge/serve`
- Import paths: только `knowledge/ingest/...`, `knowledge/serve/...`, `knowledge/connector/...`, repo `pkg/*`
- `make test-graph` зелёный

## Риски

| Риск | Митигация |
|------|-----------|
| Mass import path churn | По одному source/бинарнику за коммит |
| Docker GOWORK paths | Обновить deploy в той же фазе 4 |
| MCP/API shared code | Один модуль `serve` — общий `internal/config` для Neo4j env |
