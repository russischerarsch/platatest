CREATE TABLE IF NOT EXISTS quotes(
		id SERIAL PRIMARY KEY,
		pair VARCHAR(6) NOT NULL UNIQUE,
		price DECIMAL,
		attempts INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP,
		status VARCHAR(10) DEFAULT 'pending' CHECK (status IN ('pending','completed','failed'))
	)