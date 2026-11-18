CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    revision INTEGER NOT NULL DEFAULT 1,
    creation_token TEXT UNIQUE,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    state TEXT NOT NULL CHECK (state IN ('', 'сухой', 'сырой', 'готовый')),
    kcal REAL NOT NULL CHECK (kcal >= 0),
    protein REAL NOT NULL CHECK (protein >= 0),
    fat REAL NOT NULL CHECK (fat >= 0),
    carbs REAL NOT NULL CHECK (carbs >= 0)
) STRICT;

CREATE TABLE plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    revision INTEGER NOT NULL DEFAULT 1
) STRICT;

CREATE TABLE meals (
    id INTEGER PRIMARY KEY,
    plan_id INTEGER NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    name TEXT NOT NULL CHECK (length(trim(name)) > 0),
    UNIQUE (plan_id, position)
) STRICT;

CREATE TABLE meal_items (
    meal_id INTEGER NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    position INTEGER NOT NULL,
    grams REAL NOT NULL CHECK (grams > 0),
    PRIMARY KEY (meal_id, position)
) STRICT;

CREATE INDEX meal_items_by_product ON meal_items(product_id);

CREATE TABLE settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    active_plan_id INTEGER REFERENCES plans(id) ON DELETE SET NULL
) STRICT;

INSERT INTO settings(id) VALUES (1);
PRAGMA user_version = 1;
