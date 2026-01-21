const sessionConfig = {
  secret: process.env.SESSION_SECRET || 'knowledgebase-secret-key',
  resave: false,
  saveUninitialized: true,
  cookie: { 
    secure: process.env.NODE_ENV === 'production', 
    maxAge: 24 * 60 * 60 * 1000, // 24 hours
    httpOnly: true
  }
};

module.exports = sessionConfig;
