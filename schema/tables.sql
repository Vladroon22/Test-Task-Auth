CREATE TABLE IF NOT EXISTS sessions (
    email VARCHAR(30) NOT NULL,
    userID INT NOT NULL,                   
    userIP VARCHAR(45) NOT NULL,           
    expireAt TIMESTAMP NOT NULL,            
    refresh VARCHAR(255) NOT NULL
);

DROP TABLE IF EXISTS sessions;
