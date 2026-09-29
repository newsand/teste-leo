const express = require('express');
const usersRouter = require('./routes/users');
const { getDb, closeDb } = require('./db');

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/users', usersRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

app.use((err, req, res, next) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

if (require.main === module) {
  getDb();

  const server = app.listen(PORT, () => {
    console.log(`Server running on port ${PORT}`);
  });

  process.on('SIGTERM', () => {
    console.log('SIGTERM received, shutting down...');
    server.close(() => {
      closeDb();
      process.exit(0);
    });
  });

  process.on('SIGINT', () => {
    console.log('SIGINT received, shutting down...');
    server.close(() => {
      closeDb();
      process.exit(0);
    });
  });
}

module.exports = app;
