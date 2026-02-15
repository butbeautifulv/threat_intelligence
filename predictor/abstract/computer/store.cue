#Store: {
    name: string
    kind: "postgres" | "neo4j" | "redis" | "s3" | "kv"
    config: {...}
    // Контракты чтения/записи
    read?: [...{
        queryType: string
        resultType: string
    }]
    write?: [...{
        inputType: string
    }]
}
