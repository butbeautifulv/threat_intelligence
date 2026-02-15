#Protocol: "http" | "kafka" | "sql" | "neo4j" | "grpc" | "protobuf-rpc"

#Direction: "in" | "out"

#Port: {
    name:       string
    direction:  #Direction
    protocol:   #Protocol

    // Тип сообщения на этом порту (логический, доменный)
    messageType: string

    // Привязка к конкретному транспорту/ресурсу
    config: {
        // Для http
        path?:   string
        method?: "GET" | "POST" | "PUT" | "DELETE"

        // Для kafka
        topic?:  string
        groupId?: string

        // Для sql
        dsn?:    string
        query?:  string

        // Для neo4j
        uri?:    string
        cypher?: string
    }
}

#Adapter: {
    name: string

    // Откуда и куда адаптируем
    from: {
        protocol: #Protocol
        messageType: string
    }
    to: {
        protocol: #Protocol
        messageType: string
    }

    // Ссылка на реализацию (Go-пакет, бинарь, handler-id)
    impl: {
        language: "go" | "python" | "rust" | "cue-gen"
        ref:      string
    }
}

#DomainHandler: {
    name: string

    // Доменный контракт
    inputType:  string
    outputType: string

    // Ссылка на реализацию доменной логики
    impl: {
        language: "go" | "rust" | "cue-gen"
        ref:      string
    }
}

#Node: {
    name: string
    kind: "worker" | "api" | "scheduler" | "factory" | "adapter-only"

    // Порты
    ports: [...#Port]

    // Доменные хендлеры
    domainHandlers: [...#DomainHandler]

    // Адаптеры, привязанные к портам
    adapters: [...#Adapter]

    // Общий конфиг узла (env, ресурсы, etc.)
    config: {
        replicas?: int
        resources?: {
            cpu?: string
            mem?: string
        }
        env?: [string]: string
    }
}
