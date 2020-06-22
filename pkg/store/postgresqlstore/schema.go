package postgresqlstore

const postgresSchema = `
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(255),
    email           VARCHAR(320),
    email_verified  BOOL,
    phone           VARCHAR(18),
    phone_verified  BOOL,
    deleted         BOOL,
    admin           BOOL,
    created         TIMESTAMPTZ DEFAULT NOW(),
    github_id       TEXT,
    connect_id      TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS users_github_id_idx ON users (github_id);

CREATE TABLE IF NOT EXISTS tokens (
    token       TEXT NOT NULL PRIMARY KEY,
    resource    TEXT,
    user_id     INTEGER,
    perm_write  BOOL,
    created     TIMESTAMPTZ,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS teams (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(255),
    description  TEXT
);

CREATE TABLE IF NOT EXISTS shape_collections (
    id           SERIAL PRIMARY KEY,
    team_id      INTEGER NOT NULL,
    name         VARCHAR(255),
    description  TEXT,

    FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS shapes (
    id                  SERIAL PRIMARY KEY,
    shape_collection_id  INTEGER NOT NULL,
    name                VARCHAR(255),
    type                VARCHAR(64),
    properties          JSONB,
    shape               BYTEA,

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
    id           SERIAL PRIMARY KEY,
    team_id      INTEGER,
    name         VARCHAR(255),
    description  TEXT,

    FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS trackers(
    id              SERIAL PRIMARY KEY,
    collection_id   INTEGER NOT NULL,
    name            VARCHAR(255),
    description     TEXT,

    FOREIGN KEY(collection_id) REFERENCES collections(id)
);

CREATE TABLE IF NOT EXISTS positions(
    id           SERIAL PRIMARY KEY,
    tracker_id   INTEGER NOT NULL,
    ts           BIGINT,
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
    id                  SERIAL PRIMARY KEY,
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
    id              SERIAL PRIMARY KEY,

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
