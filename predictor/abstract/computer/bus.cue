#Bus: {
    name: string
    protocol: "http" | "kafka" | "grpc" | "sql" | "neo4j" | "file"
    messageType: string
    direction: "in" | "out"
    config: {...}
}
