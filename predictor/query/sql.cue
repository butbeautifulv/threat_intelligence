// sqlschema.cue

// Базовые типы и утилиты
#NonEmptyString: string & !=""
#Identifier: #NonEmptyString & =~"^[a-z][a-z0-9_]*$" // строгий snake_case, без заглавных

#Bool:   "BOOL" | "BOOLEAN"
#Int:    "INT" | "INTEGER" | "BIGINT" | "SMALLINT"
#Float:  "FLOAT" | "DOUBLE" | "REAL" | "DECIMAL"
#Text:   "TEXT" | "VARCHAR" | "CHAR"
#Time:   "TIMESTAMP" | "TIMESTAMPTZ" | "DATE" | "TIME"
#JSON:   "JSON" | "JSONB"

#ScalarType: #Bool | #Int | #Float | #Text | #Time | #JSON

#OnDeleteAction: "NO ACTION" | "RESTRICT" | "CASCADE" | "SET NULL" | "SET DEFAULT"
#OnUpdateAction: "NO ACTION" | "RESTRICT" | "CASCADE" | "SET NULL" | "SET DEFAULT"

// Ограничения на длину для строковых типов
#StringType: {
    type: "VARCHAR" | "CHAR"
    length?: int & >0 & <=65535
}

// Унифицированное описание типа колонки
#ColumnType: {
    // либо скалярный тип без параметров
    type: #ScalarType
} | {
    // либо строковый тип с длиной
    #StringType
} | {
    // DECIMAL(p,s)
    type: "DECIMAL"
    precision: int & >0 & <=38
    scale:     int & >=0 & <=precision
}

// Ограничения на default: очень консервативно
#DefaultLiteral: string & =~"^[A-Z0-9_ ':-]+$"

// Описание колонки
#Column: {
    name: #Identifier

    // Тип
    #ColumnType

    // NOT NULL
    notNull?: bool | *false

    // DEFAULT (строгое, без выражений)
    default?: #DefaultLiteral

    // CHECK-constraint на уровне колонки (строка, но можно ограничить паттерном)
    check?: string & =~"^[A-Z0-9_ ()<>=!'+-/*.,:]+$"

    // Комментарий
    comment?: string & !~"(--|;)" // не позволяем инъекции
}

// Primary key
#PrimaryKey: {
    name?: #Identifier
    columns: [...#Identifier] & len(columns) > 0
}

// Unique constraint
#UniqueConstraint: {
    name?: #Identifier
    columns: [...#Identifier] & len(columns) > 0
}

// Foreign key
#ForeignKey: {
    name?: #Identifier

    columns: [...#Identifier] & len(columns) > 0

    refTable:  #Identifier
    refColumns: [...#Identifier] & len(refColumns) == len(columns)

    onDelete?: #OnDeleteAction | *"NO ACTION"
    onUpdate?: #OnUpdateAction | *"NO ACTION"
}

// Index
#Index: {
    name: #Identifier
    unique?: bool | *false
    columns: [...{
        name: #Identifier
        order?: "ASC" | "DESC" | *"ASC"
    }] & len(columns) > 0
}

// Таблица
#Table: {
    name: #Identifier

    // Колонки
    columns: [...#Column] & len(columns) > 0

    // PK
    primaryKey?: #PrimaryKey

    // UNIQUE
    uniques?: [...#UniqueConstraint] | *[]

    // FK
    foreignKeys?: [...#ForeignKey] | *[]


    // Индексы
    indexes?: [...#Index] | *[]


    // Комментарий
    comment?: string & !~"(--|;)"


    // Инварианты на уровне таблицы
    // 1) Имена колонок уникальны
    _colNames: [i]: string @tag(name)
    for i, c in columns {
        _colNames[i]: c.name
    }
    if len(_colNames) != len(unique(_colNames)) {
        _error: "duplicate column names in table \(name)"
    }

    // 2) Все колонки в PK существуют
    if primaryKey != _|_ {
        for pkCol in primaryKey.columns {
            if !(pkCol in {for c in columns { c.name: true }}) {
                _error_pk: "primary key column \(pkCol) not found in table \(name)"
            }
        }
    }

    // 3) Все колонки в UNIQUE существуют
    for u in uniques {
        for uc in u.columns {
            if !(uc in {for c in columns { c.name: true }}) {
                _error_u: "unique column \(uc) not found in table \(name)"
            }
        }
    }

    // 4) Все колонки в индексах существуют
    for idx in indexes {
        for ic in idx.columns {
            if !(ic.name in {for c in columns { c.name: true }}) {
                _error_idx: "index column \(ic.name) not found in table \(name)"
            }
        }
    }
}

// Вся схема БД
schema: {
    // Версия схемы (для миграций)
    version: int & >=1

    // Имя БД (логическое)
    name: #Identifier

    // Таблицы
    tables: [...#Table] & len(tables) > 0

    // Инварианты на уровне схемы

    // 1) Имена таблиц уникальны
    _tblNames: [i]: string @tag(name)
    for i, t in tables {
        _tblNames[i]: t.name
    }
    if len(_tblNames) != len(unique(_tblNames)) {
        _error_tables: "duplicate table names in schema"
    }

    // 2) FK ссылаются на существующие таблицы и колонки
    _tableMap: {
        for t in tables {
            (t.name): {
                columns: {for c in t.columns { (c.name): true }}
            }
        }
    }

    for t in tables {
        if t.foreignKeys != _|_ {
            for fk in t.foreignKeys {
                if !(fk.refTable in _tableMap) {
                    _error_fk_table: "FK \(fk.name) in table \(t.name) references non-existing table \(fk.refTable)"
                }
                // Проверка колонок назначения
                for rc in fk.refColumns {
                    if !(_tableMap[fk.refTable].columns[rc]) {
                        _error_fk_col: "FK \(fk.name) in table \(t.name) references non-existing column \(rc) in table \(fk.refTable)"
                    }
                }
                // Проверка локальных колонок
                for lc in fk.columns {
                    if !(lc in {for c in t.columns { c.name: true }}) {
                        _error_fk_local: "FK \(fk.name) in table \(t.name) uses non-existing local column \(lc)"
                    }
                }
            }
        }
    }
}
