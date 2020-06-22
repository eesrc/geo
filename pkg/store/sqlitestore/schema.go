package sqlitestore

const schema = `
PRAGMA foreign_keys = ON;
PRAGMA defer_foreign_keys = FALSE;

CREATE TABLE IF NOT EXISTS users (
    id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name            VARCHAR(255),
    email           VARCHAR(320),
    email_verified  BOOL,
    phone           VARCHAR(18),
    phone_verified  BOOL,
    deleted         BOOL,
    admin           BOOL,
    created         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    github_id       TEXT,
    connect_id      TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS users_github_id_idx ON users (github_id);

CREATE TABLE IF NOT EXISTS tokens (
    token       TEXT NOT NULL PRIMARY KEY,
    resource    TEXT,
    user_id     INTEGER,
    perm_write  BOOL,
    created     TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS teams (
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name         VARCHAR(255),
    description  TEXT
);

CREATE TABLE IF NOT EXISTS shape_collections (
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    team_id      INTEGER NOT NULL,
    name         VARCHAR(255),
    description  TEXT,

    FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS shapes (
    id                  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    shape_collection_id INTEGER NOT NULL,
    name                VARCHAR(255),
    type                VARCHAR(64),
    properties          JSON,
    shape               BLOB,

    FOREIGN KEY(shape_collection_id) REFERENCES shape_collections(id)
);

CREATE INDEX IF NOT EXISTS idx_shapes_shapecollection ON shapes(shape_collection_id);

CREATE TABLE IF NOT EXISTS team_members (
    user_id     INTEGER NOT NULL,
    team_id     INTEGER NOT NULL,
    admin       BOOL,

    PRIMARY KEY(user_id,team_id),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS collections(
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    team_id      INTEGER,
    name         VARCHAR(255),
    description  TEXT,

    FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS trackers(
    id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    collection_id   INTEGER NOT NULL,
    name            VARCHAR(255),
    description     TEXT,

    FOREIGN KEY(collection_id) REFERENCES collections(id)
);

CREATE TABLE IF NOT EXISTS positions(
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    tracker_id   INTEGER NOT NULL,
    ts           INTEGER,
    lat          REAL,
    lon          REAL,
    alt          REAL,
    heading      REAL,
    speed        REAL,
    payload      BYTEA,
    precision    REAL,

    FOREIGN KEY(tracker_id) REFERENCES trackers(id)
);

CREATE INDEX IF NOT EXISTS idx_positions_ts ON positions(ts);

CREATE TABLE IF NOT EXISTS subscriptions(
    id                  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    team_id             INTEGER NOT NULL,
    name                VARCHAR(255),
    description         TEXT,
    active              BOOL,
    output              VARCHAR(32),
    output_config       JSONB,
    types               JSONB,
    confidences         JSONB,
    shape_collection_id INTEGER,
    trackable_type      VARCHAR(32),
    trackable_id        INTEGER,

    FOREIGN KEY(team_id) REFERENCES teams(id),
    FOREIGN KEY(shape_collection_id) REFERENCES shape_collections(id)
);

CREATE TABLE IF NOT EXISTS position_movements(
    id              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    subscription_id INTEGER NOT NULL,
    tracker_id      INTEGER NOT NULL,
    shape_id        INTEGER NOT NULL,
    position_id     INTEGER NOT NULL,
    movement        JSONB,

    FOREIGN KEY(subscription_id) REFERENCES subscriptions(id),
    FOREIGN KEY(tracker_id) REFERENCES trackers(id),
    FOREIGN KEY(position_id) REFERENCES positions(id)
);

CREATE INDEX IF NOT EXISTS idx_position_movements_subscription ON position_movements(subscription_id);
`
