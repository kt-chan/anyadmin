const sessionConfig = {
  secret: process.env.SESSION_SECRET || 'knowledgebase-secret-key',
  resave: false,
  saveUninitialized: false,
  cookie: { 
    secure: false, // Changed to false for container/proxy environments without HTTPS
    maxAge: 24 * 60 * 60 * 1000, // 24 hours
    httpOnly: true,
    sameSite: 'lax'
  }
};

module.exports = sessionConfig;
