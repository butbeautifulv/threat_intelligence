detect: {
    Sigma: {
        id:          string
        title:       string
        description: string
        level:       "low" | "medium" | "high" | "critical"
        tags?:       [...string]
        logsource?: {
            product: string
            service?: string
        }
    }

    Yara: {
        name:        string
        description?: string
        author?:      string
        tags?:        [...string]
    }

    AtomicTest: {
        id:      string
        name:    string
        tactic?: string
        technique?: string
        executor?: {
            name: string
            command?: string
        }
    }
}
