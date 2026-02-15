// http_request.cue

#NonEmptyString: string & !=""
#Method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS"

#Scheme: "http" | "https"

// Очень консервативный host
#Host: string & =~"^[a-zA-Z0-9.-]+$"

// Путь: начинаем с /, без пробелов
#Path: string & =~"^/[A-Za-z0-9._~!$&'()*+,;=:@/\\-]*$"

// Имя заголовка (упрощённо)
#HeaderName: string & =~"^[A-Za-z0-9-]+$"

// Имя query‑параметра
#QueryName: string & =~"^[A-Za-z0-9_]+$"

// Примитивные значения
#Value: string | int | float | bool

// Тип тела (логический, не raw)
#BodyType: "json" | "form" | "raw"

// JSON‑схема на уровне “ключ‑значение”
#JsonField: {
    name: #QueryName
    required?: bool | *false
    type: "string" | "number" | "boolean" | "object" | "array"
}

// Описание JSON‑тела
#JsonBodySchema: {
    fields: [...#JsonField] & len(fields) > 0

    // уникальность имён полей
    _names: [i]: string @tag(name)
    for i, f in fields { _names[i]: f.name }
    if len(_names) != len(unique(_names)) {
        _error_json_fields: "duplicate JSON field names in body schema"
    }
}

// Описание form‑body (application/x-www-form-urlencoded или multipart)
#FormField: {
    name: #QueryName
    required?: bool | *false
    multiple?: bool | *false
}

#FormBodySchema: {
    fields: [...#FormField] & len(fields) > 0

    _names: [i]: string @tag(name)
    for i, f in fields { _names[i]: f.name }
    if len(_names) != len(unique(_names)) {
        _error_form_fields: "duplicate form field names in body schema"
    }
}

// Тело запроса
#Body: {
    type: #BodyType

    json?: #JsonBodySchema
    form?: #FormBodySchema
    rawMime?: string & =~"^[a-zA-Z0-9!#$&^_.+-]+/[a-zA-Z0-9!#$&^_.+-]+$"

    // Инварианты: в зависимости от type
    if type == "json" {
        json: #JsonBodySchema
        form?: _|_
        rawMime?: _|_
    }
    if type == "form" {
        form: #FormBodySchema
        json?: _|_
        rawMime?: _|_
    }
    if type == "raw" {
        rawMime: string
        json?: _|_
        form?: _|_
    }
}

// Query‑параметр
#QueryParam: {
    name: #QueryName
    required?: bool | *false
    // список значений или одно
    multiple?: bool | *false
}

// Header
#Header: {
    name: #HeaderName
    value: #NonEmptyString
}

// Policy: разрешённые хосты/домены (можно вынести в отдельный файл)
#AllowedHosts: [...#Host] & len(#AllowedHosts) > 0

// Главный объект HTTP‑запроса
request: {
    method: #Method

    scheme: #Scheme | *"https"

    host: #Host
    if !(host in {for h in #AllowedHosts { h: true }}) {
        _error_host: "host \(host) is not allowed by policy"
    }

    port?: int & >0 & <=65535

    path: #Path

    // Query‑параметры
    query?: [...#QueryParam] | *[]

    // Headers
    headers?: [...#Header] | *[]

    // Body (для методов, которые допускают тело)
    body?: #Body & {
        if method == "GET" | method == "HEAD" | method == "DELETE" {
            _error_body: "body is not allowed for method \(method) by policy"
        }
    }

    // Таймауты и ретраи
    timeoutMs?: int & >0 & <=60000 | *5000
    maxRetries?: int & >=0 & <=5 | *0

    // Инварианты: уникальность имён заголовков и query‑параметров
    if headers != _|_ {
        _hdrNames: [i]: string @tag(name)
        for i, h in headers { _hdrNames[i]: h.name }
        if len(_hdrNames) != len(unique(_hdrNames)) {
            _error_headers: "duplicate header names"
        }
    }

    if query != _|_ {
        _qNames: [i]: string @tag(name)
        for i, q in query { _qNames[i]: q.name }
        if len(_qNames) != len(unique(_qNames)) {
            _error_query: "duplicate query parameter names"
        }
    }
}

// Глобальный policy‑whitelist хостов
#AllowedHosts: [
    "api.example.com",
    "auth.example.com",
]
