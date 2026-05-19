# Uptime Monitor Project - System Design

## Relational Database Table Structure
- users
-- id SERIAL / AUTO_INCREMENT / UUID? PK NOT NULL
-- username VARCHAR(255) NOT NULL 
-- password VARCHAR(60) NOT NULL 
-- email VARCHAR(255) NOT NULL
-- created_at TIMESTAMP DEFAULT NOW
-- updated_at TIMESTAMP DEFAULT NOW

- monitors
-- id SERIAL / AUTO_INCREMENT / UUID? PK NOT NULL
-- user_id INT FK NOT NULL
-- url VARCHAR(255) NOT NULL
-- interval INT NOT NULL
-- created_at TIMESTAMP DEFAULT NOW
-- updated_at TIMESTAMP DEFAULT NOW

- checks ?
-- id SERIAL / AUTO_INCREMENT / UUID? PK NOT NULL
-- monitor_id INT FK NOT NULL
-- status_code int NOT NULL
-- created_at TIMESTAMP DEFAULT NOW
-- updated_at TIMESTAMP DEFAULT NOW

## Endpoint Structure
POST /users/register <- for register new user
POST /users/login <- for login existing user
POST /monitors <- for user can adding url for monitoring
GET /monitors <- for user get all url which has been stored
PUT /monitors/{id} <- for user can update available url
DELETE /monitors/{id} <- for user can delete available url
GET /monitors/{id}/checks <- for user can checks the report of url

Mynotes:{
saya masih bingung standar namingnya sesuai standar industri yang real gimana, apakah bener repository? atau application aja? terus ini storage? atau ganti jadi repository? ,terus ini direktorinya http atau api? routes atau server? handlers saja atau dipisah kayak user_handlers, monitor_handlers,admin_handlers, dan sebagainya?, struct repository/application/etc yang dipake buat hanlder, mending taroh di routes/server, dihandler aja kayak gini, atau dipisah filenya? juga struct jsonResponsenya, terus juga aku masih sangat meragukan dengan strukturnya apakah ini sesusai standar industri dan namingnya juga, mungkin bisa koreksi punyaku lagi ini

}
