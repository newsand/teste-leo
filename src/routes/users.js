const express = require('express');
const bcrypt = require('bcrypt');
const { getDb } = require('../db');
const { validateCreateUser, validateUpdateUser } = require('../validators');

const router = express.Router();
const SALT_ROUNDS = 10;

function omitPassword(user) {
  if (!user) return user;
  const { password, ...rest } = user;
  return rest;
}

router.post('/', async (req, res) => {
  try {
    const validation = validateCreateUser(req.body);
    if (!validation.valid) {
      return res.status(400).json({ error: 'Validation failed', details: validation.errors });
    }

    const { name, email, password } = req.body;
    const db = getDb();

    const existing = db.prepare('SELECT id FROM users WHERE email = ?').get(email);
    if (existing) {
      return res.status(400).json({ error: 'Email already exists' });
    }

    const hashedPassword = await bcrypt.hash(password, SALT_ROUNDS);

    const stmt = db.prepare(`
      INSERT INTO users (name, email, password, created_at, updated_at)
      VALUES (?, ?, ?, datetime('now'), datetime('now'))
    `);
    const result = stmt.run(name.trim(), email.toLowerCase(), hashedPassword);

    const user = db.prepare('SELECT * FROM users WHERE id = ?').get(result.lastInsertRowid);

    res.status(201).json(omitPassword(user));
  } catch (error) {
    console.error('Error creating user:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/', (req, res) => {
  try {
    const db = getDb();

    const page = Math.max(1, parseInt(req.query.page, 10) || 1);
    const limit = Math.min(100, Math.max(1, parseInt(req.query.limit, 10) || 10));
    const offset = (page - 1) * limit;

    const countResult = db.prepare('SELECT COUNT(*) as total FROM users').get();
    const total = countResult.total;
    const totalPages = Math.ceil(total / limit);

    const users = db.prepare(`
      SELECT * FROM users
      ORDER BY created_at DESC
      LIMIT ? OFFSET ?
    `).all(limit, offset);

    res.json({
      data: users.map(omitPassword),
      pagination: {
        page,
        limit,
        total,
        totalPages,
        hasNextPage: page < totalPages,
        hasPreviousPage: page > 1
      }
    });
  } catch (error) {
    console.error('Error listing users:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/:id', (req, res) => {
  try {
    const db = getDb();
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      return res.status(400).json({ error: 'Invalid user ID' });
    }

    const user = db.prepare('SELECT * FROM users WHERE id = ?').get(id);

    if (!user) {
      return res.status(404).json({ error: 'User not found' });
    }

    res.json(omitPassword(user));
  } catch (error) {
    console.error('Error getting user:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id', async (req, res) => {
  try {
    const db = getDb();
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      return res.status(400).json({ error: 'Invalid user ID' });
    }

    const existing = db.prepare('SELECT * FROM users WHERE id = ?').get(id);
    if (!existing) {
      return res.status(404).json({ error: 'User not found' });
    }

    const validation = validateUpdateUser(req.body);
    if (!validation.valid) {
      return res.status(400).json({ error: 'Validation failed', details: validation.errors });
    }

    const updates = {};
    if (req.body.name !== undefined) updates.name = req.body.name.trim();
    if (req.body.email !== undefined) updates.email = req.body.email.toLowerCase();
    if (req.body.password !== undefined) {
      updates.password = await bcrypt.hash(req.body.password, SALT_ROUNDS);
    }

    if (updates.email && updates.email !== existing.email) {
      const emailExists = db.prepare('SELECT id FROM users WHERE email = ? AND id != ?').get(updates.email, id);
      if (emailExists) {
        return res.status(400).json({ error: 'Email already exists' });
      }
    }

    const setClauses = Object.keys(updates).map(key => `${key} = ?`).join(', ');
    const values = Object.values(updates);

    db.prepare(`UPDATE users SET ${setClauses}, updated_at = datetime('now') WHERE id = ?`).run(...values, id);

    const user = db.prepare('SELECT * FROM users WHERE id = ?').get(id);
    res.json(omitPassword(user));
  } catch (error) {
    console.error('Error updating user:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.patch('/:id', async (req, res) => {
  try {
    const db = getDb();
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      return res.status(400).json({ error: 'Invalid user ID' });
    }

    const existing = db.prepare('SELECT * FROM users WHERE id = ?').get(id);
    if (!existing) {
      return res.status(404).json({ error: 'User not found' });
    }

    const validation = validateUpdateUser(req.body);
    if (!validation.valid) {
      return res.status(400).json({ error: 'Validation failed', details: validation.errors });
    }

    const updates = {};
    if (req.body.name !== undefined) updates.name = req.body.name.trim();
    if (req.body.email !== undefined) updates.email = req.body.email.toLowerCase();
    if (req.body.password !== undefined) {
      updates.password = await bcrypt.hash(req.body.password, SALT_ROUNDS);
    }

    if (updates.email && updates.email !== existing.email) {
      const emailExists = db.prepare('SELECT id FROM users WHERE email = ? AND id != ?').get(updates.email, id);
      if (emailExists) {
        return res.status(400).json({ error: 'Email already exists' });
      }
    }

    const setClauses = Object.keys(updates).map(key => `${key} = ?`).join(', ');
    const values = Object.values(updates);

    db.prepare(`UPDATE users SET ${setClauses}, updated_at = datetime('now') WHERE id = ?`).run(...values, id);

    const user = db.prepare('SELECT * FROM users WHERE id = ?').get(id);
    res.json(omitPassword(user));
  } catch (error) {
    console.error('Error updating user:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.delete('/:id', (req, res) => {
  try {
    const db = getDb();
    const id = parseInt(req.params.id, 10);

    if (isNaN(id)) {
      return res.status(400).json({ error: 'Invalid user ID' });
    }

    const existing = db.prepare('SELECT id FROM users WHERE id = ?').get(id);
    if (!existing) {
      return res.status(404).json({ error: 'User not found' });
    }

    db.prepare('DELETE FROM users WHERE id = ?').run(id);

    res.status(200).json({ message: 'User deleted successfully' });
  } catch (error) {
    console.error('Error deleting user:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

module.exports = router;
