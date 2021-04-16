CREATE TABLE IF NOT EXISTS users(                                                                  
	id serial PRIMARY KEY,
	username VARCHAR(100) NOT NULL,
	email VARCHAR(100) NOT NULL UNIQUE,
	password VARCHAR(100) NOT NULL,
	create_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	update_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS confirmation_users(
	id serial PRIMARY KEY,
	activated BOOLEAN DEFAULT FALSE,
	user_id INT NOT NULL,
	CONSTRAINT fk_users
		FOREIGN KEY(user_id)
			REFERENCES users(id)
			ON DELETE CASCADE
);
