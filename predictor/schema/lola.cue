lola: {
    Artifact: {
        name:        string
        description: string
        os:          [...string]

        mitreID?:    string
        category?:   string
        privileges?: string

        commands: [...{
            command:     string
            description: string
            usecase?:    string
            category?:   string
            privileges?: string
            mitreID?:    string
            os?:         [...string]
            tags?:       [...string]
        }]

        paths?: [...string]

        detection?: {
            sigma?: [...string]
            yara?:  [...string]
        }

        resources?: [...{
            link: string
        }]

        acknowledgement?: [...{
            name:   string
            handle: string
        }]
    }
}
