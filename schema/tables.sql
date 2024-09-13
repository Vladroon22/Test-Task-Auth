-- Active: 1725963449851@@127.0.0.1@5431@postgres@public


CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    email VARCHAR(30) NOT NULL,
    userID INT NOT NULL,                   
    userIP VARCHAR(45) NOT NULL,           
    regTime TIMESTAMP NOT NULL,            
    expireTime TIMESTAMP NOT NULL,         
    refresh VARCHAR(255) NOT NULL
);

DROP TABLE IF EXISTS sessions;
