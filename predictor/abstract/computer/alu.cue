#ALU: {
    handlers: [...{
        name:       string
        inputType:  string
        outputType: string
        impl: {
            language: "go" | "rust" | "python"
            ref:      string
        }
    }]
}
