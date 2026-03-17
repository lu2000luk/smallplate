import { db } from "..";

function getTableColumns(tableName: string): string[] {
  const statement = db.query(`PRAGMA table_info(${tableName})`) as {
    all(): Array<{ name: string }>;
  };

  return statement.all().map((column) => column.name);
}

function ensureTableColumn(
  tableName: string,
  columnName: string,
  definition: string,
) {
  const columns = getTableColumns(tableName);
  if (columns.includes(columnName)) {
    return;
  }

  db.run(`ALTER TABLE ${tableName} ADD COLUMN ${columnName} ${definition}`);
}

export async function sql_init(users = true, plates = true) {
  if (users) {
    db.run(`
      CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        data TEXT
      );
    `);

    db.run(`
      CREATE TABLE IF NOT EXISTS user_auth_keys (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        auth_key TEXT NOT NULL UNIQUE,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
      );
    `);

    db.run(`
      CREATE INDEX IF NOT EXISTS idx_user_auth_keys_user_id
      ON user_auth_keys(user_id);
    `);

    db.run(`
      CREATE INDEX IF NOT EXISTS idx_user_auth_keys_auth_key
      ON user_auth_keys(auth_key);
    `);
  }

  if (plates) {
    db.run(`
      CREATE TABLE IF NOT EXISTS plates (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        servers TEXT,
        data TEXT,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
      );
    `);

    ensureTableColumn("plates", "servers", "TEXT");
    ensureTableColumn("plates", "data", "TEXT");

    db.run(`
      UPDATE plates
      SET servers = '{}'
      WHERE servers IS NULL;
    `);

    db.run(`
      UPDATE plates
      SET data = '{"enabled_services":[]}'
      WHERE data IS NULL;
    `);

    db.run(`
      CREATE INDEX IF NOT EXISTS idx_plates_user_id
      ON plates(user_id);
    `);

    db.run(`
      CREATE TABLE IF NOT EXISTS api_keys (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        plate_id INTEGER NOT NULL,
        api_key TEXT NOT NULL UNIQUE,
        created_at INTEGER NOT NULL,
        FOREIGN KEY (plate_id) REFERENCES plates(id) ON DELETE CASCADE
      );
    `);

    db.run(`
      CREATE INDEX IF NOT EXISTS idx_api_keys_plate_id
      ON api_keys(plate_id);
    `);

    db.run(`
      CREATE INDEX IF NOT EXISTS idx_api_keys_api_key
      ON api_keys(api_key);
    `);
  }
}
