CREATE TABLE IF NOT EXISTS users(                                                                  
	id serial PRIMARY KEY,
	username VARCHAR(100) NOT NULL,
	email VARCHAR(100) NOT NULL UNIQUE,
	password VARCHAR(100) NOT NULL,
	create_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	update_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
