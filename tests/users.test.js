const { describe, it, before, after, beforeEach } = require('node:test');
const assert = require('node:assert');
const http = require('http');
const path = require('path');
const fs = require('fs');

process.env.DB_PATH = path.join(__dirname, 'test.db');

const app = require('../src/index');
const { getDb, closeDb } = require('../src/db');

let server;
let baseUrl;

function request(method, path, body = null) {
  return new Promise((resolve, reject) => {
    const url = new URL(path, baseUrl);
    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname + url.search,
      method,
      headers: {
        'Content-Type': 'application/json'
      }
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve({ status: res.statusCode, body: data ? JSON.parse(data) : null });
        } catch {
          resolve({ status: res.statusCode, body: data });
        }
      });
    });

    req.on('error', reject);
    if (body) req.write(JSON.stringify(body));
    req.end();
  });
}

before(async () => {
  server = app.listen(0);
  await new Promise(resolve => server.on('listening', resolve));
  const port = server.address().port;
  baseUrl = `http://127.0.0.1:${port}`;
});

beforeEach(() => {
  const db = getDb();
  db.exec('DELETE FROM users');
});

after(async () => {
  if (server) {
    await new Promise(resolve => server.close(resolve));
  }
  closeDb();
  try {
    fs.unlinkSync(process.env.DB_PATH);
    fs.unlinkSync(process.env.DB_PATH + '-wal');
    fs.unlinkSync(process.env.DB_PATH + '-shm');
  } catch {}
  setTimeout(() => process.exit(0), 100);
});

describe('POST /users', () => {
  it('should create a user with valid data', async () => {
    const res = await request('POST', '/users', {
      name: 'John Doe',
      email: 'john@example.com',
      password: 'secret123'
    });

    assert.strictEqual(res.status, 201);
    assert.strictEqual(res.body.name, 'John Doe');
    assert.strictEqual(res.body.email, 'john@example.com');
    assert.ok(res.body.id);
    assert.ok(res.body.created_at);
    assert.ok(!res.body.password, 'Password should not be returned');
  });

  it('should return 400 for missing required fields', async () => {
    const res = await request('POST', '/users', { name: 'John' });

    assert.strictEqual(res.status, 400);
    assert.ok(res.body.error);
  });

  it('should return 400 for invalid email', async () => {
    const res = await request('POST', '/users', {
      name: 'John Doe',
      email: 'invalid-email',
      password: 'secret123'
    });

    assert.strictEqual(res.status, 400);
    assert.ok(res.body.details.some(d => d.includes('email')));
  });

  it('should return 400 for duplicate email', async () => {
    await request('POST', '/users', {
      name: 'John Doe',
      email: 'john@example.com',
      password: 'secret123'
    });

    const res = await request('POST', '/users', {
      name: 'Jane Doe',
      email: 'john@example.com',
      password: 'secret456'
    });

    assert.strictEqual(res.status, 400);
    assert.ok(res.body.error.includes('Email already exists'));
  });
});

describe('GET /users', () => {
  it('should return paginated list of users', async () => {
    await request('POST', '/users', { name: 'User 1', email: 'user1@example.com', password: 'pass1' });
    await request('POST', '/users', { name: 'User 2', email: 'user2@example.com', password: 'pass2' });

    const res = await request('GET', '/users');

    assert.strictEqual(res.status, 200);
    assert.ok(Array.isArray(res.body.data));
    assert.strictEqual(res.body.data.length, 2);
    assert.ok(res.body.pagination);
    assert.strictEqual(res.body.pagination.total, 2);
  });

  it('should respect pagination parameters', async () => {
    for (let i = 1; i <= 5; i++) {
      await request('POST', '/users', { name: `User ${i}`, email: `user${i}@example.com`, password: 'pass' });
    }

    const res = await request('GET', '/users?page=2&limit=2');

    assert.strictEqual(res.status, 200);
    assert.strictEqual(res.body.data.length, 2);
    assert.strictEqual(res.body.pagination.page, 2);
    assert.strictEqual(res.body.pagination.limit, 2);
    assert.strictEqual(res.body.pagination.totalPages, 3);
  });

  it('should not include password in list response', async () => {
    await request('POST', '/users', { name: 'User', email: 'user@example.com', password: 'secret' });

    const res = await request('GET', '/users');

    assert.strictEqual(res.status, 200);
    assert.ok(!res.body.data[0].password);
  });
});

describe('GET /users/:id', () => {
  it('should return user by id', async () => {
    const createRes = await request('POST', '/users', {
      name: 'John Doe',
      email: 'john@example.com',
      password: 'secret123'
    });

    const res = await request('GET', `/users/${createRes.body.id}`);

    assert.strictEqual(res.status, 200);
    assert.strictEqual(res.body.id, createRes.body.id);
    assert.strictEqual(res.body.name, 'John Doe');
    assert.ok(!res.body.password);
  });

  it('should return 404 for non-existent user', async () => {
    const res = await request('GET', '/users/99999');

    assert.strictEqual(res.status, 404);
    assert.ok(res.body.error.includes('not found'));
  });

  it('should return 400 for invalid id', async () => {
    const res = await request('GET', '/users/invalid');

    assert.strictEqual(res.status, 400);
  });
});

describe('PUT /users/:id', () => {
  it('should update user', async () => {
    const createRes = await request('POST', '/users', {
      name: 'John Doe',
      email: 'john@example.com',
      password: 'secret123'
    });

    const res = await request('PUT', `/users/${createRes.body.id}`, {
      name: 'John Updated'
    });

    assert.strictEqual(res.status, 200);
    assert.strictEqual(res.body.name, 'John Updated');
    assert.strictEqual(res.body.email, 'john@example.com');
  });

  it('should return 404 for non-existent user', async () => {
    const res = await request('PUT', '/users/99999', { name: 'Test' });

    assert.strictEqual(res.status, 404);
  });
});

describe('PATCH /users/:id', () => {
  it('should partially update user', async () => {
    const createRes = await request('POST', '/users', {
      name: 'John Doe',
      email: 'john@example.com',
      password: 'secret123'
    });

    const res = await request('PATCH', `/users/${createRes.body.id}`, {
      email: 'newemail@example.com'
    });

    assert.strictEqual(res.status, 200);
    assert.strictEqual(res.body.email, 'newemail@example.com');
    assert.strictEqual(res.body.name, 'John Doe');
  });
});

describe('DELETE /users/:id', () => {
  it('should delete user', async () => {
    const createRes = await request('POST', '/users', {
      name: 'John Doe',
      email: 'john@example.com',
      password: 'secret123'
    });

    const res = await request('DELETE', `/users/${createRes.body.id}`);

    assert.strictEqual(res.status, 200);

    const getRes = await request('GET', `/users/${createRes.body.id}`);
    assert.strictEqual(getRes.status, 404);
  });

  it('should return 404 for non-existent user', async () => {
    const res = await request('DELETE', '/users/99999');

    assert.strictEqual(res.status, 404);
  });
});
