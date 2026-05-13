ti: {
    IOC: {
        type: "ip" | "url" | "hash" | "domain"
        value: string
        confidence?: int & >=0 & <=100
        tags?: [...string]
        source?: string
    }

    Campaign: {
        id:      string
        name:    string
        actors?: [...string]
        iocs?:   [...IOC]
        summary?: string
    }

    Cluster: {
        id:      string
        name:    string
        campaigns?: [...Campaign]
        description?: string
    }
}
