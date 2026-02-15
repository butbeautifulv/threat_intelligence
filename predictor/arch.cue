package predictor

Layer: "controller" | "usecase" | "domain" | "adapter"

#AllowedDeps: {
    controller: [...Layer] & ["usecase"]
    usecase:    [...Layer] & ["domain", "adapter"]
    domain:     [...Layer] & []
    adapter:    [...Layer] & ["domain"]
}

Module: {
    name:        string
    layer:       Layer
    imports:     [...string] // список импортируемых модулей (по имени слоя)
    description: string
    version:     =~"^v[0-9]+\\.[0-9]+\\.[0-9]+$"
    owner:       string

    // правило: каждый импорт должен быть в списке разрешённых для слоя
    for i, imp in imports {
        imp: Layer
        imp: _ @allowed
        _ @allowed: #AllowedDeps[layer][i]
    }
}
