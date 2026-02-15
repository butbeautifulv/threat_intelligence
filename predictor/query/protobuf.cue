// proto_schema.cue

#Identifier: string & =~"^[A-Z][A-Za-z0-9_]*$"
#FieldName: string & =~"^[a-z][a-z0-9_]*$"
#PackageName: string & =~"^[a-z][a-z0-9_.]*$"

#ScalarType: "string" | "bytes" |
             "int32" | "int64" | "uint32" | "uint64" |
             "sint32" | "sint64" |
             "fixed32" | "fixed64" |
             "sfixed32" | "sfixed64" |
             "bool" | "float" | "double"

// Ограничения на номера полей
#FieldNumber: int & >=1 & <=536870911 &
    !(number >=19000 & number <=19999) // зарезервировано Google

// Enum value
#EnumValue: {
    name: #Identifier
    number: int
}

// Enum
#Enum: {
    name: #Identifier
    values: [...#EnumValue] & len(values) > 0

    // уникальность имён
    _names: [i]: string @tag(name)
    for i, v in values { _names[i]: v.name }
    if len(_names) != len(unique(_names)) {
        _error_enum_names: "duplicate enum value names in \(name)"
    }

    // уникальность номеров
    _nums: [i]: int @tag(num)
    for i, v in values { _nums[i]: v.number }
    if len(_nums) != len(unique(_nums)) {
        _error_enum_nums: "duplicate enum value numbers in \(name)"
    }
}

// Field
#Field: {
    name: #FieldName
    number: #FieldNumber

    // Тип: скаляр, enum или message
    type: #ScalarType | #Identifier

    // repeated
    repeated?: bool | *false

    // optional (proto3)
    optional?: bool | *false

    // JSON name
    jsonName?: string & =~"^[a-zA-Z][A-Za-z0-9_]*$"

    // Комментарий без инъекций
    comment?: string & !~"(--|;)"
}

// Message
#Message: {
    name: #Identifier

    fields: [...#Field] & len(fields) > 0

    nestedMessages?: [...#Message] | *[]
    nestedEnums?: [...#Enum] | *[]

    // уникальность имён полей
    _fieldNames: [i]: string @tag(name)
    for i, f in fields { _fieldNames[i]: f.name }
    if len(_fieldNames) != len(unique(_fieldNames)) {
        _error_field_names: "duplicate field names in message \(name)"
    }

    // уникальность номеров
    _fieldNums: [i]: int @tag(num)
    for i, f in fields { _fieldNums[i]: f.number }
    if len(_fieldNums) != len(unique(_fieldNums)) {
        _error_field_nums: "duplicate field numbers in message \(name)"
    }
}

// Главный объект .proto
proto: {
    syntax: "proto3"

    package: #PackageName

    messages: [...#Message] | *[]
    enums:    [...#Enum]    | *[]

    // Инвариант: все типы, используемые в полях, должны существовать
    _typeMap: {
        for m in messages { (m.name): "message" }
        for e in enums    { (e.name): "enum" }
    }

    for m in messages {
        for f in m.fields {
            if f.type =~"^[A-Z]" {
                if !(_typeMap[f.type]) {
                    _error_missing_type: "type \(f.type) used in \(m.name).\(f.name) is not defined"
                }
            }
        }
    }

    // Инвариант: имена сообщений уникальны
    _msgNames: [i]: string @tag(name)
    for i, m in messages { _msgNames[i]: m.name }
    if len(_msgNames) != len(unique(_msgNames)) {
        _error_msg_names: "duplicate message names"
    }

    // Инвариант: имена enum уникальны
    _enumNames: [i]: string @tag(name)
    for i, e in enums { _enumNames[i]: e.name }
    if len(_enumNames) != len(unique(_enumNames)) {
        _error_enum_names2: "duplicate enum names"
    }
}
