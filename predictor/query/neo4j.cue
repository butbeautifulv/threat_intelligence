// neo4j_query.cue

// Базовые типы
#NonEmptyString: string & !=""
#Identifier: #NonEmptyString & =~"^[a-z][a-z0-9_]*$"

// Разрешённые лейблы и типы связей (policy-слой)
#AllowedNodeLabels: [...#Identifier] & len(#AllowedNodeLabels) > 0
#AllowedRelTypes:  [...#Identifier] & len(#AllowedRelTypes) > 0

// Разрешённые свойства
#AllowedPropertyName: #Identifier

// Примитивные значения
#Value: string | int | float | bool

// Операторы сравнения
#CompareOp: "=" | "<>" | "<" | "<=" | ">" | ">=" | "IN" | "CONTAINS" | "STARTS WITH" | "ENDS WITH"

// Логические операторы
#LogicOp: "AND" | "OR"

// Условие WHERE (упрощённое, но безопасное)
#WhereExpr: {
    // Бинарное сравнение: n.prop = 42
    field: {
        alias: #Identifier
        property: #AllowedPropertyName
    }
    op: #CompareOp
    value: #Value
} | {
    // Логическое объединение
    logic: #LogicOp
    left:  #WhereExpr
    right: #WhereExpr
}

// Направление связи
#RelDirection: "->" | "<-" | "-"

// Описание узла в MATCH-паттерне
#NodePattern: {
    alias: #Identifier
    labels: [...#Identifier] & len(labels) > 0

    // policy: все лейблы должны быть из whitelist
    for l in labels {
        if !(l in {for x in #AllowedNodeLabels { x: true }}) {
            _error_label: "label \(l) is not allowed"
        }
    }
}

// Описание связи в MATCH-паттерне
#RelPattern: {
    alias?: #Identifier
    types: [...#Identifier] & len(types) > 0
    direction?: #RelDirection | *"->"
    minHops?: int & >=1
    maxHops?: int & >=minHops

    // policy: типы связей только из whitelist
    for t in types {
        if !(t in {for x in #AllowedRelTypes { x: true }}) {
            _error_rel: "relationship type \(t) is not allowed"
        }
    }

    // policy: ограничиваем глубину
    if maxHops != _|_ & maxHops > 5 {
        _error_depth: "maxHops > 5 is not allowed"
    }
}

// Один MATCH-паттерн: (a:User)-[r:FRIEND_OF]->(b:User)
#MatchPattern: {
    start: #NodePattern
    path: [...{
        rel:  #RelPattern
        node: #NodePattern
    }] & len(path) > 0
}

// RETURN-элемент
#ReturnItem: {
    alias: #Identifier
    property?: #AllowedPropertyName
    // AS-алиас для вывода
    as?: #Identifier
}

// ORDER BY
#OrderBy: {
    alias: #Identifier
    property?: #AllowedPropertyName
    direction?: "ASC" | "DESC" | *"ASC"
}

// Главный объект запроса
query: {
    // Только READ-операции: MATCH + OPTIONAL MATCH + RETURN
    // Никаких CREATE/MERGE/DELETE/SET

    // Обязательный MATCH
    match: [...#MatchPattern] & len(match) > 0

    // Необязательный WHERE
    where?: #WhereExpr

    // RETURN
    returns: [...#ReturnItem] & len(returns) > 0

    // DISTINCT
    distinct?: bool | *false

    // ORDER BY
    orderBy?: [...#OrderBy] | *[]

    // LIMIT
    limit?: int & >0 & <=1000 // policy: ограничиваем размер результата

    // SKIP
    skip?: int & >=0 | *0

    // Инварианты: все alias, используемые в WHERE/RETURN/ORDER BY, должны быть объявлены в MATCH
    _aliases: {
        for m in match {
            (m.start.alias): true
            for p in m.path {
                (p.node.alias): true
                if p.rel.alias != _|_ {
                    (p.rel.alias): true
                }
            }
        }
    }

    if where != _|_ {
        _checkWhere: {
            // рекурсивная проверка alias в where
            #Check(w: #WhereExpr): {
                if w.field != _|_ {
                    if !(_aliases[w.field.alias]) {
                        _error_where_alias: "alias \(w.field.alias) in WHERE is not defined in MATCH"
                    }
                }
                if w.logic != _|_ {
                    #Check(w.left)
                    #Check(w.right)
                }
            }
            #Check(where)
        }
    }

    for r in returns {
        if !(_aliases[r.alias]) {
            _error_return_alias: "alias \(r.alias) in RETURN is not defined in MATCH"
        }
    }

    if orderBy != _|_ {
        for o in orderBy {
            if !(_aliases[o.alias]) {
                _error_order_alias: "alias \(o.alias) in ORDER BY is not defined in MATCH"
            }
        }
    }
}

// Глобальный policy-конфиг (можно вынести в отдельный файл и импортировать)
#AllowedNodeLabels: [
    "user",
    "post",
    "comment",
]

#AllowedRelTypes: [
    "friend_of",
    "wrote",
    "commented_on",
]
