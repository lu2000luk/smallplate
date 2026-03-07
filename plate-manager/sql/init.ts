import { db } from "..";

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
  }

  if (plates) {
    db.run(`
      CREATE TABLE IF NOT EXISTS plates (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        data TEXT,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
      );
    `);

    db.run(`
      CREATE INDEX IF NOT EXISTS idx_plates_user_id
      ON plates(user_id);
    `);
  }
}
