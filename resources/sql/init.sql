CREATE TABLE IF NOT EXISTS auth_user(
	id SERIAL PRIMARY KEY,
	username TEXT NOT NULL,
	email TEXT NOT NULL,
	password TEXT,
	active BOOLEAN DEFAULT TRUE,
	verified BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP WITH TIME ZONE,
	last_login TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS auth_user_username_idx ON auth_user(username);
CREATE INDEX IF NOT EXISTS auth_user_email_idx ON auth_user(email);

CREATE TABLE IF NOT EXISTS auth_oauth(
	user_id INTEGER PRIMARY KEY REFERENCES auth_user(id) ON DELETE CASCADE,
	provider TEXT NOT NULL,
	sid TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS auth_oauth_provider_sid_idx ON auth_oauth(provider, sid);